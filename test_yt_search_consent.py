import urllib.request
import re
import urllib.parse

query = "hello adele"
searchUrl = "https://www.youtube.com/results?search_query=" + urllib.parse.quote(query)
reqS = urllib.request.Request(searchUrl, headers={
    'User-Agent': 'Mozilla/5.0',
    'Cookie': 'CONSENT=YES+cb.20210328-17-p0.en+FX+433;'
})
respS = urllib.request.urlopen(reqS)
bodyS = respS.read().decode('utf-8')

match = re.search(r'"videoId":"([^"]+)"', bodyS)
if match:
    videoID = match.group(1)
    print("Found VideoID:", videoID)
else:
    print("Not found")
