package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

type TruecallerResponse struct {
	Data []struct {
		Name   string `json:"name"`
		Phones []struct {
			E164Format string `json:"e164Format"`
			Carrier    string `json:"carrier"`
		} `json:"phones"`
		Badges []string `json:"badges"`
		Score  *struct {
			SpamScore int `json:"spamScore"`
		} `json:"score"`
	} `json:"data"`
}

func HandleCallerID(ctx *BotContext, number string) {
	token := os.Getenv("TRUECALLER_TOKEN")
	if token == "" {
		msg := `🚨 *النظام السري معطل حالياً!* 🚨

عشان أشغل لك اختراق (تروكولر VIP) بدون إعلانات وبدون حدود، لازم تعطيني تصريح الدخول (Token).

*الطريقة بـ HTTP Toolkit:*
1. افتح HTTP Toolkit بجوالك وشغل تطبيق Truecaller.
2. ابحث عن أي رقم.
3. بالـ Toolkit، دور على طلب (Request) رايح لرابط يبدأ بـ:
` + "`" + `search5-noneu.truecaller.com` + "`" + `
4. انسخ الكود الطويل اللي قدام كلمة ` + "`" + `Authorization: Bearer` + "`" + `
5. ضيفه في متغيرات Render (Environment Variables) بالاسم هذا:
` + "`" + `TRUECALLER_TOKEN` + "`" + `

بمجرد ما تحطه، الأمر راح يشتغل طيارة! 🚀`
		sendMessage(ctx, msg)
		return
	}

	if number == "" {
		sendMessage(ctx, "اكتب الرقم بعد الأمر! مثال:\n.اعرف 0555555555")
		return
	}

	sendMessage(ctx, "🕵️‍♂️ جاري اختراق قاعدة البيانات...")

	// Cleanup number
	number = strings.ReplaceAll(number, " ", "")
	number = strings.ReplaceAll(number, "+", "")
	number = strings.ReplaceAll(number, "-", "")

	url := fmt.Sprintf("https://search5-noneu.truecaller.com/v2/search?countryCode=SA&type=4&q=%s", number)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		sendMessage(ctx, "حدث خطأ داخلي في النظام.")
		return
	}

	// Format token correctly if the user accidentally copied "Bearer " with it
	token = strings.TrimPrefix(token, "Bearer ")
	
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", "Truecaller/11.75.5 (Android;10)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		sendMessage(ctx, "❌ فشل الاتصال بالسيرفر السري.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		sendMessage(ctx, "❌ التوكن (Token) حق تروكولر منتهي أو غلط! جيب توكن جديد من HTTP Toolkit وحدثه في Render.")
		return
	}

	var res TruecallerResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		sendMessage(ctx, "❌ فشل قراءة البيانات المهربة من السيرفر.")
		return
	}

	if len(res.Data) == 0 {
		sendMessage(ctx, "❓ للأسف، الرقم هذا ماله أي اسم في قاعدة البيانات (شبح 👻).")
		return
	}

	person := res.Data[0]
	name := person.Name
	
	carrier := "غير معروف"
	if len(person.Phones) > 0 && person.Phones[0].Carrier != "" {
		carrier = person.Phones[0].Carrier
	}

	spamScore := 0
	if person.Score != nil {
		spamScore = person.Score.SpamScore
	}

	isVerified := ""
	for _, badge := range person.Badges {
		if badge == "verified" {
			isVerified = " ✅ (حساب موثق)"
			break
		}
	}

	msg := fmt.Sprintf("🚨 *نتيجة الفحص السري* 🚨\n\n")
	msg += fmt.Sprintf("👤 *الاسم:* %s%s\n", name, isVerified)
	msg += fmt.Sprintf("📱 *المزود:* %s\n", carrier)
	
	if spamScore > 0 {
		msg += fmt.Sprintf("⚠️ *نسبة السبام (إزعاج):* %d%%\n", spamScore)
	}

	ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String(msg)},
	})
}
