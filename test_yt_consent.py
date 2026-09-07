import urllib.request
import re

url = "https://www.youtube.com/watch?v=XfMBdq5iFnw"
req = urllib.request.Request(url, headers={
    'User-Agent': 'Mozilla/5.0',
    'Cookie': 'CONSENT=YES+cb.20210328-17-p0.en+FX+433;'
})
resp = urllib.request.urlopen(req)
html = resp.read().decode('utf-8')

title = re.search(r'"title":"(.*?)"', html)
print("Title:", title.group(1) if title else "None")
