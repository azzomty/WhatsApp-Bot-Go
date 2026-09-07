import urllib.request
import json
import ssl

ctx = ssl.create_default_context()
ctx.check_hostname = False
ctx.verify_mode = ssl.CERT_NONE

# Fetch instances from a known list if possible, or try some known ones
instances = [
    "co.wuk.sh",
    "api.cobalt.tools",
    "cobalt.q-n-d.de",
    "cobalt.tame.gg",
    "api.cobalt.tame.gg",
    "cobalt.squair.xyz",
    "cobalt.kittycat.boo",
    "cobalt.canine.tools",
    "api.cobalt.canine.tools",
    "cobalt.mgytr.top",
    "cobalt.cjs.nz"
]

data = json.dumps({"url": "https://www.youtube.com/watch?v=XfMBdq5iFnw", "isAudioOnly": True}).encode('utf-8')

for host in instances:
    for path in ["/", "/api/json"]:
        url = f"https://{host}{path}"
        req = urllib.request.Request(url, data=data, headers={
            'Accept': 'application/json',
            'Content-Type': 'application/json',
            'User-Agent': 'Mozilla/5.0'
        })
        try:
            resp = urllib.request.urlopen(req, context=ctx, timeout=5)
            res = resp.read().decode('utf-8')
            print(f"SUCCESS {url}: {res[:100]}")
        except Exception as e:
            pass
