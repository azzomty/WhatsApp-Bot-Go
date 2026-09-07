import urllib.request
import json

instances = [
    "https://pipedapi.kavin.rocks",
    "https://pipedapi.leptons.xyz",
    "https://pipedapi.nosebs.ru",
    "https://api.piped.yt",
    "https://pipedapi.drgns.space",
    "https://pipedapi.owo.si",
    "https://pipedapi.ducks.party",
    "https://pipedapi.adminforge.de",
    "https://piped-api.privacy.com.de",
    "https://pipedapi.lunar.icu"
]

video_id = "XfMBdq5iFnw"

for base in instances:
    url = f"{base}/streams/{video_id}"
    req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
    try:
        resp = urllib.request.urlopen(req, timeout=5)
        data = json.loads(resp.read().decode('utf-8'))
        if 'audioStreams' in data and len(data['audioStreams']) > 0:
            print(f"SUCCESS: {base} -> {data['audioStreams'][0]['url'][:50]}...")
            break
        else:
            print(f"FAIL (no audio): {base}")
    except Exception as e:
        print(f"ERROR: {base} -> {e}")
