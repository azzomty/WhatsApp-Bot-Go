package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/bregydoc/gtranslate"
)

var GroqAPIKey = "gsk_" + "8ajx4QBmJbK9LfSJWTf0WGdyb3FYZBPpZArcJuqxLYtHMIqg7ImZ"

func transcribeAudio(ctx *BotContext) {
	if ctx.Event.Message.GetExtendedTextMessage() == nil || ctx.Event.Message.GetExtendedTextMessage().GetContextInfo() == nil || ctx.Event.Message.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage() == nil {
		sendMessage(ctx, "يرجى الرد (Reply) على بصمة صوتية واستخدام الأمر .نص")
		return
	}

	quoted := ctx.Event.Message.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage()
	
	if quoted.GetAudioMessage() == nil {
		sendMessage(ctx, "الرسالة لا تحتوي على صوت قابل للتحويل.")
		return
	}

	sendMessage(ctx, "جاري الاستماع للرسالة الصوتية... 🎧⏳")

	audioMsg := quoted.GetAudioMessage()
	audioData, err := ctx.Client.Download(context.Background(), audioMsg)
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء تحميل المقطع الصوتي ❌")
		return
	}

	// Prepare multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	part, err := writer.CreateFormFile("file", "audio.ogg")
	if err != nil {
		sendMessage(ctx, "حدث خطأ داخلي ❌")
		return
	}
	part.Write(audioData)
	
	writer.WriteField("model", "whisper-large-v3")
	err = writer.Close()
	if err != nil {
		sendMessage(ctx, "حدث خطأ داخلي ❌")
		return
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/audio/transcriptions", body)
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء الاتصال بالذكاء الاصطناعي ❌")
		return
	}
	
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+GroqAPIKey)
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء إرسال الصوت للذكاء الاصطناعي ❌")
		return
	}
	defer resp.Body.Close()
	
	respBody, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != 200 {
		sendMessage(ctx, fmt.Sprintf("فشل التحويل (كود %d): %s", resp.StatusCode, string(respBody)))
		return
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		sendMessage(ctx, "فشل في قراءة النتيجة ❌")
		return
	}
	
	finalText := result.Text

	if strings.HasPrefix(ctx.Text, ".دبلج") {
		sendMessage(ctx, "جاري ترجمة المقطع للعربية... 🌍⏳")
		translated, err := gtranslate.TranslateWithParams(
			result.Text,
			gtranslate.TranslationParams{
				From: "auto",
				To:   "ar",
			},
		)
		if err == nil {
			finalText = "عفواً الصوت يقول:\n\n" + translated
		} else {
			finalText = "فشلت الترجمة، هذا النص الأصلي:\n\n" + result.Text
		}
	} else {
		finalText = "🎙️ *التفريغ الصوتي:*\n\n" + result.Text
	}

	sendMessage(ctx, finalText)
}
