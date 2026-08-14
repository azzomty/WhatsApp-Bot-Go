package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

var GroqAPIKey = os.Getenv("GROQ_API_KEY")

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
	
	writer.WriteField("model", "whisper-large-v3-turbo")
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
	
	if result.Text == "" {
		sendMessage(ctx, "لم يتمكن الذكاء الاصطناعي من فهم أي كلمات في المقطع 🤫")
		return
	}

	sendMessage(ctx, "🎙️ *التفريغ الصوتي:*\n\n"+result.Text)
}
