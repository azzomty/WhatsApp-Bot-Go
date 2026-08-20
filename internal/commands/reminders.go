package commands

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

var reminderRegex = regexp.MustCompile(`^(?P<text>.*?)\s+(?P<time>\d+)(?P<unit>[mshdمسثد])$`)

func HandleReminder(ctx *BotContext, command string) {
	if ctx.Event.Info.IsGroup {
		sendMessage(ctx, "أوامر التذكير متاحة فقط في الخاص لحماية إزعاج القروبات.")
		return
	}

	if command == ".ذكرني اتصال" {
		sendMessage(ctx, "للأسف ميزة الاتصال (المكالمات) غير مدعومة حالياً من طرف مكتبة الواتساب لأنها تتطلب إعدادات WebRTC كاملة. لكن تم تحويل طلبك إلى (تذكير رسالة) مؤقتاً! ")
	}

	// Remove the command prefix
	parts := strings.SplitN(ctx.Text, " ", 3)
	if len(parts) < 3 {
		sendMessage(ctx, "الصيغة غير صحيحة! جرب مثلاً:\n.ذكرني رسالة اقفل الفرن 10m\n(m: دقايق, s: ثواني, h: ساعات)")
		return
	}

	// The text part is everything after the command
	fullArgs := strings.Join(parts[2:], " ")

	matches := reminderRegex.FindStringSubmatch(fullArgs)
	if len(matches) == 0 {
		sendMessage(ctx, "ما قدرت أفهم الوقت! جرب الصيغة هذي:\n.ذكرني رسالة اقفل الفرن 10m\nأو\n.ذكرني رسالة اقفل الفرن 10د")
		return
	}

	reminderText := strings.TrimSpace(matches[1])
	timeValueStr := matches[2]
	timeUnit := matches[3]

	timeValue, err := strconv.Atoi(timeValueStr)
	if err != nil {
		sendMessage(ctx, "رقم الوقت غير صحيح.")
		return
	}

	var duration time.Duration
	switch timeUnit {
	case "s", "ث":
		duration = time.Duration(timeValue) * time.Second
	case "m", "د":
		duration = time.Duration(timeValue) * time.Minute
	case "h", "س":
		duration = time.Duration(timeValue) * time.Hour
	default:
		duration = time.Duration(timeValue) * time.Minute
	}

	if duration <= 0 || duration > 24*time.Hour {
		sendMessage(ctx, "الوقت لازم يكون بين ثانية واحدة و 24 ساعة.")
		return
	}

	sendMessage(ctx, fmt.Sprintf("تم! راح أذكرك بـ (%s) بعد %d %s ", reminderText, timeValue, getUnitName(timeUnit)))

	go func(jid string, text string, d time.Duration) {
		time.Sleep(d)
		msg := fmt.Sprintf("⏰ *تــذكــيــر* ⏰\n\n%s", text)
		ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
			ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String(msg)},
		})
	}(ctx.Event.Info.Chat.String(), reminderText, duration)
}

func getUnitName(unit string) string {
	switch unit {
	case "s", "ث":
		return "ثانية"
	case "m", "د":
		return "دقيقة"
	case "h", "س":
		return "ساعة"
	}
	return ""
}
