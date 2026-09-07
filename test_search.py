import requests
from bs4 import BeautifulSoup
import sys

url = "https://4h.b9p2m6c.shop/?search_param=animes&s=conan"
r = requests.get(url, headers={'User-Agent': 'Mozilla/5.0'})
print(r.status_code)
print(r.text[:500])
