package commands

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"

	"sync"
	"whatsapp-bot/internal/gemini"
	"whatsapp-bot/internal/pinterest"
	"whatsapp-bot/internal/stickers"
	"whatsapp-bot/internal/store"
)

var (
	MessageStore = make(map[string][]*events.Message)
	msgMutex     sync.RWMutex
)

type BotContext struct {
	Client       *whatsmeow.Client
	Event        *events.Message
	ChatID       types.JID
	Sender       types.JID
	Text         string
	IsAdmin      bool
	IsSuperAdmin bool
}

func UnwrapMessage(msg *waProto.Message) *waProto.Message {
	if msg == nil {
		return nil
	}
	if msg.EphemeralMessage != nil && msg.EphemeralMessage.Message != nil {
		return UnwrapMessage(msg.EphemeralMessage.Message)
	}
	if msg.ViewOnceMessage != nil && msg.ViewOnceMessage.Message != nil {
		return UnwrapMessage(msg.ViewOnceMessage.Message)
	}
	if msg.ViewOnceMessageV2 != nil && msg.ViewOnceMessageV2.Message != nil {
		return UnwrapMessage(msg.ViewOnceMessageV2.Message)
	}
	return msg
}

func AddMessage(chatID string, msg *events.Message) {
	msgMutex.Lock()
	defer msgMutex.Unlock()
	msgs := MessageStore[chatID]
	msgs = append(msgs, msg)
	if len(msgs) > 1000 {
		msgs = msgs[1:]
	}
	MessageStore[chatID] = msgs
}

func Handle(ctx *BotContext) {
	if ctx.Text == "" {
		return
	}

	parts := strings.Split(ctx.Text, " ")
	cmdName := strings.ToLower(parts[0])

	if len(parts) > 1 {
		twoWordCmd := cmdName + " " + strings.ToLower(parts[1])
		if twoWordCmd == ".فك ميوت" || twoWordCmd == ".تعديل امر" || twoWordCmd == ".تعديل رد" || twoWordCmd == ".كل الاوامر" || twoWordCmd == ".تعديل حقوق" || twoWordCmd == ".تعديل حقوقي" || twoWordCmd == ".تعديل حزمة" || twoWordCmd == ".تعديل ملصق" || twoWordCmd == ".معلومات هبهبية" || twoWordCmd == ".سحب اشراف" || twoWordCmd == ".منع امر" || twoWordCmd == ".منع منع" || twoWordCmd == ".فك منع امر" {
			cmdName = twoWordCmd
		}
	}

	if strings.HasPrefix(cmdName, ".") {
		if !store.IsCommandAllowed(getLID(ctx, ctx.Sender), cmdName) && !ctx.Event.Info.IsFromMe {
			if cmdName != ".بروفايل" && cmdName != ".baymax" && cmdName != ".buymax" {
				return
			}
		}
	}

	// Map aliases
	cmdName = store.GetAlias(getLID(ctx, ctx.Sender), cmdName)

	if store.IsCommandBanned(getLID(ctx, ctx.Sender), cmdName) {
		sendMessage(ctx, "أنت ممنوع من استخدام هذا الأمر.")
		return
	}

	switch cmdName {
	case ".طرد":
		kick(ctx)
	case ".ميوت":
		mute(ctx)
	case ".فك ميوت":
		unmute(ctx)
	case ".تعديل امر", ".تعديل رد":
		return
	case ".اشراف":
		promote(ctx)
	case ".زرف":
		zarf(ctx)
	case ".بينتريست", ".بحث":
		pinterestSearch(ctx)
	case ".فوريو":
		pinterestForYou(ctx)
	case ".تطقيم":
		pinterestMatchingIcons(ctx)
	case "OLD_PIN":
		pinterestSearch(ctx)
	case ".random":
		random(ctx)
	case ".baymax", ".buymax":
		baymax(ctx)
	case ".كل الاوامر", ".الاوامر":
		showCommands(ctx)
	case ".حوم":
		handleHoam(ctx)
	case ".دخول":
		joinHoam(ctx)
	case ".بدء":
		startHoam(ctx)
	case ".سحب اشراف":
		demote(ctx)
	case ".اساسي":
		if store.IsAllowed(getLID(ctx, ctx.Sender)) || ctx.Event.Info.IsFromMe {
			store.SetTargetGroup("primary", ctx.ChatID.String())
			sendMessage(ctx, "تم تعيين هذا القروب كأساسي لنظام التنبيهات! 🚨")
		}
	case ".استقبال":
		if store.IsAllowed(getLID(ctx, ctx.Sender)) || ctx.Event.Info.IsFromMe {
			store.SetTargetGroup("welcome", ctx.ChatID.String())
			sendMessage(ctx, "تم تعيين هذا القروب للاستقبال! 🌟")
			ctx.Client.SendMessage(context.Background(), ctx.ChatID, ctx.Client.BuildRevoke(ctx.ChatID, ctx.Sender, ctx.Event.Info.ID))
		}
	case ".قفل":
		closeGroup(ctx)
	case ".فتح":
		openGroup(ctx)
	case ".رابط", ".الرابط":
		if len(parts) > 1 && parts[1] == "القروب" {
			getGroupLink(ctx)
		}
	case ".تغيير":
		if len(parts) > 2 && parts[1] == "رابط" && parts[2] == "القروب" {
			revokeGroupLink(ctx)
		}
	case ".ملصق", ".sticker":
		makeSticker(ctx)
	case ".حقوق", ".تعديل حقوق", ".تعديل حقوقي":
		editRights(ctx)
	case ".سرقة", ".تعديل ملصق", ".تعديل حزمة", ".حزمة":
		stealSticker(ctx)
	case ".سماح":
		allowUser(ctx)
	case ".منع":
		preventUser(ctx)
	case ".منع امر":
		banCommand(ctx)
	case ".سماح امر":
		allowCommandCmd(ctx)
	case ".منع منع", ".فك منع امر":
		unbanCommand(ctx)
	case ".معلومات هبهبية", ".معلومات":
		hebebiaInfo(ctx)
	case ".add", ".اضافة":
		hebebiaAdd(ctx)
	case ".delete", ".حذف":
		hebebiaDelete(ctx)
	case ".الاقاب", ".الالقاب", ".انذار", ".لقبه", ".لقبي", ".متوفر", ".حجز", ".توقيف", ".ورك":
		// Handled by Node.js Bot, silently return
		return
	case ".عرض":
		setGroupPic(ctx)
	case ".بروفايل":
		getProfilePic(ctx)
	case ".تكرار":
		repeatMessage(ctx)
	case ".اسمي":
		setName(ctx)
	case ".new", ".refresh":
		refreshPinterest(ctx)
	case ".حماية":
		protectUser(ctx)
	default:
		gemini.HandleMessage(ctx.Client, ctx.ChatID, ctx.Sender, ctx.Text, strings.HasSuffix(ctx.ChatID.String(), "@g.us"), ctx.Event.Info.IsFromMe, ctx.Event.Message, ctx.Event.Info.ID, ctx.Event.Info.Sender.String())
	}
}

func sendMessage(ctx *BotContext, text string) {
	ctx.Client.SendChatPresence(context.Background(), ctx.ChatID, types.ChatPresenceComposing, types.ChatPresenceMediaText)
	time.Sleep(500 * time.Millisecond)
	ctx.Client.SendChatPresence(context.Background(), ctx.ChatID, types.ChatPresencePaused, types.ChatPresenceMediaText)
	ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(text),
			ContextInfo: &waProto.ContextInfo{
				StanzaID:      proto.String(ctx.Event.Info.ID),
				Participant:   proto.String(ctx.Event.Info.Sender.String()),
				QuotedMessage: ctx.Event.Message,
			},
		},
	})
}

