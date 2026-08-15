package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

type TikwmResponse struct {
	Code int `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Play string `json:"play"`
	} `json:"data"`
}

func universalDownload(ctx *BotContext) {
	query := strings.TrimSpace(strings.TrimPrefix(ctx.Text, strings.Split(ctx.Text, " ")[0]))
	if query == "" {
		sendMessage(ctx, "اكتب رابط المقطع مع الأمر! مثلاً:\n.تحميل https://www.tiktok.com/...")
		return
	}

	sendMessage(ctx, "جاري سحب المقطع من الرابط...")

	// If it's TikTok, use tikwm API because yt-dlp gets IP blocked
	if strings.Contains(query, "tiktok.com") {
		downloadTikTok(ctx, query)
		return
	}

	// For all other platforms, use yt-dlp
	downloadWithYtDlp(ctx, query, false)
}

func downloadTikTok(ctx *BotContext, link string) {
	apiURL := "https://www.tikwm.com/api/?url=" + url.QueryEscape(link)
	resp, err := http.Get(apiURL)
	if err != nil {
		sendMessage(ctx, "فشل الاتصال بسيرفر تيك توك!")
		return
	}
	defer resp.Body.Close()

	var result TikwmResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		sendMessage(ctx, "فشل تحليل استجابة السيرفر!")
		return
	}

	if result.Code != 0 || result.Data.Play == "" {
		sendMessage(ctx, "ما قدرت أسحب المقطع، تأكد من الرابط أو جرب مقطع ثاني!")
		return
	}

	mediaResp, err := http.Get(result.Data.Play)
	if err != nil || mediaResp.StatusCode != 200 {
		sendMessage(ctx, "فشل تحميل المقطع بعد استخراجه!")
		return
	}
	defer mediaResp.Body.Close()

	mediaData, _ := io.ReadAll(mediaResp.Body)
	sendMediaData(ctx, mediaData, "video/mp4", whatsmeow.MediaVideo)
}

func downloadWithYtDlp(ctx *BotContext, link string, silentFail bool) {
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("dl_%d.mp4", time.Now().UnixNano()))
	
	cmd := exec.Command("./yt-dlp", "-f", "b", "-o", tmpFile, link)
	err := cmd.Run()
	if err != nil {
		if !silentFail {
			sendMessage(ctx, "فشل التحميل! ممكن الموقع غير مدعوم أو الرابط غلط.")
		}
		return
	}
	defer os.Remove(tmpFile)

	mediaData, err := os.ReadFile(tmpFile)
	if err != nil {
		if !silentFail {
			sendMessage(ctx, "فشل قراءة الملف بعد التحميل!")
		}
		return
	}

	sendMediaData(ctx, mediaData, "video/mp4", whatsmeow.MediaVideo)
}

func sendMediaData(ctx *BotContext, data []byte, mimeType string, mediaType whatsmeow.MediaType) {
	uploadedMedia, err := ctx.Client.Upload(context.Background(), data, mediaType)
	if err != nil {
		sendMessage(ctx, "فشل رفع المقطع للواتساب!")
		return
	}

	finalMsg := &waProto.Message{
		VideoMessage: &waProto.VideoMessage{
			URL:           proto.String(uploadedMedia.URL),
			DirectPath:    proto.String(uploadedMedia.DirectPath),
			MediaKey:      uploadedMedia.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploadedMedia.FileEncSHA256,
			FileSHA256:    uploadedMedia.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
		},
	}

	ctx.Client.SendMessage(context.Background(), ctx.ChatID, finalMsg)
}
