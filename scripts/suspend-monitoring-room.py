#!/usr/bin/env python
# Query
import sys
import requests
import json

username = sys.argv[1]
password = sys.argv[2]
roomName = '@option.RoomName@'

elkAddr = "http://oit-elk-kibana6:9200"
RoomIndex = "oit-static-av-rooms"
RoomType = "room"
room = roomName.upper()

url = "/".join([elkAddr, RoomIndex, RoomType, room, "_update"])

splitRoom = room.split("-")
building = splitRoom[0]
room = splitRoom[1]


content = {'doc': {'notifications-suppressed': False}}

payload = json.dumps(content)
print(payload.decode('utf-8'))

r = requests.post(url, data=payload, auth=(username, password))
print(r.status_code)
print(r.text)


