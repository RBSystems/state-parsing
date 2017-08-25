#!/usr/bin/env python

import requests
import json
from string import digits
from datetime import datetime
import time
import os

username = os.environ['ELK_SA_USERNAME']
password = os.environ['ELK_SA_PASSWORD']


queryurl = "http://oit-elk-kibana6:9200/oit-av%2A/_search"

elkAddr = "http://oit-elk-kibana6:9200"
index = "oit-static-av-devices"

TranslationTable = {
        'D': 'display',
        'CP': 'control-processor',
        'DSP': 'digital-signal-processor',
        'PC': 'general-computer',
        'SW': 'video-switcher'
        }

querypayload = '''{
"query": {
    "bool": {
      "must": {
        "match": {
          "event-type-string": "DETAILSTATE"
        }
      },
      "filter": {
        "range": {
          "timestamp": {
            "gte": "now-210s",
            "lte": "now"
          }
        }
      }
    }
},
"size": 0,
  "aggs": {
    "building": {
      "terms": {
        "field": "building.raw",
        "size": 100
      },
      "aggs": {
        "room": {
          "terms": {
            "field": "room.raw",
            "size": 100
          },
          "aggs": {
            "device": {
              "terms": {
                "field": "event.device.raw",
                "size": 100
              },
              "aggs": {
                "event-key": {
                  "terms": {
                    "field": "event.eventInfoKey.raw",
                    "size": 50
                  },
                  "aggs": {
                    "data": {
                      "top_hits": {
                        "size": 1,
                        "_source": [
                          "event.eventInfoKey",
                          "event.eventInfoValue",
                          "event.device",
                          "event-cause-string",
                          "@timestamp"
                        ],
                        "sort": {
                          "@timestamp": "desc"
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}
'''
headers = {
    'content-type': "application/json",
    'cache-control': "no-cache"
    }

response = requests.request("POST", queryurl, data=querypayload, headers=headers, auth=(username, password))

#print(response.text)

#
# Handle response from elk
#

searchresults = json.loads(response.text)

if "aggregations" not in searchresults:
    print("aggregatios not in search results")
    print(searchresults)
    exit(1)

for bucket in searchresults["aggregations"]["building"]["buckets"]:
    building = bucket["key"]
    for roomBucket in bucket["room"]["buckets"]:
            room = roomBucket["key"]
            for deviceBucket in roomBucket["device"]["buckets"]:
                device = deviceBucket["key"]

                roomName = building + "-" + room
                hostname = roomName + "-" + device

                for i in range(len(device)):
                    if device[i].isdigit():
                        devType = device[:i]

                if devType in TranslationTable:
                    devType = TranslationTable[devType]

                url = elkAddr + "/" + index + "/" + devType + "/" + hostname

                r = requests.get(url, auth = (username, password))



                if r.status_code == 404:
                    content = {"room": roomName, "hostname": hostname}
                elif r.status_code == 200:
                    value = r.content.decode('utf-8')
                    content = json.loads(value)["_source"]
                else:
                    print("Invalid response: " + str(r.status_code))
                    continue

                for eventBucket in deviceBucket["event-key"]["buckets"]:
                    event = eventBucket["data"]["hits"]["hits"][0]["_source"]["event"]
                    if event["eventInfoValue"].isdigit():
                        content[event["eventInfoKey"]] = int(event["eventInfoValue"])
                    else:
                        content[event["eventInfoKey"]] = event["eventInfoValue"]
                    if eventBucket["data"]["hits"]["hits"][0]["_source"]["event-cause-string"] == "USERINPUT":
                        content['last-user-input'] = eventBucket["data"]["hits"]["hits"][0]["_source"]["@timestamp"]


                payload = json.dumps(content)
                print(content)

                sentr = requests.put(url,data=payload,auth = (username, password))

