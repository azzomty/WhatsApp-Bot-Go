import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'r') as f:
    content = f.read()

old_func = """func downloadStardimaMovieWithQuality(ctx *BotContext, selected StardimaVideo, qualityFmt string, splitIfLarge bool) {
	hyperURL, err := GetStardimaHyperwatchingURL(selected.URL)
	if err != nil {
		sendMessage(ctx, "فشل العثور على رابط المشاهدة.")
		return
	}
	uqloadEmbed, err := GetUqloadEmbedURL(hyperURL)
	if err != nil {
		sendMessage(ctx, "السيرفر الأساسي غير متوفر.")
		return
	}
	m3u8URL, err := GetUqloadM3U8(uqloadEmbed)
	if err != nil {
		sendMessage(ctx, "فشل في فك تشفير السيرفر.")
		return
	}
	data, err := DownloadM3U8WithQuality(m3u8URL, qualityFmt)
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء التحميل: "+err.Error())
		return
	}
	sendVideoDataWithSplit(ctx, data, selected.Title, "فيلم", splitIfLarge)
}"""

new_func = """func downloadStardimaMovieWithQuality(ctx *BotContext, selected StardimaVideo, qualityFmt string, splitIfLarge bool) {
	hyperURL, err := GetStardimaHyperwatchingURL(selected.URL)
	if err != nil {
		sendMessage(ctx, "فشل العثور على رابط المشاهدة.")
		return
	}
	m3u8URL, err := GetBestM3U8(hyperURL)
	if err != nil {
		sendMessage(ctx, "خطأ في السيرفر: "+err.Error())
		return
	}
	data, err := DownloadM3U8WithQuality(m3u8URL, qualityFmt)
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء التحميل: "+err.Error())
		return
	}
	sendVideoDataWithSplit(ctx, data, selected.Title, "فيلم", splitIfLarge)
}"""

content = content.replace(old_func, new_func)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'w') as f:
    f.write(content)
