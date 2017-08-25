
import requests
import os
import json

username = os.environ['ELK_SA_USERNAME']
password = os.environ['ELK_SA_PASSWORD']

url = "http://oit-elk-kibana6:9200/oit-static-av-rooms/ITB/1101"

payload ='''
{
    "query": {},
    "size": 100
}
'''
headers = {
    'content-type': "application/json",
    'cache-control': "no-cache"
    }

response = requests.request("GET", url, headers=headers, auth=(username, password))
searchresults = json.loads(response.text)
print response.text

vals = searchresults['_source']
baseURL = "http://oit-elk-kibana6:9200/"

content = vals
content['alerts'] = {"notify": False}

payload = json.dumps(content)

print url

requests.put(url,data=payload,auth = (username, password))
