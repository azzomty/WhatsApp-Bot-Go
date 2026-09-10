package commands

import (
	"context"
	"strings"

	"whatsapp-bot/internal/stickers"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

func HandleFontImageCommand(ctx *BotContext) bool {
	text := strings.TrimSpace(ctx.Text)

	if strings.HasPrefix(text, ".صورة_خط") || strings.HasPrefix(text, ".ملصق_خط") || strings.HasPrefix(text, ".خط") {
		cmdPrefix := ""
		if strings.HasPrefix(text, ".صورة_خط") { cmdPrefix = ".صورة_خط" }
		if strings.HasPrefix(text, ".ملصق_خط") { cmdPrefix = ".ملصق_خط" }
		if strings.HasPrefix(text, ".خط") { cmdPrefix = ".خط" }
		
		text = strings.TrimPrefix(text, cmdPrefix)
		text = strings.TrimSpace(text)
		
		parts := strings.Split(text, "|")
		if len(parts) < 2 {
			sendMessage(ctx, "طريقة الاستخدام:\n.خط النص | اسم الخط | لون الخلفية | لون النص\n\nمثال:\n.خط I Love Assuity | Creattion Demo | transparent | gold")
			return true
		}
		
		textContent := strings.TrimSpace(parts[0])
		fontName := strings.TrimSpace(parts[1])
		bgColor := "transparent"
		fgColor := "black"
		
		if len(parts) > 2 {
			bgColor = strings.TrimSpace(parts[2])
		}
		if len(parts) > 3 {
			fgColor = strings.TrimSpace(parts[3])
		}

		imgBytes, err := GenerateTextImage(textContent, fontName, bgColor, fgColor)
		if err != nil {
			sendMessage(ctx, "حدث خطأ: " + err.Error())
			return true
		}
		
		// Convert PNG bytes to WhatsApp WebP Sticker
		webpData, err := stickers.GenerateSticker(imgBytes, false, "خطوط البوت", "Assuity")
		if err != nil {
			sendMessage(ctx, "فشل صنع الملصق.")
			return true
		}

		resp, err := ctx.Client.Upload(context.Background(), webpData, whatsmeow.MediaImage)
		if err != nil {
			sendMessage(ctx, "فشل في رفع الصورة.")
			return true
		}

		msg := &waProto.Message{
			StickerMessage: &waProto.StickerMessage{
				URL:           proto.String(resp.URL),
				DirectPath:    proto.String(resp.DirectPath),
				MediaKey:      resp.MediaKey,
				Mimetype:      proto.String("image/webp"),
				FileEncSHA256: resp.FileEncSHA256,
				FileSHA256:    resp.FileSHA256,
				FileLength:    proto.Uint64(uint64(len(webpData))),
			},
		}

		_, _ = ctx.Client.SendMessage(context.Background(), ctx.Event.Info.Chat, msg)
		return true
	}
	
	return false
}
