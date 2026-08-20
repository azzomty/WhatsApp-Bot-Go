package gemini

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

var (
	CurrentModel        = "gemini-3.6-flash"
	IsExtended          = false
	geminiCookie1PSID   = "g.a000BgkyVrxC8cwsVYW5dLfLHd-KlWxi4gxc32Uv8-Ee8FKXA51i5udy6dsl2ETs9tvosAI61QACgYKATASARUSFQHGX2MivIgMfl2Oqgtdx_Ae5yGcQBoVAUF8yKrTAw1qZVJgsEXATsPto0cv0076"
	geminiCookie1PSIDTS = "sidts-CjEBPWEu2Qdx4paX5rcNpwudpzTgTl1EKLjkScswFeHZj8FjKmm0dgbCVl54AbeKfwDnEAA"
	geminiCookie1PSIDCC = "AKEyXzXKC-oKhMpOZ8e3mYtnJ-ShXiMOZLMrrkDJJ49zPPaJcPJm5Zoy3sPfmZ6sPFG3KkEQfZRm"
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
	} else if lowerText == "/extended" {
		IsExtended = true
		sendMessage(clientWA, chatID, "تم تفعيل الوضع الموسع (Extended Mode) بنجاح ", msg, stanzaID, participant)
		return true
	} else if lowerText == "/unextended" {
		IsExtended = false
		sendMessage(clientWA, chatID, "تم إيقاف الوضع الموسع (Extended Mode) بنجاح ", msg, stanzaID, participant)
		return true
	} else if lowerText == ".new chat" {
		sendMessage(clientWA, chatID, "تم بدء محادثة جديدة مع جيميناي", msg, stanzaID, participant)
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
			sendMessage(clientWA, chatID, "جاري التفكير...", msg, stanzaID, participant)

			finalPrompt := prompt
			if CurrentModel != "" {
				finalPrompt = "[System Note: Act as " + CurrentModel + ".]\n" + finalPrompt
			}
			if IsExtended {
				finalPrompt = "[System Note: Please provide a highly detailed, comprehensive, and extended answer.]\n" + finalPrompt
			}

			cmd := exec.Command("./gemini_cli", geminiCookie1PSID, geminiCookie1PSIDTS, geminiCookie1PSIDCC, finalPrompt)
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

			parts := strings.Split(response, "---MEDIA---")
			textContent := strings.TrimSpace(parts[0])

			if textContent != "" {
				sendMessage(clientWA, chatID, textContent, msg, stanzaID, participant)
			} else {
				sendMessage(clientWA, chatID, "تم توليد الوسائط:", msg, stanzaID, participant)
			}

			if len(parts) > 1 {
				mediaUrls := strings.Split(strings.TrimSpace(parts[1]), "\n")
				for _, mUrl := range mediaUrls {
					mUrl = strings.TrimSpace(mUrl)
					if mUrl == "" {
						continue
					}

					// Download the image
					resp, err := http.Get(mUrl)
					if err == nil {
						imgData, _ := ioutil.ReadAll(resp.Body)
						resp.Body.Close()

						if len(imgData) > 0 {
							uploadResp, err := clientWA.Upload(context.Background(), imgData, whatsmeow.MediaImage)
							if err == nil {
								imgMsg := &waProto.ImageMessage{
									URL:           proto.String(uploadResp.URL),
									DirectPath:    proto.String(uploadResp.DirectPath),
									MediaKey:      uploadResp.MediaKey,
									Mimetype:      proto.String("image/jpeg"),
									FileEncSHA256: uploadResp.FileEncSHA256,
									FileSHA256:    uploadResp.FileSHA256,
									FileLength:    proto.Uint64(uint64(len(imgData))),
									ContextInfo: &waProto.ContextInfo{
										StanzaID:      proto.String(stanzaID),
										Participant:   proto.String(participant),
										QuotedMessage: msg,
									},
								}
								clientWA.SendMessage(context.Background(), chatID, &waProto.Message{
									ImageMessage: imgMsg,
								})
							}
						}
					}
				}
			}

		} else if hasGeminiCommand {
			sendMessage(clientWA, chatID, "يرجى كتابة طلبك بعد الأمر.", msg, stanzaID, participant)
		}
		return true
	}
	return false
}
