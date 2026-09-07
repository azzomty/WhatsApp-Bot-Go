import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

content = content.replace("	AutoJoinGroups = false\n\nvar (", "	AutoJoinGroups = false\n)\n\nvar (")

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
