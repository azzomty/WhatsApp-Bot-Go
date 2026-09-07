import urllib.request
import re
import json

url = "https://www.youtube.com/watch?v=XfMBdq5iFnw"
req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
try:
    resp = urllib.request.urlopen(req)
    html = resp.read().decode('utf-8')
    
    match = re.search(r'ytInitialPlayerResponse\s*=\s*({.+?});var', html)
    if match:
        data = json.loads(match.group(1))
        details = data.get('videoDetails', {})
        print("Title:", details.get('title'))
        print("Author:", details.get('author'))
        print("Views:", details.get('viewCount'))
        print("Duration:", details.get('lengthSeconds'))
        print("Thumb:", details.get('thumbnail', {}).get('thumbnails', [-1])[-1].get('url'))
    else:
        print("No match")
except Exception as e:
    print("Error:", e)
