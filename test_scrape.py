import requests
from bs4 import BeautifulSoup
import sys

q = "conan"
url = f"https://4h.b9p2m6c.shop/?search_param=animes&s={q}"
r = requests.get(url, headers={'User-Agent': 'Mozilla/5.0'})
soup = BeautifulSoup(r.text, 'html.parser')

cards = soup.find_all('div', class_='anime-card-container')
for c in cards[:5]:
    a = c.find('a')
    if a:
        print(a.get('href'), a.text.strip())

if not cards:
    print("No cards found. Let's try finding any links containing 'anime'.")
    links = soup.find_all('a')
    for l in links[:10]:
        print(l.get('href'))
