package commands

import (
	"context"
	"strings"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"

	"whatsapp-bot/internal/store"
)

type ExchangeSession struct {
	ChatID   types.JID
	Messages []*events.Message
	Timer    *time.Timer
}

var (
	exchangeSessions = make(map[string]*ExchangeSession)
	exchangeMu       sync.Mutex
)

func BroadcastExchange(client *whatsmeow.Client) {
	go func() {
		favs := store.GetFavorites()
		for _, favStr := range favs {
			if store.GetStrike(favStr) >= 3 {
				continue // Skip users who ignored 3 times
			}
			
			favJid, err := types.ParseJID(favStr)
			if err != nil {
				continue
			}
			
			store.IncrementStrike(favStr)
			
			msg := "للتبادل اكتب .تبادل"
			client.SendMessage(context.Background(), favJid, &waProto.Message{
				Conversation: proto.String(msg),
			})
			time.Sleep(15 * time.Second) // Wait between sends to avoid spam/ban
		}
	}()
}

func HandleExchangeMessage(ctx *BotContext) bool {


	chatStr := ctx.ChatID.String()

	// Handle saving my exchange link
	if ctx.Text == ".رابطي" || ctx.Text == ".روابطي" {
		qMsg := ctx.Event.Message.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage()
		if qMsg != nil {
			data, err := proto.Marshal(qMsg)
			if err == nil {
				store.AddMyExchangeMsg(data)
				sendMessage(ctx, "تم إضافة هذه الرسالة لقائمة روابط التبادل الخاصة بك بنجاح!")
			}
		} else {
			sendMessage(ctx, "لازم تسوي ريبلاي على رسالتك وتكتب .رابطي أو .روابطي")
		}
		return true
	}
	
	if ctx.Text == ".حذف روابطي" {
	    store.ClearMyExchangeMsgs()
	    sendMessage(ctx, "تم حذف جميع روابط التبادل الخاصة بك!")
	    return true
	}

	// Handle toggling favorite
	if ctx.Text == ".مفضلة" {
		status := store.ToggleFavorite(chatStr)
		if status {
			sendMessage(ctx, "تم إضافة هذا الشخص لقائمة المفضلة لنظام التبادل.")
		} else {
			sendMessage(ctx, "تم إزالة هذا الشخص من قائمة المفضلة.")
		}
		return true
	}

	// Handle .نشر broadcast
	if ctx.Text == ".نشر" {
		sendMessage(ctx, "جاري إرسال رسالة التبادل لجميع الأرقام في المفضلة...")
		BroadcastExchange(ctx.Client)
		return true
	}

	// Handle setting exchange group
	if ctx.Text == "!تبادل" && ctx.Event.Info.IsGroup {
		store.SetExchangeGroup(chatStr)
		sendMessage(ctx, "تم تعيين هذا القروب كقروب التبادل الأساسي! سيتم تحويل الروابط هنا.")
		return true
	}

	// Handle starting exchange session
	if ctx.Text == ".تبادل" {
		store.ResetStrike(chatStr)
		exchangeMu.Lock()
		if _, exists := exchangeSessions[chatStr]; exists {
			exchangeMu.Unlock()
			sendMessage(ctx, "أنت بالفعل في جلسة تبادل! أرسل الروابط الخاصة بك.")
			return true
		}

		session := &ExchangeSession{
			ChatID:   ctx.ChatID,
			Messages: make([]*events.Message, 0),
		}
		
		session.Timer = time.AfterFunc(10*time.Second, func() {
			finishExchangeSession(ctx.Client, session)
		})
		exchangeSessions[chatStr] = session
		exchangeMu.Unlock()

		sendMessage(ctx, "تم بدء التبادل! أرسل روابطك الآن (معاك 10 ثواني بعد كل رابط)، أو اكتب .انتهيت إذا خلصت.")
		return true
	}

	// Check if user is in a session
	exchangeMu.Lock()
	session, exists := exchangeSessions[chatStr]
	exchangeMu.Unlock()

	if exists {
		if ctx.Text == ".انتهيت" {
			session.Timer.Stop()
			finishExchangeSession(ctx.Client, session)
			return true
		}

		// Reset timer
		session.Timer.Reset(10 * time.Second)

		// Check if message has a link
		if strings.Contains(ctx.Text, "chat.whatsapp.com/") {
			exchangeMu.Lock()
			session.Messages = append(session.Messages, ctx.Event)
			exchangeMu.Unlock()
		}
		return true // Consume message so it doesn't trigger gemini
	}

	return false
}

func finishExchangeSession(client *whatsmeow.Client, session *ExchangeSession) {
	exchangeMu.Lock()
	delete(exchangeSessions, session.ChatID.String())
	exchangeMu.Unlock()

	egStr := store.GetExchangeGroup()
	var groupJid types.JID
	if egStr != "" {
		groupJid, _ = types.ParseJID(egStr)
	}

	// Forward their valid link messages to the group
	forwardedCount := 0
	if groupJid.User != "" {
		for _, msgEvent := range session.Messages {
			msgCopy := proto.Clone(msgEvent.Message).(*waProto.Message)
			
			// Build forward context
			ctxInfo := &waProto.ContextInfo{
				IsForwarded: proto.Bool(true),
			}
			
			// Attach to the right message type
			if msgCopy.ExtendedTextMessage != nil {
				msgCopy.ExtendedTextMessage.ContextInfo = ctxInfo
			} else if msgCopy.Conversation != nil {
				msgCopy.ExtendedTextMessage = &waProto.ExtendedTextMessage{
					Text:        msgCopy.Conversation,
					ContextInfo: ctxInfo,
				}
				msgCopy.Conversation = nil
			}
			
			client.SendMessage(context.Background(), groupJid, msgCopy)
			forwardedCount++
			time.Sleep(1 * time.Second) // Sleep to avoid ban
		}
	}

	// Send ALL my saved exchange messages to them
	myMsgsData := store.GetMyExchangeMsgs()
	for _, myMsgData := range myMsgsData {
		var myMsg waProto.Message
		if err := proto.Unmarshal(myMsgData, &myMsg); err == nil {
			msgCopy := proto.Clone(&myMsg).(*waProto.Message)
			// Add forward flag
			ctxInfo := &waProto.ContextInfo{
				IsForwarded: proto.Bool(true),
			}
			if msgCopy.ExtendedTextMessage != nil {
				msgCopy.ExtendedTextMessage.ContextInfo = ctxInfo
			} else if msgCopy.ImageMessage != nil {
				msgCopy.ImageMessage.ContextInfo = ctxInfo
			} else if msgCopy.VideoMessage != nil {
				msgCopy.VideoMessage.ContextInfo = ctxInfo
			} else if msgCopy.DocumentMessage != nil {
			    msgCopy.DocumentMessage.ContextInfo = ctxInfo
			} else if msgCopy.Conversation != nil {
				msgCopy.ExtendedTextMessage = &waProto.ExtendedTextMessage{
					Text:        msgCopy.Conversation,
					ContextInfo: ctxInfo,
				}
				msgCopy.Conversation = nil
			}

			client.SendMessage(context.Background(), session.ChatID, msgCopy)
			time.Sleep(1 * time.Second)
		}
	}

	// Final reply
	client.SendMessage(context.Background(), session.ChatID, &waProto.Message{
		Conversation: proto.String("نشرت روابطك انشر روابطي من فضلك"),
	})
}
