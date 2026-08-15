package commands

import (
	"context"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

func convertToMp3(ctx *BotContext) {
	msg := ctx.Event.Message
	if msg == nil {
		sendMessage(ctx, "لازم ترد على مقطع فيديو عشان أحوله لصوت!")
		return
	}

	var videoMsg *waProto.VideoMessage
	if msg.ExtendedTextMessage != nil && msg.ExtendedTextMessage.ContextInfo != nil && msg.ExtendedTextMessage.ContextInfo.QuotedMessage != nil {
		quoted := msg.ExtendedTextMessage.ContextInfo.QuotedMessage
		if quoted.VideoMessage != nil {
			videoMsg = quoted.VideoMessage
		}
	}

	if videoMsg == nil {
		sendMessage(ctx, "لازم ترد على مقطع فيديو عشان أحوله لصوت!")
		return
	}

	sendMessage(ctx, "جاري تحويل الفيديو إلى صوت...")

	// Download the video
	data, err := ctx.Client.Download(context.Background(), videoMsg)
	if err != nil {
		sendMessage(ctx, "فشل تحميل الفيديو للتحويل!")
		return
	}

	// Upload as audio
	uploadedMedia, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaAudio)
	if err != nil {
		sendMessage(ctx, "فشل رفع الصوت للواتساب!")
		return
	}

	finalMsg := &waProto.Message{
		AudioMessage: &waProto.AudioMessage{
			URL:           proto.String(uploadedMedia.URL),
			DirectPath:    proto.String(uploadedMedia.DirectPath),
			MediaKey:      uploadedMedia.MediaKey,
			Mimetype:      proto.String("audio/mp4"),
			FileEncSHA256: uploadedMedia.FileEncSHA256,
			FileSHA256:    uploadedMedia.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
			PTT:           proto.Bool(false),
			ContextInfo: &waProto.ContextInfo{
				StanzaID:    proto.String(ctx.Event.Info.ID),
				Participant: proto.String(ctx.Sender.ToNonAD().String()),
			},
		},
	}

	ctx.Client.SendMessage(context.Background(), ctx.ChatID, finalMsg)
}
