#!/usr/bin/env bash
# Install yt-dlp
curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o yt-dlp
chmod a+rx yt-dlp

# Install ffmpeg
curl -L https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz -o ffmpeg.tar.xz
tar -xf ffmpeg.tar.xz
mv ffmpeg-*-amd64-static/ffmpeg ffmpeg
mv ffmpeg-*-amd64-static/ffprobe ffprobe
chmod +x ffmpeg ffprobe
rm -rf ffmpeg.tar.xz ffmpeg-*-amd64-static

# Build the Go app
go build -o bot_go_new main.go
