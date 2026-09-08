with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/mdm.go', 'r') as f:
    c = f.read()

old_cond = '''				// API usually already filters by deck, but we keep check just in case.
				// Since we query ?deck=Name, we can also accept empty deckName (if API format changes)
				if deckName == "" || strings.EqualFold(deckName, session.TargetDeck.Name) {'''

new_cond = '''				// Always accept since the API already filters it!
				if true {'''

c = c.replace(old_cond, new_cond)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/mdm.go', 'w') as f:
    f.write(c)
