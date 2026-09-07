import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

old_logic = """	// Choose first fembed or ok.ru link, or just first link
	targetLink := links[0]
	
	if strings.Contains(targetLink, "mediafire.com") {
		// Replace file_premium with file
		targetLink = strings.Replace(targetLink, "file_premium", "file", 1)
		targetLink = resolveMediaFire(targetLink)
	}
	
	// Download using yt-dlp via existing function DownloadM3U8WithQuality (works for any url yt-dlp supports)
	data, err := DownloadM3U8WithQuality(targetLink, "bestvideo[height<=720]+bestaudio/best[height<=720]")
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء التحميل: " + err.Error())
		return
	}
	sendVideoDataWithSplit(ctx, data, anime.AnimeName, epNum, false)
}"""

new_logic = """	var data []byte
	var success bool

	for i, targetLink := range links {
		if strings.Contains(targetLink, "mediafire.com") {
			targetLink = strings.Replace(targetLink, "file_premium", "file", 1)
			targetLink = resolveMediaFire(targetLink)
		}
		
		if i > 0 {
			sendMessage(ctx, fmt.Sprintf("السيرفر السابق محذوف، جاري تجربة سيرفر بديل (%d/%d)...", i+1, len(links)))
		}
		
		data, err = DownloadM3U8WithQuality(targetLink, "bestvideo[height<=720]+bestaudio/best[height<=720]")
		if err == nil {
			success = true
			break
		}
	}
	
	if !success {
		sendMessage(ctx, "فشلت جميع السيرفرات في التحميل. (ربما تم حذف الحلقة من جميع المصادر بسبب حقوق النشر)")
		return
	}
	
	sendVideoDataWithSplit(ctx, data, anime.AnimeName, epNum, false)
}"""

content = content.replace(old_logic, new_logic)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
    f.write(content)
