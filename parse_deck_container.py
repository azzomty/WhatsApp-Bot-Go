from bs4 import BeautifulSoup
with open("yubel.html", "r") as f:
    soup = BeautifulSoup(f, "html.parser")

containers = soup.find_all(class_=lambda c: c and "deck-container" in c)
print(f"Found {len(containers)} deck containers.")
for i, c in enumerate(containers):
    imgs = c.find_all("img", class_=lambda cls: cls and "card-img" in cls)
    names = [img.get("alt", "") for img in imgs if img.get("alt", "") not in ["cp-ur", "cp-sr", "cp-r", "cp-n"]]
    print(f"Container {i} has {len(names)} cards.")
    if i == 0:
        print("Cards:", names)
