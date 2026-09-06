package commands

import (
	"context"
	"fmt"
"net/http"
	"net/url"
	"regexp"
"io"
	"os"
"path/filepath"
"io/ioutil"
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
			// 1. Fetch Soundcloud Client ID
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
				
								authRe := regexp.MustCompile(`"track_authorization":"([^"]+)"`)
				auth := ""
				if m := authRe.FindStringSubmatch(bodyStr); len(m) > 1 {
					auth = m[1]
				}
				progUrl := ""
				isHLS := false
				if m := progRe.FindStringSubmatch(bodyStr); len(m) > 1 {
					progUrl = m[1]
				} else {
					hlsRe := regexp.MustCompile(`"url":"([^"]+)","preset":"[^"]+","duration":[0-9]+,"snipped":false,"format":{"protocol":"hls"`)
					if m := hlsRe.FindStringSubmatch(bodyStr); len(m) > 1 {
						progUrl = m[1]
						isHLS = true
					}
				}
				
				if progUrl == "" {
					sendMessage(ctx, "المقطع محمي أو غير متوفر للتحميل.")
					return
				}
				
				// 3. Get MP3 Direct URL
				reqP, _ := http.NewRequest("GET", progUrl + "?client_id=" + clientID + "&track_authorization=" + auth, nil)
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
					snippet := string(bodyP)
					if len(snippet) > 50 { snippet = snippet[:50] }
					sendMessage(ctx, "فشل استخراج رابط التحميل. السبب: " + snippet)
					return
				}
				dlUrl := dlM[1]
				
				caption := fmt.Sprintf("*%s*\n\nاستماعات: %s\nإعجابات: %s\nتاريخ الرفع: %s", title, views, likes, date)
				if thumb != "" {
					sendImageFromURL(ctx, thumb, caption)
				} else {
					sendMessage(ctx, caption)
				}
				
				// 4. Download & Send MP3
					var audioData []byte
					if isHLS {
						tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("sc_%d.mp3", time.Now().UnixNano()))
						cmd := exec.Command("./ffmpeg", "-y", "-i", dlUrl, "-c:a", "libmp3lame", "-q:a", "2", tmpFile)
						err := cmd.Run()
						if err != nil {
							sendMessage(ctx, "حدث خطأ أثناء تحويل الملف الصوتي.")
							return
						}
						audioData, _ = ioutil.ReadFile(tmpFile)
						os.Remove(tmpFile)
					} else {
						respDL, errDL := http.Get(dlUrl)
						if errDL != nil {
							sendMessage(ctx, "فشل تحميل الملف.")
							return
						}
						defer respDL.Body.Close()
						audioData, _ = io.ReadAll(respDL.Body)
					}
				
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
