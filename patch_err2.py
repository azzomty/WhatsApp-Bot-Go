import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'r') as f:
    content = f.read()

old_err = """			m3u8URL, err := GetBestM3U8(watchURL)
			if err != nil {
				sendMessage(ctx, "عذراً، لم أتمكن من العثور على سيرفر يعمل.")
				return
			}"""

new_err = """			m3u8URL, err := GetBestM3U8(watchURL)
			if err != nil {
				sendMessage(ctx, "خطأ في السيرفر: " + err.Error())
				return
			}"""

content = content.replace(old_err, new_err)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'w') as f:
    f.write(content)
