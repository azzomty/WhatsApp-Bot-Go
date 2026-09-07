with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    content = f.read()

# Remove the extractor args completely
content = content.replace(', "--extractor-args", "youtube:player_client=android,ios"', '')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(content)
print("Removed extractor args")
