#!/usr/bin/env python

#
# Query
#
import requests
import os

username = os.environ['ELK_SA_USERNAME']
password = os.environ['ELK_SA_PASSWORD']

url = "http://oit-elk-kibana6:9200/oit-static-av-devices/_search"

payload = '''
{
  "_source": [
    "hostname",
    "last-heartbeat"
  ],
  "query": {
    "bool": {
      "must": [
        {
          "match": {
            "_type": "control-processor"
          }
        },
        {
          "match": {
            "alerts.lost-heartbeat.alerting": true
          }
        },
        {
          "match": {
            "alerting": true
          }
        }
      ],
      "filter": {
        "range": {
          "last-heartbeat": {
            "gte": "now-30s"
          }
        }
      }
    }
  },
  "size": -1
}
'''
headers = {
    'content-type': "application/json",
    'cache-control': "no-cache"
    }
print username
print password
response = requests.request("POST", url, data=payload, headers=headers, auth=(username, password))

print(response.text)

#
# Handle response from elk
#
import json
import sys
import requests
import time

from datetime import datetime

searchresults = json.loads(response.text)
if searchresults["hits"]["total"] == 0:
    print("No heartbeats have been restored.")
    sys.exit()
slackchannel = 'T0311JJTE/B67FHGZ0W/hMtIq8eppuGsaZnTOblwFOut'

httpProxy = "http://east.byu.edu:3128"
httpsProxy = "https://east.byu.edu:3128"

proxyDict = {
    "http": httpProxy,
    "https": httpsProxy
}

TranslationTable = {
    'D': 'display',
    'CP': 'control-processor',
    'DSP': 'digital-signal-processor',
    'PC': 'general-computer'
}

elkAddr = "http://oit-elk-kibana6:9200"
index = "oit-static-av-devices"

alertHeader = "alerts"
errorTypeString = "lost-heartbeat"
errorStringHeader = 'message'
lastUpdateHeader = 'alert-sent'
alertingHeader = 'alerting'

errorString = "Too much time elapsed since last heartbeat. Last heartbeat was at : "

timestring = "%Y-%m-%dT%H:%M:%S.%f%Z"

print(searchresults)

url = "https://hooks.slack.com/services/" + slackchannel


def datetime_from_utc_to_local(utc_datetime):
    now_timestamp = time.time()
    offset = datetime.fromtimestamp(
        now_timestamp) - datetime.utcfromtimestamp(now_timestamp)
    return utc_datetime + offset


for hit in searchresults["hits"]["hits"]:
    heartbeatPreParse = hit["_source"]["last-heartbeat"]
    device = hit["_source"]["hostname"]

    splitHostname = device.split("-")

    for i in range(len(splitHostname[2])):
        if splitHostname[2][i].isdigit():
            devType = splitHostname[2][:i]

    if devType in TranslationTable:
        devType = TranslationTable[devType]

    elkurl = elkAddr + "/" + index + "/" + devType + "/" + device

    heartbeat = datetime.strptime(heartbeatPreParse, '%Y-%m-%dT%H:%M:%S.%fZ')
    lastHeartbeat = datetime_from_utc_to_local(
        heartbeat).strftime("%Y-%m-%d %H:%M:%S")

    body = '''{"mrkdwn": true,"text": "Good news! A device has restored connection! \n*Device*:\t''' + \
        device + '\n*Last-Heartbeat*:\t' + lastHeartbeat + '"}'
    print "Sending for device " + device
    r = requests.put(url, data=body, proxies=proxyDict)

    # Set the status in ELK. in Rundeck 2.9 we can call out to another job for
    # this.

    r = requests.get(elkurl, auth=(username, password))
    if r.status_code != 200:
        print "non-200 code: " + str(r.status_code)
        continue

    val = r.content.decode('utf-8')
    content = json.loads(val)['_source']

    if errorTypeString not in content:
        content[alertHeader] = {}
    if errorTypeString not in content[alertHeader]:
        content[alertHeader][errorTypeString] = {}

    content[alertHeader][errorTypeString][errorStringHeader] = ""
    content[alertHeader][errorTypeString][alertingHeader] = False

    content[alertingHeader] = False

    # we need to check for all the other keys and see if this was the last alert to clear.
    for key in content[alertHeader]:
        if content[alertHeader][key][alertingHeader] == True:
            content[alertingHeader] = True
            break


    payload = json.dumps(content)
    print(payload.decode('utf-8'))
    print elkurl

    r = requests.put(elkurl, data=payload, auth=(username, password))
