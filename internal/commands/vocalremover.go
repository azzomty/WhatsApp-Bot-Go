package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

func HandleRemoveMusic(ctx *BotContext) {
	quoted := ctx.Event.Message.ExtendedTextMessage
	if quoted == nil || quoted.ContextInfo == nil || quoted.ContextInfo.QuotedMessage == nil {
		sendMessage(ctx, "يرجى الرد على فيديو بكتابة: .كتم")
		return
	}

	qMsg := quoted.ContextInfo.QuotedMessage

	var data []byte
	var err error
	var ext string

	if qMsg.VideoMessage != nil {
		data, err = ctx.Client.Download(context.Background(), qMsg.VideoMessage)
		ext = ".mp4"
	} else {
		sendMessage(ctx, "يرجى الرد على فيديو فقط لكتم صوته!")
		return
	}

	if err != nil {
		sendMessage(ctx, "فشل تحميل المقطع.")
		return
	}

	sendMessage(ctx, "⏳ جاري كتم الصوت من الفيديو...")

	tmpDir, _ := os.MkdirTemp("", "vocal_remover")
	defer os.RemoveAll(tmpDir)

	inputFile := filepath.Join(tmpDir, "input"+ext)
	os.WriteFile(inputFile, data, 0644)

	outputFile := filepath.Join(tmpDir, "output_muted"+ext)

	ffmpegPath := "/home/lennox/Desktop/اهها/Go_Bot/node_modules/ffmpeg-static/ffmpeg"
	
	// Remove audio completely using -an
	cmd := exec.Command(ffmpegPath, "-y", "-i", inputFile, "-c:v", "copy", "-an", outputFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Mute Error:", string(out))
		sendMessage(ctx, "حدث خطأ أثناء معالجة الفيديو.")
		return
	}

	outData, err := os.ReadFile(outputFile)
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء قراءة الملف المعالج.")
		return
	}

	resp, err := ctx.Client.Upload(context.Background(), outData, whatsmeow.MediaVideo) 
	if err != nil {
		sendMessage(ctx, "فشل رفع الملف.")
		return
	}

	msg := &waProto.Message{}
	msg.VideoMessage = &waProto.VideoMessage{
		URL:           proto.String(resp.URL),
		DirectPath:    proto.String(resp.DirectPath),
		MediaKey:      resp.MediaKey,
		FileEncSHA256: resp.FileEncSHA256,
		FileSHA256:    resp.FileSHA256,
		FileLength:    proto.Uint64(uint64(len(outData))),
		Mimetype:      proto.String("video/mp4"),
		Caption:       proto.String("🔇 تم كتم الصوت من الفيديو بنجاح!"),
	}

	ctx.Client.SendMessage(context.Background(), ctx.ChatID, msg)
}