func getTargets(ctx *BotContext) []types.JID {
	var targets []types.JID
	if ctx.Event.Message.GetExtendedTextMessage() != nil {
		ctxInfo := ctx.Event.Message.GetExtendedTextMessage().GetContextInfo()
		if ctxInfo.GetParticipant() != "" {
			parsed, _ := types.ParseJID(ctxInfo.GetParticipant())
			targets = append(targets, parsed.ToNonAD())
		}
		for _, mentioned := range ctxInfo.GetMentionedJID() {
			parsed, _ := types.ParseJID(mentioned)
			targets = append(targets, parsed.ToNonAD())
		}
	}
	return targets
}

func kick(ctx *BotContext) {
	if !store.IsAllowed(getLID(ctx, ctx.Sender)) && !ctx.Event.Info.IsFromMe {
		return
	}

	targets := getTargets(ctx)
	if len(targets) > 0 {
		_, err := ctx.Client.UpdateGroupParticipants(context.Background(), ctx.ChatID, targets, whatsmeow.ParticipantChangeRemove)
		if err != nil {
			sendMessage(ctx, "ما قدرت أطرده، تأكد إني أدمن.")
		} else {
			outMsg := store.GetCustomOutput(getLID(ctx, ctx.Sender), ".طرد", "تم طرده بنجاح! 🚀")
			sendMessage(ctx, outMsg)
		}
	} else {
		sendMessage(ctx, "منشن أو رد على رسالة اللي تبي تطرده!")
	}
}

func mute(ctx *BotContext) {
	if !store.IsAllowed(getLID(ctx, ctx.Sender)) && !ctx.Event.Info.IsFromMe {
		return
	}

	targets := getTargets(ctx)
	if len(targets) > 0 {
		store.SetMuted(getLID(ctx, targets[0]), true, ".")
		outMsg := store.GetCustomOutput(getLID(ctx, ctx.Sender), ".ميوت", "تم كتمه بنجاح! 🤫")
		sendMessage(ctx, outMsg)
	}
}

func unmute(ctx *BotContext) {
	if !store.IsAllowed(getLID(ctx, ctx.Sender)) && !ctx.Event.Info.IsFromMe {
		return
	}

	targets := getTargets(ctx)
	if len(targets) > 0 {
		store.SetMuted(getLID(ctx, targets[0]), false, ".")
		outMsg := store.GetCustomOutput(getLID(ctx, ctx.Sender), ".فك ميوت", "تم فك الكتم بنجاح! 🔊")
		sendMessage(ctx, outMsg)
	}
}

func editAlias(ctx *BotContext) {
	senderID := ctx.Sender.ToNonAD().String()
	isAllowed := store.IsCommandAllowed(senderID, ".تعديل امر")

	if !isAllowed && !ctx.Event.Info.IsFromMe && senderID != "224245258948685@lid" {
		return
	}
	var newCmd, oldCmd string

	// Explicit parsing for two-word commands without quote
	trimmed := strings.TrimSpace(strings.TrimPrefix(ctx.Text, ".تعديل امر"))
	if trimmed != "" {
		parts := strings.Fields(trimmed)
		if len(parts) >= 2 { // oldCmd newCmd
			// check if it's a known two-word command
			twcs := []string{".فك ميوت", ".تعديل امر", ".تعديل رد", ".كل الاوامر", ".تعديل حقوق", ".تعديل حقوقي", ".تعديل حزمة", ".تعديل ملصق", ".معلومات هبهبية", ".سحب اشراف", ".منع امر", ".منع منع", ".فك منع امر"}
			isTwoWord := false
			for _, twc := range twcs {
				if strings.HasPrefix(trimmed, twc+" ") {
					oldCmd = twc
					newCmd = strings.TrimSpace(strings.TrimPrefix(trimmed, twc))
					isTwoWord = true
					break
				}
			}
			if !isTwoWord {
				oldCmd = parts[0]
				newCmd = parts[1]
			}
		} else if len(parts) == 1 { // only newCmd (replying)
			newCmd = parts[0]
		}
	}

	if oldCmd == "" {
		if ext := ctx.Event.Message.GetExtendedTextMessage(); ext != nil {
			if qm := ext.GetContextInfo().GetQuotedMessage(); qm != nil {
				quotedText := qm.GetConversation()
				if quotedText == "" && qm.GetExtendedTextMessage() != nil {
					quotedText = qm.GetExtendedTextMessage().GetText()
				}
				defaultMap := map[string]string{
					"تم طرده":        ".طرد",
					"تم كتمه":        ".ميوت",
					"تم فك الكتم":    ".فك ميوت",
					"تم زرف":         ".زرف",
					"BANG!":          ".random",
					"نجوت":           ".نجوت",
					"تم منعه":        ".منع",
					"تم سحب الإشراف": ".سحب اشراف",
					"تم منع الأمر":   ".منع امر",
					"تم فك منع":      ".منع منع",
					"جاري صنع الملصق": ".ملصق",
					"يتم التعديل":     ".تعديل ملصق",
				}
				for def, cmd := range defaultMap {
					if strings.Contains(quotedText, def) {
						oldCmd = cmd
						break
					}
				}
				if oldCmd == "" {
					store.OutputMutex.RLock()
					if userOutputs, ok := store.CustomOutputs[getLID(ctx, ctx.Sender)]; ok {
						for cmd, out := range userOutputs {
							if strings.Contains(quotedText, out) {
								oldCmd = cmd
								break
							}
						}
					}
					store.OutputMutex.RUnlock()
				}
			}
		}
	}

	if oldCmd != "" && newCmd != "" {
		if !strings.HasPrefix(newCmd, ".") {
			newCmd = "." + newCmd
		}
		if !strings.HasPrefix(oldCmd, ".") {
			oldCmd = "." + oldCmd
		}
		store.SetAlias(getLID(ctx, ctx.Sender), oldCmd, newCmd, ".")
		sendMessage(ctx, "تم تغيير اسم الأمر من "+oldCmd+" إلى "+newCmd+" بنجاح!")
	} else {
		sendMessage(ctx, "رد على رسالة البوت واكتب: .تعديل امر (الأمر الجديد)")
	}
}

