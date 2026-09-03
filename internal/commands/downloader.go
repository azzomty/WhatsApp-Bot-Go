package commands

import (
	"context"
	"fmt"
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
	if strings.HasPrefix(text, ".تحميل") {
		parts := strings.Fields(text)
		if len(parts) < 2 {
			sendMessage(ctx, "يرجى وضع رابط بعد الأمر.\nمثال: .تحميل https://youtu.be/...")
			return true
		}
		url := parts[1]
		
		downloadSessions[ctx.Sender.User] = &DownloaderSession{URL: url}
		sendMessage(ctx, "هل تريد تحميله كـ صوت 🎵 أم كـ فيديو 🎬؟\n\nأرسل:\n.صوت\nأو\n.فيديو")
		return true
	}
	
	session, exists := downloadSessions[ctx.Sender.User]
	if exists {
		if text == ".صوت" || text == ".فيديو" {
			delete(downloadSessions, ctx.Sender.User)
			if text == ".صوت" {
				sendMessage(ctx, "⏳ جاري تحميل الصوت، يرجى الانتظار...")
				go processDownload(ctx, session.URL, "audio")
			} else {
				sendMessage(ctx, "⏳ جاري تحميل الفيديو، يرجى الانتظار...")
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
		cmd = exec.Command("./yt-dlp", "--ffmpeg-location", "./ffmpeg", "-N", "4", "--no-check-certificate", "-f", "bestvideo[height<=720][ext=mp4]+bestaudio[ext=m4a]/best[height<=720][ext=mp4]/best", "--merge-output-format", "mp4", url, "-o", finalFile)
	} else {
		finalFile = tmpFile + ".mp3"
		cmd = exec.Command("./yt-dlp", "--ffmpeg-location", "./ffmpeg", "-N", "4", "--no-check-certificate", "-f", "bestaudio/best", "-x", "--audio-format", "mp3", url, "-o", finalFile)
	}
	
	out, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "Unsupported URL") {
			sendMessage(ctx, "❌ الرابط غير مدعوم أو غير صحيح.")
		} else if strings.Contains(string(out), "Sign in to confirm") || strings.Contains(string(out), "HTTP Error 403") || strings.Contains(string(out), "blocked") {
			sendMessage(ctx, "❌ تم حظر تحميل هذا الرابط مؤقتاً من قبل الموقع (حماية ضد البوتات). جرب رابط آخر أو موقع آخر.")
		} else {
			outStr := string(out)
			if len(outStr) > 300 {
				outStr = outStr[:300]
			}
			sendMessage(ctx, "❌ حدث خطأ أثناء التحميل:\n" + outStr)
		}
		return
	}
	
	data, err := os.ReadFile(finalFile)
	if err != nil {
		sendMessage(ctx, "❌ فشل في قراءة الملف بعد التحميل.")
		return
	}
	
	if mode == "video" {
		err = sendVideoDataGeneric(ctx, data)
	} else {
		err = sendAudioData(ctx, data)
	}
	
	if err != nil {
		sendMessage(ctx, fmt.Sprintf("❌ فشل إرسال الملف (قد يكون حجمه كبيراً جداً للواتساب). الخطأ: %v", err))
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
