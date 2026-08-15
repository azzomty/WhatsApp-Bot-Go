package commands

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"whatsapp-bot/internal/youtube"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

func interactiveYoutube(ctx *BotContext) {
	query := strings.TrimSpace(strings.TrimPrefix(ctx.Text, strings.Split(ctx.Text, " ")[0]))
	if query == "" {
		sendMessage(ctx, "اكتب اسم المقطع للبحث! مثلاً:\n.يوتيوب رابح صقر")
		return
	}

	videoIDs, _, err := youtube.SearchVideos(query, 5, "")
	if err != nil || len(videoIDs) == 0 {
		sendMessage(ctx, "ما قدرت ألقى نتائج، تأكد من مفتاح الـ API أو حاول باسم ثاني!")
		return
	}

	sessionKey := ctx.ChatID.String() + "_" + ctx.Sender.String()
	youtube.SetSession(sessionKey, &youtube.InteractiveSession{
		Query:        query,
		VideoIDs:     videoIDs,
		CurrentIndex: 0,
	})

	sendInteractiveResult(ctx, sessionKey)
}

func sendInteractiveResult(ctx *BotContext, sessionKey string) {
	session := youtube.GetSession(sessionKey)
	if session == nil {
		return
	}

	if session.CurrentIndex >= len(session.VideoIDs) {
		sendMessage(ctx, "خلصت كل النتائج! جرب تبحث بكلمات مختلفة.")
		youtube.DeleteSession(sessionKey)
		return
	}

	videoID := session.VideoIDs[session.CurrentIndex]

	info, err := youtube.GetVideoDetails(videoID)
	if err != nil {
		session.CurrentIndex++
		sendInteractiveResult(ctx, sessionKey)
		return
	}

	var thumbData []byte
	if info.Thumbnail != "" {
		if resp, err := http.Get(info.Thumbnail); err == nil {
			if resp.StatusCode != 200 {
				resp.Body.Close()
				if fallback, err2 := http.Get(fmt.Sprintf("https://i.ytimg.com/vi/%s/hqdefault.jpg", videoID)); err2 == nil {
					data, _ := io.ReadAll(fallback.Body)
					fallback.Body.Close()
					thumbData, _ = youtube.CropTo16x9(data)
				}
			} else {
				data, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				thumbData, _ = youtube.CropTo16x9(data)
			}
		}
	}

	caption := youtube.FormatCaption(info)
	caption += "\n\nهل هذا المقطع اللي تبغاه؟\nرد بـ (يب) للتحميل، أو (لا) للمقطع اللي بعده."

	var thumbMsgID string
	if len(thumbData) > 0 {
		uploadedThumb, err := ctx.Client.Upload(context.Background(), thumbData, whatsmeow.MediaImage)
		if err == nil {
			imgMsg := &waProto.ImageMessage{
				URL:           proto.String(uploadedThumb.URL),
				DirectPath:    proto.String(uploadedThumb.DirectPath),
				MediaKey:      uploadedThumb.MediaKey,
				Mimetype:      proto.String("image/jpeg"),
				FileEncSHA256: uploadedThumb.FileEncSHA256,
				FileSHA256:    uploadedThumb.FileSHA256,
				FileLength:    proto.Uint64(uint64(len(thumbData))),
				Caption:       proto.String(caption),
			}
			resp, _ := ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{ImageMessage: imgMsg})
			thumbMsgID = resp.ID
		}
	}

	if thumbMsgID == "" {
		ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
			ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String(caption)},
		})
	}
}

// handleInteractiveReply checks if the user is replying "يب" or "لا" to an interactive session
func handleInteractiveReply(ctx *BotContext) bool {
	msg := ctx.Event.Message
	if msg == nil {
		return false
	}

	var text string
	if msg.Conversation != nil {
		text = *msg.Conversation
	} else if msg.ExtendedTextMessage != nil && msg.ExtendedTextMessage.Text != nil {
		text = *msg.ExtendedTextMessage.Text
	} else {
		return false
	}

	text = strings.TrimSpace(strings.ToLower(text))
	if text != "يب" && text != "نعم" && text != "لا" {
		return false
	}

	sessionKey := ctx.ChatID.String() + "_" + ctx.Sender.String()
	session := youtube.GetSession(sessionKey)
	if session == nil {
		return false
	}

	if text == "لا" {
		session.CurrentIndex++
		sendInteractiveResult(ctx, sessionKey)
		return true
	}

	// Yes, download it!
	videoID := session.VideoIDs[session.CurrentIndex]
	youtube.DeleteSession(sessionKey)

	// We pass a dummy text to processYoutubeMedia so it uses the videoID
	// Actually we can't use processYoutubeMedia because it searches by query.
	// We need to download it directly.
	downloadInteractiveMedia(ctx, videoID)
	return true
}

func downloadInteractiveMedia(ctx *BotContext, videoID string) {
	sendMessage(ctx, "جاري تحميل المقطع...")
	link := "https://youtu.be/" + videoID
	downloadWithYtDlp(ctx, link, true)
}
