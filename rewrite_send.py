import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'r') as f:
    content = f.read()

func_pattern = re.compile(r'func sendVideoData\(ctx \*BotContext, data \[\]byte, animeName, epNum string\).*?\n\}', re.DOTALL)

new_func = """func sendVideoData(ctx *BotContext, data []byte, animeName, epNum string) {
	// Just send it normally without splitting
	resp, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaVideo)
	if err != nil {
		sendMessage(ctx, "فشل في رفع الحلقة للواتساب.")
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
}"""

content = func_pattern.sub(new_func, content)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'w') as f:
    f.write(content)
