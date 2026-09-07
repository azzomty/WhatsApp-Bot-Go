import requests
from bs4 import BeautifulSoup

def search_ddg(query):
    url = "https://html.duckduckgo.com/html/"
    headers = {"User-Agent": "Mozilla/5.0"}
    data = {"q": query}
    res = requests.post(url, headers=headers, data=data)
    soup = BeautifulSoup(res.text, 'html.parser')
    links = []
    for a in soup.find_all('a', class_='result__url'):
        links.append(a.get('href'))
    return links

print(search_ddg('site:4h.b9p2m6c.shop "detective conan" "الحلقة 1"'))
