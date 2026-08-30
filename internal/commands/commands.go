package commands

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"

	"whatsapp-bot/internal/gemini"
	"whatsapp-bot/internal/pinterest"
	"whatsapp-bot/internal/stickers"
	"whatsapp-bot/internal/store"
	"whatsapp-bot/internal/youtube"
)

var (
	MessageStore = make(map[string][]*events.Message)
	msgMutex     sync.RWMutex
)

var (
	startTime      time.Time
	messageCount   int
	commandsCount  int
	IsBotEnabled   = true
	AutoJoinGroups = false
	seenGroupLinks = make(map[string]bool)
	seenLinksMutex sync.Mutex
)

type BotContext struct {
	Client       *whatsmeow.Client
	Event        *events.Message
	ChatID       types.JID
	Sender       types.JID
	Text         string
	MentionedJid []string
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

var lastValidCommand = make(map[string]string)

func Handle(ctx *BotContext) {
	
	if ctx.Text == ".bot off" {
		IsBotEnabled = false
		sendMessage(ctx, "تم إيقاف البوت بالكامل!")
		return
	}
	if ctx.Text == ".bot on" {
		IsBotEnabled = true
		sendMessage(ctx, "تم تفعيل البوت بالكامل!")
		return
	}


	if !IsBotEnabled {
		return
	}

	// Hibi Check for Saudi Number
	if ctx.Event.Message != nil && ctx.Event.Message.ExtendedTextMessage != nil && ctx.Event.Message.ExtendedTextMessage.ContextInfo != nil {
		qMsg := ctx.Event.Message.ExtendedTextMessage.ContextInfo.QuotedMessage
		if qMsg != nil {
			participant := ctx.Event.Message.ExtendedTextMessage.ContextInfo.GetParticipant()
			myJid := ctx.Client.Store.ID.ToNonAD().String()
			
			if participant == myJid {
				qText := ""
				if qMsg.ExtendedTextMessage != nil {
					qText = qMsg.ExtendedTextMessage.GetText()
				} else if qMsg.Conversation != nil {
					qText = qMsg.GetConversation()
				}
				qText = strings.TrimSpace(qText)
				if qText == "وش لقبك" || qText == "وش لقبي" {
					sendMessage(ctx, "New character unlock hibi💫🔓")
					return
				}
			}
		}
	}



	if handleInteractiveReply(ctx) {
		return
	}
	if HandleStickerPackSession(ctx) {
		return
	}
	if HandleExchangeMessage(ctx) {
		return
	}
	if ctx.Text == "" {
		return
	}
	
	if !store.IsAllowed(getLID(ctx, ctx.Sender)) && !ctx.Event.Info.IsFromMe {
		return
	}

	parts := strings.Split(ctx.Text, " ")
	cmdName := strings.ToLower(parts[0])

	if len(parts) > 1 {
		twoWordCmd := cmdName + " " + strings.ToLower(parts[1])
		if twoWordCmd == ".مواعيد صلاة" || twoWordCmd == ".مشرفين القروبات" || twoWordCmd == ".تنظيف القروبات" || twoWordCmd == ".مواعيد الصلاة" || twoWordCmd == ".فك ميوت" || twoWordCmd == ".تعديل امر" || twoWordCmd == ".تعديل رد" || twoWordCmd == ".روليت" || twoWordCmd == ".اعرف الرقم" || twoWordCmd == ".كل الاوامر" || twoWordCmd == ".تعديل حقوق" || twoWordCmd == ".تعديل حقوقي" || twoWordCmd == ".تعديل حزمة" || twoWordCmd == ".تعديل ملصق" || twoWordCmd == ".معلومات هبهبية" || twoWordCmd == ".سحب اشراف" || twoWordCmd == ".منع امر" || twoWordCmd == ".منع منع" || twoWordCmd == ".فك منع امر" || twoWordCmd == ".فك كومنت" || twoWordCmd == ".عمل حزمة" || twoWordCmd == ".صنع حزمة" || twoWordCmd == ".انهاء الحزمة" || twoWordCmd == ".إلغاء الحزمة" || twoWordCmd == ".الغاء الحزمة" || twoWordCmd == ".ذكرني اتصال" || twoWordCmd == ".ذكرني رسالة" || twoWordCmd == ".قائمة الاحاديث" || twoWordCmd == ".قائمة الأحاديث" || twoWordCmd == ".اسم pdf" || twoWordCmd == ".بحث قسم" || twoWordCmd == ".كل الاقسام" || twoWordCmd == ".كل الأقسام" || twoWordCmd == ".قائمة الكراتين" {
			cmdName = twoWordCmd
		}
	}

	if cmdName == ".new" {
		if lastCmd, ok := lastValidCommand[ctx.Sender.User]; ok {
			ctx.Text = lastCmd
			parts = strings.Split(ctx.Text, " ")
			cmdName = strings.ToLower(parts[0])
			if len(parts) > 1 {
				twoWordCmd := cmdName + " " + strings.ToLower(parts[1])
				if twoWordCmd == ".مواعيد صلاة" || twoWordCmd == ".مشرفين القروبات" || twoWordCmd == ".تنظيف القروبات" || twoWordCmd == ".مواعيد الصلاة" || twoWordCmd == ".فك ميوت" || twoWordCmd == ".تعديل امر" || twoWordCmd == ".تعديل رد" || twoWordCmd == ".روليت" || twoWordCmd == ".اعرف الرقم" || twoWordCmd == ".كل الاوامر" || twoWordCmd == ".تعديل حقوق" || twoWordCmd == ".تعديل حقوقي" || twoWordCmd == ".تعديل حزمة" || twoWordCmd == ".تعديل ملصق" || twoWordCmd == ".معلومات هبهبية" || twoWordCmd == ".سحب اشراف" || twoWordCmd == ".منع امر" || twoWordCmd == ".منع منع" || twoWordCmd == ".فك منع امر" || twoWordCmd == ".فك كومنت" || twoWordCmd == ".عمل حزمة" || twoWordCmd == ".صنع حزمة" || twoWordCmd == ".انهاء الحزمة" || twoWordCmd == ".إلغاء الحزمة" || twoWordCmd == ".الغاء الحزمة" || twoWordCmd == ".ذكرني اتصال" || twoWordCmd == ".ذكرني رسالة" || twoWordCmd == ".قائمة الاحاديث" || twoWordCmd == ".قائمة الأحاديث" || twoWordCmd == ".اسم pdf" || twoWordCmd == ".بحث قسم" || twoWordCmd == ".كل الاقسام" || twoWordCmd == ".كل الأقسام" || twoWordCmd == ".قائمة الكراتين" {
					cmdName = twoWordCmd
				}
			}
		} else {
			return
		}
	} else if strings.HasPrefix(cmdName, ".") {
		lastValidCommand[ctx.Sender.User] = ctx.Text
	}

	if store.IsCommandDisabled(cmdName) {
		if cmdName != ".فك كومنت" { // ensure .فك كومنت itself can't be locked out forever if disabled by mistake
			return
		}
	}

	if strings.HasPrefix(cmdName, ".") {
		if false {
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
	case ".تعديل امر":
		editAlias(ctx)
	case ".تعديل رد":
		editOutput(ctx)
	case ".اشراف":
		promote(ctx)
	case ".زرف":
		zarf(ctx)
	case ".طقس":
		GetWeather(ctx)
	case ".mp3":
		convertToMp3(ctx)
	// case ".يوتيوب":
	// 	interactiveYoutube(ctx)
	// case ".تحميل":
	// 	universalDownload(ctx)
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
	case ".تحليل":
		if ctx.Event.Message.GetExtendedTextMessage() != nil && ctx.Event.Message.GetExtendedTextMessage().GetContextInfo() != nil {
			quoted := ctx.Event.Message.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage()
			if quoted != nil {
				if doc := quoted.GetDocumentMessage(); doc != nil {
					data, err := ctx.Client.Download(context.Background(), doc)
					if err == nil {
						_ = os.WriteFile("har.json", data, 0644)
						sendMessage(ctx, "تم حفظ الملف بنجاح! المبرمج سيقوم بتحليله الآن.")
					} else {
						sendMessage(ctx, "فشل تحميل الملف.")
					}
				} else if img := quoted.GetImageMessage(); img != nil {
					data, err := ctx.Client.Download(context.Background(), img)
					if err == nil {
						_ = os.WriteFile("screenshot.jpg", data, 0644)
						sendMessage(ctx, "تم حفظ الصورة بنجاح! المبرمج يراها الآن.")
					} else {
						sendMessage(ctx, "فشل تحميل الصورة.")
					}
				}
			}
		}
	case ".baymax", ".buymax":
		baymax(ctx)
	case ".كل الاوامر", ".الاوامر":
		showCommands(ctx)
	case ".حوم":
		handleHoam(ctx)
	case ".دخلني قروبات":
		if ctx.ChatID.String() == "120363402487910101@g.us" { // The specific exchange group
			return
		}
		sendMessage(ctx, "https://chat.whatsapp.com/EDxI9u5H0P84tGk06Rj9C2\nhttps://chat.whatsapp.com/G98mP1O19Kq7z4aW4rT7yE\nhttps://chat.whatsapp.com/Drd16tUuA56JqU7iI5XJ9q\nhttps://chat.whatsapp.com/C5uG5k3X3j00wZ9q0E9q0E")
	case ".دخول":
		joinHoam(ctx)
	case ".بدء":
		startHoam(ctx)
	case ".سحب اشراف":
		demote(ctx)
	case ".اساسي":
		if store.IsAllowed(getLID(ctx, ctx.Sender)) || ctx.Event.Info.IsFromMe {
			store.SetTargetGroup("primary", ctx.ChatID.String())
			sendMessage(ctx, "تم تعيين هذا القروب كأساسي لنظام التنبيهات!")
		}
	case ".تفعيل":
		if store.IsAllowed(getLID(ctx, ctx.Sender)) || ctx.Event.Info.IsFromMe {
			enableCommand(ctx, parts)
		}
	case ".عمل حزمة", ".صنع حزمة":
		CreateStickerPackCommand(ctx)
	case ".انهاء الحزمة":
		FinishStickerPackCommand(ctx)
	case ".إلغاء الحزمة", ".الغاء الحزمة":
		CancelStickerPackCommand(ctx)
	case ".ذكرني اتصال", ".ذكرني رسالة", ".تذكير":
		HandleReminder(ctx, cmdName)
	case ".حديث":
		HandleHadith(ctx)
	case ".قائمة الاحاديث", ".قائمة الأحاديث":
		HandleHadithMenu(ctx)
	case ".بحث قسم":
		HandleCategorySearch(ctx)
	case ".كل الاقسام", ".كل الأقسام":
		HandleAllCategories(ctx)
	case ".ايميل", ".إيميل":
		HandleTempMail(ctx)
	case ".pdf":
		HandlePDF(ctx)
	case ".اسم pdf":
		HandleRenamePDF(ctx)
	case ".فلم", ".فيلم", ".مسلسل", ".انمي", ".أنمي", ".مانجا", ".مانهاوا", ".كرتون", ".انمي_مدبلج":
		HandleMediaCommand(ctx, cmdName)
	case ".الجزء", ".جزء":
		if activeSource[ctx.Sender.User] == "stardima" {
			parts := strings.Split(ctx.Text, " ")
			if len(parts) > 1 {
				idx, _ := strconv.Atoi(parts[1])
				HandleStardimaPart(ctx, idx)
			} else {
				sendMessage(ctx, "يرجى تحديد الرقم، مثال: .جزء 1")
			}
		} else {
			HandlePartCommand(ctx)
		}
	case ".قائمة الكراتين", ".قائمة":
		HandleCartoonList(ctx)
	case ".ستارديما":
		HandleStardimaCommand(ctx)
	case ".رقم":
		HandleNumberSelect(ctx)
				case ".حلقة":
		if activeSource[ctx.Sender.User] == "stardima" {
			parts := strings.Split(ctx.Text, " ")
			if len(parts) > 1 {
				idx, _ := strconv.Atoi(parts[1])
				HandleStardimaEpisode(ctx, idx)
			} else {
				sendMessage(ctx, "يرجى تحديد الرقم، مثال: .حلقة 1")
			}
		} else {
			HandleEpisodeCommand(ctx)
		}

	case ".قفل":
		closeGroup(ctx)
	case ".فتح":
		openGroup(ctx)
	case ".الرابط", ".رابط":
		if len(parts) > 1 && parts[1] == "القروب" {
			getGroupLink(ctx)
		}
	case ".نص", ".لخص", ".دبلج":
		targetLang := "ar"
		if len(parts) > 1 {
			langMap := map[string]string{
				"عربي":    "ar",
				"انجليزي": "en",
				"فرنسي":   "fr",
				"اسباني":  "es",
				"ياباني":  "ja",
				"كوري":    "ko",
				"روسي":    "ru",
				"صيني":    "zh",
				"تركي":    "tr",
				"الماني":  "de",
				"هندي":    "hi",
			}
			if val, ok := langMap[parts[1]]; ok {
				targetLang = val
			}
		}
		transcribeAudio(ctx, targetLang)
	case ".ترجم", ".عربي":
		translateMessage(ctx, "ar")
	case ".انجليزي":
		translateMessage(ctx, "en")
	case ".ياباني":
		translateMessage(ctx, "ja")
	case ".كوري":
		translateMessage(ctx, "ko")
	case ".صيني":
		translateMessage(ctx, "zh")
	case ".روسي":
		translateMessage(ctx, "ru")
	case ".فرنسي":
		translateMessage(ctx, "fr")
	case ".اسباني":
		translateMessage(ctx, "es")
	case ".تركي":
		translateMessage(ctx, "tr")
	case ".الماني":
		translateMessage(ctx, "de")
	case ".هندي":
		translateMessage(ctx, "hi")
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
		handleDeleteMsgOrHebebia(ctx)
	case ".الاقاب", ".الالقاب", ".انذار", ".لقبه", ".لقبي", ".متوفر", ".حجز", ".توقيف", ".ورك":
		// Handled by Node.js Bot, silently return
		return
	case ".عرض":
		setGroupPic(ctx)
	case ".بروفايل":
		getProfilePic(ctx)
	case ".تكرار":
		repeatMessage(ctx)
	// case ".اغنية":
	// 	processYoutubeMedia(ctx, true)
	// case ".فيديو":
	// 	multiVideoSearch(ctx)
	case ".react":
		reactMessage(ctx)
	case ".اسمي":
		setName(ctx)
	case ".new", ".refresh":
		handleNewCommand(ctx)
	case ".مواعيد صلاة", ".مواعيد الصلاة":
		address := ""
		if len(parts) > 2 {
			address = strings.Join(parts[2:], " ")
		}
		HandlePrayerTimes(ctx, address)
	case ".توقيت":
		address := ""
		if len(parts) > 1 {
			address = strings.Join(parts[1:], " ")
		}
		HandleCurrentTime(ctx, address)
	case ".حماية":
		protectUser(ctx)
	case ".كومنت":
		disableCommand(ctx, parts)
	case ".فك كومنت":
		enableCommand(ctx, parts)
	case ".رياكت":
		setAutoReact(ctx, parts)
	case ".استقبال":
		setWelcomeFeature(ctx)
	case ".كتم", ".حذف الاغنية":
		HandleRemoveMusic(ctx)
	case ".اكيناتور":
		HandleAkinator(ctx)
	default:
		if HandleAkinatorAnswer(ctx) {
			return
		}
		gemini.HandleMessage(ctx.Client, ctx.ChatID, ctx.Sender, ctx.Text, strings.HasSuffix(ctx.ChatID.String(), "@g.us"), ctx.Event.Info.IsFromMe, ctx.Event.Message, ctx.Event.Info.ID, ctx.Event.Info.Sender.String())
	}
}

func sendMessage(ctx *BotContext, text string) {
	fmt.Println("SENT:", text)
	text = strings.ReplaceAll(text, "...", "")
	text = strings.ReplaceAll(text, "..", "")
	text = strings.ReplaceAll(text, ",,,", "")
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
			outMsg := store.GetCustomOutput(getLID(ctx, ctx.Sender), ".طرد", "تم طرده بنجاح!")
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
		outMsg := store.GetCustomOutput(getLID(ctx, ctx.Sender), ".ميوت", "تم كتمه بنجاح!")
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
		outMsg := store.GetCustomOutput(getLID(ctx, ctx.Sender), ".فك ميوت", "تم فك الكتم بنجاح!")
		sendMessage(ctx, outMsg)
	}
}

func editAlias(ctx *BotContext) {
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
		outMsg := store.GetCustomOutput(getLID(ctx, ctx.Sender), ".random", "BANG!")
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
		outMsg := store.GetCustomOutput(getLID(ctx, ctx.Sender), ".نجوت", "نجوت هالمرة!")
		sendMessage(ctx, outMsg)
	}
}

