with open('/home/lennox/Desktop/اهها/Go_Bot/main.go', 'r') as f:
    content = f.read()

import re

# We want to remove everything from `if reactText != "" {` to the end of the `if v.Message.GetReactionMessage() != nil` block.
start = content.find('if reactText != "" {')
end = content.find('\n\t\t\treturn\n\t\t}\n\t\tif v.Info.IsFromMe {')

if start != -1 and end != -1:
    content = content[:start] + content[end:]

with open('/home/lennox/Desktop/اهها/Go_Bot/main.go', 'w') as f:
    f.write(content)
