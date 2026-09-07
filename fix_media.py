import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'r') as f:
    content = f.read()

# Replace Episode handler logic
old_ep_logic = """		uqloadEmbed, err := GetUqloadEmbedURL(watchURL)
		if err != nil {
			sendMessage(ctx, "السيرفر الأساسي غير متوفر لهذه الحلقة حالياً.")
			return
		}
		
		m3u8URL, err := GetUqloadM3U8(uqloadEmbed)
		if err != nil {
			sendMessage(ctx, "فشل في فك تشفير السيرفر.")
			return
		}"""

new_ep_logic = """		m3u8URL, err := GetBestM3U8(watchURL)
		if err != nil {
			sendMessage(ctx, "عذراً، لم أتمكن من العثور على أي سيرفر يعمل لهذه الحلقة حالياً.")
			return
		}"""

content = content.replace(old_ep_logic, new_ep_logic)

# Replace Movie handler logic
old_movie_logic = """		uqloadEmbed, err := GetUqloadEmbedURL(hyperURL)
		if err != nil {
			sendMessage(ctx, "السيرفر الأساسي غير متوفر لهذا الفيلم حالياً.")
			return
		}
		
		m3u8URL, err := GetUqloadM3U8(uqloadEmbed)
		if err != nil {
			sendMessage(ctx, "فشل في فك تشفير السيرفر.")
			return
		}"""

new_movie_logic = """		m3u8URL, err := GetBestM3U8(hyperURL)
		if err != nil {
			sendMessage(ctx, "عذراً، لم أتمكن من العثور على أي سيرفر يعمل لهذا الفيلم حالياً.")
			return
		}"""

content = content.replace(old_movie_logic, new_movie_logic)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'w') as f:
    f.write(content)
print("Replaced in media.go")
