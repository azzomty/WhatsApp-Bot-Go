from bs4 import BeautifulSoup
import requests

html = requests.get('https://stardima-37.cartoon.com.im/tvshow/68cd2ad5a2e0a').text
soup = BeautifulSoup(html, 'html.parser')

print(soup.get_text()[:2000])
