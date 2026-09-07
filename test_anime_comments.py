import urllib.request
import json
import urllib.parse

params = {
    "_order_by": "latest_first",
    "hide_irrelevant": "Yes",
    "anime_id": 1460, # One piece anime_id
    "_limit": 10,
    "myfirst": "Yes",
    "_offset": 0
}
encoded = urllib.parse.quote(json.dumps(params))
url = "https://anslayer.com/anime/public/anime-comments/get-anime-comments?json=" + encoded
req = urllib.request.Request(url)
req.add_header('Accept', 'application/json')
req.add_header('Client-Id', 'android-app2')
req.add_header('Authorization', 'Bearer a458a73d7d556e1f5474a8c50cf1fb17d54314a3')
req.add_header('User-Agent', 'okhttp/3.12.13')

try:
    resp = urllib.request.urlopen(req)
    print("ANIME COMMENTS WORK:")
    print(resp.read().decode('utf-8')[:300])
except Exception as e:
    print("Failed anime comments:", e)
