import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'r') as f:
    content = f.read()

# Modify DownloadM3U8 to return stderr
old_ytdlp = """	cmd := exec.Command("yt-dlp", m3u8URL, "-o", tmpFile)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("yt-dlp failed: %v", err)
	}"""

new_ytdlp = """	cmd := exec.Command("yt-dlp", m3u8URL, "-o", tmpFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp failed: %v\\nOutput: %s", err, string(out))
	}"""

content = content.replace(old_ytdlp, new_ytdlp)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'w') as f:
    f.write(content)
