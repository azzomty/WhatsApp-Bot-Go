import re

with open('internal/commands/media.go', 'r') as f:
    content = f.read()

# Replace sendVideoData to handle > 64MB as Document
new_func = """
func sendVideoData(ctx *BotContext, data []byte, animeName, epNum string) {
	if len(data) > 64*1024*1024 {
		// Send as Document
		resp, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaDocument)
		if err != nil {
			sendMessage(ctx, "حدث خطأ أثناء رفع الحلقة للواتساب.")
			return
		}

		docMsg := &waProto.DocumentMessage{
			URL:           proto.String(resp.URL),
			DirectPath:    proto.String(resp.DirectPath),
			MediaKey:      resp.MediaKey,
			Mimetype:      proto.String("video/mp4"),
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
			Title:         proto.String(fmt.Sprintf("%s - الحلقة %s.mp4", animeName, epNum)),
			FileName:      proto.String(fmt.Sprintf("%s - الحلقة %s.mp4", animeName, epNum)),
			Caption:       proto.String(fmt.Sprintf("*%s* - الحلقة %s", animeName, epNum)),
		}

		ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
			DocumentMessage: docMsg,
		})
	} else {
		// Send as Video
		resp, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaVideo)
		if err != nil {
			sendMessage(ctx, "حدث خطأ أثناء رفع الحلقة للواتساب.")
			return
		}

		vidMsg := &waProto.VideoMessage{
			URL:           proto.String(resp.URL),
			DirectPath:    proto.String(resp.DirectPath),
			MediaKey:      resp.MediaKey,
			Mimetype:      proto.String("video/mp4"),
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
			Caption:       proto.String(fmt.Sprintf("*%s* - الحلقة %s", animeName, epNum)),
		}

		ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
			VideoMessage: vidMsg,
		})
	}
}
"""

content = re.sub(r'func sendVideoData\(ctx \*BotContext, data \[\]byte, animeName, epNum string\) \{.*?\n\}\n', new_func, content, flags=re.DOTALL)

with open('internal/commands/media.go', 'w') as f:
    f.write(content)
