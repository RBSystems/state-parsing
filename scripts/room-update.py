#!/usr/bin/env python

import requests
import json
import os

username = os.environ['ELK_SA_USERNAME']
password = os.environ['ELK_SA_PASSWORD']


queryurl = "http://oit-elk-kibana6:9200/oit-static-av-devices,oit-static-av-rooms/_search"

elkAddr = "http://oit-elk-kibana6:9200"
index = "oit-static-av-rooms"
recType = "room"

powerQuery = '''
{
  "size": 0,
  "query": {
    "bool": {
      "must": [
        {
          "match": {
            "power": "on"
          }
        }
      ]
    }
  },
  "aggs": {
    "rooms": {
      "terms": {
        "field": "room"
      },
    
    "aggs": {
      "index": {
        "terms": {
          "field": "_index"
        }
      }
    }}
  }
}
'''

alertingQuery = '''
{
  "size": 0,
  "query": {
    "bool": {
      "must": [
        {
          "match": {
            "alerting": true
          }
        }
      ]
    }
  },
  "aggs": {
    "rooms": {
      "terms": {
        "field": "room"
      },
    
    "aggs": {
      "index": {
        "terms": {
          "field": "_index"
        }
      }
    }}
  }
}
'''

def TrueUpState(query, attribute):
    toSend = ""
    r = requests.post(queryurl, data=query, auth=(username,password))

    if r.status_code /100 != 2:
        return 

    content = json.loads(r.content.decode("utf-8"))
    
    #We get the aggregations, then go through each room to see if there's something that needs to be updated in the room index
    for room in content["aggregations"]["rooms"]["buckets"]:
        if len(room["index"]["buckets"]) == 2:
            #The room index reflects at least on device in the room
            continue
        print("[room-updater] Updating room", room["key"])
        #We need to update
        header = { "update": { "_index": index, "_type": recType, "_id": room["key"]}}

        if room["index"]["buckets"][0]["key"] == "oit-static-av-devices":
            #We need to update the room index to be 'on' or 'alerting'
            if attribute == "power": 
                print("[room-updater] Updating setting the room power state to on")
                body = {"doc": {"power": "on"}, "doc_as_upsert": True}
            elif attribute == "alerting": 
                print("[room-updater] Updating setting the room alerting state to true")
                body = {"doc": {"alerting":True }, "doc_as_upsert": True}
            else:
                continue
        elif room["index"]["buckets"][0]["key"] == "oit-static-av-rooms":
            if attribute == "power": 
                print("[room-updater] Updating setting the room power state to standby")
                body = {"doc": {"power": "standby"}, "doc_as_upsert": True}
            elif attribute == "alerting": 
                body = {"doc": {"alerting": False}, "doc_as_upsert": True}
                print("[room-updater] Updating setting the room alerting state to false")
            else:
                continue
        toSend = toSend + json.dumps(header) + "\n" + json.dumps(body) + "\n"
        
    url = elkAddr + "/_bulk"
    sentr = requests.post(url,data=toSend,auth = (username, password))
  

TrueUpState(powerQuery, "power")
TrueUpState(alertingQuery, "alerting")
