with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

content = content.replace("package commands\n\nimport (", "package commands\n\nimport (\n\t\"sync\"\n)\n\nvar (\n\tseenGroupLinks = make(map[string]bool)\n\tseenLinksMutex sync.Mutex\n)\n\nimport (")

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
