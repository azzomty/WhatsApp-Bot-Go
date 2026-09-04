package commands

import (
	"context"
	"fmt"
"net/http"
	"net/url"
	"regexp"
"io"
	"os"
	"os/exec"
	"strings"
	"time"
	"strconv"
	
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

type DownloaderSession struct {
	URL string
}

var downloadSessions = make(map[string]*DownloaderSession)

func HandleDownloadCommand(ctx *BotContext) bool {
	text := strings.TrimSpace(ctx.Text)
		if strings.HasPrefix(text, ".اغنية") || strings.HasPrefix(text, ".أغنية") {
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			sendMessage(ctx, "يرجى كتابة اسم الأغنية للبحث عنها.\nمثال: .اغنية hello adele")
			return true
		}
		query := parts[1]
		sendMessage(ctx, "جاري البحث...")
		
		go func() {
			// 1. Search YouTube natively
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
				progressURL := strings.ReplaceAll(pMatches[1], "\\/", "/")
				
				title := "مقطع صوتي"
				if tM := titleRe.FindStringSubmatch(string(bodyL)); len(tM) > 1 {
					title = tM[1]
				}
				
				thumb := ""
				if imgM := imgRe.FindStringSubmatch(string(bodyL)); len(imgM) > 1 {
					thumb = strings.ReplaceAll(imgM[1], "\\/", "/")
				}
				
				caption := fmt.Sprintf("*%s*\n\nجاري تجهيز المقطع الصوتي... ⏳", title)
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
							downloadURL = strings.ReplaceAll(dMatches[1], "\\/", "/")
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
		parts := strings.Fields(text)
		if len(parts) < 2 {
			sendMessage(ctx, "يرجى وضع رابط بعد الأمر.\nمثال: .تحميل https://youtu.be/...")
			return true
		}
		url := parts[1]
		
		downloadSessions[ctx.Sender.User] = &DownloaderSession{URL: url}
		sendMessage(ctx, "هل تريد تحميله كصوت أم كفيديو؟\nأرسل:\n.صوت\nأو\n.فيديو")
		return true
	}
	
	session, exists := downloadSessions[ctx.Sender.User]
	if exists {
		if text == ".صوت" || text == ".فيديو" {
			delete(downloadSessions, ctx.Sender.User)
			if text == ".صوت" {
				sendMessage(ctx, "جاري تحميل الصوت، يرجى الانتظار...")
				go processDownload(ctx, session.URL, "audio")
			} else {
				sendMessage(ctx, "جاري تحميل الفيديو، يرجى الانتظار...")
				go processDownload(ctx, session.URL, "video")
			}
			return true
		}
	}
	
	return false
}

func processDownload(ctx *BotContext, url string, mode string) {
	tmpFile := "/tmp/dl_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	defer os.Remove(tmpFile + ".mp4")
	defer os.Remove(tmpFile + ".mp3")
	defer os.Remove(tmpFile + ".webm")
	defer os.Remove(tmpFile + ".m4a")
	
	var cmd *exec.Cmd
	var finalFile string
	
	if mode == "video" {
		finalFile = tmpFile + ".mp4"
				args := []string{"--ffmpeg-location", "./ffmpeg", "-N", "4", "--no-check-certificate", "-f", "bestvideo[height<=720][ext=mp4]+bestaudio[ext=m4a]/best[height<=720][ext=mp4]/best", "--merge-output-format", "mp4", url, "-o", finalFile}
		if _, err := os.Stat("cookies.txt"); err == nil {
			args = append([]string{"--cookies", "cookies.txt"}, args...)
		}
		cmd = exec.Command("./yt-dlp", args...)
	} else {
		finalFile = tmpFile + ".mp3"
				args := []string{"--ffmpeg-location", "./ffmpeg", "-N", "4", "--no-check-certificate", "-f", "bestaudio/best", "-x", "--audio-format", "mp3", url, "-o", finalFile}
		if _, err := os.Stat("cookies.txt"); err == nil {
			args = append([]string{"--cookies", "cookies.txt"}, args...)
		}
		cmd = exec.Command("./yt-dlp", args...)
	}
	
	out, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "Unsupported URL") {
			sendMessage(ctx, "الرابط غير مدعوم أو غير صحيح.")
		} else if strings.Contains(string(out), "Sign in to confirm") || strings.Contains(string(out), "HTTP Error 403") || strings.Contains(string(out), "blocked") {
			sendMessage(ctx, "تم حظر تحميل هذا الرابط مؤقتاً من قبل الموقع. جرب رابط آخر.")
		} else {
			outStr := string(out)
			if len(outStr) > 300 {
				outStr = outStr[:300]
			}
			sendMessage(ctx, "حدث خطأ أثناء التحميل:\n" + outStr)
		}
		return
	}
	
	data, err := os.ReadFile(finalFile)
	if err != nil {
		sendMessage(ctx, "فشل في قراءة الملف بعد التحميل.")
		return
	}
	
	if mode == "video" {
		err = sendVideoDataGeneric(ctx, data)
	} else {
		err = sendAudioData(ctx, data)
	}
	
	if err != nil {
		sendMessage(ctx, fmt.Sprintf("فشل إرسال الملف. الخطأ: %v", err))
	}
}

func sendAudioData(ctx *BotContext, data []byte) error {
	uploaded, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaAudio)
	if err != nil {
		return err
	}
	msg := &waProto.Message{
		AudioMessage: &waProto.AudioMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String("audio/mpeg"),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
		},
	}
	_, err = ctx.Client.SendMessage(context.Background(), ctx.ChatID, msg)
	return err
}

func sendVideoDataGeneric(ctx *BotContext, data []byte) error {
	uploaded, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaVideo)
	if err != nil {
		return err
	}
	msg := &waProto.Message{
		VideoMessage: &waProto.VideoMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String("video/mp4"),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
		},
	}
	_, err = ctx.Client.SendMessage(context.Background(), ctx.ChatID, msg)
	return err
}

func sendImageFromURL(ctx *BotContext, imgUrl string, caption string) error {
	resp, err := http.Get(imgUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	
	uploaded, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaImage)
	if err != nil {
		return err
	}
	
	msg := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			Caption:       proto.String(caption),
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String("image/jpeg"),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
		},
	}
	_, err = ctx.Client.SendMessage(context.Background(), ctx.ChatID, msg)
	return err
}
