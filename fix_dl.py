import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/download.go', 'r') as f:
    content = f.read()

old_logic = """	fileInfo, err := os.Stat(videoPath)
	if fileInfo.Size() > 64*1024*1024 { // 64MB limit
		sendMessage(ctx, "حجم الفيديو كبير جداً (أكثر من 64 ميجا)، جاري الإرسال كملف...")
		sendMediaData(ctx, mediaData, "video/mp4", whatsmeow.MediaDocument)
		return
	}"""

new_logic = """	fileInfo, err := os.Stat(videoPath)
	if err == nil && fileInfo.Size() > 64*1024*1024 { // Just inform them it's big, but send as video anyway
		// sendMessage(ctx, "حجم الفيديو كبير، جاري الرفع كفيديو...")
	}"""

content = content.replace(old_logic, new_logic)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/download.go', 'w') as f:
    f.write(content)
