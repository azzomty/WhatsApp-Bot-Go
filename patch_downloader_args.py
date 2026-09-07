import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    content = f.read()

# Fix command building to dynamically add cookies and extractor args
new_video = """		args := []string{"--ffmpeg-location", "./ffmpeg", "-N", "4", "--no-check-certificate", "-f", "bestvideo[height<=720][ext=mp4]+bestaudio[ext=m4a]/best[height<=720][ext=mp4]/best", "--merge-output-format", "mp4", "--extractor-args", "youtube:player_client=android,ios", url, "-o", finalFile}
		if _, err := os.Stat("cookies.txt"); err == nil {
			args = append([]string{"--cookies", "cookies.txt"}, args...)
		}
		cmd = exec.Command("./yt-dlp", args...)"""

new_audio = """		args := []string{"--ffmpeg-location", "./ffmpeg", "-N", "4", "--no-check-certificate", "-f", "bestaudio/best", "-x", "--audio-format", "mp3", "--extractor-args", "youtube:player_client=android,ios", url, "-o", finalFile}
		if _, err := os.Stat("cookies.txt"); err == nil {
			args = append([]string{"--cookies", "cookies.txt"}, args...)
		}
		cmd = exec.Command("./yt-dlp", args...)"""

new_info = """			args := []string{videoURL, "--no-warnings", "--print", "%(id)s|%(title)s|%(uploader)s|%(view_count)s|%(like_count)s|%(duration_string)s|%(upload_date)s|%(thumbnail)s", "--extractor-args", "youtube:player_client=android,ios"}
			if _, err := os.Stat("cookies.txt"); err == nil {
				args = append([]string{"--cookies", "cookies.txt"}, args...)
			}
			cmd := exec.Command("./yt-dlp", args...)"""

# Let's replace the hardcoded exec.Command with these.
# 1. Video
content = re.sub(r'cmd = exec.Command\("\./yt-dlp", "--cookies", "cookies\.txt", "--ffmpeg-location", "\./ffmpeg", "-N", "4", "--no-check-certificate", "-f", "bestvideo\[height<=720\]\[ext=mp4\]\+bestaudio\[ext=m4a\]/best\[height<=720\]\[ext=mp4\]/best", "--merge-output-format", "mp4", url, "-o", finalFile\)', new_video, content)
# 2. Audio
content = re.sub(r'cmd = exec.Command\("\./yt-dlp", "--cookies", "cookies\.txt", "--ffmpeg-location", "\./ffmpeg", "-N", "4", "--no-check-certificate", "-f", "bestaudio/best", "-x", "--audio-format", "mp3", url, "-o", finalFile\)', new_audio, content)
# 3. Info
content = re.sub(r'cmd := exec.Command\("\./yt-dlp", "--cookies", "cookies\.txt", videoURL, "--no-warnings", "--print", "%\(id\)s\|%\(title\)s\|%\(uploader\)s\|%\(view_count\)s\|%\(like_count\)s\|%\(duration_string\)s\|%\(upload_date\)s\|%\(thumbnail\)s"\)', new_info, content)


with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(content)
print("Done")