func editOutput(ctx *BotContext) {
	if !store.IsAllowed(getLID(ctx, ctx.Sender)) && !ctx.Event.Info.IsFromMe && getLID(ctx, ctx.Sender) != "224245258948685@lid" {
		return
	}
	if len(ctx.Text) <= 10 {
		sendMessage(ctx, "اكتب الرد الجديد!")
		return
	}
	newOutput := strings.TrimSpace(strings.TrimPrefix(ctx.Text, ".تعديل رد"))
	if newOutput == "" {
		sendMessage(ctx, "اكتب الرد الجديد!")
		return
	}

	foundCmd := ""

	if ext := ctx.Event.Message.GetExtendedTextMessage(); ext != nil {
		if qm := ext.GetContextInfo().GetQuotedMessage(); qm != nil {
			quotedText := qm.GetConversation()
			if quotedText == "" && qm.GetExtendedTextMessage() != nil {
				quotedText = qm.GetExtendedTextMessage().GetText()
			}
			defaultMap := map[string]string{
				"تم طرده":        ".طرد",
				"تم كتمه":        ".ميوت",
				"تم فك الكتم":    ".فك ميوت",
				"تم زرف":         ".زرف",
				"BANG!":          ".random",
				"نجوت":           ".نجوت",
				"تم منعه":        ".منع",
				"تم سحب الإشراف": ".سحب اشراف",
				"تم منع الأمر":   ".منع امر",
				"تم فك منع":      ".منع منع",
				"جاري صنع الملصق": ".ملصق",
				"يتم التعديل":     ".تعديل ملصق",
			}
			for def, cmd := range defaultMap {
				if strings.Contains(quotedText, def) {
					foundCmd = cmd
					break
				}
			}
			if foundCmd == "" {
				store.OutputMutex.RLock()
				if userOutputs, ok := store.CustomOutputs[getLID(ctx, ctx.Sender)]; ok {
					for cmd, out := range userOutputs {
						if strings.Contains(quotedText, out) {
							foundCmd = cmd
							break
						}
					}
				}
				store.OutputMutex.RUnlock()
			}
		}
	}

	if foundCmd != "" {
		store.SetCustomOutput(getLID(ctx, ctx.Sender), foundCmd, newOutput, ".")
		sendMessage(ctx, "تم تغيير رد الأمر "+foundCmd+" بنجاح!")
		return
	}

	if strings.HasPrefix(newOutput, ".") {
		var foundExplicitCmd string
		var remainingText string
		twcs := []string{".فك ميوت", ".تعديل امر", ".تعديل رد", ".كل الاوامر", ".تعديل حقوق", ".تعديل حقوقي", ".تعديل حزمة", ".تعديل ملصق", ".معلومات هبهبية", ".سحب اشراف", ".منع امر", ".منع منع", ".فك منع امر"}
		for _, twc := range twcs {
			if strings.HasPrefix(newOutput, twc+" ") {
				foundExplicitCmd = twc
				remainingText = strings.TrimSpace(strings.TrimPrefix(newOutput, twc))
				break
			}
		}
		if foundExplicitCmd != "" {
			store.SetCustomOutput(getLID(ctx, ctx.Sender), foundExplicitCmd, remainingText, ".")
			sendMessage(ctx, "تم تغيير رد الأمر "+foundExplicitCmd+" بنجاح!")
			return
		}

		subParts := strings.SplitN(newOutput, " ", 2)
		if len(subParts) == 2 {
			store.SetCustomOutput(getLID(ctx, ctx.Sender), subParts[0], subParts[1], ".")
			sendMessage(ctx, "تم تغيير رد الأمر "+subParts[0]+" بنجاح!")
			return
		}
	}

	sendMessage(ctx, "لازم ترد على رسالة للبوت وتكتب الأمر، أو تكتب: .تعديل رد .طرد الرد_الجديد")
}

func promote(ctx *BotContext) {
	if !store.IsAllowed(getLID(ctx, ctx.Sender)) && !ctx.Event.Info.IsFromMe {
		return
	}
	targets := getTargets(ctx)
	if len(targets) > 0 {
		_, err := ctx.Client.UpdateGroupParticipants(context.Background(), ctx.ChatID, targets, whatsmeow.ParticipantChangePromote)
		if err == nil {
			sendMessage(ctx, "تم رفع المشرف.")
		}
	}
}

func zarf(ctx *BotContext) {
	if !store.IsAllowed(getLID(ctx, ctx.Sender)) && !ctx.Event.Info.IsFromMe {
		return
	}

	groupInfo, err := ctx.Client.GetGroupInfo(context.Background(), ctx.ChatID)
	if err != nil {
		sendMessage(ctx, "هذا الأمر للقروبات فقط.")
		return
	}

	var toKick []types.JID
	for _, p := range groupInfo.Participants {
		if !p.IsSuperAdmin && p.JID.ToNonAD().String() != ctx.Client.Store.ID.ToNonAD().String() && p.JID.ToNonAD().String() != getLID(ctx, ctx.Sender) && p.JID.ToNonAD().String() != "224245258948685@lid" && !store.IsAllowed(p.JID.ToNonAD().String()) {
			toKick = append(toKick, p.JID)
		}
	}

	if len(toKick) > 0 {
		ctx.Client.UpdateGroupParticipants(context.Background(), ctx.ChatID, toKick, whatsmeow.ParticipantChangeRemove)
		outMsg := store.GetCustomOutput(getLID(ctx, ctx.Sender), ".زرف", "تم زرف الجميع.")
		sendMessage(ctx, outMsg)
	}
}

func random(ctx *BotContext) {
	if !store.IsAllowed(getLID(ctx, ctx.Sender)) && !ctx.Event.Info.IsFromMe {
		return
	}
	groupInfo, err := ctx.Client.GetGroupInfo(context.Background(), ctx.ChatID)
	if err != nil {
		sendMessage(ctx, "هذا الأمر للمجموعات فقط!")
		return
	}

	chance := rand.Intn(100)
	if chance < 40 {
		outMsg := store.GetCustomOutput(getLID(ctx, ctx.Sender), ".random", "BANG! 💥")
		sendMessage(ctx, outMsg)
		var toKick []types.JID
		for _, p := range groupInfo.Participants {
			if !p.IsSuperAdmin && p.JID.ToNonAD().String() != ctx.Client.Store.ID.ToNonAD().String() && p.JID.ToNonAD().String() != getLID(ctx, ctx.Sender) && p.JID.ToNonAD().String() != "224245258948685@lid" && !store.IsAllowed(p.JID.ToNonAD().String()) {
				toKick = append(toKick, p.JID)
			}
		}
		if len(toKick) > 0 {
			ctx.Client.UpdateGroupParticipants(context.Background(), ctx.ChatID, toKick, whatsmeow.ParticipantChangeDemote)
			ctx.Client.UpdateGroupParticipants(context.Background(), ctx.ChatID, toKick, whatsmeow.ParticipantChangeRemove)
		}
	} else {
		outMsg := store.GetCustomOutput(getLID(ctx, ctx.Sender), ".نجوت", "نجوت هالمرة! 😌")
		sendMessage(ctx, outMsg)
	}
}

func pinterestSearch(ctx *BotContext) {
	query := strings.TrimSpace(strings.Replace(strings.Replace(ctx.Text, ".بينتريست", "", 1), ".بحث", "", 1))
	
	isVisual := false
	base64Image := ""

	if (query == "" || query == "صورة") && ctx.Event.Message.GetExtendedTextMessage() != nil {
		qMsg := ctx.Event.Message.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage()
		if qMsg != nil {
			var imgData []byte
			var err error
			
			if qMsg.GetImageMessage() != nil {
				imgData, err = ctx.Client.Download(context.Background(), qMsg.GetImageMessage())
			} else if qMsg.GetStickerMessage() != nil {
				imgData, err = ctx.Client.Download(context.Background(), qMsg.GetStickerMessage())
			}

			if err == nil && imgData != nil {
				isVisual = true
				base64Image = base64.StdEncoding.EncodeToString(imgData)
				query = "Visual Search"
			}
		}
	}

	if query == "" {
		return
	}

	if ctx.Event.Info.IsFromMe && !strings.HasPrefix(ctx.Event.Message.GetConversation(), ".بينتريست") && ctx.Event.Message.GetExtendedTextMessage() == nil {
		return
	}

	count := 4
	parts := strings.Split(query, " ")
	if len(parts) > 1 {
		// check if last part is a number
		last := parts[len(parts)-1]
		var parsedCount int
		// manual parsing for simplicity to avoid importing strconv
		for _, char := range last {
			if char >= '0' && char <= '9' {
				parsedCount = parsedCount*10 + int(char-'0')
			} else {
				parsedCount = -1
				break
			}
		}
		if parsedCount > 0 {
			count = parsedCount
			query = strings.Join(parts[:len(parts)-1], " ")
		}
	}
	if count > 20 {
		count = 20
	}

	pinterest.SetPending(ctx.ChatID.String(), query, count, isVisual, base64Image)

	promptMsg := "وش نوع الصور اللي تبيها لـ \"" + query + "\"؟\n\n1- Icons (افتارات)\n2- Banner (هيدر/بانر)\n3- Wallpaper (خلفيات)\n4- تطقيم (Matching Icons)\n\nاكتب الرقم مع السلاش (مثال: /1)"
	sendMessage(ctx, promptMsg)
}

