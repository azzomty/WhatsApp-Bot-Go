from bs4 import BeautifulSoup

with open("mdm.html", "r") as f:
    soup = BeautifulSoup(f, "html.parser")

tiers = soup.find_all(string=lambda t: t and "Tier" in t)
for t in tiers:
    print(t)
