with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

content = content.replace('os.ReadFile("baymax_sticker.webp"\n', 'os.ReadFile("baymax_sticker.webp")\n')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