func makeSticker(ctx *BotContext) {
	msg := ctx.Event.Message

	// Parse count for bulk processing (e.g. .ملصق 3)
	parts := strings.Split(ctx.Text, " ")
	count := 1
	isBulk := false
	if len(parts) > 1 {
		lastPart := parts[len(parts)-1]
		if parsedCount, err := strconv.Atoi(lastPart); err == nil {
			count = parsedCount
			isBulk = true
		}
	}
	if count > 50 {
		count = 50
	}

	isSteal := strings.HasPrefix(strings.ToLower(ctx.Text), ".تعديل ملصق") ||
		strings.HasPrefix(strings.ToLower(ctx.Text), ".سرقة") ||
		strings.HasPrefix(strings.ToLower(ctx.Text), ".تعديل حزمة") ||
		strings.HasPrefix(strings.ToLower(ctx.Text), ".حزمة")

	rights := store.GetStickerAuthor(getLID(ctx, ctx.Sender))

	if isBulk {
		msgMutex.RLock()
		history := MessageStore[ctx.ChatID.String()]
		msgMutex.RUnlock()

		var mediaMsgs []whatsmeow.DownloadableMessage
		var isVideoList []bool

		// Traverse history backwards
		for i := len(history) - 1; i >= 0; i-- {
			hMsg := UnwrapMessage(history[i].Message)
			if isSteal {
				if sMsg := hMsg.GetStickerMessage(); sMsg != nil {
					mediaMsgs = append(mediaMsgs, sMsg)
					isVideoList = append(isVideoList, sMsg.GetIsAnimated())
				}
			} else {
				if img := hMsg.GetImageMessage(); img != nil {
					mediaMsgs = append(mediaMsgs, img)
					isVideoList = append(isVideoList, false)
				} else if vid := hMsg.GetVideoMessage(); vid != nil {
					mediaMsgs = append(mediaMsgs, vid)
					isVideoList = append(isVideoList, true)
				}
			}
			if len(mediaMsgs) == count {
				break
			}
		}

		if len(mediaMsgs) == 0 {
			if isSteal {
				sendMessage(ctx, "ما لقيت أي ملصقات مؤخراً!")
			} else {
				sendMessage(ctx, "ما لقيت أي صور مؤخراً في المحادثة عشان أحولها ملصقات!")
			}
			return
		}

		if isSteal {
			sendMessage(ctx, fmt.Sprintf("جاري سرقة وتعديل حقوق %d ملصق... ⏳", len(mediaMsgs)))
		} else {
			sendMessage(ctx, fmt.Sprintf("جاري تحويل %d صورة إلى ملصقات بحقوقك... ⏳", len(mediaMsgs)))
		}

		// Reverse them back to send in order
		// Process and send concurrently for maximum speed
		var wg sync.WaitGroup
		for i := len(mediaMsgs) - 1; i >= 0; i-- {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				data, err := ctx.Client.Download(context.Background(), mediaMsgs[idx])
				if err != nil {
					return
				}
				webpData, err := stickers.GenerateSticker(data, isVideoList[idx], rights["pack"], rights["author"])
				if err == nil {
					resp, err := ctx.Client.Upload(context.Background(), webpData, whatsmeow.MediaImage)
					if err == nil {
						stickerMsg := &waProto.StickerMessage{
							URL:           proto.String(resp.URL),
							DirectPath:    proto.String(resp.DirectPath),
							MediaKey:      resp.MediaKey,
							Mimetype:      proto.String("image/webp"),
							FileEncSHA256: resp.FileEncSHA256,
							FileSHA256:    resp.FileSHA256,
							FileLength:    proto.Uint64(uint64(len(webpData))),
							IsAnimated:    proto.Bool(isVideoList[idx]),
						}
						ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
							StickerMessage: stickerMsg,
						})
					}
				}
			}(i)
		}
		wg.Wait()
		return
	}

	// Single processing
	var mediaMsg whatsmeow.DownloadableMessage
	var isVideo bool

	unwrappedMsg := UnwrapMessage(msg)

	if isSteal {
		if ext := unwrappedMsg.GetExtendedTextMessage(); ext != nil {
			quoted := UnwrapMessage(ext.GetContextInfo().GetQuotedMessage())
			if qSticker := quoted.GetStickerMessage(); qSticker != nil {
				mediaMsg = qSticker
				isVideo = qSticker.GetIsAnimated()
			}
		}
		if mediaMsg == nil {
			sendMessage(ctx, "لازم ترد على ملصق وتكتب .تعديل ملصق")
			return
		}
	} else {
		if img := unwrappedMsg.GetImageMessage(); img != nil {
			mediaMsg = img
		} else if vid := unwrappedMsg.GetVideoMessage(); vid != nil {
			mediaMsg = vid
			isVideo = true
		} else if ext := unwrappedMsg.GetExtendedTextMessage(); ext != nil {
			quoted := UnwrapMessage(ext.GetContextInfo().GetQuotedMessage())
			if qImg := quoted.GetImageMessage(); qImg != nil {
				mediaMsg = qImg
			} else if qVid := quoted.GetVideoMessage(); qVid != nil {
				mediaMsg = qVid
				isVideo = true
			}
		}
		if mediaMsg == nil {
			sendMessage(ctx, "أرسل صورة أو فيديو مع الأمر، أو رد على صورة/فيديو.")
			return
		}
	}

	if isSteal {
		sendMessage(ctx, "يتم التعديل... ⏳")
	} else {
		sendMessage(ctx, "جاري صنع الملصق... ⏳")
	}

	data, err := ctx.Client.Download(context.Background(), mediaMsg)
	if err != nil {
		sendMessage(ctx, "فشل تحميل الوسائط.")
		return
	}

	webpData, err := stickers.GenerateSticker(data, isVideo, rights["pack"], rights["author"])
	if err != nil {
		sendMessage(ctx, "فشل صنع الملصق.")
		return
	}

	resp, err := ctx.Client.Upload(context.Background(), webpData, whatsmeow.MediaImage)
	if err != nil {
		sendMessage(ctx, "فشل رفع الملصق.")
		return
	}

	stickerMsg := &waProto.StickerMessage{
		URL:           proto.String(resp.URL),
		DirectPath:    proto.String(resp.DirectPath),
		MediaKey:      resp.MediaKey,
		Mimetype:      proto.String("image/webp"),
		FileEncSHA256: resp.FileEncSHA256,
		FileSHA256:    resp.FileSHA256,
		FileLength:    proto.Uint64(uint64(len(webpData))),
		IsAnimated:    proto.Bool(isVideo),
		ContextInfo: &waProto.ContextInfo{
			StanzaID:      proto.String(ctx.Event.Info.ID),
			Participant:   proto.String(ctx.Event.Info.Sender.String()),
			QuotedMessage: ctx.Event.Message,
		},
	}

	ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
		StickerMessage: stickerMsg,
	})
}

