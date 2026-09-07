import requests
from bs4 import BeautifulSoup
r = requests.get("https://4h.b9p2m6c.shop/", headers={'User-Agent': 'Mozilla/5.0'})
soup = BeautifulSoup(r.text, 'html.parser')
for a in soup.find_all('a', href=True):
    if 'anime/' in a['href'] or 'episode/' in a['href']:
        print(a['href'])
        break
