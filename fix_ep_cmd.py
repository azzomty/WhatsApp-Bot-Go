import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'r') as f:
    code = f.read()

old_ep_logic = '''		// Download embed link using yt-dlp
		outPath, err := youtube.DownloadDirectURL(embedLink)
		if err != nil {
			sendMessage(ctx, "حدث خطأ أثناء تحميل الحلقة من السيرفر المباشر.")
			return
		}
		defer os.Remove(outPath)
		
		data, err := os.ReadFile(outPath)
		if err != nil {
			sendMessage(ctx, "حدث خطأ أثناء قراءة الملف.")
			return
		}
		sendVideoData(ctx, data, showName, epNum)'''

new_ep_logic = '''		// embedLink is a comma-separated list of links
		links := strings.Split(embedLink, ",")
		var outPath string
		var dlErr error
		
		// Try downloading from each link until one works
		for _, link := range links {
			if strings.TrimSpace(link) == "" {
				continue
			}
			sendMessage(ctx, "جاري تجربة السيرفر المباشر... ⏳")
			outPath, dlErr = youtube.DownloadDirectURL(link)
			if dlErr == nil && outPath != "" {
				break
			}
		}

		if dlErr != nil || outPath == "" {
			sendMessage(ctx, "حدث خطأ: السيرفرات المباشرة لا تستجيب، سأحاول جلب الحلقة من يوتيوب...")
			searchQuery := fmt.Sprintf("%s حلقة %s مدبلج بالعربي", showName, epNum)
			videoID, err := youtube.SearchVideo(searchQuery)
			if err != nil {
				sendMessage(ctx, "عذراً، الحلقة غير متوفرة.")
				return
			}
			
			data, err := youtube.DownloadMedia(videoID, false)
			if err != nil {
				sendMessage(ctx, fmt.Sprintf("حدث خطأ أثناء تحميل الحلقة: %v", err))
				return
			}
			sendVideoData(ctx, data, showName, epNum)
			return
		}
		defer os.Remove(outPath)
		
		data, err := os.ReadFile(outPath)
		if err != nil {
			sendMessage(ctx, "حدث خطأ أثناء قراءة الملف.")
			return
		}
		sendVideoData(ctx, data, showName, epNum)'''

code = code.replace(old_ep_logic, new_ep_logic)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'w') as f:
    f.write(code)

print("Updated episode retry logic")