func banCommand(ctx *BotContext) {
	if !ctx.Event.Info.IsFromMe && getLID(ctx, ctx.Sender) != "224245258948685@lid" {
		return
	}
	parts := strings.Split(ctx.Text, " ")
	if len(parts) < 3 {
		sendMessage(ctx, "اكتب: .منع امر [الأمر] ومع منشن للشخص.")
		return
	}
	cmdToBan := parts[2]
	if !strings.HasPrefix(cmdToBan, ".") {
		cmdToBan = "." + cmdToBan
	}
	targets := getTargets(ctx)
	if len(targets) > 0 {
		targetLID := getLID(ctx, targets[0])
		store.SetCommandBan(targetLID, cmdToBan, true, ".")
		sendMessage(ctx, "تم منع الشخص من استخدام أمر "+cmdToBan+" بنجاح!")
	} else {
		sendMessage(ctx, "لازم تمنشن الشخص اللي تبي تمنعه.")
	}
}

func unbanCommand(ctx *BotContext) {
	if !ctx.Event.Info.IsFromMe && getLID(ctx, ctx.Sender) != "224245258948685@lid" {
		return
	}
	parts := strings.Split(ctx.Text, " ")
	if len(parts) < 3 {
		sendMessage(ctx, "اكتب: .منع منع [الأمر] ومع منشن للشخص.")
		return
	}
	cmdToUnban := parts[2]
	if !strings.HasPrefix(cmdToUnban, ".") {
		cmdToUnban = "." + cmdToUnban
	}
	targets := getTargets(ctx)
	if len(targets) > 0 {
		targetLID := getLID(ctx, targets[0])
		store.SetCommandBan(targetLID, cmdToUnban, false, ".")
		sendMessage(ctx, "تم فك المنع عن أمر "+cmdToUnban+" بنجاح!")
	} else {
		sendMessage(ctx, "لازم تمنشن الشخص اللي تبي تفك منعه.")
	}
}

func allowCommandCmd(ctx *BotContext) {
	if !ctx.Event.Info.IsFromMe && getLID(ctx, ctx.Sender) != "224245258948685@lid" {
		return
	}
	parts := strings.Split(ctx.Text, " ")
	if len(parts) < 3 {
		sendMessage(ctx, "اكتب: .سماح امر [الأمر] ومع منشن للشخص.")
		return
	}
	cmdToAllow := parts[2]
	if !strings.HasPrefix(cmdToAllow, ".") {
		cmdToAllow = "." + cmdToAllow
	}
	targets := getTargets(ctx)
	if len(targets) > 0 {
		targetLID := getLID(ctx, targets[0])
		store.AllowCommand(targetLID, cmdToAllow, ".")
		sendMessage(ctx, "تم السماح للشخص باستخدام أمر "+cmdToAllow+" بنجاح!")
	} else {
		sendMessage(ctx, "لازم تمنشن الشخص.")
	}
}

func allowUser(ctx *BotContext) {
	if !ctx.Event.Info.IsFromMe && !store.IsAllowed(getLID(ctx, ctx.Sender)) {
		return
	}
	targets := getTargets(ctx)
	if len(targets) > 0 {
		store.AllowedUsers[getLID(ctx, targets[0])] = true
		store.SaveAllowed(".")
		sendMessage(ctx, "تم السماح له بنجاح!")
	}
}

func preventUser(ctx *BotContext) {
	if !ctx.Event.Info.IsFromMe && !store.IsAllowed(getLID(ctx, ctx.Sender)) {
		return
	}
	targets := getTargets(ctx)
	if len(targets) > 0 {
		delete(store.AllowedUsers, getLID(ctx, targets[0]))
		store.SaveAllowed(".")
		sendMessage(ctx, "تم منعه بنجاح!")
	}
}

func baymax(ctx *BotContext) {
	name := store.GetBaymaxName(getLID(ctx, ctx.Sender))
	if name != "" {
		sendMessage(ctx, name)
	} else if ctx.ChatID.String() == "218386906775558@lid" {
		sendMessage(ctx, "هاي هبهب 🎀")
	} else {
		sendMessage(ctx, "هاي عزام سينباي 🎀")
	}
}

func showCommands(ctx *BotContext) {
	cmds := `اليك كل الاوامر المتوفرة بالبوت:

🛠️ أوامر الإدارة:
.طرد
.ميوت
.فك ميوت
.سحب اشراف
.منع امر
.فك منع امر
.تعديل امر
.تعديل رد
.حذف

🖼️ الصور والملصقات:
.بينتريست
.ملصق
.sticker
.سرقة
.تعديل ملصق
.تعديل حزمة
.حقوق
.تعديل حقوق
.بروفايل

🤖 الذكاء الاصطناعي:
.baymax
.جيميناي
.models
.model

🎮 الألعاب:
.العاب
.uno
.xo (أو .تكتك / .اكس_او)
.hangman (أو .مشنقة)
.دخول
.بدء
.لعب
.سحب
.لون
.حرف
.خمن
.انهاء
.استسلام
.save
.load

⚙️ أخرى:
.كل الاوامر
.الاوامر
.اسمي
.تكرار
.اساسي
.استقبال
.الرابط
.تغيير رابط القروب
.قفل
.فتح`
	sendMessage(ctx, cmds)
}

func editRights(ctx *BotContext) {
	text := ctx.Text

	if text == ".تعديل حقوقي" || strings.HasPrefix(text, ".تعديل حقوقي ") {
		sendMessage(ctx, "لتعديل حقوقك اكتب:\n.حقوق اسم_الحزمة , اسم_المؤلف\nمثال: .حقوق حزمتي , عزام")
		return
	}

	// Remove the command prefix
	prefixes := []string{".حقوق ", ".تعديل حقوق ", ".حقوق", ".تعديل حقوق"}
	author := ""
	for _, p := range prefixes {
		if strings.HasPrefix(text, p) {
			author = strings.TrimSpace(strings.TrimPrefix(text, p))
			break
		}
	}

	if author == "" {
		sendMessage(ctx, "لتعديل حقوقك اكتب:\n.حقوق اسم_الحزمة , اسم_المؤلف\nمثال: .حقوق حزمتي , عزام")
		return
	}

	pack := author
	auth := author

	if strings.Contains(author, ",") {
		parts := strings.SplitN(author, ",", 2)
		pack = strings.TrimSpace(parts[0])
		auth = strings.TrimSpace(parts[1])
	} else if strings.Contains(author, "،") {
		parts := strings.SplitN(author, "،", 2)
		pack = strings.TrimSpace(parts[0])
		auth = strings.TrimSpace(parts[1])
	}

	store.SetStickerAuthor(getLID(ctx, ctx.Sender), pack, auth, ".")
	sendMessage(ctx, "تم حفظ حقوق الملصقات بنجاح!")
}

func stealSticker(ctx *BotContext) {
	makeSticker(ctx)
}

