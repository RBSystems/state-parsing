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

payload = "{\n\t\"query\": {\n\t\t\"bool\": {\n\t\t\t\"must\": {\n\t\t\t\t\"match\": {\n\t\t\t\t\t\"hosttype\": \"control processor\"\n\t\t\t\t}\n\t\t\t},\n\t\t\t\"filter\": {\n\t\t\t\t\"range\": {\n\t\t\t\t\t\"timestamp\": {\n\t\t\t\t\t\t\"gte\": \"now-17s\"\n\t\t\t\t\t}\n\t\t\t\t}\n\t\t\t}\n\t\t}\n\t},\n\t\"_source\": [\n\t\t\"hostname\",\n\t\t\"@timestamp\"\n\t],\n\t\"size\": 0,\n\t\"aggs\": {\n\t\t\"unique_hostname\": {\n\t\t\t\"terms\": {\n\t\t\t\t\"field\": \"hostname.raw\",\n\t\t\t\t\"size\": 1000\n\t\t\t},\n\t\t\t\"aggs\": {\n\t\t\t\t\"last_timestamp\": {\n\t\t\t\t\t\"max\": {\n\t\t\t\t\t\t\"field\": \"@timestamp\"\n\t\t\t\t\t}\n\t\t\t\t}\n\t\t\t}\n\t\t},\n\t\t\"num_hostnames\": {\n\t\t\t\"cardinality\": {\n\t\t\t\t\"field\": \"hostname.raw\"\n\t\t\t}\n\t\t}\n\t}\n}"
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
        content = {fieldToUpdate: bucket["last_timestamp"]["value_as_string"], "room": room, "hostname": hostname}
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
