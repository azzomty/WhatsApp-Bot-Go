with open('/home/lennox/Desktop/اهها/Go_Bot/main.go', 'r') as f:
    content = f.read()

content = content.replace('ytCookies := os.Getenv("YOUTUBE_COOKIES")', 'ytCookies := os.Getenv("COOKIES_TXT")\n\tif ytCookies == "" { ytCookies = os.Getenv("YOUTUBE_COOKIES") }')

with open('/home/lennox/Desktop/اهها/Go_Bot/main.go', 'w') as f:
    f.write(content)
