import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'r') as f:
    code = f.read()

old_func = '''func HandleEpisodeCommand(ctx *BotContext) {
	parts := strings.Split(ctx.Text, " ")
	if len(parts) < 2 {
		sendMessage(ctx, "يرجى كتابة رقم الحلقة، مثال: .حلقة 40")
		return
	}
	epNum := parts[1]

	cartoonMutex.Lock()
	showName, ok := cartoonSessions[ctx.Sender.User]
	cartoonMutex.Unlock()

	if !ok || showName == "" {
		sendMessage(ctx, "لم تقم بالبحث عن أي كرتون مسبقاً. ابحث أولاً باستخدام أمر .كرتون (مثال: .كرتون سبونج بوب)")
		return
	}

	sendMessage(ctx, "جاري البحث عن الحلقة...")

	// Try to find dubbed first
	searchQuery := fmt.Sprintf("%s حلقة %s مدبلج بالعربي", showName, epNum)
	videoID, err := youtube.SearchVideo(searchQuery)
	if err != nil {
		// Try without dubbed
		searchQuery = fmt.Sprintf("%s حلقة %s مترجم بالعربي", showName, epNum)
		videoID, err = youtube.SearchVideo(searchQuery)
		if err != nil {
			sendMessage(ctx, "للأسف، لم أتمكن من العثور على الحلقة على يوتيوب.")
			return
		}
	}

	sendMessage(ctx, "جاري تحميل الحلقة وإرسالها... (قد يستغرق الأمر بعض الوقت حسب حجم الفيديو)")

	go func() {
		data, err := youtube.DownloadMedia(videoID, false)
		if err != nil {
			sendMessage(ctx, fmt.Sprintf("حدث خطأ أثناء تحميل الحلقة: %v", err))
			return
		}

		resp, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaVideo)
		if err != nil {
			sendMessage(ctx, fmt.Sprintf("حدث خطأ أثناء رفع الحلقة للواتساب: %v", err))
			return
		}

		vidMsg := &waProto.VideoMessage{
			URL:           proto.String(resp.URL),
			DirectPath:    proto.String(resp.DirectPath),
			MediaKey:      resp.MediaKey,
			Mimetype:      proto.String("video/mp4"),
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
			Caption:       proto.String(fmt.Sprintf("*%s* - الحلقة %s", showName, epNum)),
		}

		ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
			VideoMessage: vidMsg,
		})
	}()
}'''

new_func = '''func HandleEpisodeCommand(ctx *BotContext) {
	parts := strings.Split(ctx.Text, " ")
	if len(parts) < 2 {
		sendMessage(ctx, "يرجى كتابة رقم الحلقة، مثال: .حلقة 40")
		return
	}
	epNum := parts[1]

	cartoonMutex.Lock()
	showName, ok := cartoonSessions[ctx.Sender.User]
	cartoonMutex.Unlock()

	if !ok || showName == "" {
		sendMessage(ctx, "لم تقم بالبحث عن أي كرتون مسبقاً. ابحث أولاً باستخدام أمر .كرتون (مثال: .كرتون سبونج بوب)")
		return
	}

	sendMessage(ctx, fmt.Sprintf("جاري جلب الحلقة %s من %s (من السيرفرات المباشرة)... ⏳", epNum, showName))

	go func() {
		epNumInt, _ := strconv.Atoi(epNum)
		embedLink, err := ScrapeWitanimeEpisode(showName, epNumInt)
		
		if err != nil || embedLink == "" {
			// Fallback to youtube search
			sendMessage(ctx, "لم أتمكن من إيجاد سيرفر مباشر، سأحاول جلبها من يوتيوب... ⏳")
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

		// Download embed link using yt-dlp
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
		sendVideoData(ctx, data, showName, epNum)
	}()
}

func sendVideoData(ctx *BotContext, data []byte, animeName, epNum string) {
	resp, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaVideo)
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء رفع الحلقة للواتساب.")
		return
	}

	vidMsg := &waProto.VideoMessage{
		URL:           proto.String(resp.URL),
		DirectPath:    proto.String(resp.DirectPath),
		MediaKey:      resp.MediaKey,
		Mimetype:      proto.String("video/mp4"),
		FileEncSHA256: resp.FileEncSHA256,
		FileSHA256:    resp.FileSHA256,
		FileLength:    proto.Uint64(uint64(len(data))),
		Caption:       proto.String(fmt.Sprintf("*%s* - الحلقة %s", animeName, epNum)),
	}

	ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
		VideoMessage: vidMsg,
	})
}
'''

if old_func in code:
    code = code.replace(old_func, new_func)
    with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'w') as f:
        f.write(code)
    print("Replaced HandleEpisodeCommand successfully")
else:
    print("Could not find HandleEpisodeCommand")
