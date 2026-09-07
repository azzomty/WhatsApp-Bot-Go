import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    content = f.read()

start_str = '			cmd := exec.Command("./yt-dlp", "scsearch1:" + query'
end_str = '			processDownload(ctx, scURL, "audio")'

if start_str in content and end_str in content:
    start_idx = content.find(start_str)
    end_idx = content.find(end_str) + len(end_str)

    new_code = """			// 1. Search YouTube natively
				searchUrl := "https://www.youtube.com/results?search_query=" + url.QueryEscape(query)
				reqS, _ := http.NewRequest("GET", searchUrl, nil)
				reqS.Header.Set("User-Agent", "Mozilla/5.0")
				reqS.Header.Set("Cookie", "CONSENT=YES+cb.20210328-17-p0.en+FX+433;")
				respS, errS := http.DefaultClient.Do(reqS)
				if errS != nil {
					sendMessage(ctx, "حدث خطأ أثناء البحث في يوتيوب.")
					return
				}
				defer respS.Body.Close()
				bodyS, _ := io.ReadAll(respS.Body)
				
				reSearch := regexp.MustCompile(`"videoId":"([^"]+)"`)
				matches := reSearch.FindStringSubmatch(string(bodyS))
				if len(matches) < 2 {
					sendMessage(ctx, "لم يتم العثور على نتائج.")
					return
				}
				videoURL := "https://www.youtube.com/watch?v=" + matches[1]
				
				// 2. Call loader.to API
				loaderAPI := "https://loader.to/ajax/download.php?format=mp3&url=" + url.QueryEscape(videoURL)
				reqL, _ := http.NewRequest("GET", loaderAPI, nil)
				reqL.Header.Set("User-Agent", "Mozilla/5.0")
				respL, errL := http.DefaultClient.Do(reqL)
				if errL != nil {
					sendMessage(ctx, "حدث خطأ في خدمة التحميل.")
					return
				}
				defer respL.Body.Close()
				bodyL, _ := io.ReadAll(respL.Body)
				
				progressUrlRe := regexp.MustCompile(`"progress_url":"([^"]+)"`)
				titleRe := regexp.MustCompile(`"title":"([^"]+)"`)
				imgRe := regexp.MustCompile(`"image":"([^"]+)"`)
				
				pMatches := progressUrlRe.FindStringSubmatch(string(bodyL))
				if len(pMatches) < 2 {
					sendMessage(ctx, "فشل في تجهيز الأغنية من السيرفر.")
					return
				}
				progressURL := strings.ReplaceAll(pMatches[1], "\\\\/", "/")
				
				title := "مقطع صوتي"
				if tM := titleRe.FindStringSubmatch(string(bodyL)); len(tM) > 1 {
					title = tM[1]
				}
				
				thumb := ""
				if imgM := imgRe.FindStringSubmatch(string(bodyL)); len(imgM) > 1 {
					thumb = strings.ReplaceAll(imgM[1], "\\\\/", "/")
				}
				
				caption := fmt.Sprintf("*%s*\\n\\nجاري تجهيز المقطع الصوتي... ⏳", title)
				if thumb != "" {
					sendImageFromURL(ctx, thumb, caption)
				} else {
					sendMessage(ctx, caption)
				}
				
				// 4. Poll progress API
				downloadURL := ""
				dlRe := regexp.MustCompile(`"download_url":"([^"]+)"`)
				for i := 0; i < 20; i++ {
					time.Sleep(3 * time.Second)
					reqP, _ := http.NewRequest("GET", progressURL, nil)
					reqP.Header.Set("User-Agent", "Mozilla/5.0")
					respP, errP := http.DefaultClient.Do(reqP)
					if errP == nil {
						bodyP, _ := io.ReadAll(respP.Body)
						respP.Body.Close()
						if dMatches := dlRe.FindStringSubmatch(string(bodyP)); len(dMatches) > 1 {
							downloadURL = strings.ReplaceAll(dMatches[1], "\\\\/", "/")
							break
						}
					}
				}
				
				if downloadURL == "" {
					sendMessage(ctx, "❌ استغرق التجهيز وقتاً طويلاً. جرب مجدداً.")
					return
				}
				
				// 5. Download the final MP3
				respDL, errDL := http.Get(downloadURL)
				if errDL != nil {
					sendMessage(ctx, "❌ حدث خطأ أثناء تحميل الملف الصوتي.")
					return
				}
				defer respDL.Body.Close()
				
				audioData, _ := io.ReadAll(respDL.Body)
				
				// 6. Send the Audio!
				respUL, errUL := ctx.Client.Upload(context.Background(), audioData, whatsmeow.MediaAudio)
				if errUL != nil {
					sendMessage(ctx, "❌ فشل رفع المقطع إلى واتساب.")
					return
				}
				
				msg := &waProto.Message{
					AudioMessage: &waProto.AudioMessage{
						Url:           proto.String(respUL.URL),
						DirectPath:    proto.String(respUL.DirectPath),
						MediaKey:      respUL.MediaKey,
						Mimetype:      proto.String("audio/mpeg"),
						FileEncSha256: respUL.FileEncSHA256,
						FileSha256:    respUL.FileSHA256,
						FileLength:    proto.Uint64(uint64(len(audioData))),
						Ptt:           proto.Bool(false),
					},
				}
				
				_, _ = ctx.Client.SendMessage(context.Background(), ctx.Event.Info.Chat, msg)"""

    content = content[:start_idx] + new_code + content[end_idx:]

    # Add missing imports for url and regexp
    if '"net/url"' not in content:
        content = content.replace('"net/http"', '"net/http"\n\t"net/url"\n\t"regexp"')
    elif '"regexp"' not in content:
        content = content.replace('"net/url"', '"net/url"\n\t"regexp"')

    with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
        f.write(content)
    print("Replaced with loader.to API successfully")
else:
    print("Could not find start or end")
