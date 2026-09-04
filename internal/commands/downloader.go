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
			searchUrl := "https://www.youtube.com/results?search_query=" + url.QueryEscape(query)
			reqS, _ := http.NewRequest("GET", searchUrl, nil)
			reqS.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
			respS, errS := http.DefaultClient.Do(reqS)
			if errS != nil {
				sendMessage(ctx, "حدث خطأ أثناء البحث في يوتيوب.")
				return
			}
			defer respS.Body.Close()
			bodyS, _ := io.ReadAll(respS.Body)
			
			re := regexp.MustCompile(`"videoId":"([^"]+)"`)
			matches := re.FindStringSubmatch(string(bodyS))
			if len(matches) < 2 {
				sendMessage(ctx, "لم يتم العثور على نتائج.")
				return
			}
			videoID := matches[1]
			videoURL := "https://www.youtube.com/watch?v=" + videoID
			
						args := []string{videoURL, "--no-warnings", "--print", "%(id)s|%(title)s|%(uploader)s|%(view_count)s|%(like_count)s|%(duration_string)s|%(upload_date)s|%(thumbnail)s", "--extractor-args", "youtube:player_client=android,ios"}
			if _, err := os.Stat("cookies.txt"); err == nil {
				args = append([]string{"--cookies", "cookies.txt"}, args...)
			}
			cmd := exec.Command("./yt-dlp", args...)
			out, err := cmd.CombinedOutput()
			if err != nil {
				sendMessage(ctx, "حدث خطأ أثناء جلب بيانات الأغنية:\n" + string(out)[:min(len(string(out)), 300)])
				return
			}
			
			lines := strings.Split(string(out), "\n")
			var dataLine string
			for _, line := range lines {
				if strings.Contains(line, "|") && !strings.Contains(line, "Deprecated") && !strings.Contains(line, "WARNING") {
					dataLine = line
					break
				}
			}
			
			if dataLine == "" {
				sendMessage(ctx, "لم يتم العثور على نتائج.")
				return
			}
			
			d := strings.Split(dataLine, "|")
			if len(d) < 8 {
				sendMessage(ctx, "خطأ في قراءة بيانات الأغنية.")
				return
			}
			
			id, title, uploader, views, likes, duration, date, thumb := d[0], d[1], d[2], d[3], d[4], d[5], d[6], d[7]
			
			caption := fmt.Sprintf("*%s*\n\nالقناة: %s\nالمشاهدات: %s\nالإعجابات: %s\nالمدة: %s\nتاريخ الرفع: %s", title, uploader, views, likes, duration, date)
			
			// Send thumbnail
			errThumb := sendImageFromURL(ctx, thumb, caption)
			if errThumb != nil {
				sendMessage(ctx, caption) // Fallback if image fails
			}
			
			// Now download audio
			url := "https://www.youtube.com/watch?v=" + id
			processDownload(ctx, url, "audio")
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
				args := []string{"--ffmpeg-location", "./ffmpeg", "-N", "4", "--no-check-certificate", "-f", "bestvideo[height<=720][ext=mp4]+bestaudio[ext=m4a]/best[height<=720][ext=mp4]/best", "--merge-output-format", "mp4", "--extractor-args", "youtube:player_client=android,ios", url, "-o", finalFile}
		if _, err := os.Stat("cookies.txt"); err == nil {
			args = append([]string{"--cookies", "cookies.txt"}, args...)
		}
		cmd = exec.Command("./yt-dlp", args...)
	} else {
		finalFile = tmpFile + ".mp3"
				args := []string{"--ffmpeg-location", "./ffmpeg", "-N", "4", "--no-check-certificate", "-f", "bestaudio/best", "-x", "--audio-format", "mp3", "--extractor-args", "youtube:player_client=android,ios", url, "-o", finalFile}
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
