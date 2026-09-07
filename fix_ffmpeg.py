import os
import glob

# Search all go files
for root, _, files in os.walk('/home/lennox/Desktop/اهها/Go_Bot/internal'):
    for file in files:
        if file.endswith('.go'):
            path = os.path.join(root, file)
            with open(path, 'r') as f:
                content = f.read()
            
            orig = content
            
            # fix stardima.go
            if "stardima.go" in path:
                content = content.replace(
                    'exec.Command("./yt-dlp", "-N", "16", "--no-check-certificate", "-f", "best[height<=480]/best", m3u8URL, "-o", tmpFile)',
                    'exec.Command("./yt-dlp", "--ffmpeg-location", "./ffmpeg", "-N", "16", "--no-check-certificate", "-f", "best[height<=480]/best", m3u8URL, "-o", tmpFile)'
                )
                content = content.replace(
                    'exec.Command("./yt-dlp", "-N", "16", "--no-check-certificate", "-f", quality, m3u8URL, "-o", tmpFile)',
                    'exec.Command("./yt-dlp", "--ffmpeg-location", "./ffmpeg", "-N", "16", "--no-check-certificate", "-f", quality, m3u8URL, "-o", tmpFile)'
                )
            
            # fix download.go
            if "download.go" in path:
                content = content.replace(
                    'exec.Command("./yt-dlp", "-N", "16", "--no-check-certificate", "--cookies", "cookies.txt", "-f", "b", "-o", tmpFile, link)',
                    'exec.Command("./yt-dlp", "--ffmpeg-location", "./ffmpeg", "-N", "16", "--no-check-certificate", "--cookies", "cookies.txt", "-f", "b", "-o", tmpFile, link)'
                )
                
            # fix media.go
            if "media.go" in path:
                content = content.replace('ffmpegPath := "ffmpeg"', 'ffmpegPath := "./ffmpeg"')
                
            # fix youtube.go
            if "youtube.go" in path:
                content = content.replace('ffmpegPath = "ffmpeg"', 'ffmpegPath = "./ffmpeg"')
                
            # fix vocalremover.go
            if "vocalremover.go" in path:
                content = content.replace('ffmpegPath := "/home/lennox/Desktop/اهها/Go_Bot/node_modules/ffmpeg-static/ffmpeg"', 'ffmpegPath := "./ffmpeg"')

            if orig != content:
                with open(path, 'w') as f:
                    f.write(content)
                print(f"Patched {path}")

