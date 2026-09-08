package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

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
			// Search and extract details using yt-dlp
			cmdSearch := exec.Command("./yt-dlp", "ytsearch1:"+query, "--dump-json", "--extractor-args", "youtube:player_client=android,web")
			out, errS := cmdSearch.CombinedOutput()
			if errS != nil || len(out) < 10 {
				sendMessage(ctx, "لم يتم العثور على الأغنية.")
				return
			}

			// Try to find the JSON line in the output (yt-dlp prints warnings)
			var jsonStr string
			for _, line := range strings.Split(string(out), "\n") {
				if strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}") {
					jsonStr = line
					break
				}
			}

			if jsonStr == "" {
				sendMessage(ctx, "فشل جلب تفاصيل الأغنية.")
				return
			}

			var video struct {
				Title      string `json:"title"`
				Channel    string `json:"uploader"`
				Views      int    `json:"view_count"`
				Thumbnail  string `json:"thumbnail"`
				WebpageURL string `json:"webpage_url"`
			}

			if err := json.Unmarshal([]byte(jsonStr), &video); err != nil {
				sendMessage(ctx, "فشل قراءة تفاصيل الأغنية.")
				return
			}

			// Extract Details
			caption := fmt.Sprintf("*%s*\n\nالقناة: %s\nالمشاهدات: %d", video.Title, video.Channel, video.Views)

			if video.Thumbnail != "" {
				sendImageFromURL(ctx, video.Thumbnail, caption)
			} else {
				sendMessage(ctx, caption)
			}

			// Download Audio
			processDownload(ctx, video.WebpageURL, "audio")
		}()
		return true
	}

	if strings.HasPrefix(text, ".تحميل") {
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			sendMessage(ctx, "يرجى كتابة اسم المقطع أو الرابط.\nمثال: .تحميل ملخص مباراة")
			return true
		}

		sendMessage(ctx, "جاري التنزيل...")
		go processDownload(ctx, parts[1], "video")
		return true
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
		args := []string{"--ffmpeg-location", "./ffmpeg", "-N", "4", "--no-check-certificate", "--extractor-args", "youtube:player_client=android,web", "-f", "bestvideo[height<=720][ext=mp4]+bestaudio[ext=m4a]/best[height<=720][ext=mp4]/best", "--merge-output-format", "mp4", url, "-o", finalFile}
		if _, err := os.Stat("cookies.txt"); err == nil {
			args = append([]string{"--cookies", "cookies.txt"}, args...)
		}
		cmd = exec.Command("./yt-dlp", args...)
	} else {
		finalFile = tmpFile + ".mp3"
		args := []string{"--ffmpeg-location", "./ffmpeg", "-N", "4", "--no-check-certificate", "--extractor-args", "youtube:player_client=android,web", "-f", "bestaudio/best", "-x", "--audio-format", "mp3", url, "-o", finalFile}
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
			sendMessage(ctx, "حدث خطأ أثناء التحميل:\n"+outStr)
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
