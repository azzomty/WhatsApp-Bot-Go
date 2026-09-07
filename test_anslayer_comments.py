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

params = {
    "_order_by": "latest_first",
    "hide_irrelevant": "Yes",
    "episode_id": 67885.0,
    "_limit": 10,
    "myfirst": "Yes",
    "_offset": 0
}
url = f'https://anslayer.com/anime/public/episode-comments/get-episode-comments?json={json.dumps(params)}'
r = requests.get(url, headers=headers)
print("Comments status:", r.status_code)
if r.status_code == 200:
    data = r.json()
    comments = data.get('response', {}).get('data', [])
    for c in comments[:3]:
        print(f"Comment ID: {c.get('comment_id')} - User: {c.get('comment_user_id')} - Text: {c.get('comment_text')}")
