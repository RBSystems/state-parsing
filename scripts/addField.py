
import requests
import os
import json

username = os.environ['ELK_SA_USERNAME']
password = os.environ['ELK_SA_PASSWORD']

url = "http://oit-elk-kibana6:9200/oit-static-av-devices/"

payload ='''
    {
    "query" : {
            "match_all" : {}
        },
      "size": 1000
    }
'''
headers = {
    'content-type': "application/json",
    'cache-control': "no-cache"
    }

response = requests.request("POST", url + "_search", headers=headers, auth=(username, password), data=payload)
searchresults = json.loads(response.text)

print searchresults['hits']['total']
print len(searchresults['hits']['hits'])
for v in searchresults['hits']['hits']:
    vals = v['_source']
    baseURL = "http://oit-elk-kibana6:9200/"

    content = vals
    content['enable-notifications'] = content['hostname']

    payload = json.dumps(content)

    postURL = url + v['_type'] + "/" + content['hostname']
    print postURL

    requests.put(postURL,data=payload,auth = (username, password))
