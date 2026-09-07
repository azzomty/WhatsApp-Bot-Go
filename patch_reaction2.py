with open('/home/lennox/Desktop/اهها/Go_Bot/main.go', 'r') as f:
    content = f.read()

import re
# Remove the block from `if reactText != "" {` to `}` just before `return`
pattern = re.compile(r'if reactText != "" \{\n\s+fmt\.Println\("REACTION DETECTED.*?\}\n\s+\}\n', re.DOTALL)
new_content = pattern.sub('', content)

with open('/home/lennox/Desktop/اهها/Go_Bot/main.go', 'w') as f:
    f.write(new_content)
