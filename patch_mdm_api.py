with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/mdm.go', 'r') as f:
    c = f.read()

# Replace the broken apiURL
old_api = '''apiURL := "https://www.masterduelmeta.com/api/v1/top-decks?created[$gte]=(days-30)&limit=0"'''
new_api = '''apiURL := "https://www.masterduelmeta.com/api/v1/top-decks?deck=" + url.QueryEscape(session.TargetDeck.Name)'''

c = c.replace(old_api, new_api)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/mdm.go', 'w') as f:
    f.write(c)
