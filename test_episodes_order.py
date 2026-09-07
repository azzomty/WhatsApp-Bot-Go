import urllib.request
import json

url = "https://anslayer.com/anime/public/anime/get-anime-details?anime_id=1460&fetch_episodes=Yes&more_info=No"
req = urllib.request.Request(url)
req.add_header('Accept', 'application/json')
req.add_header('Client-Id', 'android-app2')
req.add_header('Authorization', 'Bearer a458a73d7d556e1f5474a8c50cf1fb17d54314a3')
req.add_header('User-Agent', 'okhttp/3.12.13')

try:
    resp = urllib.request.urlopen(req)
    data = json.loads(resp.read().decode('utf-8'))
    episodes = data['response']['episodes']['data']
    print(f"Total episodes: {len(episodes)}")
    if len(episodes) > 0:
        print(f"First episode in list: {episodes[0]['episode_number']}")
        print(f"Last episode in list: {episodes[-1]['episode_number']}")
except Exception as e:
    print("Failed:", e)
