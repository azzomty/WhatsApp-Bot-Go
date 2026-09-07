import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/download.go', 'r') as f:
    content = f.read()

old_cmd = 'cmd := exec.Command("./yt-dlp", "--cookies", "cookies.txt", "-f", "b", "-o", tmpFile, link)'
new_cmd = 'cmd := exec.Command("./yt-dlp", "-N", "16", "--no-check-certificate", "--cookies", "cookies.txt", "-f", "b", "-o", tmpFile, link)'

content = content.replace(old_cmd, new_cmd)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/download.go', 'w') as f:
    f.write(content)
