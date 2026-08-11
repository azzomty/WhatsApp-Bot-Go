package gemini

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/proto"
)

var (
	client       *genai.Client
	CurrentModel = "gemini-3.6-flash"
	chatSessions = make(map[string]*genai.ChatSession)
	geminiAPIKey = os.Getenv("GEMINI_API_KEY")
)

func init() {
	ctx := context.Background()
	c, err := genai.NewClient(ctx, option.WithAPIKey(geminiAPIKey))
	if err != nil {
		log.Printf("Failed to create genai client: %v", err)
	}
	client = c
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
			if client == nil {
				sendMessage(clientWA, chatID, "Gemini client not initialized", msg, stanzaID, participant)
				return true
			}
			chatStr := chatID.String()
			session, ok := chatSessions[chatStr]
			if !ok {
				model := client.GenerativeModel(CurrentModel)
				model.SystemInstruction = &genai.Content{
					Parts: []genai.Part{genai.Text("أنت ذكاء اصطناعي اسمك جيميناي. أجب باختصار وبشكل مفيد.")},
				}
				session = model.StartChat()
				chatSessions[chatStr] = session
			}

			resp, err := session.SendMessage(context.Background(), genai.Text(prompt))
			if err != nil {
				log.Printf("Gemini Error: %v", err)
				delete(chatSessions, chatStr)
				sendMessage(clientWA, chatID, fmt.Sprintf("حدث خطأ في الذكاء الاصطناعي:\n%v", err), msg, stanzaID, participant)
				return true
			}

			if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
				if part, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
					sendMessage(clientWA, chatID, string(part), msg, stanzaID, participant)
				}
			}
		} else if hasGeminiCommand {
			sendMessage(clientWA, chatID, "يرجى كتابة سؤالك.", msg, stanzaID, participant)
		}
		return true
	}
	return false
}

func GeneratePinterestQueryFromImage(imageBytes []byte) (string, error) {
	if client == nil {
		return "", fmt.Errorf("Gemini client not initialized")
	}
	model := client.GenerativeModel("gemini-3.6-flash")
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text("You are an image analyzer. Return ONLY a short English search query (2-4 words) that best describes the main subject of this image for a Pinterest search. Do not include quotes, periods, or any other text.")},
	}
	resp, err := model.GenerateContent(context.Background(), genai.ImageData("jpeg", imageBytes))
	if err != nil {
		return "", err
	}
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if part, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			return strings.TrimSpace(string(part)), nil
		}
	}
	return "", fmt.Errorf("No response from Gemini")
}