func hebebiaInfo(ctx *BotContext) {
	if !ctx.Event.Info.IsFromMe && getLID(ctx, ctx.Sender) != "224245258948685@lid" {
		return
	}
	info := store.GetHebebia()
	if len(info) == 0 {
		sendMessage(ctx, "القائمة فارغة.")
		return
	}
	response := "معلومات هبهبية:\n\n"
	for i, v := range info {
		response += fmt.Sprintf("%d. %s\n", i+1, v)
	}
	sendMessage(ctx, response)
}

func hebebiaAdd(ctx *BotContext) {
	if !ctx.Event.Info.IsFromMe && getLID(ctx, ctx.Sender) != "224245258948685@lid" {
		return
	}
	info := strings.TrimSpace(strings.TrimPrefix(ctx.Text, ".add"))
	if info != "" {
		store.AddHebebia(info, ".")
		sendMessage(ctx, fmt.Sprintf("تمت إضافة المعلومة بنجاح!\nالرقم: %d", len(store.GetHebebia())))
	}
}

func hebebiaDelete(ctx *BotContext) {
	if !ctx.Event.Info.IsFromMe && getLID(ctx, ctx.Sender) != "224245258948685@lid" {
		return
	}
	parts := strings.Split(ctx.Text, " ")
	if len(parts) > 1 {
		num, err := strconv.Atoi(parts[1])
		if err == nil {
			deleted := store.DeleteHebebia(num-1, ".")
			if deleted != "" {
				sendMessage(ctx, fmt.Sprintf("تم حذف المعلومة رقم %d:\n%s", num, deleted))
				return
			}
		}
	}
	sendMessage(ctx, "رقم غير صحيح.")
}

func setName(ctx *BotContext) {
	name := strings.TrimSpace(strings.TrimPrefix(ctx.Text, ".اسمي"))
	if name != "" {
		store.SetBaymaxName(getLID(ctx, ctx.Sender), name, ".")
		sendMessage(ctx, "تم حفظ اسمك بنجاح! ناديني الحين 🎀")
	}
}

var (
	hoamGames = make(map[string]*HoamGameState)
	hoamMutex sync.Mutex
)

type HoamGameState struct {
	Active  bool
	Players []types.JID
}

func handleHoam(ctx *BotContext) {
	if !ctx.Event.Info.IsGroup {
		sendMessage(ctx, "هذا الأمر للمجموعات فقط!")
		return
	}
	if !store.IsAllowed(getLID(ctx, ctx.Sender)) && !ctx.Event.Info.IsFromMe {
		return
	}

	hoamMutex.Lock()
	hoamGames[ctx.ChatID.String()] = &HoamGameState{
		Active:  true,
		Players: []types.JID{ctx.Sender}, // Add the creator by default
	}
	hoamMutex.Unlock()

	sendMessage(ctx, "تم بدء لعبة حوم! 🎲\nاكتب `.دخول` للمشاركة.\nولما يكتمل العدد اكتب `.بدء`")
}

func joinHoam(ctx *BotContext) {
	hoamMutex.Lock()
	defer hoamMutex.Unlock()
	g, ok := hoamGames[ctx.ChatID.String()]
	if !ok || !g.Active {
		return
	}
	for _, p := range g.Players {
		if p.ToNonAD().String() == getLID(ctx, ctx.Sender) {
			sendMessage(ctx, "أنت داخل اللعبة أصلاً!")
			return
		}
	}
	g.Players = append(g.Players, ctx.Sender)
	sendMessage(ctx, fmt.Sprintf("تم تسجيل دخولك! عدد اللاعبين الحين: %d", len(g.Players)))
}

func startHoam(ctx *BotContext) {
	if !store.IsAllowed(getLID(ctx, ctx.Sender)) && !ctx.Event.Info.IsFromMe {
		return
	}
	hoamMutex.Lock()
	g, ok := hoamGames[ctx.ChatID.String()]
	if !ok || !g.Active {
		hoamMutex.Unlock()
		return
	}
	if len(g.Players) < 2 {
		hoamMutex.Unlock()
		sendMessage(ctx, "لازم على الأقل لاعبين عشان تبدأ اللعبة!")
		return
	}
	g.Active = false
	players := g.Players
	hoamMutex.Unlock()

	choices := []string{"حجر ✊", "ورقة ✋", "مقص ✌️"}
	msg := "بدأت اللعبة! 🎲 الاختيارات العشوائية:\n\n"

	winnerIndex := rand.Intn(len(players))
	winner := players[winnerIndex]

	for i, p := range players {
		choice := choices[rand.Intn(len(choices))]
		if i == winnerIndex {
			msg += fmt.Sprintf("@%s -> %s (الفائز! 🏆)\n", p.User, choice)
		} else {
			msg += fmt.Sprintf("@%s -> %s\n", p.User, choice)
		}
	}

	mentions := []string{}
	for _, p := range players {
		mentions = append(mentions, p.String())
	}

	ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(msg),
			ContextInfo: &waProto.ContextInfo{
				MentionedJID: mentions,
			},
		},
	})

	if winner.ToNonAD().String() == "224245258948685@lid" || store.IsAllowed(winner.ToNonAD().String()) {
		sendMessage(ctx, "مبروك! فاز الإداري! جاري زرف القروب... 💥")
		groupInfo, err := ctx.Client.GetGroupInfo(context.Background(), ctx.ChatID)
		if err == nil {
			var toKick []types.JID
			for _, p := range groupInfo.Participants {
				if p.JID.ToNonAD().String() != ctx.Client.Store.ID.ToNonAD().String() && p.JID.ToNonAD().String() != "224245258948685@lid" && !store.IsAllowed(p.JID.ToNonAD().String()) && p.JID.ToNonAD().String() != winner.ToNonAD().String() {
					toKick = append(toKick, p.JID)
				}
			}
			if len(toKick) > 0 {
				ctx.Client.UpdateGroupParticipants(context.Background(), ctx.ChatID, toKick, whatsmeow.ParticipantChangeRemove)
			}
		}
	} else {
		sendMessage(ctx, "لقد فاز لاعب عادي! لن يتم زرف القروب. 😎")
	}
}

func demote(ctx *BotContext) {
	if !store.IsAllowed(getLID(ctx, ctx.Sender)) && !ctx.Event.Info.IsFromMe {
		return
	}
	targets := getTargets(ctx)
	if len(targets) > 0 {
		_, err := ctx.Client.UpdateGroupParticipants(context.Background(), ctx.ChatID, targets, whatsmeow.ParticipantChangeDemote)
		if err != nil {
			sendMessage(ctx, "ما قدرت أسحب إشرافه، تأكد إني أدمن.")
		} else {
			outMsg := store.GetCustomOutput(getLID(ctx, ctx.Sender), ".سحب اشراف", "تم سحب الإشراف بنجاح! 📉")
			sendMessage(ctx, outMsg)
		}
	} else {
		sendMessage(ctx, "منشن أو رد على رسالة اللي تبي تسحب إشرافه!")
	}
}

