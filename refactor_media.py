import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'r') as f:
    content = f.read()

# Add new state variables
state_vars = """var stardimaSelectedSeason = make(map[string]StardimaSeason)

type PendingStardima struct {
	Type   string
	Video  StardimaVideo
	Season StardimaSeason
	EpNum  string
}
var stardimaPending = make(map[string]PendingStardima)
"""
content = content.replace("var stardimaSelectedSeason = make(map[string]StardimaSeason)", state_vars)

# Modify HandleNumberSelect
old_handle_num_movie = """	} else {
		// It's a movie, download immediately
		sendMessage(ctx, fmt.Sprintf("تم اختيار الفيلم: *%s*\\nجاري التجهيز والتحميل...", selected.Title))
		go downloadStardimaMovie(ctx, selected)
	}"""
new_handle_num_movie = """	} else {
		// It's a movie, ask for quality
		pending := PendingStardima{
			Type: "movie",
			Video: selected,
		}
		cartoonMutex.Lock()
		stardimaPending[ctx.Sender.User] = pending
		cartoonMutex.Unlock()
		
		msg := "يرجى اختيار الجودة المطلوبة:\\n1. جودة 1080p (الأعلى - سيتم تقسيمها لأجزاء لو حجمها كبير)\\n2. جودة 720p (عالية - فيديو واحد)\\n3. جودة 480p (متوسطة وسريعة - فيديو واحد)\\n\\nللاختيار اكتب: `.جودة` متبوعاً بالرقم (مثال: `.جودة 1`)"
		sendMessage(ctx, msg)
	}"""
content = content.replace(old_handle_num_movie, new_handle_num_movie)

# Modify HandleStardimaEpisode
old_handle_ep = """		if watchURL == "" {
			sendMessage(ctx, "لم يتم العثور على الحلقة المطلوبة.")
			return
		}
		
		m3u8URL, err := GetBestM3U8(watchURL)
		if err != nil {
			fmt.Println("Error:", err)
			sendMessage(ctx, "عذراً، لم أتمكن من العثور على أي سيرفر يعمل لهذه الحلقة حالياً.")
			return
		}
		sendMessage(ctx, "جاري جلب الحلقة وتحميلها...")
		
		data, err := DownloadM3U8(m3u8URL)
		if err != nil {
			sendMessage(ctx, "حدث خطأ أثناء التحميل: "+err.Error())
			return
		}
		
		sendVideoData(ctx, data, selectedShow.Title+" - "+selSeason.Name, strconv.Itoa(epNum))
	}()"""
new_handle_ep = """		if watchURL == "" {
			sendMessage(ctx, "لم يتم العثور على الحلقة المطلوبة.")
			return
		}
		
		pending := PendingStardima{
			Type: "episode",
			Video: selectedShow,
			Season: selSeason,
			EpNum: strconv.Itoa(epNum),
		}
		cartoonMutex.Lock()
		stardimaPending[ctx.Sender.User] = pending
		cartoonMutex.Unlock()
		
		msg := "يرجى اختيار الجودة المطلوبة:\\n1. جودة 1080p (الأعلى - سيتم تقسيمها لأجزاء لو حجمها كبير)\\n2. جودة 720p (عالية - فيديو واحد)\\n3. جودة 480p (متوسطة وسريعة - فيديو واحد)\\n\\nللاختيار اكتب: `.جودة` متبوعاً بالرقم (مثال: `.جودة 1`)"
		sendMessage(ctx, msg)
	}()"""
if old_handle_ep in content:
    content = content.replace(old_handle_ep, new_handle_ep)
else:
    # Try a more robust regex replacement for HandleStardimaEpisode
    pattern = re.compile(r'if watchURL == "" \{\s*sendMessage\(ctx, "لم يتم العثور على الحلقة المطلوبة\."\)\s*return\s*\}.*?sendVideoData\(ctx, data, selectedShow\.Title\+" - "\+selSeason\.Name, strconv\.Itoa\(epNum\)\)\s*\}\(\)', re.DOTALL)
    content = pattern.sub(new_handle_ep, content)

