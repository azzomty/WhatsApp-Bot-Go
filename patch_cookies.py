import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    content = f.read()

# Add cookies to processDownload (video)
content = content.replace(
    'cmd = exec.Command("./yt-dlp", "--ffmpeg-location", "./ffmpeg", "-N", "4", "--no-check-certificate", "-f", "bestvideo[height<=720][ext=mp4]+bestaudio[ext=m4a]/best[height<=720][ext=mp4]/best", "--merge-output-format", "mp4", url, "-o", finalFile)',
    'cmd = exec.Command("./yt-dlp", "--cookies", "cookies.txt", "--ffmpeg-location", "./ffmpeg", "-N", "4", "--no-check-certificate", "-f", "bestvideo[height<=720][ext=mp4]+bestaudio[ext=m4a]/best[height<=720][ext=mp4]/best", "--merge-output-format", "mp4", url, "-o", finalFile)'
)

# Add cookies to processDownload (audio)
content = content.replace(
    'cmd = exec.Command("./yt-dlp", "--ffmpeg-location", "./ffmpeg", "-N", "4", "--no-check-certificate", "-f", "bestaudio/best", "-x", "--audio-format", "mp3", url, "-o", finalFile)',
    'cmd = exec.Command("./yt-dlp", "--cookies", "cookies.txt", "--ffmpeg-location", "./ffmpeg", "-N", "4", "--no-check-certificate", "-f", "bestaudio/best", "-x", "--audio-format", "mp3", url, "-o", finalFile)'
)

# Add cookies to YouTube info fetch
content = content.replace(
    'cmd := exec.Command("./yt-dlp", videoURL, "--no-warnings", "--print", "%(id)s|%(title)s|%(uploader)s|%(view_count)s|%(like_count)s|%(duration_string)s|%(upload_date)s|%(thumbnail)s")',
    'cmd := exec.Command("./yt-dlp", "--cookies", "cookies.txt", videoURL, "--no-warnings", "--print", "%(id)s|%(title)s|%(uploader)s|%(view_count)s|%(like_count)s|%(duration_string)s|%(upload_date)s|%(thumbnail)s")'
)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(content)
print("Patched cookies")