func sendAndCacheImage(ctx *BotContext, chatID types.JID, imgMsg *waProto.ImageMessage) (*whatsmeow.SendResponse, error) {
	msg := &waProto.Message{ImageMessage: imgMsg}
	sendResp, err := ctx.Client.SendMessage(context.Background(), chatID, msg)
	if err == nil {
		dummy := &events.Message{
			Info: types.MessageInfo{
				ID:        sendResp.ID,
				MessageSource: types.MessageSource{Chat: chatID, IsFromMe: true},
				
				Timestamp: sendResp.Timestamp,
			},
			Message: msg,
		}
		AddMessage(chatID.String(), dummy)
	}
	return &sendResp, err
}

func pinterestSearch(ctx *BotContext) {
	query := strings.TrimSpace(strings.Replace(strings.Replace(ctx.Text, ".بينتريست", "", 1), ".بحث", "", 1))
	
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
	if count > 100 {
		count = 100
	}

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

	pinterest.SetPending(ctx.ChatID.String(), query, count, isVisual, base64Image, "")

	promptMsg := "وش نوع الصور اللي تبيها لـ \"" + query + "\"؟\n\n1- Icons\n2- Banner\n3- Wallpaper\n4- Matching Icons\n5- GIF\n6- Video\n\nاكتب الرقم مع السلاش (مثال: /1)"
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
		parsedCount := 0
		isValid := false
		for _, char := range lastPart {
			if char >= '0' && char <= '9' {
				parsedCount = parsedCount*10 + int(char-'0')
				isValid = true
			} else if char >= '٠' && char <= '٩' {
				parsedCount = parsedCount*10 + int(char-'٠')
				isValid = true
			} else {
				isValid = false
				break
			}
		}
		if isValid && parsedCount > 0 {
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
				} else if doc := hMsg.GetDocumentMessage(); doc != nil {
					if strings.HasPrefix(doc.GetMimetype(), "image/") {
						mediaMsgs = append(mediaMsgs, doc)
						isVideoList = append(isVideoList, false)
					} else if strings.HasPrefix(doc.GetMimetype(), "video/") {
						mediaMsgs = append(mediaMsgs, doc)
						isVideoList = append(isVideoList, true)
					}
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
			sendMessage(ctx, fmt.Sprintf("جاري سرقة وتعديل حقوق %d ملصق", len(mediaMsgs)))
		} else {
			sendMessage(ctx, fmt.Sprintf("جاري تحويل %d صورة إلى ملصقات بحقوقك", len(mediaMsgs)))
		}

		// Reverse them back to send in order
		// Limit concurrency to 2 to avoid overloading the CPU with ffmpeg
		var wg sync.WaitGroup
		sem := make(chan struct{}, 2)
		for i := len(mediaMsgs) - 1; i >= 0; i-- {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				
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
		} else if doc := unwrappedMsg.GetDocumentMessage(); doc != nil && (strings.HasPrefix(doc.GetMimetype(), "image/") || strings.HasPrefix(doc.GetMimetype(), "video/")) {
			mediaMsg = doc
			isVideo = strings.HasPrefix(doc.GetMimetype(), "video/")
		} else if ext := unwrappedMsg.GetExtendedTextMessage(); ext != nil {
			quoted := UnwrapMessage(ext.GetContextInfo().GetQuotedMessage())
			if qImg := quoted.GetImageMessage(); qImg != nil {
				mediaMsg = qImg
			} else if qVid := quoted.GetVideoMessage(); qVid != nil {
				mediaMsg = qVid
				isVideo = true
			} else if qDoc := quoted.GetDocumentMessage(); qDoc != nil && (strings.HasPrefix(qDoc.GetMimetype(), "image/") || strings.HasPrefix(qDoc.GetMimetype(), "video/")) {
				mediaMsg = qDoc
				isVideo = strings.HasPrefix(qDoc.GetMimetype(), "video/")
			}
		}
		if mediaMsg == nil {
			sendMessage(ctx, "أرسل صورة أو فيديو مع الأمر، أو رد على صورة/فيديو.")
			return
		}
	}

	if isSteal {
		outMsg := store.GetCustomOutput(getLID(ctx, ctx.Sender), ".تعديل ملصق", "يتم التعديل")
		sendMessage(ctx, outMsg)
	} else {
		outMsg := store.GetCustomOutput(getLID(ctx, ctx.Sender), ".ملصق", "جاري صنع الملصق")
		sendMessage(ctx, outMsg)
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
		outMsg := store.GetCustomOutput(getLID(ctx, ctx.Sender), ".منع امر", "تم منع الأمر بنجاح!")
		sendMessage(ctx, outMsg)
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
		outMsg := store.GetCustomOutput(getLID(ctx, ctx.Sender), ".منع منع", "تم فك منع الأمر بنجاح!")
		sendMessage(ctx, outMsg)
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
		outMsg := store.GetCustomOutput(getLID(ctx, ctx.Sender), ".منع", "تم منعه بنجاح!")
		sendMessage(ctx, outMsg)
	}
}

