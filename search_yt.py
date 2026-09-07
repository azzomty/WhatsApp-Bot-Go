import urllib.request
import re

url = "https://www.youtube.com/results?search_query=hello+adele"
req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
resp = urllib.request.urlopen(req)
html = resp.read().decode('utf-8')
match = re.search(r'\"videoId\":\"(.*?)\"', html)
print(match.group(1) if match else "Not found")
