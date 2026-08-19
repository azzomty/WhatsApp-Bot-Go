import re

with open("internal/pinterest/pinterest.go", "r") as f:
    content = f.read()

target = """func SearchPinterestGifs(query string, count int) []PinResult {"""

# Replace the whole block of SearchPinterestGifs
content = re.sub(r'func SearchPinterestGifs\(.*?\)\s*\[\]PinResult\s*\{.*?\n\}\n\nfunc parsePinterestData\(', 'func parsePinterestData(', content, flags=re.DOTALL)

with open("internal/pinterest/pinterest.go", "w") as f:
    f.write(content)

with open("main.go", "r") as f:
    main_content = f.read()

main_content = main_content.replace('results = pinterest.SearchPinterestGifs(req.Query, overrideCount)', 'results = pinterest.SearchPinterest(req.Query+" gif", "gif", overrideCount)')

with open("main.go", "w") as f:
    f.write(main_content)

