package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

type GuerrillaMailResponse struct {
	EmailAddr string `json:"email_addr"`
	Alias     string `json:"alias"`
	SIDToken  string `json:"sid_token"`
}

type GuerrillaMailListResponse struct {
	List []struct {
		MailID      int    `json:"mail_id"`
		MailFrom    string `json:"mail_from"`
		MailSubject string `json:"mail_subject"`
		MailExcerpt string `json:"mail_excerpt"`
		MailBody    string `json:"mail_body"`
	} `json:"list"`
}

func HandleTempMail(ctx *BotContext) {
	sendMessage(ctx, "جاري إنشاء إيميلك المؤقت... ")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.guerrillamail.com/ajax.php?f=get_email_address")
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء إنشاء الإيميل، جرب مرة ثانية.")
		return
	}
	defer resp.Body.Close()

	var gmResp GuerrillaMailResponse
	if err := json.NewDecoder(resp.Body).Decode(&gmResp); err != nil || gmResp.EmailAddr == "" {
		sendMessage(ctx, "فشل في توليد الإيميل.")
		return
	}

	msg := fmt.Sprintf("*تم إنشاء إيميلك المؤقت بنجاح!*\n\nالإيميل: `%s`\n\nالإيميل بيكون شغال لمدة 15 دقيقة، وأي رسالة بتوصل عليه (مثل كود تفعيل) راح أحولها لك هنا فوراً!", gmResp.EmailAddr)
	ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String(msg)},
	})

	// Start background polling
	go pollTempMail(ctx.Client, ctx.ChatID, gmResp.SIDToken)
}

func pollTempMail(client *whatsmeow.Client, chatID types.JID, sidToken string) {
	seen := make(map[int]bool)
	endTime := time.Now().Add(15 * time.Minute)
	httpClient := &http.Client{Timeout: 10 * time.Second}

	for time.Now().Before(endTime) {
		time.Sleep(10 * time.Second) // Check every 10 seconds

		resp, err := httpClient.Get(fmt.Sprintf("https://api.guerrillamail.com/ajax.php?f=get_email_list&offset=0&sid_token=%s", sidToken))
		if err != nil {
			continue
		}

		var listResp GuerrillaMailListResponse
		if err := json.NewDecoder(resp.Body).Decode(&listResp); err == nil {
			for _, m := range listResp.List {
				// ID 1 is usually the welcome message
				if m.MailID <= 1 {
					continue
				}

				if !seen[m.MailID] {
					seen[m.MailID] = true
					
					// We only get excerpt in the list, we need to fetch the full body
					contentResp, err := httpClient.Get(fmt.Sprintf("https://api.guerrillamail.com/ajax.php?f=fetch_email&email_id=%d&sid_token=%s", m.MailID, sidToken))
					if err == nil {
						var content struct {
							MailBody string `json:"mail_body"`
						}
						json.NewDecoder(contentResp.Body).Decode(&content)
						contentResp.Body.Close()

						body := content.MailBody
						if body == "" {
							body = m.MailExcerpt
						} else {
							// Strip simple HTML tags if present
							body = strings.ReplaceAll(body, "<br />", "\n")
							body = strings.ReplaceAll(body, "<br>", "\n")
						}

						alertMsg := fmt.Sprintf("*رسالة جديدة وصلت لإيميلك!*\n\n*من:* %s\n*الموضوع:* %s\n\n*الرسالة:*\n%s", m.MailFrom, m.MailSubject, body)
						client.SendMessage(context.Background(), chatID, &waProto.Message{
							ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String(alertMsg)},
						})
					}
				}
			}
		}
		resp.Body.Close()
	}
	
	expiryMsg := "⌛ *انتهت صلاحية إيميلك المؤقت (15 دقيقة).* لو احتجت إيميل ثاني أرسل `.ايميل` من جديد."
	client.SendMessage(context.Background(), chatID, &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String(expiryMsg)},
	})
}
