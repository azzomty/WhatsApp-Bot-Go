from bs4 import BeautifulSoup
import requests

html = requests.get('https://stardima-37.cartoon.com.im/search?q=%D8%AF%D8%A7%D9%86%D9%8A').text
soup = BeautifulSoup(html, 'html.parser')

print("Titles found:")
# Look for standard grid items
for item in soup.select('a[href*="/tvshow/"], a[href*="/movie/"]'):
    title = item.text.strip()
    if title:
        print(f"- {title} : {item.get('href')}")
