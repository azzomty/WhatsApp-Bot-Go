import urllib.request
import json
import urllib.parse

url = "https://anslayer.com/anime/public/animes/get-published-animes?json=%7B%22_offset%22%3A0%2C%22_limit%22%3A30%2C%22_order_by%22%3A%22latest_first%22%2C%22list_type%22%3A%22favorites%22%2C%22just_info%22%3A%22Yes%22%2C%22user_id%22%3A9174886%7D"
req = urllib.request.Request(url)
req.add_header('Accept', 'application/json')
req.add_header('Client-Id', 'android-app2')
req.add_header('Authorization', 'Bearer a458a73d7d556e1f5474a8c50cf1fb17d54314a3')
req.add_header('User-Agent', 'okhttp/3.12.13')

try:
    resp = urllib.request.urlopen(req)
    print("FAVORITES:")
    print(resp.read().decode('utf-8')[:500])
except Exception as e:
    print("Failed favorites:", e)
