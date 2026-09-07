import urllib.request
import json
import ssl

ctx = ssl.create_default_context()
ctx.check_hostname = False
ctx.verify_mode = ssl.CERT_NONE

instances = [
    "api.cobalt.tame.gg",
    "api.cobalt.canine.tools",
    "api.cobalt.cjs.nz",
    "cobalt.kittycat.boo",
    "api.cobalt.mgytr.top",
    "api.cobalt.squair.xyz"
]

data = json.dumps({"url": "https://www.youtube.com/watch?v=XfMBdq5iFnw", "isAudioOnly": True}).encode('utf-8')

for host in instances:
    for protocol in ['https']:
        url = f"{protocol}://{host}/"
        req = urllib.request.Request(url, data=data, headers={
            'Accept': 'application/json',
            'Content-Type': 'application/json',
            'Origin': f'https://{host.replace("api.", "")}',
            'User-Agent': 'Mozilla/5.0'
        })
        try:
            resp = urllib.request.urlopen(req, context=ctx, timeout=5)
            print(f"SUCCESS {url}: {resp.read()[:100]}")
        except Exception as e:
            print(f"FAIL {url}: {e}")
