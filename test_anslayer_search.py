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

search_json = {
    "_offset": 0,
    "_limit": 10,
    "_order_by": "latest_first",
    "list_type": "filter",
    "anime_name": "one piece",
    "just_info": "Yes"
}
search_url = f'https://anslayer.com/anime/public/animes/get-published-animes?json={json.dumps(search_json)}'
r = requests.get(search_url, headers=headers)
print("Search status:", r.status_code)
if r.status_code == 200:
    data = r.json()
    animes = data.get('response', {}).get('data', [])
    for a in animes[:5]:
        print(f"ID: {a.get('anime_id')} - Name: {a.get('anime_name')}")

