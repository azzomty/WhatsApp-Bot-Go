package commands

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"whatsapp-bot/internal/stickers"
	"whatsapp-bot/internal/store"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

type StickerPackSession struct {
	Images   [][]byte
	Author   string
	Title    string
	MaxLimit int
	Mu       sync.Mutex
}

var (
	packSessions = make(map[string]*StickerPackSession)
	packMu       sync.RWMutex
)

func CreateStickerPackCommand(ctx *BotContext) {
	packMu.Lock()
	defer packMu.Unlock()

	sender := ctx.Event.Info.Sender.ToNonAD().String()
	
	rights := store.GetStickerAuthor(getLID(ctx, ctx.Sender))
	packName := "My Sticker Pack"
	authorName := "Antigravity Bot"
	if rights["pack"] != "" {
		packName = rights["pack"]
	}
	if rights["author"] != "" {
		authorName = rights["author"]
	}

	limit := 30
	parts := strings.Split(ctx.Text, " ")
	if len(parts) > 2 {
		if parsedLimit, err := strconv.Atoi(parts[2]); err == nil && parsedLimit > 0 {
			if parsedLimit > 100 {
				parsedLimit = 100
			}
			limit = parsedLimit
		}
	} else if len(parts) == 2 && parts[0] != ".عمل" && parts[0] != ".صنع" {
		// e.g. if the user sent .حزمة 25 without space? Wait, .عمل حزمة is two words.
	}

	packSessions[sender] = &StickerPackSession{
		Author:   authorName,
		Title:    packName,
		MaxLimit: limit,
	}

	sendMessage(ctx, fmt.Sprintf("تم بدء إنشاء حزمة ملصقات جديدة! أرسل الصور أو الملصقات الآن (كحد أقصى %d).\nولما تخلص أرسل `.انهاء الحزمة` أو `.إلغاء الحزمة`.", limit))
}

func FinishStickerPackCommand(ctx *BotContext) {
	packMu.Lock()
	sender := ctx.Event.Info.Sender.ToNonAD().String()
	session, ok := packSessions[sender]
	if !ok {
		packMu.Unlock()
		sendMessage(ctx, "ما عندك حزمة قيد الإنشاء! ابدأ حزمة جديدة بكتابة `.عمل حزمة`")
		return
	}
	delete(packSessions, sender)
	packMu.Unlock()

	session.Mu.Lock()
	images := session.Images
	session.Mu.Unlock()

	if len(images) == 0 {
		sendMessage(ctx, "الحزمة فاضية! ما أرسلت أي صور.")
		return
	}

	sendMessage(ctx, "جاري تجميع الحزمة وتجهيزها...")

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	w, _ := zipWriter.Create("title.txt")
	w.Write([]byte(session.Title))

	w, _ = zipWriter.Create("author.txt")
	w.Write([]byte(session.Author))

	for i, img := range images {
		w, _ = zipWriter.Create(fmt.Sprintf("%d.webp", i+1))
		w.Write(img)
	}
	zipWriter.Close()

	uploaded, err := ctx.Client.Upload(context.Background(), buf.Bytes(), whatsmeow.MediaDocument)
	if err != nil {
		sendMessage(ctx, "فشل رفع الحزمة!")
		return
	}

	fileName := session.Title + ".wastickers"
	msg := &waProto.Message{
		DocumentMessage: &waProto.DocumentMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String("application/zip"),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(buf.Len())),
			FileName:      proto.String(fileName),
		},
	}
	ctx.Client.SendMessage(context.Background(), ctx.Event.Info.Chat, msg)
}

func CancelStickerPackCommand(ctx *BotContext) {
	packMu.Lock()
	sender := ctx.Event.Info.Sender.ToNonAD().String()
	_, ok := packSessions[sender]
	if ok {
		delete(packSessions, sender)
		sendMessage(ctx, "تم إلغاء الحزمة!")
	} else {
		sendMessage(ctx, "ما عندك حزمة قيد الإنشاء أصلاً!")
	}
	packMu.Unlock()
}

func HandleStickerPackSession(ctx *BotContext) bool {
	if ctx.Text == ".انهاء الحزمة" || ctx.Text == ".إلغاء الحزمة" || ctx.Text == ".عمل حزمة" || ctx.Text == ".صنع حزمة" || ctx.Text == ".الغاء الحزمة" {
		return false
	}

	packMu.RLock()
	sender := ctx.Event.Info.Sender.ToNonAD().String()
	session, ok := packSessions[sender]
	packMu.RUnlock()

	if !ok {
		return false
	}

	var imgData []byte
	var err error
	if ctx.Event.Message.GetImageMessage() != nil {
		data, err := ctx.Client.Download(context.Background(), ctx.Event.Message.GetImageMessage())
		if err == nil {
			webpData, err := stickers.GenerateSticker(data, false, session.Title, session.Author)
			if err == nil {
				imgData = webpData
			}
		}
	} else if ctx.Event.Message.GetStickerMessage() != nil {
		data, err := ctx.Client.Download(context.Background(), ctx.Event.Message.GetStickerMessage())
		if err == nil {
			imgData = data
		}
	} else if ctx.Event.Message.GetVideoMessage() != nil {
		data, err := ctx.Client.Download(context.Background(), ctx.Event.Message.GetVideoMessage())
		if err == nil {
			webpData, err := stickers.GenerateSticker(data, true, session.Title, session.Author)
			if err == nil {
				imgData = webpData
			}
		}
	}

	if len(imgData) > 0 {
		session.Mu.Lock()
		limit := session.MaxLimit
		if limit == 0 {
			limit = 30
		}
		if len(session.Images) < limit {
			session.Images = append(session.Images, imgData)
			sendMessage(ctx, fmt.Sprintf("تم إضافة الملصق (%d/%d)", len(session.Images), limit))
		} else {
			sendMessage(ctx, fmt.Sprintf("وصلت للحد الأقصى (%d ملصق)! أرسل `.انهاء الحزمة` لإنشائها.", limit))
		}
		session.Mu.Unlock()
		return true
	}

	if err != nil {
		return false
	}

	// We're in a session, but they sent text. Just ignore it and don't process commands unless it starts with a dot?
	// Returning true means "handled", false means "let normal commands run".
	// If they send normal chat, we shouldn't block it.
	return false
}
