import requests
from bs4 import BeautifulSoup
import base64

url = "https://4h.b9p2m6c.shop/episode/%d8%a7%d9%86%d9%85%d9%8a-black-torch-%d8%a7%d9%84%d8%ad%d9%84%d9%82%d8%a9-1-%d9%85%d8%aa%d8%b1%d8%ac%d9%85%d8%a9/"
r = requests.get(url, headers={'User-Agent': 'Mozilla/5.0'})
soup = BeautifulSoup(r.text, 'html.parser')

servers = soup.find_all('ul', id='episode-servers')
if servers:
    for a in servers[0].find_all('a'):
        # Just print the whole A tag to see what attributes it has
        print(a)
