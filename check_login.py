import urllib.request
import urllib.parse
import json
import random

url = "https://anslayer.com/anime/public/oauth/login"
data = urllib.parse.urlencode({
    "username": "bademail@gmail.com",
    "password": "badpassword",
    "device_id": "48c6467c5039bf74"
}).encode('utf-8')

req = urllib.request.Request(url, data=data)
req.add_header('Accept', 'application/json')
req.add_header('Client-Id', 'android-app2')
req.add_header('Client-Secret', '7befba6263cc14c90d2f1d6da2c5cf9b251bfbbd')
req.add_header('User-Agent', 'okhttp/3.12.13')

try:
    resp = urllib.request.urlopen(req)
    print(resp.read().decode('utf-8'))
except urllib.error.HTTPError as e:
    print(e.code)
    print(e.read().decode('utf-8'))
