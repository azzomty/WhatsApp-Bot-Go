import requests, json, re

show_url = "https://stardima-37.cartoon.com.im/tvshow/69dcda2b0ff3b"
r = requests.get(show_url)
m = re.search(r'play/(\d+)', r.text)
play_url = show_url + "/play/" + m.group(1)
r2 = requests.get(play_url)
m2 = re.search(r'data-season-id="(\d+)"', r2.text)
season_id = m2.group(1)

r3 = requests.get("https://stardima-37.cartoon.com.im/series/season/" + season_id, headers={"X-Requested-With": "XMLHttpRequest"})
eps = r3.json()["episodes"]
ep5 = next(e for e in eps if e["episode_number"] == 5)
print("Ep 5 watch URL:", ep5["watch_url"])

r4 = requests.get(ep5["watch_url"])
m3 = re.search(r'data-page="([^"]+)"', r4.text)
data = json.loads(m3.group(1).replace("&quot;", '"'))
servers = data["props"]["video"]["servers"]
print("Servers:")
for s in servers:
    print(s["name"], s["id"])
