with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

content = content.replace('ioutil.WriteFile("baymax_sticker.webp"', 'os.WriteFile("baymax_sticker.webp"')
content = content.replace('ioutil.ReadFile("baymax_sticker.webp")', 'os.ReadFile("baymax_sticker.webp"')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
