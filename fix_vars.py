import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

content = content.replace("\tseenGroupLinks = make(map[string]bool)\n\tseenLinksMutex sync.Mutex\n", "")
content = content.replace("var AutoJoinGroups bool", "var (\n\tAutoJoinGroups bool\n\tseenGroupLinks = make(map[string]bool)\n\tseenLinksMutex sync.Mutex\n)")

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