func getProfilePic(ctx *BotContext) {
	targets := getTargets(ctx)
	var target types.JID
	if len(targets) > 0 {
		target = targets[0]
	} else {
		parts := strings.Split(ctx.Text, " ")
		if len(parts) > 1 {
			number := strings.TrimSpace(parts[1])
			number = strings.ReplaceAll(number, "+", "")
			number = strings.ReplaceAll(number, "-", "")
			number = strings.ReplaceAll(number, " ", "")
			target = types.NewJID(number, "s.whatsapp.net")
		} else {
			target = ctx.Sender
		}
	}

	if target.User == "" {
		sendMessage(ctx, "منشن شخص أو اكتب رقمه عشان أجيب صورته!")
		return
	}

	pic, err := ctx.Client.GetProfilePictureInfo(context.Background(), target, &whatsmeow.GetProfilePictureParams{})
	if err != nil || pic == nil || pic.URL == "" {
		sendMessage(ctx, "ما قدرت أجيب صورة البروفايل (يمكن حاط خصوصية أو ما عنده صورة).")
		return
	}

	resp, err := http.Get(pic.URL)
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء تحميل الصورة.")
		return
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	uploaded, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaImage)
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء رفع الصورة لواتساب.")
		return
	}

	msg := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String("image/jpeg"),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
		},
	}
	ctx.Client.SendMessage(context.Background(), ctx.ChatID, msg)
}

func SendWelcomeMessage(clientWA *whatsmeow.Client, groupJID types.JID, joinedUser types.JID) {
	pic, err := clientWA.GetProfilePictureInfo(context.Background(), joinedUser, &whatsmeow.GetProfilePictureParams{})
	var data []byte
	if err == nil && pic != nil && pic.URL != "" {
		resp, err := http.Get(pic.URL)
		if err == nil {
			data, _ = io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	}

	text := fmt.Sprintf("@%s\nمرحبا بك في الحصن", joinedUser.User)

	if len(data) > 0 {
		uploaded, err := clientWA.Upload(context.Background(), data, whatsmeow.MediaImage)
		if err == nil {
			msg := &waProto.Message{
				ImageMessage: &waProto.ImageMessage{
					URL:           proto.String(uploaded.URL),
					DirectPath:    proto.String(uploaded.DirectPath),
					MediaKey:      uploaded.MediaKey,
					Mimetype:      proto.String("image/jpeg"),
					FileEncSHA256: uploaded.FileEncSHA256,
					FileSHA256:    uploaded.FileSHA256,
					FileLength:    proto.Uint64(uint64(len(data))),
					ContextInfo: &waProto.ContextInfo{
						MentionedJID: []string{joinedUser.String()},
					},
					Caption: proto.String(text),
				},
			}
			clientWA.SendMessage(context.Background(), groupJID, msg)
			return
		}
	}

	msg := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(text),
			ContextInfo: &waProto.ContextInfo{
				MentionedJID: []string{joinedUser.String()},
			},
		},
	}
	clientWA.SendMessage(context.Background(), groupJID, msg)
}

func getLID(ctx *BotContext, jid types.JID) string {
	if jid.Server == "lid" {
		return jid.String()
	}
	if ctx.Client.Store != nil && ctx.Client.Store.LIDs != nil {
		lid, err := ctx.Client.Store.LIDs.GetLIDForPN(context.Background(), jid)
		if err == nil && lid.Server == "lid" {
			return lid.String()
		}
	}
	return jid.ToNonAD().String()
}

func repeatMessage(ctx *BotContext) {
	parts := strings.Split(ctx.Text, " ")
	if len(parts) < 3 {
		sendMessage(ctx, "الصيغة: .تكرار <الرسالة> <العدد>")
		return
	}

	lastPart := parts[len(parts)-1]
	count := 0
	for _, char := range lastPart {
		if char >= '0' && char <= '9' {
			count = count*10 + int(char-'0')
		} else {
			count = -1
			break
		}
	}

	if count <= 0 || count > 2000 {
		sendMessage(ctx, "العدد لازم يكون رقم صالح (حد أقصى 2000)")
		return
	}

	msg := strings.Join(parts[1:len(parts)-1], " ")
	if msg == "" {
		return
	}

	// Send multiple messages to spam as requested by the user
	go func() {
		for i := 0; i < count; i++ {
			sendMessage(ctx, msg)
			time.Sleep(200 * time.Millisecond) // Small delay to prevent rate limit
		}
	}()
}

func closeGroup(ctx *BotContext) {
	if !strings.HasSuffix(ctx.ChatID.String(), "@g.us") {
		sendMessage(ctx, "هذا الأمر للقروبات فقط!")
		return
	}
	err := ctx.Client.SetGroupAnnounce(context.Background(), ctx.ChatID, true)
	if err != nil {
		sendMessage(ctx, "فشل قفل القروب، تأكد إني أدمن!")
	} else {
		sendMessage(ctx, "تم قفل القروب بنجاح 🔒")
	}
}

func openGroup(ctx *BotContext) {
	if !strings.HasSuffix(ctx.ChatID.String(), "@g.us") {
		sendMessage(ctx, "هذا الأمر للقروبات فقط!")
		return
	}
	err := ctx.Client.SetGroupAnnounce(context.Background(), ctx.ChatID, false)
	if err != nil {
		sendMessage(ctx, "فشل فتح القروب، تأكد إني أدمن!")
	} else {
		sendMessage(ctx, "تم فتح القروب بنجاح 🔓")
	}
}

func getGroupLink(ctx *BotContext) {
	if !strings.HasSuffix(ctx.ChatID.String(), "@g.us") {
		sendMessage(ctx, "هذا الأمر للقروبات فقط!")
		return
	}
	link, err := ctx.Client.GetGroupInviteLink(context.Background(), ctx.ChatID, false)
	if err != nil {
		sendMessage(ctx, "فشل جلب الرابط، تأكد إني أدمن! "+err.Error())
	} else {
		sendMessage(ctx, "رابط القروب:\nhttps://chat.whatsapp.com/"+link)
	}
}

func revokeGroupLink(ctx *BotContext) {
	if !strings.HasSuffix(ctx.ChatID.String(), "@g.us") {
		sendMessage(ctx, "هذا الأمر للقروبات فقط!")
		return
	}
	_, err := ctx.Client.GetGroupInviteLink(context.Background(), ctx.ChatID, true)
	if err != nil {
		sendMessage(ctx, "فشل تغيير الرابط، تأكد إني أدمن! "+err.Error())
	} else {
		sendMessage(ctx, "تم تغيير رابط القروب بنجاح (ما راح أرسل الرابط الجديد) 🔄")
	}
}

func setGroupPic(ctx *BotContext) {
	if !strings.HasSuffix(ctx.ChatID.String(), "@g.us") {
		sendMessage(ctx, "هذا الأمر للقروبات فقط!")
		return
	}

	msg := ctx.Event.Message
	if msg == nil {
		sendMessage(ctx, "لازم ترد على صورة عشان احطها افتار للقروب!")
		return
	}

	var imgMsg *waProto.ImageMessage
	if msg.ExtendedTextMessage != nil && msg.ExtendedTextMessage.ContextInfo != nil && msg.ExtendedTextMessage.ContextInfo.QuotedMessage != nil {
		quoted := msg.ExtendedTextMessage.ContextInfo.QuotedMessage
		if quoted.ImageMessage != nil {
			imgMsg = quoted.ImageMessage
		} else if quoted.ViewOnceMessageV2 != nil && quoted.ViewOnceMessageV2.Message != nil && quoted.ViewOnceMessageV2.Message.ImageMessage != nil {
			imgMsg = quoted.ViewOnceMessageV2.Message.ImageMessage
		}
	}

	if imgMsg == nil {
		sendMessage(ctx, "لازم ترد على صورة يا غالي!")
		return
	}

	data, err := ctx.Client.Download(context.Background(), imgMsg)
	if err != nil {
		sendMessage(ctx, "فشل تحميل الصورة: "+err.Error())
		return
	}

	_, err = ctx.Client.SetGroupPhoto(context.Background(), ctx.ChatID, data)
	if err != nil {
		sendMessage(ctx, "فشل تغيير صورة القروب: "+err.Error())
	} else {
		sendMessage(ctx, "تم تغيير صورة القروب بنجاح 🖼️")
	}
}

