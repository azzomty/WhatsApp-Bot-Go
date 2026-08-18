package commands

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

func HandlePDF(ctx *BotContext) {
	parts := strings.Split(ctx.Text, " ")
	if len(parts) < 2 {
		sendMessage(ctx, "الصيغة غير صحيحة! جرب مثلاً:\n.pdf 5\n(عشان اجمع لك آخر 5 صور في الشات بملف PDF واحد)")
		return
	}

	count, err := strconv.Atoi(parts[1])
	if err != nil || count <= 0 {
		sendMessage(ctx, "الرقم غير صحيح! حط رقم صحيح، مثلاً 5")
		return
	}

	if count > 30 {
		sendMessage(ctx, "الحد الأقصى هو 30 صورة في ملف واحد عشان ما يعلق البوت!")
		return
	}

	sendMessage(ctx, fmt.Sprintf("جاري تجميع آخر %d صور وتحويلها لـ PDF... ⏳", count))

	msgMutex.Lock()
	msgs := MessageStore[ctx.ChatID.String()]
	msgMutex.Unlock()

	var imagesData [][]byte

	// Iterate backwards over messages
	for i := len(msgs) - 1; i >= 0; i-- {
		m := msgs[i]
		if m.Info.ID == ctx.Event.Info.ID {
			continue // Skip the command message itself
		}

		uMsg := UnwrapMessage(m.Message)
		if uMsg != nil && uMsg.GetImageMessage() != nil {
			data, err := ctx.Client.Download(context.Background(), uMsg.GetImageMessage())
			if err == nil {
				imagesData = append([][]byte{data}, imagesData...) // Prepend to keep chronological order
			}
			if len(imagesData) == count {
				break
			}
		}
	}

	if len(imagesData) == 0 {
		sendMessage(ctx, "مالقيت أي صور في الشات السابقة! أرسل الصور أولاً وبعدين جرب الأمر.")
		return
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	
	for i, data := range imagesData {
		img, format, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			continue
		}

		// A4 size is 210x297mm
		pageWidth := 210.0
		pageHeight := 297.0

		// Calculate scaled dimensions fitting A4
		imgRatio := float64(img.Width) / float64(img.Height)
		pageRatio := pageWidth / pageHeight

		var finalW, finalH float64
		if imgRatio > pageRatio {
			finalW = pageWidth
			finalH = pageWidth / imgRatio
		} else {
			finalH = pageHeight
			finalW = pageHeight * imgRatio
		}

		xPos := (pageWidth - finalW) / 2
		yPos := (pageHeight - finalH) / 2

		pdf.AddPage()
		
		opt := gofpdf.ImageOptions{
			ImageType: strings.ToUpper(format),
			ReadDpi:   false,
		}

		// Use memory reader for gofpdf
		pdf.RegisterImageOptionsReader(fmt.Sprintf("img_%d", i), opt, bytes.NewReader(data))
		pdf.ImageOptions(fmt.Sprintf("img_%d", i), xPos, yPos, finalW, finalH, false, opt, 0, "")
	}

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء صناعة ملف الـ PDF.")
		return
	}

	pdfData := buf.Bytes()
	resp, err := ctx.Client.Upload(context.Background(), pdfData, whatsmeow.MediaDocument)
	if err != nil {
		sendMessage(ctx, "فشل في رفع الملف للواتساب.")
		return
	}

	docMsg := &waProto.DocumentMessage{
		URL:           proto.String(resp.URL),
		DirectPath:    proto.String(resp.DirectPath),
		MediaKey:      resp.MediaKey,
		Mimetype:      proto.String("application/pdf"),
		FileEncSHA256: resp.FileEncSHA256,
		FileSHA256:    resp.FileSHA256,
		FileLength:    proto.Uint64(uint64(len(pdfData))),
		Title:         proto.String(fmt.Sprintf("صور_%d.pdf", time.Now().Unix())),
		FileName:      proto.String(fmt.Sprintf("Images_%d.pdf", count)),
	}

	ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
		DocumentMessage: docMsg,
	})
}

func HandleRenamePDF(ctx *BotContext) {
	parts := strings.Split(ctx.Text, " ")
	if len(parts) < 3 {
		sendMessage(ctx, "يرجى كتابة الاسم الجديد بعد الأمر، مثال:\n.اسم pdf بحث التخرج")
		return
	}

	newName := strings.Join(parts[2:], " ")
	if !strings.HasSuffix(strings.ToLower(newName), ".pdf") {
		newName += ".pdf"
	}

	uMsg := UnwrapMessage(ctx.Event.Message)
	if uMsg == nil || uMsg.GetExtendedTextMessage() == nil || uMsg.GetExtendedTextMessage().GetContextInfo() == nil {
		sendMessage(ctx, "لازم ترد على ملف PDF عشان أقدر أغير اسمه!")
		return
	}

	quoted := uMsg.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage()
	if quoted == nil || quoted.GetDocumentMessage() == nil {
		sendMessage(ctx, "الرسالة اللي رديت عليها مو ملف! رد على ملف PDF وجرب ثاني.")
		return
	}

	docMsg := quoted.GetDocumentMessage()
	if !strings.HasSuffix(strings.ToLower(docMsg.GetFileName()), ".pdf") && docMsg.GetMimetype() != "application/pdf" {
		sendMessage(ctx, "الملف اللي رديت عليه مو بصيغة PDF!")
		return
	}

	sendMessage(ctx, "جاري تغيير اسم الملف وإعادة إرساله... ⏳")

	data, err := ctx.Client.Download(context.Background(), docMsg)
	if err != nil {
		sendMessage(ctx, "فشل في تحميل الملف الأساسي.")
		return
	}

	resp, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaDocument)
	if err != nil {
		sendMessage(ctx, "فشل في رفع الملف الجديد.")
		return
	}

	newDocMsg := &waProto.DocumentMessage{
		URL:           proto.String(resp.URL),
		DirectPath:    proto.String(resp.DirectPath),
		MediaKey:      resp.MediaKey,
		Mimetype:      proto.String("application/pdf"),
		FileEncSHA256: resp.FileEncSHA256,
		FileSHA256:    resp.FileSHA256,
		FileLength:    proto.Uint64(uint64(len(data))),
		Title:         proto.String(newName),
		FileName:      proto.String(newName),
	}

	ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
		DocumentMessage: newDocMsg,
	})
}