func baymax(ctx *BotContext) {
	out := store.GetCustomOutput(getLID(ctx, ctx.Sender), ".baymax", "")
	if out == "" {
		out = store.GetCustomOutput(getLID(ctx, ctx.Sender), ".buymax", "")
	}
	if out != "" {
		sendMessage(ctx, out)
		return
	}

	name := store.GetBaymaxName(getLID(ctx, ctx.Sender))
	if name != "" {
		sendMessage(ctx, name)
	} else if ctx.Sender.User == "966508364121" || ctx.Event.Info.IsFromMe {
		sendMessage(ctx, "هاي هبهب ")
	} else {
		sendMessage(ctx, "هاي عزام سينباي ")
	}
}

func showCommands(ctx *BotContext) {
	cmds := `[ الدليل الشامل لجميع أوامر وخصائص البوت ]
====================

[ أوامر الإدارة والحماية ]
.حماية (لحماية أدمن معين بالرد عليه، إذا انطرد أو انسحب إشرافه بينسحب إشراف الكل)
.طرد (لطرد شخص بالرد على رسالته)
.ميوت (لكتم شخص بالرد عليه، أي رسالة يرسلها تنحذف فورا)
.فك ميوت (لفك الكتم عن شخص)
.سحب اشراف (لسحب إشراف الأدمنز بالقروب)
.حذف (لحذف رسالة بالرد عليها)
.كومنت (لمنع الأعضاء من إرسال رسائل - يقفل الشات)
.فك كومنت (للسماح للأعضاء بإرسال رسائل)
يا معين* (تفعيل نظام الحماية: أي شخص يرسل جهة اتصال ينطرد فوراً)

[ أوامر التفعيل والتعطيل ]
.baymax أو .buymax (لتفعيل البوت في القروب عشان يبدأ يستجيب للأوامر)
.bot off (لتعطيل البوت في القروب)
.دخلني قروبات (لفتح/إغلاق ميزة الدخول التلقائي للقروبات من الروابط)\n.تفعيل التبادل / .ايقاف التبادل (لتفعيل أو إيقاف نظام التبادل بالكامل)

[ أوامر التبادل (تعمل دائماً) ]
.رابطي أو .روابطي (بالرد على رسالتك لحفظها كرسالة تبادل)
.حذف روابطي (لحذف جميع روابطك المحفوظة)
.تبادل (لبدء جلسة التبادل، ترسل روابطك وتُرسل للقروب الأساسي)
.انتهيت (لإنهاء جلسة التبادل مبكراً)
.نشر (لنشر رسائل التبادل الخاصة بك لكل الأرقام في المفضلة)
.مفضلة (لإضافة/إزالة رقم من قائمة مفضلة التبادل)
!تبادل (لتعيين القروب الحالي كقروب تبادل أساسي)

[ أوامر التحميل والميديا ]
.اغنية (لتحميل المقاطع كصوتيات مع التفاصيل)
.يوتيوب (للبحث التفاعلي في اليوتيوب)
.تحميل (لتحميل المقاطع من تيك توك، إنستقرام، إكس، سناب شات)
.فيديو [كلمة] [عدد] (للبحث وتنزيل مقاطع فيديو بصمت)
.new (لجلب مقاطع جديدة من بحثك السابق)
.mp3 (تحويل أي فيديو إلى صوت بالرد عليه)
.بينتريست أو .بحث (للبحث عن صور)
/5 [كلمة] (للبحث عن صور متحركة GIF من تينور)
.فوريو (صور عشوائية)
.تطقيم (صور متطابقة)

[ أوامر الملصقات ]
.ملصق (لتحويل الصور والفيديوهات لملصقات)
.عمل حزمة (لتحويل مجموعة صور/فيديوهات إلى حزمة ملصقات)
.انهاء الحزمة (لإنهاء جلسة صنع الحزمة وإرسالها)
.تعديل حقوق (لتغيير حقوق الحزمة الافتراضية)
.حقوق (لتغيير حقوق ملصق معين)
.تعديل ملصق (للتعديل على ملصق موجود)
.تعديل حزمة (للتعديل على حزمة موجودة)
.زرف (لسحب ملصق وتغيير حقوقه للحقوق الخاصة بك)
.استهبال (يخلي البوت يضحك على رسالة معينة)

[ أوامر الذكاء الاصطناعي والأدوات ]
.جيميناي (للتحدث مع الذكاء الاصطناعي)
.ترجم (لترجمة نص)
.تفريغ (لتفريغ المقاطع الصوتية إلى نص)
.عزل (لعزل الموسيقى عن الصوت)
.طقس [المدينة] (لمعرفة حالة الطقس)
.ايميل (لإنشاء إيميل مؤقت واستقبال الرسائل عليه)
.حديث [رقم] (لاستخراج أحاديث نبوية)
.مواعيد الصلاة (لمعرفة مواعيد الصلاة)
.pdf (لإنشاء ملف PDF من الصور)

[ أوامر الألعاب والترفيه ]
.العاب (لعرض قائمة الألعاب مثل اونو، اكس او، والمشنقة)
.مارد أو .المارد (للعب مع المارد السحري أكيناتور)
.حوم (لعبة الحوم)
.كت، .لو، .نسبة، .سؤال، .هل، .صراحة، .فعالية، .عقاب، .انمي، .اقتباس

[ أوامر التخصيص ]
.منع امر (لمنع أمر معين عن شخص)
.فك منع امر (لإلغاء المنع)
.تعديل امر (لتغيير اسم أمر)
.تعديل رد (لتغيير رد البوت الافتراضي على أمر معين)
.بروفايل (لعرض بروفايلك)
.وش لقبك (لمعرفة اللقب الخاص بك - هبة)

====================
ملاحظة: البوت الآن مبرمج للعمل بهدوء بدون أي علامات ترقيم مزعجة.`
	
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

func handleDeleteMsgOrHebebia(ctx *BotContext) {
	parts := strings.Fields(ctx.Text)

	if ctx.Event.Message.GetExtendedTextMessage() != nil {
		ctxInfo := ctx.Event.Message.GetExtendedTextMessage().GetContextInfo()
		if ctxInfo.GetStanzaID() != "" {
			if !store.IsAllowed(getLID(ctx, ctx.Sender)) && !ctx.Event.Info.IsFromMe {
				return
			}
			count := 1
			if len(parts) > 1 {
				parsedCount, err := strconv.Atoi(parts[1])
				if err == nil && parsedCount > 0 {
					count = parsedCount
				}
			}

			targetJID, _ := types.ParseJID(ctxInfo.GetParticipant())
			if count == 1 {
				ctx.Client.SendMessage(context.Background(), ctx.ChatID, ctx.Client.BuildRevoke(ctx.ChatID, targetJID, ctxInfo.GetStanzaID()))
				return
			}

			targetID := targetJID.ToNonAD().String()
			history := store.GetHistory(ctx.ChatID.String())
			var toDelete []string
			for i := len(history) - 1; i >= 0; i-- {
				if history[i].Sender == targetID {
					toDelete = append(toDelete, history[i].ID)
					if len(toDelete) >= count {
						break
					}
				}
			}
			if len(toDelete) > 0 {
				go func() {
					for _, msgID := range toDelete {
						ctx.Client.SendMessage(context.Background(), ctx.ChatID, ctx.Client.BuildRevoke(ctx.ChatID, targetJID, msgID))
					}
				}()
			}
			return
		} else if len(ctxInfo.GetMentionedJID()) > 0 {
			if !store.IsAllowed(getLID(ctx, ctx.Sender)) && !ctx.Event.Info.IsFromMe {
				return
			}
			count := 1
			if len(parts) > 1 {
				parsedCount, err := strconv.Atoi(parts[1])
				if err == nil && parsedCount > 0 {
					count = parsedCount
				}
			}
			targetJID, _ := types.ParseJID(ctxInfo.GetMentionedJID()[0])
			targetID := targetJID.ToNonAD().String()
			history := store.GetHistory(ctx.ChatID.String())
			var toDelete []string
			for i := len(history) - 1; i >= 0; i-- {
				if history[i].Sender == targetID {
					toDelete = append(toDelete, history[i].ID)
					if len(toDelete) >= count {
						break
					}
				}
			}
			if len(toDelete) > 0 {
				go func() {
					for _, msgID := range toDelete {
						ctx.Client.SendMessage(context.Background(), ctx.ChatID, ctx.Client.BuildRevoke(ctx.ChatID, targetJID, msgID))
					}
				}()
			}
			return
		}
	}

	if !ctx.Event.Info.IsFromMe && getLID(ctx, ctx.Sender) != "224245258948685@lid" {
		return
	}
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
	sendMessage(ctx, "رقم غير صحيح أو ما منشنت/رديت على أحد عشان أحذف رسائله.")
}

func setName(ctx *BotContext) {
	name := strings.TrimSpace(strings.TrimPrefix(ctx.Text, ".اسمي"))
	if name != "" {
		store.SetBaymaxName(getLID(ctx, ctx.Sender), name, ".")
		sendMessage(ctx, "تم حفظ اسمك بنجاح! ناديني الحين ")
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

	sendMessage(ctx, "تم بدء لعبة حوم! \nاكتب `.دخول` للمشاركة.\nولما يكتمل العدد اكتب `.بدء`")
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

	choices := []string{"حجر", "ورقة", "مقص"}
	msg := "بدأت اللعبة! الاختيارات العشوائية:\n\n"

	winnerIndex := rand.Intn(len(players))
	winner := players[winnerIndex]

	for i, p := range players {
		choice := choices[rand.Intn(len(choices))]
		if i == winnerIndex {
			msg += fmt.Sprintf("@%s -> %s (الفائز! )\n", p.User, choice)
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
		sendMessage(ctx, "مبروك! فاز الإداري! جاري زرف القروب")
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
		sendMessage(ctx, "لقد فاز لاعب عادي! لن يتم زرف القروب.")
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
			outMsg := store.GetCustomOutput(getLID(ctx, ctx.Sender), ".سحب اشراف", "تم سحب الإشراف بنجاح!")
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
	customText, ok := store.GetWelcomeGroup(groupJID.String())
	if !ok || customText == "" {
		return // Not enabled or no text for this group
	}

	pic, err := clientWA.GetProfilePictureInfo(context.Background(), joinedUser, &whatsmeow.GetProfilePictureParams{})
	var data []byte
	if err == nil && pic != nil && pic.URL != "" {
		resp, err := http.Get(pic.URL)
		if err == nil {
			data, _ = io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	}

	text := fmt.Sprintf("@%s\n%s", joinedUser.User, customText)

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
	// First check if there's a quoted message we should repeat
	quotedText := ""
	unwrappedMsg := UnwrapMessage(ctx.Event.Message)
	if ext := unwrappedMsg.GetExtendedTextMessage(); ext != nil {
		quoted := UnwrapMessage(ext.GetContextInfo().GetQuotedMessage())
		if quoted.GetConversation() != "" {
			quotedText = quoted.GetConversation()
		} else if quoted.GetExtendedTextMessage() != nil {
			quotedText = quoted.GetExtendedTextMessage().GetText()
		}
	}

	// Remove the command prefix (.تكرار)
	text := ctx.Text
	fields := strings.Fields(text)
	if len(fields) > 0 {
		text = strings.TrimSpace(strings.TrimPrefix(text, fields[0]))
	}

	parts := strings.Fields(text)
	if len(parts) == 0 {
		sendMessage(ctx, "الصيغة: .تكرار <الرسالة> <العدد>\nأو رد على رسالة واكتب: .تكرار <العدد>")
		return
	}

	// The last part should be the count
	lastPart := parts[len(parts)-1]
	count, err := strconv.Atoi(lastPart)
	if err != nil || count <= 0 || count > 2000 {
		sendMessage(ctx, "العدد لازم يكون رقم صحيح في نهاية الرسالة (حد أقصى 2000)")
		return
	}

	// If there's quoted text and only the count was provided in the command
	var msgToRepeat string
	if quotedText != "" && len(parts) == 1 {
		msgToRepeat = quotedText
	} else {
		// Extract everything before the last number
		idx := strings.LastIndex(text, lastPart)
		if idx == -1 {
			return
		}
		msgToRepeat = strings.TrimSpace(text[:idx])
	}

	if msgToRepeat == "" {
		sendMessage(ctx, "وين الكلام اللي تبيني أكرره؟")
		return
	}

	repeatedMsg := strings.Repeat(msgToRepeat+" ", count)
	sendMessage(ctx, strings.TrimSpace(repeatedMsg))
}

func processYoutubeMedia(ctx *BotContext, isAudio bool) {
	cmdName := ".اغنية"
	if !isAudio {
		cmdName = ".فيديو"
	}
	query := strings.TrimSpace(strings.TrimPrefix(ctx.Text, strings.Split(ctx.Text, " ")[0]))
	if query == "" {
		sendMessage(ctx, fmt.Sprintf("اكتب اسم المقطع مع الأمر! مثلاً:\n%s رابح صقر", cmdName))
		return
	}

	videoID, err := youtube.SearchVideo(query)
	if err != nil {
		sendMessage(ctx, "ما قدرت ألقى المقطع، تأكد من مفتاح الـ API أو حاول باسم ثاني!")
		return
	}

	var infoWg sync.WaitGroup
	var mediaWg sync.WaitGroup
	var info *youtube.VideoInfo
	var infoErr error
	var thumbData []byte
	var mediaData []byte
	var mediaErr error

	infoWg.Add(1)
	go func() {
		defer infoWg.Done()
		info, infoErr = youtube.GetVideoDetails(videoID)
		if infoErr == nil && info.Thumbnail != "" {
			if resp, err := http.Get(info.Thumbnail); err == nil {
				if resp.StatusCode != 200 {
					resp.Body.Close()
					// Fallback to hqdefault
					if fallback, err2 := http.Get(fmt.Sprintf("https://i.ytimg.com/vi/%s/hqdefault.jpg", videoID)); err2 == nil {
						data, _ := io.ReadAll(fallback.Body)
						fallback.Body.Close()
						thumbData, _ = youtube.CropTo16x9(data)
					}
				} else {
					data, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					thumbData, _ = youtube.CropTo16x9(data)
				}
			}
		}
	}()

	mediaWg.Add(1)
	go func() {
		defer mediaWg.Done()
		mediaData, mediaErr = youtube.DownloadMedia(videoID, isAudio)
	}()

	// Wait ONLY for info and thumbnail so we can send it instantly!
	infoWg.Wait()

	if infoErr != nil {
		sendMessage(ctx, fmt.Sprintf("جبت المقطع بس فشلت في استخراج تفاصيله!\nالسبب: %s", infoErr.Error()))
		return
	}

	caption := youtube.FormatCaption(info)

	var thumbMsgID string
	if len(thumbData) > 0 {
		uploadedThumb, err := ctx.Client.Upload(context.Background(), thumbData, whatsmeow.MediaImage)
		if err == nil {
			imgMsg := &waProto.ImageMessage{
				URL:           proto.String(uploadedThumb.URL),
				DirectPath:    proto.String(uploadedThumb.DirectPath),
				MediaKey:      uploadedThumb.MediaKey,
				Mimetype:      proto.String("image/jpeg"),
				FileEncSHA256: uploadedThumb.FileEncSHA256,
				FileSHA256:    uploadedThumb.FileSHA256,
				FileLength:    proto.Uint64(uint64(len(thumbData))),
				Caption:       proto.String(caption),
			}
			resp, _ := ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{ImageMessage: imgMsg})
			thumbMsgID = resp.ID
		}
	}

	if thumbMsgID == "" {
		resp, _ := ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
			ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String(caption)},
		})
		thumbMsgID = resp.ID
	}

	// Now wait for the actual media download to finish!
	mediaWg.Wait()

	if mediaErr != nil {
		fmt.Printf("DownloadMedia Error for %s: %v\n", videoID, mediaErr)
		sendMessage(ctx, "فشل تحميل المقطع، ممكن يكون طويل جداً أو فيه مشكلة بالشبكة!")
		return
	}

	mediaType := whatsmeow.MediaAudio
	mimeType := "audio/mp4"
	if !isAudio {
		mediaType = whatsmeow.MediaVideo
		mimeType = "video/mp4"
	}

	uploadedMedia, err := ctx.Client.Upload(context.Background(), mediaData, mediaType)
	if err != nil {
		sendMessage(ctx, "فشل رفع المقطع للواتساب!")
		return
	}

	finalMsg := &waProto.Message{}
	if isAudio {
		finalMsg.AudioMessage = &waProto.AudioMessage{
			URL:           proto.String(uploadedMedia.URL),
			DirectPath:    proto.String(uploadedMedia.DirectPath),
			MediaKey:      uploadedMedia.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploadedMedia.FileEncSHA256,
			FileSHA256:    uploadedMedia.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(mediaData))),
			ContextInfo: &waProto.ContextInfo{
				StanzaID:    proto.String(thumbMsgID),
				Participant: proto.String(ctx.Client.Store.ID.String()),
				QuotedMessage: &waProto.Message{
					ImageMessage: &waProto.ImageMessage{Caption: proto.String(caption)},
				},
			},
		}
	} else {
		finalMsg.VideoMessage = &waProto.VideoMessage{
			URL:           proto.String(uploadedMedia.URL),
			DirectPath:    proto.String(uploadedMedia.DirectPath),
			MediaKey:      uploadedMedia.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploadedMedia.FileEncSHA256,
			FileSHA256:    uploadedMedia.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(mediaData))),
			ContextInfo: &waProto.ContextInfo{
				StanzaID:    proto.String(thumbMsgID),
				Participant: proto.String(ctx.Client.Store.ID.String()),
				QuotedMessage: &waProto.Message{
					ImageMessage: &waProto.ImageMessage{Caption: proto.String(caption)},
				},
			},
		}
	}

	ctx.Client.SendMessage(context.Background(), ctx.ChatID, finalMsg)
}

