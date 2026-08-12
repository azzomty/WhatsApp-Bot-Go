package gemini

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

var (
	CurrentModel = "gemini-web"
	geminiCookie1PSID   = "g.a000BgkyVrxC8cwsVYW5dLfLHd-KlWxi4gxc32Uv8-Ee8FKXA51i5udy6dsl2ETs9tvosAI61QACgYKATASARUSFQHGX2MivIgMfl2Oqgtdx_Ae5yGcQBoVAUF8yKrTAw1qZVJgsEXATsPto0cv0076"
	geminiCookie1PSIDTS = "sidts-CjEBPWEu2XHmxY-HfLlcBIfHKBw-4VRrbeyhKEIUv87IgE2p0KtL0uLMwSUi2xWV51NgEAA"
)

func init() {
	// Web API doesn't need init
}

func sendMessage(clientWA *whatsmeow.Client, chatID types.JID, text string, quotedMsg *waProto.Message, stanzaID string, participant string) {
	msg := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(text),
			ContextInfo: &waProto.ContextInfo{
				StanzaID:      proto.String(stanzaID),
				Participant:   proto.String(participant),
				QuotedMessage: quotedMsg,
			},
		},
	}
	clientWA.SendMessage(context.Background(), chatID, msg)
}

func HandleMessage(clientWA *whatsmeow.Client, chatID types.JID, sender types.JID, text string, isGroup bool, isFromMe bool, msg *waProto.Message, stanzaID string, participant string) bool {
	lowerText := strings.ToLower(text)

	if lowerText == ".models" || lowerText == ".model" {
		modelsList := "النماذج المتوفرة:\n1- gemini-3.5-flash-lite (Fastest answers)\n2- gemini-3.6-flash (All-around help)\n3- gemini-3.1-pro\n\nلتغيير النموذج اكتب داش مع الرقم، مثال:\n-2"
		sendMessage(clientWA, chatID, modelsList, msg, stanzaID, participant)
		return true
	} else if lowerText == "-1" || lowerText == "-2" || lowerText == "-3" {
		if lowerText == "-1" {
			CurrentModel = "gemini-3.5-flash-lite"
		} else if lowerText == "-2" {
			CurrentModel = "gemini-3.6-flash"
		} else if lowerText == "-3" {
			CurrentModel = "gemini-3.1-pro"
		}
		sendMessage(clientWA, chatID, "تم تغيير النموذج بنجاح إلى: "+CurrentModel, msg, stanzaID, participant)
		return true
	}

	hasGeminiCommand := strings.Contains(lowerText, ".جيميناي") || strings.Contains(lowerText, "جيميناي.") || strings.Contains(lowerText, ".حيميناي") || strings.Contains(lowerText, "حيميناي.")

	isMentioned := false
	if msg != nil && msg.ExtendedTextMessage != nil && msg.ExtendedTextMessage.ContextInfo != nil {
		for _, jid := range msg.ExtendedTextMessage.ContextInfo.MentionedJID {
			parsed, _ := types.ParseJID(jid)
			if parsed.User == clientWA.Store.ID.User {
				isMentioned = true
				break
			}
		}
	}

	prompt := text
	if hasGeminiCommand {
		prompt = strings.ReplaceAll(prompt, ".جيميناي", "")
		prompt = strings.ReplaceAll(prompt, "جيميناي.", "")
		prompt = strings.ReplaceAll(prompt, ".حيميناي", "")
		prompt = strings.ReplaceAll(prompt, "حيميناي.", "")
	}
	prompt = strings.TrimSpace(prompt)

	if isGroup && !isFromMe {
		if hasGeminiCommand || isMentioned {
			// Do not process Gemini commands for others in groups
			return false
		}
	}

	if hasGeminiCommand || isMentioned {
		if msg != nil && msg.ExtendedTextMessage != nil && msg.ExtendedTextMessage.ContextInfo != nil && msg.ExtendedTextMessage.ContextInfo.QuotedMessage != nil {
			qMsg := msg.ExtendedTextMessage.ContextInfo.QuotedMessage
			qText := qMsg.GetConversation()
			if qText == "" && qMsg.GetExtendedTextMessage() != nil {
				qText = qMsg.GetExtendedTextMessage().GetText()
			}
			if qText != "" {
				prompt += "\n\n[الرسالة المقتبسة]:\n" + qText
			}
		}

		prompt = strings.TrimSpace(prompt)

		if prompt != "" {
			sendMessage(clientWA, chatID, "⏳ جاري التفكير...", msg, stanzaID, participant)

			cmd := exec.Command("./gemini_cli", geminiCookie1PSID, geminiCookie1PSIDTS, prompt)
			out, err := cmd.CombinedOutput()
			
			if err != nil {
				log.Printf("Gemini Error: %v\nOutput: %s", err, string(out))
				sendMessage(clientWA, chatID, fmt.Sprintf("حدث خطأ في الذكاء الاصطناعي:\n%v", err), msg, stanzaID, participant)
				return true
			}

			response := strings.TrimSpace(string(out))
			if response == "" {
				response = "عذراً، لم أتمكن من توليد رد."
			}
			
			sendMessage(clientWA, chatID, response, msg, stanzaID, participant)

		} else if hasGeminiCommand {
			sendMessage(clientWA, chatID, "يرجى كتابة سؤالك.", msg, stanzaID, participant)
		}
		return true
	}
	return false
}
