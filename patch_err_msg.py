with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    content = f.read()

import re

err_msg_old = '''				if len(dlM) < 2 {
					sendMessage(ctx, "فشل استخراج رابط التحميل.")
					return
				}'''

err_msg_new = '''				if len(dlM) < 2 {
					snippet := string(bodyP)
					if len(snippet) > 50 { snippet = snippet[:50] }
					sendMessage(ctx, "فشل استخراج رابط التحميل. السبب: " + snippet)
					return
				}'''

content = content.replace(err_msg_old, err_msg_new)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(content)
print("Patched err msg")