func reactMessage(ctx *BotContext) {
	emoji := strings.TrimSpace(strings.TrimPrefix(ctx.Text, strings.Split(ctx.Text, " ")[0]))
	if emoji == "" {
		sendMessage(ctx, "اكتب الإيموجي مع الأمر! مثلاً:\n.react 💔")
		return
	}

	// Must be replying to a message
	var targetStanzaID string
	var targetParticipant string
	
	unwrappedMsg := UnwrapMessage(ctx.Event.Message)
	if ext := unwrappedMsg.GetExtendedTextMessage(); ext != nil {
		if ctxInfo := ext.GetContextInfo(); ctxInfo != nil {
			if ctxInfo.StanzaID != nil && ctxInfo.Participant != nil {
				targetStanzaID = *ctxInfo.StanzaID
				targetParticipant = *ctxInfo.Participant
			}
		}
	}

	if targetStanzaID == "" {
		sendMessage(ctx, "لازم ترد على رسالة الشخص اللي تبي تحط عليها رياكت!")
		return
	}

	targetJID, _ := types.ParseJID(targetParticipant)
	
	ctx.Client.SendMessage(context.Background(), ctx.ChatID, ctx.Client.BuildReaction(ctx.ChatID, targetJID, targetStanzaID, emoji))
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
		sendMessage(ctx, "تم قفل القروب بنجاح ")
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
		sendMessage(ctx, "تم فتح القروب بنجاح ")
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
		sendMessage(ctx, "تم تغيير رابط القروب بنجاح (ما راح أرسل الرابط الجديد)")
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
		sendMessage(ctx, "تم تغيير صورة القروب بنجاح")
	}
}