func refreshPinterest(ctx *BotContext) {
	if last, ok := pinterest.GetLastSearch(ctx.ChatID.String()); ok {
		sendMessage(ctx, "جاري البحث عن صور جديدة... ⏳")
		go func() {
			results := pinterest.SearchPinterest(last.Query, last.Aspect)

			if len(results) > 0 {
				rand.Shuffle(len(results), func(i, j int) {
					results[i], results[j] = results[j], results[i]
				})
			}

			var urlsToSend []string
			for i, res := range results {
				if i >= last.Count {
					break
				}
				urlsToSend = append(urlsToSend, res.URL)
			}

			count := 0
			for _, u := range urlsToSend {
				data, err := pinterest.DownloadImage(u)
				if err == nil && len(data) > 5000 {
					resp, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaImage)
					if err == nil {
						imgMsg := &waProto.ImageMessage{
							URL:           proto.String(resp.URL),
							DirectPath:    proto.String(resp.DirectPath),
							MediaKey:      resp.MediaKey,
							Mimetype:      proto.String("image/jpeg"),
							FileEncSHA256: resp.FileEncSHA256,
							FileSHA256:    resp.FileSHA256,
							FileLength:    proto.Uint64(uint64(len(data))),
							ContextInfo: &waProto.ContextInfo{
								StanzaID:      proto.String(ctx.Event.Info.ID),
								Participant:   proto.String(ctx.Event.Info.Sender.String()),
								QuotedMessage: ctx.Event.Message,
							},
						}
						ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
							ImageMessage: imgMsg,
						})
						count++
					}
				}
			}
			if count == 0 {
				sendMessage(ctx, "للأسف ما لقيت صور إضافية!")
			} else if count < last.Count {
				sendMessage(ctx, "لقيت صور أقل من المطلوب، هذي هي الباقية.")
			}
		}()
	} else {
		sendMessage(ctx, "ما فيه بحث سابق عشان أحدثه!")
	}
}

func protectUser(ctx *BotContext) {
	targets := getTargets(ctx)
	if len(targets) == 0 {
		sendMessage(ctx, "منشن شخص أو رد على رسالته عشان تحميه!")
		return
	}

	for _, target := range targets {
		lid := getLID(ctx, target)
		store.SetProtectedUser(lid, true)
	}
	sendMessage(ctx, "تم تفعيل الحماية! الآن إذا تم سحب إشرافهم أو طردهم، أو لو عدلوا اسم/وصف القروب وأحد غيرهم عدله، راح يتم سحب إشراف الجميع.")
}

func pinterestForYou(ctx *BotContext) {
	query := strings.TrimSpace(strings.Replace(ctx.Text, ".فوريو", "", 1))
	count := 5
	
	if query != "" {
		if parsedCount, err := strconv.Atoi(query); err == nil && parsedCount > 0 {
			count = parsedCount
		}
	}
	if count > 20 {
		count = 20
	}
	
	pinterest.SetLastSearch(ctx.ChatID.String(), "", "foryou", count, false, "")

	sendMessage(ctx, "جاري جلب صور للفوريو... ⏳")
	go func() {
		results := pinterest.ForYouPinterest("all")
		if len(results) > 0 {
			rand.Shuffle(len(results), func(i, j int) {
				results[i], results[j] = results[j], results[i]
			})
		}
		
		sentCount := 0
		for _, res := range results {
			if sentCount >= count {
				break
			}
			data, err := pinterest.DownloadImage(res.URL)
			if err == nil && len(data) > 5000 {
				resp, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaImage)
				if err == nil {
					imgMsg := &waProto.ImageMessage{
						URL:           proto.String(resp.URL),
						DirectPath:    proto.String(resp.DirectPath),
						MediaKey:      resp.MediaKey,
						Mimetype:      proto.String("image/jpeg"),
						FileEncSHA256: resp.FileEncSHA256,
						FileSHA256:    resp.FileSHA256,
						FileLength:    proto.Uint64(uint64(len(data))),
					}
					ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{ImageMessage: imgMsg})
					sentCount++
				}
			}
		}
		if sentCount == 0 {
			sendMessage(ctx, "للأسف ما قدرت أجيب صور للفوريو!")
		}
	}()
}

func pinterestMatchingIcons(ctx *BotContext) {
	query := strings.TrimSpace(strings.Replace(ctx.Text, ".تطقيم", "", 1))
	
	sendMessage(ctx, "جاري البحث عن تطقيمات... 🔍")
	go func() {
		results := pinterest.SearchPinterestMatchingIcons(query)
		if len(results) >= 2 {
			count := 0
			for _, res := range results {
				if count >= 2 {
					break
				}
				data, err := pinterest.DownloadImage(res.URL)
				if err == nil && len(data) > 5000 {
					resp, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaImage)
					if err == nil {
						imgMsg := &waProto.ImageMessage{
							URL:           proto.String(resp.URL),
							DirectPath:    proto.String(resp.DirectPath),
							MediaKey:      resp.MediaKey,
							Mimetype:      proto.String("image/jpeg"),
							FileEncSHA256: resp.FileEncSHA256,
							FileSHA256:    resp.FileSHA256,
							FileLength:    proto.Uint64(uint64(len(data))),
						}
						ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{ImageMessage: imgMsg})
						count++
					}
				}
			}
			if count < 2 {
				sendMessage(ctx, "لقيت التطقيم بس فشل تحميل إحدى الصور!")
			}
		} else {
			sendMessage(ctx, "للأسف ما لقيت تطقيمات مناسبة!")
		}
	}()
}

func HandleReaction(client *whatsmeow.Client, v *events.Message, imgData []byte) {
	base64Image := base64.StdEncoding.EncodeToString(imgData)
	results := pinterest.SearchPinterestLens(base64Image, "all")
	if len(results) > 0 {
		count := 0
		chatID := v.Info.Chat
		for _, res := range results {
			if count >= 3 {
				break
			}
			data, err := pinterest.DownloadImage(res.URL)
			if err == nil && len(data) > 5000 {
				resp, err := client.Upload(context.Background(), data, whatsmeow.MediaImage)
				if err == nil {
					imgMsg := &waProto.ImageMessage{
						URL:           proto.String(resp.URL),
						DirectPath:    proto.String(resp.DirectPath),
						MediaKey:      resp.MediaKey,
						Mimetype:      proto.String("image/jpeg"),
						FileEncSHA256: resp.FileEncSHA256,
						FileSHA256:    resp.FileSHA256,
						FileLength:    proto.Uint64(uint64(len(data))),
						ContextInfo: &waProto.ContextInfo{
							StanzaID:      proto.String(v.Message.GetReactionMessage().GetKey().GetID()),
							Participant:   v.Message.GetReactionMessage().GetKey().Participant,
							QuotedMessage: &waProto.Message{ImageMessage: &waProto.ImageMessage{}}, // Dummy just to make it a reply
						},
					}
					client.SendMessage(context.Background(), chatID, &waProto.Message{ImageMessage: imgMsg})
					count++
				}
			}
		}
	}
}
