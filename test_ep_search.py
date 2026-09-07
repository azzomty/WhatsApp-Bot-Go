import requests
from bs4 import BeautifulSoup
import sys
q = "detective conan"
url = f"https://4h.b9p2m6c.shop/?search_param=episodes&s={q}"
r = requests.get(url, headers={'User-Agent': 'Mozilla/5.0'})
soup = BeautifulSoup(r.text, 'html.parser')
for a in soup.find_all('a', href=True):
    if 'episode' in a['href']:
        print(a.text.strip(), a['href'])
