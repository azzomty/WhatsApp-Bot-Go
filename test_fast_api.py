import urllib.request
import json
import time

instances = [
    "https://cobalt.tame.gg",
    "https://cobalt.squair.xyz",
    "https://cobalt.kittycat.boo",
    "https://cobalt.mgytr.top",
    "https://cobalt.canine.tools",
    "https://cobalt.cjs.nz",
    "https://co.wuk.sh",
    "https://api.cobalt.best"
]

data = json.dumps({
    "url": "https://www.youtube.com/watch?v=YQHsXMglC9A",
    "isAudioOnly": True,
    "aFormat": "mp3"
}).encode('utf-8')

for host in instances:
    url = f"{host}/api/json" if "co.wuk.sh" not in host else host
    # For v10, it's just '/'
    urls_to_try = [f"{host}", f"{host}/api/json"]
    
    for u in urls_to_try:
        req = urllib.request.Request(u, data=data, headers={
            'Accept': 'application/json',
            'Content-Type': 'application/json',
            'User-Agent': 'Mozilla/5.0'
        })
        try:
            start = time.time()
            resp = urllib.request.urlopen(req, timeout=3)
            res = resp.read().decode('utf-8')
            if 'url' in res or 'status' in res:
                print(f"SUCCESS {u} in {time.time()-start:.2f}s: {res[:100]}")
                break
        except Exception as e:
            pass
