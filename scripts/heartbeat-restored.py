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
    "last-heartbeat" ], "query": {
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

print(searchresults)

url = "https://hooks.slack.com/services/" + slackchannel


def datetime_from_utc_to_local(utc_datetime):
    now_timestamp = time.time()
    offset = datetime.fromtimestamp(
        now_timestamp) - datetime.utcfromtimestamp(now_timestamp)
    return utc_datetime + offset

roomsToDo = {}

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
    valDecoded = json.loads(val)
    content = valDecoded['_source']

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

    r = requests.put(elkurl, data=payload, auth=(username, password))

    ##We may need to clear the error in the room index as well.
    print content[alertingHeader]
    print (content["room"] in roomsToDo)

    if not content[alertingHeader] :
        if content["room"] not in roomsToDo:
            print "adding remove all"
            roomsToDo[content["room"]] = {"all": True}
    else: 
        roomsToDo[content["room"]] = {"all": False}


countURL = elkAddr + "/" + index + "/_count"

print "giving ELK time to Update"
time.sleep(5)
print "checking for room alerts to clear"
print len(roomsToDo), " rooms to check"

for room in roomsToDo:
    if roomsToDo[room]["all"]:
        print "Removing alerting status for room " + room
        ##check if any devices in the room (besides this one) are still alerting
        query = '''
        {
          "query": {
            "bool": {
              "must": [
                {
                  "match": {
                    "room": "'''+ room + '''"
                  }
                },
                {
                  "match": {
                    "alerting": true
                }
                }
              ]
            }
          }
        }
        '''
        r = requests.post(countURL, data=query, auth=(username, password))

        if r.status_code != 200:
            print "non-200 code: " + str(r.status_code)
            continue
        val = json.loads(r.content.decode('utf-8'))
        print val

        if val["count"] == 0:
            #Get the split hostname  
            splitRoom = room.split("-")
            
            #If count = 0; get the room, set alerting to false and then continue
            roomURL = elkAddr + "/" + roomIndex + "/" + splitRoom[0] + "/" + splitRoom[1]
            print roomURL

            r = requests.get(roomURL, auth=(username, password))

            if r.status_code != 200:
                print "non-200 code: " + str(r.status_code)
                continue

            val = r.content.decode('utf-8')
            valDecoded = json.loads(val)
            
            content = valDecoded['_source']
          
            content[alertingHeader] = False
            
            if alertHeader not in content:
                content[alertHeader] = {}
            if errorTypeString not in content[alertHeader]:
                content[alertHeader][errorTypeString]= {}

            content[alertHeader][errorTypeString][alertingHeader] = False

            payload = json.dumps(content)
            
            print "removing alerting status from ", room
            r = requests.put(roomURL, data=payload, auth=(username,password))
        else:
            print "there is another alert in room ", room
    else:
        #since all the alerts didn't clear, we need to check if at least the heatbeat alerts have cleared
        #Check if any devices, (besides this one) are alerting with a heartbeat-lost
        query = '''
        {
          "query": {
            "bool": {
              "must": [
                {
                  "match": {
                    "room": "'''+ room + '''"
                   }
                },
                {
                  "match": {
                    "alerts.lost-heartbeat.alerting": true
                }
                }
              ]
            }
          }
        }
        '''

        r = requests.post(countURL, data=query, auth=(username, password))

        if r.status_code != 200:
            print "non-200 code: " + str(r.status_code)
            continue

        if json.loads(r.content.decode('utf-8'))["count"] == 0:
            splitRoom = room.split("-")
            
            #If count = 0; get the room, set alerting to false and then continue
            roomURL = elkAddr + "/" + roomIndex + "/" + splitRoom[0] + "/" + splitRoom[1]
            r = requests.get(roomURL, auth=(username, password))

            if r.status_code != 200:
                print "non-200 code: " + str(r.status_code)
                continue

            val = r.content.decode('utf-8')
            valDecoded = json.loads(val)
            content = valDecoded['_source']

            if alertHeader not in content:
                content[alertHeader] = {}
            if errorTypeString not in content[alertHeader]:
                content[alertHeader][errorTypeString]= {}

            content[alertHeader][errorTypeString][alertingHeader] = False

            payload = json.dumps(content)

            print "removing heartbeat alerts for ", room
            r = requests.put(roomURL, data=payload, auth=(username,password))
        else:
            print "there is another heartbeat alert in room ", room
