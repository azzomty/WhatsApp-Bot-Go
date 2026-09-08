with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/mdm.go', 'r') as f:
    c = f.read()

import re
c = re.sub(r'deckName := ""\s*switch v := d\.DeckType\.\(type\) \{\s*case string:\s*deckName = v\s*case map\[string\]interface\{\}:\s*if n, ok := v\["name"\]\.\(string\); ok \{ deckName = n \}\s*\}', '', c)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/mdm.go', 'w') as f:
    f.write(c)
