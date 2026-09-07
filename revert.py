import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'r') as f:
    content = f.read()

old_episode = """		pending := PendingStardima{
			Type: "episode",
			Video: selectedShow,
			Season: selSeason,
			EpNum: strconv.Itoa(epNum),
		}
		cartoonMutex.Lock()
		stardimaPending[ctx.Sender.User] = pending
		cartoonMutex.Unlock()
		
		msg := `يرجى اختيار الجودة المطلوبة:
1. جودة 1080p (الأعلى - سيتم تقسيمها لأجزاء لو حجمها كبير)
2. جودة 720p (عالية - فيديو واحد)
3. جودة 480p (متوسطة وسريعة - فيديو واحد)

للاختيار اكتب: .جودة متبوعاً بالرقم (مثال: .جودة 1)`
		sendMessage(ctx, msg)"""

new_episode = """		m3u8URL, err := GetBestM3U8(watchURL)
		if err != nil {
			sendMessage(ctx, "خطأ في السيرفر: " + err.Error())
			return
		}
		data, err := DownloadM3U8WithQuality(m3u8URL, "bestvideo[height<=720]+bestaudio/best[height<=720]")
		if err != nil {
			sendMessage(ctx, "حدث خطأ أثناء التحميل: " + err.Error())
			return
		}
		sendVideoDataWithSplit(ctx, data, selectedShow.Title+" - "+selSeason.Name, strconv.Itoa(epNum), false)"""

content = content.replace(old_episode, new_episode)

old_movie = """			pending := PendingStardima{
				Type: "movie",
				Video: selected,
			}
			cartoonMutex.Lock()
			stardimaPending[ctx.Sender.User] = pending
			cartoonMutex.Unlock()
			
			msg := `يرجى اختيار الجودة المطلوبة:
1. جودة 1080p (الأعلى - سيتم تقسيمها لأجزاء لو حجمها كبير)
2. جودة 720p (عالية - فيديو واحد)
3. جودة 480p (متوسطة وسريعة - فيديو واحد)

للاختيار اكتب: .جودة متبوعاً بالرقم (مثال: .جودة 1)`
			sendMessage(ctx, msg)
		}
	} else if action == "stardima_season" {"""

new_movie = """			sendMessage(ctx, "جاري جلب الفيلم وتحميله...")
			go downloadStardimaMovieWithQuality(ctx, selected, "bestvideo[height<=720]+bestaudio/best[height<=720]", false)
		}
	} else if action == "stardima_season" {"""

content = content.replace(old_movie, new_movie)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'w') as f:
    f.write(content)
