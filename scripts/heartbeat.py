#!/usr/bin/env python

#
# Query
#
import requests
import os
import sys

username = os.environ['ELK_SA_USERNAME']
password = os.environ['ELK_SA_PASSWORD']

url = "http://oit-elk-kibana6:9200/oit-heartbeat-av%2A/_search"

payload ='''
{
	"query": {
		"bool": {
			"must": {
				"match": {
					"hosttype": "control processor"
				}
			},
			"filter": {
				"range": {
					"timestamp": {
						"gte": "now-15s"
					}
				}
			}
		} }, "_source": [
		"hostname",
		"@timestamp"
	],
	"size": 0,
	"aggs": {
		"unique_hostname": {
			"terms": { "field": "hostname.raw", "size": 1000
			},
			"aggs": {
				"last_timestamp": {
					"max": {
						"field": "@timestamp"
					}
				}
			}
		},
		"num_hostnames": {
			"cardinality": {
				"field": "hostname.raw"
			}
		}
	}
}
'''
headers = {
    'content-type': "application/json",
    'cache-control': "no-cache"
    }

response = requests.request("POST", url, data=payload, headers=headers, auth=(username, password))

#
# Handle response from elk
#
import json
import sys
import requests
from string import digits

searchresults = json.loads(response.text)
fieldToUpdate = "last-heartbeat"

elkAddr = "http://oit-elk-kibana6:9200"
index = "oit-static-av-devices"

TranslationTable = {
        'D': 'display',
        'CP': 'control-processor',
        'DSP': 'digital-signal-processor',
        'PC': 'general-computer'
        }



for bucket in searchresults["aggregations"]["unique_hostname"]["buckets"]:
    #we need to get the old information in the bucket in ES so we can update the value

    hostname = bucket["key"]
    splitHostname = hostname.split("-")
    room = splitHostname[0] + "-" + splitHostname[1]

    if len(splitHostname) < 3:
        print(hostname + " not a valid host.")
        continue


    for i in range(len(splitHostname[2])):
        if splitHostname[2][i].isdigit():
            devType = splitHostname[2][:i]

    if devType in TranslationTable:
        devType = TranslationTable[devType]

    url = elkAddr + "/" + index + "/" + devType + "/" + hostname

    r = requests.get(url, auth = (username, password))

    if r.status_code == 404:
        print("Not Found")
        content = {fieldToUpdate: bucket["last_timestamp"]["value_as_string"], "room": room, "hostname": hostname, "suppress-notifications": hostname, "enable-notifications": hostname}
    elif r.status_code == 200:
        print("200")
        val = r.content.decode('utf-8')
        resp = json.loads(val)
        content = resp['_source']
        content[fieldToUpdate] = bucket["last_timestamp"]["value_as_string"]
    elif r.status_code == 401:
        print("Authorization error for user:" + username)
        continue
    else:
        print("Other error: " + str(r.status_code))
        continue
    payload = json.dumps(content)
    print(payload.decode('utf-8'))

    r = requests.put(url,data=payload,auth = (username, password))