func refreshPinterest(ctx *BotContext) {
	if last, ok := pinterest.GetLastSearch(ctx.ChatID.String()); ok {
		sendMessage(ctx, "جاري البحث عن صور جديدة")
		go func() {
			var results []pinterest.PinResult
			if last.Aspect == "foryou" {
				results = pinterest.ForYouPinterest("all")
			} else if last.Aspect == "matching" {
				results = pinterest.SearchPinterestMatchingIcons(last.Query)
			} else if last.IsVisual && last.Base64Image != "" {
				results = pinterest.SearchPinterestLens(last.Base64Image, last.Aspect, last.Count)
			} else {
				var newBm string
				results, newBm = pinterest.SearchPinterest(last.Query, last.Aspect, last.Count, last.Bookmark)
				pinterest.SetLastSearch(ctx.ChatID.String(), last.Query, last.Aspect, last.Count, last.IsVisual, last.Base64Image, newBm)
			}

			if len(results) > 0 && last.Aspect != "matching" {
				rand.Shuffle(len(results), func(i, j int) {
					results[i], results[j] = results[j], results[i]
				})
			}

			var pinsToSend []pinterest.PinResult
			for i, res := range results {
				if i >= last.Count {
					break
				}
				pinsToSend = append(pinsToSend, res)
			}

			count := 0
			for _, res := range pinsToSend {
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
							ContextInfo: &waProto.ContextInfo{
								StanzaID:      proto.String(ctx.Event.Info.ID),
								Participant:   proto.String(ctx.Event.Info.Sender.String()),
								QuotedMessage: ctx.Event.Message,
							},
						}
						sendResp, err := sendAndCacheImage(ctx, ctx.ChatID, imgMsg)
						if err == nil && res.ID != "" {
							pinterest.SaveMessagePin(sendResp.ID, res.ID)
						}
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
	
	pinterest.SetLastSearch(ctx.ChatID.String(), "", "foryou", count, false, "", "")

	sendMessage(ctx, "جاري جلب صور للفوريو")
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
					sendResp, err := sendAndCacheImage(ctx, ctx.ChatID, imgMsg)
					if err == nil && res.ID != "" {
						pinterest.SaveMessagePin(sendResp.ID, res.ID)
					}
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
	
	sendMessage(ctx, "جاري البحث عن تطقيمات")
	go func() {
		results := pinterest.SearchPinterestMatchingIcons(query)
		pairs := pinterest.GetMatchingPairs(results, 1)
		
		if len(pairs) >= 2 {
			count := 0
			for _, imgUrl := range pairs {
				if count >= 2 {
					break
				}
				data, err := pinterest.DownloadImage(imgUrl)
				if err == nil && len(data) > 100 {
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
						sendResp, err := sendAndCacheImage(ctx, ctx.ChatID, imgMsg)
						if err == nil {
							// We can't save PinID for matching icons pairs trivially since they are just strings, so skip it.
							// The reaction will fallback to visual search lens.
							_ = sendResp
						}
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
	fmt.Println("HandleReaction TRIGGERED!")
	
	if imgData == nil {
		fmt.Println("imgData is nil, cannot do visual search")
		return
	}
	base64Image := base64.StdEncoding.EncodeToString(imgData)
	results := pinterest.SearchPinterestLens(base64Image, "all", 10)

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
					msg := &waProto.Message{ImageMessage: imgMsg}
					sendResp, err := client.SendMessage(context.Background(), chatID, msg)
					if err == nil {
						dummy := &events.Message{
							Info: types.MessageInfo{ID: sendResp.ID, MessageSource: types.MessageSource{Chat: chatID, IsFromMe: true}, Timestamp: sendResp.Timestamp},
							Message: msg,
						}
						AddMessage(chatID.String(), dummy)
					}
					if err == nil && res.ID != "" {
						pinterest.SaveMessagePin(sendResp.ID, res.ID)
					}
					count++
				}
			}
		}
	}
}

func disableCommand(ctx *BotContext, parts []string) {
	if len(parts) < 2 {
		sendMessage(ctx, "الصيغة: .كومنت <الأمر>\nمثال: .كومنت .ميوت")
		return
	}
	cmd := strings.ToLower(parts[1])
	if !strings.HasPrefix(cmd, ".") {
		cmd = "." + cmd
	}
	store.SetCommandDisabled(cmd, true, ".")
	sendMessage(ctx, fmt.Sprintf("تم إيقاف الأمر %s بنجاح!", cmd))
}

func enableCommand(ctx *BotContext, parts []string) {
	if len(parts) < 3 {
		sendMessage(ctx, "الصيغة: .فك كومنت <الأمر>\nمثال: .فك كومنت .ميوت")
		return
	}
	cmd := strings.ToLower(parts[2])
	if !strings.HasPrefix(cmd, ".") {
		cmd = "." + cmd
	}
	store.SetCommandDisabled(cmd, false, ".")
	sendMessage(ctx, fmt.Sprintf("تم تفعيل الأمر %s بنجاح! ", cmd))
}

func setAutoReact(ctx *BotContext, parts []string) {
	if len(parts) < 2 {
		sendMessage(ctx, "الصيغة: .رياكت <ايموجي> (مع منشن أو ريبلاي)\nأو .رياكت مسح (للإلغاء)")
		return
	}
	
	emoji := parts[1]
	
	targets := getTargets(ctx)
	if len(targets) == 0 {
		sendMessage(ctx, "لازم تسوي منشن أو ريبلاي على الشخص!")
		return
	}
	
	targetID := getLID(ctx, targets[0])
	
	if emoji == "مسح" {
		store.SetAutoReact(targetID, "", ".")
		sendMessage(ctx, "تم إزالة التفاعل التلقائي عن هذا الشخص.")
		return
	}

	store.SetAutoReact(targetID, emoji, ".")
	sendMessage(ctx, fmt.Sprintf("تم! الحين أي رسالة يرسلها راح يتفاعل عليها البوت بـ %s", emoji))
}

func setWelcomeFeature(ctx *BotContext) {
	if !ctx.Event.Info.IsGroup {
		sendMessage(ctx, "هذا الأمر للقروبات فقط!")
		return
	}
	
	// text can be just ".استقبال" or ".استقبال كلام طويل..."
	text := strings.TrimSpace(strings.TrimPrefix(ctx.Text, strings.Split(ctx.Text, " ")[0]))
	
	if text == "مسح" || text == "ايقاف" || text == "إيقاف" {
		store.SetWelcomeGroup(ctx.ChatID.String(), "", ".")
		sendMessage(ctx, "تم إيقاف الاستقبال التلقائي في هذا القروب.")
		return
	}

	if text == "" {
		text = "مرحبا بك في الحصن"
	}
	
	store.SetWelcomeGroup(ctx.ChatID.String(), text, ".")
	sendMessage(ctx, "تم تفعيل الاستقبال في هذا القروب بنجاح! أي شخص بيدخل راح تترسل صورته مع الكابشن اللي اخترته.")
}

func HandleMoroccan(ctx *BotContext) {
	if ctx.Event.Message != nil && ctx.Event.Message.ExtendedTextMessage != nil && ctx.Event.Message.ExtendedTextMessage.ContextInfo != nil {
		qMsg := ctx.Event.Message.ExtendedTextMessage.ContextInfo.QuotedMessage
		if qMsg != nil {
			participant := ctx.Event.Message.ExtendedTextMessage.ContextInfo.GetParticipant()
			myJid := ctx.Client.Store.ID.ToNonAD().String()
			
			if participant == myJid {
				qText := ""
				if qMsg.ExtendedTextMessage != nil {
					qText = qMsg.ExtendedTextMessage.GetText()
				} else if qMsg.Conversation != nil {
					qText = qMsg.GetConversation()
				}
				qText = strings.TrimSpace(qText)
				if qText == "وش لقبك" || qText == "وش لقبي" {
					sendMessage(ctx, "New character unlock hibi💫🔓")
				}
			}
		}
	}
}

func HandleSyrian(ctx *BotContext) {
	if ctx.Text == ".دخلني قروبات" {
		AutoJoinGroups = !AutoJoinGroups
		// Silently return, no message!
		return
	}

	if AutoJoinGroups && strings.Contains(ctx.Text, "chat.whatsapp.com/") {
		parts := strings.Split(ctx.Text, "chat.whatsapp.com/")
		if len(parts) > 1 {
			code := strings.FieldsFunc(parts[1], func(r rune) bool {
				return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_')
			})
			if len(code) > 0 {
				linkCode := code[0]
				link := "https://chat.whatsapp.com/" + linkCode
				
				seenLinksMutex.Lock()
				isSeen := seenGroupLinks[linkCode]
				if !isSeen {
					seenGroupLinks[linkCode] = true
				}
				seenLinksMutex.Unlock()
				
				if !isSeen {
					go func() {
						myJID := ctx.Client.Store.ID.ToNonAD()
						ctx.Client.SendMessage(context.Background(), myJID, &waProto.Message{
							// Just the raw link, no extra text
							Conversation: proto.String(link),
						})
					}()
				}
			}
		}
	}
	HandleExchangeMessage(ctx)
}
