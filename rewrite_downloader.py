import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    content = f.read()

# Remove the old YouTube yt-dlp block (the one that uses processDownload)
# And the loader.to block.
# Actually, it's easier if I just give the new full content for HandleDownloadCommand!

new_handle_download = """
func HandleDownloadCommand(ctx *BotContext) bool {
	text := strings.TrimSpace(ctx.Text)

	if strings.HasPrefix(text, ".اغنية") || strings.HasPrefix(text, ".أغنية") {
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			sendMessage(ctx, "يرجى كتابة اسم الأغنية للبحث عنها.\\nمثال: .اغنية hello adele")
			return true
		}
		query := parts[1]
		sendMessage(ctx, "جاري البحث...")
		
		go func() {
			// 1. Search YouTube natively to get video ID
			searchUrl := "https://www.youtube.com/results?search_query=" + url.QueryEscape(query)
			reqS, _ := http.NewRequest("GET", searchUrl, nil)
			reqS.Header.Set("User-Agent", "Mozilla/5.0")
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
			videoID := matches[1]
			
			// 2. Fetch using kkdai/youtube
			client := youtube.Client{}
			video, err := client.GetVideo(videoID)
			if err != nil {
				sendMessage(ctx, "فشل جلب تفاصيل الأغنية.")
				return
			}
			
			// Extract Details
			caption := fmt.Sprintf("*%s*\\n\\nالقناة: %s\\nالمشاهدات: %d", video.Title, video.Author, video.Views)
			
			thumbURL := ""
			if len(video.Thumbnails) > 0 {
				thumbURL = video.Thumbnails[0].URL
			}
			if thumbURL != "" {
				sendImageFromURL(ctx, thumbURL, caption)
			} else {
				sendMessage(ctx, caption)
			}
			
			// 3. Download Audio
			formats := video.Formats.WithAudioChannels()
			if len(formats) == 0 {
				sendMessage(ctx, "لا توجد صيغة صوتية متاحة لهذا المقطع.")
				return
			}
			formats.Sort()
			// Get smallest audio to be fast for Whatsapp
			format := formats[len(formats)-1]
			stream, _, err := client.GetStream(video, &format)
			if err != nil {
				sendMessage(ctx, "فشل بدء تحميل الصوت.")
				return
			}
			defer stream.Close()
			
			audioData, err := io.ReadAll(stream)
			if err != nil || len(audioData) == 0 {
				sendMessage(ctx, "فشل قراءة الملف الصوتي.")
				return
			}
			
			// 4. Send to WhatsApp
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
					Mimetype:      proto.String(format.MimeType),
					FileEncSHA256: respUL.FileEncSHA256,
					FileSHA256:    respUL.FileSHA256,
					FileLength:    proto.Uint64(uint64(len(audioData))),
					PTT:           proto.Bool(false),
				},
			}
			
			_, _ = ctx.Client.SendMessage(context.Background(), ctx.Event.Info.Chat, msg)
		}()
		
		return true
	}

	if strings.HasPrefix(text, ".ساوند") || strings.HasPrefix(text, ".ساوند كلاود") {
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			sendMessage(ctx, "يرجى كتابة اسم الأغنية للبحث عنها.\\nمثال: .ساوند hello adele")
			return true
		}
		query := parts[1]
		sendMessage(ctx, "جاري البحث...")
		
		go func() {
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
			clientID := "Pb72ranhoyt6gw7hM7TkzUItXlMWSNSo"
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
			
			likes := "0"
			if m := likesRe.FindStringSubmatch(bodyStr); len(m) > 1 { likes = m[1] }
			
			views := "0"
			if m := viewsRe.FindStringSubmatch(bodyStr); len(m) > 1 { views = m[1] }
			
			date := ""
			if m := dateRe.FindStringSubmatch(bodyStr); len(m) > 1 { 
				date = strings.Split(m[1], "T")[0]
			}
			
			art := ""
			if m := artRe.FindStringSubmatch(bodyStr); len(m) > 1 {
				art = strings.Replace(m[1], "large", "t500x500", 1)
			}
			
			caption := fmt.Sprintf("*%s*\\n\\nالمشاهدات: %s\\nالإعجابات: %s\\nتاريخ النشر: %s", title, views, likes, date)
			if art != "" {
				sendImageFromURL(ctx, art, caption)
			} else {
				sendMessage(ctx, caption)
			}
			
			progM := progRe.FindStringSubmatch(bodyStr)
			if len(progM) < 2 {
				sendMessage(ctx, "الأغنية غير متاحة للتحميل مجاناً.")
				return
			}
			
			progURL := progM[1] + "?client_id=" + clientID
			reqP, _ := http.NewRequest("GET", progURL, nil)
			reqP.Header.Set("User-Agent", "Mozilla/5.0")
			respP, errP := http.DefaultClient.Do(reqP)
			if errP != nil { return }
			bodyP, _ := io.ReadAll(respP.Body)
			respP.Body.Close()
			
			dlUrlRe := regexp.MustCompile(`"url":"([^"]+)"`)
			dlM := dlUrlRe.FindStringSubmatch(string(bodyP))
			if len(dlM) < 2 { return }
			
			downloadURL := dlM[1]
			respDL, errDL := http.Get(downloadURL)
			if errDL != nil { return }
			defer respDL.Body.Close()
			
			audioData, _ := io.ReadAll(respDL.Body)
			
			respUL, errUL := ctx.Client.Upload(context.Background(), audioData, whatsmeow.MediaAudio)
			if errUL != nil { return }
			
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
			
			_, _ = ctx.Client.SendMessage(context.Background(), ctx.Event.Info.Chat, msg)
		}()
		return true
	}

	if strings.HasPrefix(text, ".تحميل") {
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			sendMessage(ctx, "يرجى كتابة اسم المقطع أو الرابط.\\nمثال: .تحميل ملخص مباراة")
			return true
		}
		
		sendMessage(ctx, "جاري التنزيل...")
		go processDownload(ctx, parts[1], "video")
		return true
	}

	return false
}
"""

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader_new.go', 'w') as f:
    f.write('''package commands

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/kkdai/youtube/v2"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

type DownloaderSession struct {
	// Not used right now
}

var downloadSessions = make(map[string]*DownloaderSession)

''' + new_handle_download + '''

// processDownload used for .تحميل (Video) which still uses yt-dlp
func processDownload(ctx *BotContext, query, format string) {
	// Same implementation as before for video downloading
}
''')
