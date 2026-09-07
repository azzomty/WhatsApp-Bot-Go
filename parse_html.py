from bs4 import BeautifulSoup
import requests

html = requests.get('https://stardima-37.cartoon.com.im/').text
soup = BeautifulSoup(html, 'html.parser')

print("Forms:")
for f in soup.find_all('form'):
    print(f.get('action'), f.get('method'))
    for i in f.find_all('input'):
        print('  ', i.get('name'))

print("\nLinks:")
links = set([a.get('href') for a in soup.find_all('a') if a.get('href')])
for l in links:
    if 'search' in l or '?' in l or 'cartoon' in l:
        print(l)
