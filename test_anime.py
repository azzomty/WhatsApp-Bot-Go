import requests
from bs4 import BeautifulSoup

url = "https://4h.b9p2m6c.shop/anime/black-torch/"
r = requests.get(url, headers={'User-Agent': 'Mozilla/5.0'})
soup = BeautifulSoup(r.text, 'html.parser')
for a in soup.find_all('a', href=True):
    if 'episode/' in a['href']:
        print(a['href'])
