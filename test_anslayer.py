import requests
import json

headers = {
    'Accept': 'application/json',
    'Accept-Encoding': 'gzip',
    'Authorization': 'Bearer a458a73d7d556e1f5474a8c50cf1fb17d54314a3',
    'Client-Id': 'android-app2',
    'Client-Secret': '7befba6263cc14c90d2f1d6da2c5cf9b251bfbbd',
    'User-Agent': 'okhttp/3.12.13'
}

# Try to get details with fetch_episodes=Yes
url = 'https://anslayer.com/anime/public/anime/get-anime-details?anime_id=2025&fetch_episodes=Yes&more_info=No'
r = requests.get(url, headers=headers)
print("Details status:", r.status_code)
if r.status_code == 200:
    data = r.json()
    print("Episodes found in details:", len(data.get('response', {}).get('episodes', [])))

# Try to search
search_json = {"list_type":"search","anime_name":"ون بيس"}
search_url = f'https://anslayer.com/anime/public/animes/get-published-animes?json={json.dumps(search_json)}'
r2 = requests.get(search_url, headers=headers)
print("Search status:", r2.status_code)
if r2.status_code == 200:
    animes = r2.json().get('response', {}).get('data', [])
    print("Animes found:", len(animes))
    if animes:
        print("First anime:", animes[0].get('anime_name'), "ID:", animes[0].get('anime_id'))

