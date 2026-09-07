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

url = 'https://anslayer.com/anime/public/anime/get-anime-details?anime_id=3071&fetch_episodes=Yes&more_info=No'
r = requests.get(url, headers=headers)
if r.status_code == 200:
    data = r.json()
    eps = data.get('response', {}).get('episodes', {}).get('data', [])
    if eps:
        urls = eps[0].get('episode_urls', [])
        for u in urls:
            print("Server:", u.get('episode_server_name'))
            print("URL:", u.get('episode_url'))
            
            # Try to fetch the URL
            r2 = requests.get(u.get('episode_url'), headers=headers)
            print("Fetch status:", r2.status_code)
            print("Response:", r2.text[:200])
