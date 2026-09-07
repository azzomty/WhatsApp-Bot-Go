import socket
import sys

# Monkey patch getaddrinfo
orig_getaddrinfo = socket.getaddrinfo
def new_getaddrinfo(*args, **kwargs):
    if args[0] == 'mp4upload.com':
        return orig_getaddrinfo('188.114.96.6', *args[1:], **kwargs)
    return orig_getaddrinfo(*args, **kwargs)
socket.getaddrinfo = new_getaddrinfo

import yt_dlp

ydl_opts = {
    'quiet': False,
    'outtmpl': 'test_bypass.mp4'
}
with yt_dlp.YoutubeDL(ydl_opts) as ydl:
    ydl.download(['https://mp4upload.com/embed-sz9m015l6jjv.html'])
