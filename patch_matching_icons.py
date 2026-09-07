with open('/home/lennox/Desktop/اهها/Go_Bot/internal/pinterest/pinterest.go', 'r') as f:
    content = f.read()

content = content.replace('func SearchPinterestMatchingIcons(query string) []PinResult {', 'func SearchPinterestMatchingIcons(query string, count int) []PinResult {')
content = content.replace('pins, _ := SearchPinterest("matching icons "+query, "all", 10, "")', 'pins, _ := SearchPinterest("matching icons "+query, "all", count, "")')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/pinterest/pinterest.go', 'w') as f:
    f.write(content)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

content = content.replace('results = pinterest.SearchPinterestMatchingIcons(last.Query)', 'results = pinterest.SearchPinterestMatchingIcons(last.Query, last.Count)')
content = content.replace('results := pinterest.SearchPinterestMatchingIcons(query)', 'results := pinterest.SearchPinterestMatchingIcons(query, count)')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
print("Patched matching icons count")
