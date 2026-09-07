import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/youtube/youtube.go', 'r') as f:
    code = f.read()

old_cmd2 = '''	cmd := exec.Command(ytdlp, "-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best", "-o", outPath, url)'''
new_cmd2 = '''	ffmpegPath := "node_modules/ffmpeg-static/ffmpeg"
	if _, err := os.Stat(ffmpegPath); os.IsNotExist(err) {
		ffmpegPath = "ffmpeg" // fallback to system ffmpeg
	}
	if runtime.GOOS == "windows" && ffmpegPath == "node_modules/ffmpeg-static/ffmpeg" {
		if _, err := os.Stat("node_modules/ffmpeg-static/ffmpeg.exe"); err == nil {
			ffmpegPath = "node_modules/ffmpeg-static/ffmpeg.exe"
		}
	}
	cmd := exec.Command(ytdlp, "--ffmpeg-location", ffmpegPath, "-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best", "-o", outPath, url)'''

code = code.replace(old_cmd2, new_cmd2)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/youtube/youtube.go', 'w') as f:
    f.write(code)

print("Patched DownloadDirectURL to use ffmpeg path")
