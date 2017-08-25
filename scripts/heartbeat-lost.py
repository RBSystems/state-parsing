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
  "query": {
    "bool": {
      "must": [
        {
          "match": {
            "_type": "control-processor"
          }
        }
      ],
      "should": [
        {
          "range": {
            "alerts.lost-heartbeat.alert-sent": {
              "lte": "now-20m"
            }
          }
        },
        {
          "bool": {
            "must_not": {
              "exists": {
                "field": "alerts.lost-heartbeat.alert-sent"
              }
            }
          }
        }
      ],
      "minimum_should_match": 1,
      "filter": {
        "range": {
          "last-heartbeat": {
            "lte": "now-120s"
          }
        }
      }
    }
  }
}
'''

print(username)
print(password)

headers = {
    'content-type': "application/json",
    'cache-control': "no-cache"
    }

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
    print("No missing heartbeats to report.")
    sys.exit()

slackchannel = os.environ['SLACK_HEARTBEAT_CHANNEL']

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
roomIndex = "oit-static-av-rooms"

alertHeader = "alerts"
errorTypeString = "lost-heartbeat"
errorStringHeader = 'message'
lastUpdateHeader = 'alert-sent'
alertingHeader = 'alerting'

errorString = "Too much time elapsed since last heartbeat. Last heartbeat was at : "

timestring = "%Y-%m-%dT%H:%M:%S.%f%Z"

#We'll cache the room information so that if we need it we don't have to go back up to ELK to get it 
roomInfo = {} 

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


    # Set the status in ELK. in Rundeck 2.9 we can call out to another job for
    # this.

    r = requests.get(elkurl, auth=(username, password))
    if r.status_code != 200:
        print "non-200 code: " + str(r.status_code)
        continue

    val = r.content.decode('utf-8')
    content = json.loads(val)['_source']

    if alertHeader not in content:
        content[alertHeader] = {}
    if errorTypeString not in content[alertHeader]:
        content[alertHeader][errorTypeString] = {}

    content[alertHeader][errorTypeString][errorStringHeader] = errorString + lastHeartbeat
    content[alertHeader][errorTypeString][alertingHeader] = True
    content[alertHeader][errorTypeString][lastUpdateHeader]  = datetime.utcnow().strftime(timestring)
    content[alertingHeader] = True

    payload = json.dumps(content)
    #print(payload.decode('utf-8'))
    #print elkurl

    r = requests.put(elkurl, data=payload, auth=(username, password))

   ### We need to update the room, mark it as alerting
    if content["room"] in roomInfo:
        curRoom = roomInfo[content["room"]]
    else:
        #We need to get the information
        splitRoom = content["room"].split("-")

        roomURL = elkAddr +  "/" + roomIndex + "/" + splitRoom[0] + "/" + splitRoom[1]

        print "roomURL: " + roomURL
        r = requests.get(roomURL, auth=(username, password))
        val = r.content.decode('utf-8')

        print val

        curRoom = json.loads(val)['_source']
        roomInfo[content["room"]] = curRoom
   
    if alertHeader not in roomInfo[content["room"]] or (errorTypeString not in roomInfo[content["room"]][alertHeader] or alertingHeader not in roomInfo[content["room"]][alertHeader][errorTypeString] or not roomInfo[content["room"]][alertHeader][errorTypeString][alertingHeader]) or (alertingHeader not in roomInfo[content["room"]] or not roomInfo[content["room"]][alertingHeader]):

        if alertHeader not in roomInfo[content["room"]]:
            roomInfo[content["room"]][alertHeader] = {}

        if errorTypeString not in roomInfo[content["room"]][alertHeader]:
            roomInfo[content["room"]][alertHeader][errorTypeString] = {}

        roomInfo[content["room"]][alertHeader][errorTypeString][alertingHeader] = True
        roomInfo[content["room"]][alertingHeader] = True

        payload = json.dumps(roomInfo[content["room"]])
        r = requests.put(roomURL, auth=(username, password), data=payload)

   ###---------------------------------------------------------------------------------
   ### NOTIFICATIONS ------------------------------------------------------------------
   ###---------------------------------------------------------------------------------
    #Send a slack notification
    
    #We need to check to see if alerts have been suppressed at the room levelj
    #Check if we're suppressing alerts
    
    print roomInfo[content["room"]]

    if ('notify' in content[alertHeader][errorTypeString] and not content[alertHeader][errorTypeString]) or ('notify' in content[alertHeader] and not content[alertHeader]['notify']) or ('notify' in roomInfo[content["room"]][alertHeader] and not roomInfo[content["room"]][alertHeader]['notify']):
        print("Alerts suppressed for this device " + device)
        continue

    body = '''{"mrkdwn": true,"text": "A device hasn't sent a heartbeat in a while, you might want to check it out.\n*Device*:\t''' + \
        device + '\n*Last-Heartbeat*:\t' + lastHeartbeat + '"}'
    print "Sending for device " + device
    r = requests.put(url, data=body, proxies=proxyDict)
