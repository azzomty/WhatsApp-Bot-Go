with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    content = f.read()

content = content.replace('⏳', '')
content = content.replace('❌', '')
content = content.replace('🎵', '')
content = content.replace('🎬', '')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(content)
print("Removed Emojis")
