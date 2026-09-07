import urllib.request
import re
import json

url = "https://www.youtube.com/watch?v=XfMBdq5iFnw"
req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
resp = urllib.request.urlopen(req)
html = resp.read().decode('utf-8')

title = re.search(r'"title":"(.*?)"', html)
author = re.search(r'"author":"(.*?)"', html)
view = re.search(r'"viewCount":"(.*?)"', html)
length = re.search(r'"lengthSeconds":"(.*?)"', html)
    
print("Title:", title.group(1) if title else "None")
print("Author:", author.group(1) if author else "None")
print("Views:", view.group(1) if view else "None")
print("Length:", length.group(1) if length else "None")