# Now inject HandleStardimaQuality
quality_func = """
func HandleStardimaQuality(ctx *BotContext, choice int) {
	cartoonMutex.Lock()
	pending, ok := stardimaPending[ctx.Sender.User]
	cartoonMutex.Unlock()

	if !ok {
		sendMessage(ctx, "يرجى اختيار الفيلم أو الحلقة أولاً قبل اختيار الجودة.")
		return
	}

	qualityFmt := "best[height<=1080]/best"
	splitIfLarge := true
	if choice == 2 {
		qualityFmt = "best[height<=720]/best"
		splitIfLarge = false
	} else if choice == 3 {
		qualityFmt = "best[height<=480]/best"
		splitIfLarge = false
	} else if choice != 1 {
		sendMessage(ctx, "رقم الجودة غير صحيح. يرجى اختيار 1 أو 2 أو 3.")
		return
	}

	sendMessage(ctx, "جاري تجهيز وتحميل المقطع بالجودة المطلوبة... (قد يستغرق بعض الوقت)")

	go func() {
		if pending.Type == "movie" {
			downloadStardimaMovieWithQuality(ctx, pending.Video, qualityFmt, splitIfLarge)
		} else {
			epNumInt, _ := strconv.Atoi(pending.EpNum)
			episodes, err := GetStardimaEpisodes(pending.Season.ID)
			if err != nil {
				return
			}
			var watchURL string
			for _, e := range episodes {
				if e.EpisodeNumber == epNumInt {
					watchURL = e.WatchURL
					break
				}
			}
			m3u8URL, err := GetBestM3U8(watchURL)
			if err != nil {
				sendMessage(ctx, "عذراً، لم أتمكن من العثور على سيرفر يعمل.")
				return
			}
			data, err := DownloadM3U8WithQuality(m3u8URL, qualityFmt)
			if err != nil {
				sendMessage(ctx, "حدث خطأ أثناء التحميل: "+err.Error())
				return
			}
			sendVideoDataWithSplit(ctx, data, pending.Video.Title+" - "+pending.Season.Name, pending.EpNum, splitIfLarge)
		}
	}()
}

func downloadStardimaMovieWithQuality(ctx *BotContext, selected StardimaVideo, qualityFmt string, splitIfLarge bool) {
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
}

func sendVideoDataWithSplit(ctx *BotContext, data []byte, animeName, epNum string, splitIfLarge bool) {
	if !splitIfLarge || len(data) <= 64*1024*1024 {
		resp, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaVideo)
		if err != nil {
			fmt.Println("UPLOAD ERROR:", err)
			sendMessage(ctx, "فشل في رفع المقطع للواتساب: حجمه كبير جداً للواتساب.")
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
		ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{VideoMessage: vidMsg})
		return
	}

	sendMessage(ctx, "الحجم ضخم جداً للواتساب (أكثر من 64 ميجا) وتم طلب جودة 1080p، جاري التقسيم...")
	tempDir, err := os.MkdirTemp("", "video_split")
	if err != nil { return }
	defer os.RemoveAll(tempDir)
	inputPath := tempDir + "/input.mp4"
	os.WriteFile(inputPath, data, 0644)
	outPattern := tempDir + "/part_%03d.mp4"
	
	ffmpegPath := "ffmpeg"
	if _, err := os.Stat("node_modules/ffmpeg-static/ffmpeg"); err == nil {
		ffmpegPath = "node_modules/ffmpeg-static/ffmpeg"
	}
	
	cmd := exec.Command(ffmpegPath, "-i", inputPath, "-c", "copy", "-f", "segment", "-segment_time", "600", "-reset_timestamps", "1", outPattern)
	if err := cmd.Run(); err != nil {
		sendMessage(ctx, "فشل تقسيم الفيديو.")
		return
	}
	
	files, _ := os.ReadDir(tempDir)
	var parts []string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "part_") {
			parts = append(parts, tempDir+"/"+f.Name())
		}
	}
	for i, partPath := range parts {
		partData, _ := os.ReadFile(partPath)
		resp, err := ctx.Client.Upload(context.Background(), partData, whatsmeow.MediaVideo)
		if err != nil { continue }
		caption := fmt.Sprintf("*%s* - الحلقة %s\\n(الجزء %d من %d)", animeName, epNum, i+1, len(parts))
		vidMsg := &waProto.VideoMessage{
			URL:           proto.String(resp.URL),
			DirectPath:    proto.String(resp.DirectPath),
			MediaKey:      resp.MediaKey,
			Mimetype:      proto.String("video/mp4"),
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(partData))),
			Caption:       proto.String(caption),
		}
		ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{VideoMessage: vidMsg})
	}
}
"""
content = content + "\n" + quality_func

# Replace old downloadStardimaMovie
# Wait, downloadStardimaMovie is still in the file but unused now. I can leave it or regex it out.
# Let's just leave it, it's not referenced from HandleNumberSelect anymore.

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'w') as f:
    f.write(content)
