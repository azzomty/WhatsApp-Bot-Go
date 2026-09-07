import re
import os

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/youtube/youtube.go', 'r') as f:
    code = f.read()

# Make yt-dlp path dynamic
# Add runtime package import if not there
if '"runtime"' not in code:
    code = code.replace('"os/exec"', '"os/exec"\n\t"runtime"')

old_cmd1 = 'cmd := exec.Command("./yt-dlp", "--ffmpeg-location", "node_modules/ffmpeg-static/ffmpeg", "--merge-output-format", ext, "-f", format, "-o", tmpFile, link)'
new_cmd1 = '''	ytdlp := "./yt-dlp"
	if runtime.GOOS == "windows" {
		ytdlp = "yt-dlp.exe"
	} else if _, err := os.Stat(ytdlp); os.IsNotExist(err) {
		ytdlp = "yt-dlp" // Try system path if local ./yt-dlp doesn't exist
	}
	
	ffmpegPath := "node_modules/ffmpeg-static/ffmpeg"
	if _, err := os.Stat(ffmpegPath); os.IsNotExist(err) {
		ffmpegPath = "ffmpeg" // fallback to system ffmpeg
	}
	if runtime.GOOS == "windows" && ffmpegPath == "node_modules/ffmpeg-static/ffmpeg" {
		if _, err := os.Stat("node_modules/ffmpeg-static/ffmpeg.exe"); err == nil {
			ffmpegPath = "node_modules/ffmpeg-static/ffmpeg.exe"
		}
	}

	cmd := exec.Command(ytdlp, "--ffmpeg-location", ffmpegPath, "--merge-output-format", ext, "-f", format, "-o", tmpFile, link)'''

code = code.replace(old_cmd1, new_cmd1)

old_cmd2 = 'cmd := exec.Command("./yt-dlp", "-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best", "-o", outPath, url)'
new_cmd2 = '''	ytdlp := "./yt-dlp"
	if runtime.GOOS == "windows" {
		ytdlp = "yt-dlp.exe"
	} else if _, err := os.Stat(ytdlp); os.IsNotExist(err) {
		ytdlp = "yt-dlp"
	}
	cmd := exec.Command(ytdlp, "-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best", "-o", outPath, url)'''

code = code.replace(old_cmd2, new_cmd2)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/youtube/youtube.go', 'w') as f:
    f.write(code)

print("Patched yt-dlp for Windows support")
