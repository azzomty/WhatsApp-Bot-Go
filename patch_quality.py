import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'r') as f:
    content = f.read()

old_cmd = 'cmd := exec.Command("yt-dlp", "-N", "16", "--no-check-certificate", m3u8URL, "-o", tmpFile)'
new_cmd = 'cmd := exec.Command("yt-dlp", "-N", "16", "--no-check-certificate", "-f", "best[height<=480]/best", m3u8URL, "-o", tmpFile)'

content = content.replace(old_cmd, new_cmd)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'w') as f:
    f.write(content)
