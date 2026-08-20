package commands

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"net/http"
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
			if parsedLimit > 1000 {
				parsedLimit = 1000
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




	// We need to split into multiple packs if > 30 stickers because WhatsApp has a hard limit of 30 per pack
	numPacks := (len(images) + 29) / 30

	for packIdx := 0; packIdx < numPacks; packIdx++ {
		startIdx := packIdx * 30
		endIdx := startIdx + 30
		if endIdx > len(images) {
			endIdx = len(images)
		}
		
		packImages := images[startIdx:endIdx]
		
		buf := new(bytes.Buffer)
		zipWriter := zip.NewWriter(buf)
		
		packTitle := session.Title
		if numPacks > 1 {
			packTitle = fmt.Sprintf("%s Part %d", session.Title, packIdx+1)
		}

		w, _ := zipWriter.Create("title.txt")
		w.Write([]byte(packTitle))

		w, _ = zipWriter.Create("author.txt")
		w.Write([]byte(session.Author))

		wTray, _ := zipWriter.Create("tray.png")
		if len(packImages) > 0 {
			trayResp, err := http.Post("http://127.0.0.1:4321/tray", "application/octet-stream", bytes.NewReader(packImages[0]))
			if err == nil && trayResp.StatusCode == 200 {
				trayImg, _ := ioutil.ReadAll(trayResp.Body)
				wTray.Write(trayImg)
			} else {
				imgTray := image.NewRGBA(image.Rect(0, 0, 96, 96))
				png.Encode(wTray, imgTray)
			}
		}

		for i, img := range packImages {
			w, _ = zipWriter.Create(fmt.Sprintf("%d.webp", i+1))
			w.Write(img)
		}
		zipWriter.Close()

		uploaded, err := ctx.Client.Upload(context.Background(), buf.Bytes(), whatsmeow.MediaDocument)
		if err != nil {
			sendMessage(ctx, fmt.Sprintf("فشل رفع الحزمة %d!", packIdx+1))
			continue
		}

		fileName := packTitle + ".wastickers"
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

	session.Mu.Lock()
	defer session.Mu.Unlock()

	var imgData []byte
	var err error
	
	uMsg := UnwrapMessage(ctx.Event.Message)
	if uMsg == nil {
		return false
	}

	if uMsg.GetImageMessage() != nil {
		data, err := ctx.Client.Download(context.Background(), uMsg.GetImageMessage())
		if err == nil {
			webpData, err := stickers.GenerateSticker(data, false, session.Title, session.Author)
			if err == nil {
				imgData = webpData
			}
		}
	} else if uMsg.GetStickerMessage() != nil {
		data, err := ctx.Client.Download(context.Background(), uMsg.GetStickerMessage())
		if err == nil {
			imgData = data
		}
	} else if uMsg.GetVideoMessage() != nil {
		data, err := ctx.Client.Download(context.Background(), uMsg.GetVideoMessage())
		if err == nil {
			webpData, err := stickers.GenerateSticker(data, true, session.Title, session.Author)
			if err == nil {
				imgData = webpData
			}
		}
	} else if uMsg.GetDocumentMessage() != nil {
		doc := uMsg.GetDocumentMessage()
		if strings.HasPrefix(doc.GetMimetype(), "image/") || strings.HasPrefix(doc.GetMimetype(), "video/") {
			data, err := ctx.Client.Download(context.Background(), doc)
			if err == nil {
				isVideo := strings.HasPrefix(doc.GetMimetype(), "video/")
				webpData, err := stickers.GenerateSticker(data, isVideo, session.Title, session.Author)
				if err == nil {
					imgData = webpData
				}
			}
		}
	}

	if len(imgData) > 0 {
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
		return true
	}

	if err != nil {
		sendMessage(ctx, "عذراً، فشل تحويل أحد الملفات (تأكد إنه مدعوم وغير تالف).")
		return true // Return true so it doesn't fall through to other commands
	}

	// We're in a session, but they sent text. Just ignore it and don't process commands unless it starts with a dot?
	// Returning true means "handled", false means "let normal commands run".
	// If they send normal chat, we shouldn't block it.
	return false
}
