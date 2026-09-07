import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    content = f.read()

start_str = '			// 1. Search YouTube natively'
end_str = '			_, _ = ctx.Client.SendMessage(context.Background(), ctx.Event.Info.Chat, msg)'

if start_str in content and end_str in content:
    start_idx = content.find(start_str)
    end_idx = content.find(end_str) + len(end_str)

    new_code = """			// 1. Fetch Soundcloud Client ID
				reqC, _ := http.NewRequest("GET", "https://soundcloud.com", nil)
				reqC.Header.Set("User-Agent", "Mozilla/5.0")
				respC, errC := http.DefaultClient.Do(reqC)
				if errC != nil {
					sendMessage(ctx, "حدث خطأ في الاتصال.")
					return
				}
				bodyC, _ := io.ReadAll(respC.Body)
				respC.Body.Close()
				
				jsRe := regexp.MustCompile(`https://a-v2\.sndcdn\.com/assets/[a-zA-Z0-9-]+\.js`)
				jsMatches := jsRe.FindAllString(string(bodyC), 5)
				clientID := "Pb72ranhoyt6gw7hM7TkzUItXlMWSNSo" // fallback
				for _, jsUrl := range jsMatches {
					reqJ, _ := http.NewRequest("GET", jsUrl, nil)
					reqJ.Header.Set("User-Agent", "Mozilla/5.0")
					respJ, errJ := http.DefaultClient.Do(reqJ)
					if errJ == nil {
						bodyJ, _ := io.ReadAll(respJ.Body)
						respJ.Body.Close()
						cRe := regexp.MustCompile(`client_id:"([^"]+)"`)
						if m := cRe.FindStringSubmatch(string(bodyJ)); len(m) > 1 {
							clientID = m[1]
							break
						}
					}
				}
				
				// 2. Search Soundcloud
				searchURL := "https://api-v2.soundcloud.com/search/tracks?q=" + url.QueryEscape(query) + "&client_id=" + clientID + "&limit=1"
				reqS, _ := http.NewRequest("GET", searchURL, nil)
				reqS.Header.Set("User-Agent", "Mozilla/5.0")
				respS, errS := http.DefaultClient.Do(reqS)
				if errS != nil {
					sendMessage(ctx, "فشل البحث.")
					return
				}
				bodyS, _ := io.ReadAll(respS.Body)
				respS.Body.Close()
				
				// Parse JSON manually using regex to avoid structs
				titleRe := regexp.MustCompile(`"title":"([^"]+)"`)
				likesRe := regexp.MustCompile(`"likes_count":([0-9]+)`)
				viewsRe := regexp.MustCompile(`"playback_count":([0-9]+)`)
				dateRe := regexp.MustCompile(`"created_at":"([^"]+)"`)
				artRe := regexp.MustCompile(`"artwork_url":"([^"]+)"`)
				progRe := regexp.MustCompile(`"url":"([^"]+)","preset":"[^"]+","duration":[0-9]+,"snipped":false,"format":{"protocol":"progressive"`)
				
				bodyStr := string(bodyS)
				titleM := titleRe.FindStringSubmatch(bodyStr)
				if len(titleM) < 2 {
					sendMessage(ctx, "لم يتم العثور على الأغنية.")
					return
				}
				title := titleM[1]
				
				likes := "غير معروف"
				if m := likesRe.FindStringSubmatch(bodyStr); len(m) > 1 { likes = m[1] }
				views := "غير معروف"
				if m := viewsRe.FindStringSubmatch(bodyStr); len(m) > 1 { views = m[1] }
				date := "غير معروف"
				if m := dateRe.FindStringSubmatch(bodyStr); len(m) > 1 { 
					dateParts := strings.Split(m[1], "T")
					if len(dateParts) > 0 { date = dateParts[0] }
				}
				thumb := ""
				if m := artRe.FindStringSubmatch(bodyStr); len(m) > 1 { 
					thumb = strings.ReplaceAll(m[1], "-large.jpg", "-t500x500.jpg")
				}
				
				progUrl := ""
				if m := progRe.FindStringSubmatch(bodyStr); len(m) > 1 {
					progUrl = m[1]
				}
				
				if progUrl == "" {
					sendMessage(ctx, "المقطع محمي أو غير متوفر للتحميل.")
					return
				}
				
				// 3. Get MP3 Direct URL
				reqP, _ := http.NewRequest("GET", progUrl + "?client_id=" + clientID, nil)
				reqP.Header.Set("User-Agent", "Mozilla/5.0")
				respP, errP := http.DefaultClient.Do(reqP)
				if errP != nil {
					sendMessage(ctx, "فشل تجهيز المقطع.")
					return
				}
				bodyP, _ := io.ReadAll(respP.Body)
				respP.Body.Close()
				
				dlUrlRe := regexp.MustCompile(`"url":"([^"]+)"`)
				dlM := dlUrlRe.FindStringSubmatch(string(bodyP))
				if len(dlM) < 2 {
					sendMessage(ctx, "فشل استخراج رابط التحميل.")
					return
				}
				dlUrl := dlM[1]
				
				caption := fmt.Sprintf("*%s*\\n\\nاستماعات: %s\\nإعجابات: %s\\nتاريخ الرفع: %s", title, views, likes, date)
				if thumb != "" {
					sendImageFromURL(ctx, thumb, caption)
				} else {
					sendMessage(ctx, caption)
				}
				
				// 4. Download & Send MP3
				respDL, errDL := http.Get(dlUrl)
				if errDL != nil {
					sendMessage(ctx, "فشل تحميل الملف.")
					return
				}
				defer respDL.Body.Close()
				audioData, _ := io.ReadAll(respDL.Body)
				
				respUL, errUL := ctx.Client.Upload(context.Background(), audioData, whatsmeow.MediaAudio)
				if errUL != nil {
					sendMessage(ctx, "فشل رفع المقطع إلى واتساب.")
					return
				}
				
				msg := &waProto.Message{
					AudioMessage: &waProto.AudioMessage{
						URL:           proto.String(respUL.URL),
						DirectPath:    proto.String(respUL.DirectPath),
						MediaKey:      respUL.MediaKey,
						Mimetype:      proto.String("audio/mpeg"),
						FileEncSHA256: respUL.FileEncSHA256,
						FileSHA256:    respUL.FileSHA256,
						FileLength:    proto.Uint64(uint64(len(audioData))),
						PTT:           proto.Bool(false),
					},
				}
				
				_, _ = ctx.Client.SendMessage(context.Background(), ctx.Event.Info.Chat, msg)"""

    content = content[:start_idx] + new_code + content[end_idx:]

    with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
        f.write(content)
    print("Replaced with native SoundCloud API")
else:
    print("Could not find bounds")
