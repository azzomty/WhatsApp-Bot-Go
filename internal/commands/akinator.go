package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"context"
	
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

type AkiSession struct {
	SessionID string
	Step      int
	ChatID    string
	Player    string
	IsActive  bool
	LastUpdate time.Time
	Guessing bool
}

var activeAkiSessions = struct {
	sync.RWMutex
	m map[string]*AkiSession // map chatID -> session
}{m: make(map[string]*AkiSession)}

const akiDeviceID = "1869b43f-89bb-48d8-bfaf-0653c9466114" // From user intercept

func HandleAkinator(ctx *BotContext) {
	activeAkiSessions.Lock()
	defer activeAkiSessions.Unlock()

	chat := ctx.ChatID.String()

	if sess, exists := activeAkiSessions.m[chat]; exists && sess.IsActive {
		sendMessage(ctx, "يوجد لعبة أكيناتور شغالة في هذا الشات حالياً!\nجاوب على السؤال الأخير أو اكتب (.ايقاف_اكيناتور).")
		return
	}

	sendMessage(ctx, "المارد أكيناتور: جاري تحضير اللعبة، فكر بشخصية في عقلك...")

	// Create new session
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("queue_prio", "0")
	_ = writer.WriteField("partner_id", "413")
	_ = writer.WriteField("device_id", akiDeviceID)
	_ = writer.WriteField("session_prio", "0")
	_ = writer.WriteField("sensitivity_mode", "1")
	_ = writer.WriteField("learning_mode", "0")
	_ = writer.WriteField("origin", "US")
	_ = writer.WriteField("ft_hard_constraint", "ETAT<>'AV'")
	writer.Close()

	req, _ := http.NewRequest("POST", "https://srv17.akinator.com:9641/game/sessions", payload)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "okhttp/5.4.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		sendMessage(ctx, "حدث خطأ في الاتصال بسيرفر أكيناتور.")
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Result struct {
			SessionID       string `json:"sessionId"`
			StepInformation struct {
				Question string `json:"question"`
			} `json:"stepInformation"`
		} `json:"result"`
	}
	
	if err := json.Unmarshal(body, &result); err != nil || result.Result.SessionID == "" {
		fmt.Println("Aki New Session Body:", string(body))
		sendMessage(ctx, "فشل إنشاء جلسة مع أكيناتور. السيرفر معطل مؤقتاً.")
		return
	}

	activeAkiSessions.m[chat] = &AkiSession{
		SessionID: result.Result.SessionID,
		Step:      1,
		ChatID:    chat,
		Player:    ctx.Sender.String(),
		IsActive:  true,
		LastUpdate: time.Now(),
	}

	sendAkiQuestion(ctx, result.Result.StepInformation.Question, 1)
}

func sendAkiImage(ctx *BotContext, imgUrl string, caption string) {
	resp, err := http.Get(imgUrl)
	if err == nil {
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err == nil {
			upResp, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaImage)
			if err == nil {
				imgMsg := &waProto.ImageMessage{
					URL:           proto.String(upResp.URL),
					DirectPath:    proto.String(upResp.DirectPath),
					MediaKey:      upResp.MediaKey,
					Mimetype:      proto.String("image/jpeg"),
					FileEncSHA256: upResp.FileEncSHA256,
					FileSHA256:    upResp.FileSHA256,
					FileLength:    proto.Uint64(uint64(len(data))),
					Caption:       proto.String(caption),
				}
				ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{ImageMessage: imgMsg})
				return
			}
		}
	}
	sendMessage(ctx, caption)
}

func sendAkiQuestion(ctx *BotContext, question string, step int) {
	txt := fmt.Sprintf("*السؤال رقم %d:*\n\n%s\n\n1. نعم\n2. لا\n3. أنا لا أعلم\n4. من الممكن\n5. الظاهر لا", step, question)
	sendMessage(ctx, txt)
}

func HandleAkinatorAnswer(ctx *BotContext) bool {
	activeAkiSessions.Lock()
	sess, exists := activeAkiSessions.m[ctx.ChatID.String()]
	if !exists || !sess.IsActive {
		activeAkiSessions.Unlock()
		return false // Not handled
	}
	activeAkiSessions.Unlock()

	if ctx.Text == ".انسحاب" || ctx.Text == ".ايقاف_اكيناتور" {
		sess.IsActive = false
		sendMessage(ctx, "تم إنهاء لعبة أكيناتور في هذا القروب.")
		return true
	}

	// Only allow the person who started to answer
	if sess.Player != ctx.Sender.String() {
		return false // Ignore
	}

	if sess.Guessing {
		txt := strings.TrimSpace(ctx.Text)
		if txt == "1" || txt == "١" || txt == "نعم" || txt == "صح" {
			sendMessage(ctx, "لقد فزت مرة أخرى! شكراً للعبك معي.")
			activeAkiSessions.Lock()
			sess.IsActive = false
			activeAkiSessions.Unlock()
			return true
		} else if txt == "2" || txt == "٢" || txt == "لا" || txt == "غلط" || txt == "خطأ" {
			// User rejected the guess, continue game
			activeAkiSessions.Lock()
			sess.Guessing = false
			sess.Step++
			currentStep := strconv.Itoa(sess.Step)
			activeAkiSessions.Unlock()

			payload := &bytes.Buffer{}
			writer := multipart.NewWriter(payload)
			_ = writer.WriteField("queue_prio", "0")
			_ = writer.WriteField("step", currentStep)
			writer.Close()

			url := fmt.Sprintf("https://srv17.akinator.com:9641/game/sessions/%s/resume", sess.SessionID)
			req, _ := http.NewRequest("POST", url, payload)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.Header.Set("Accept", "application/json")
			req.Header.Set("User-Agent", "okhttp/5.4.0")

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				sendMessage(ctx, "انقطع الاتصال، حاول الإجابة مرة أخرى.")
				return true
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			var resumeResult struct {
				Result struct {
					HasQuestion int    `json:"hasQuestion"`
					Question    string `json:"question"`
				} `json:"result"`
			}
			json.Unmarshal(body, &resumeResult)

			if resumeResult.Result.Question != "" {
				sendAkiQuestion(ctx, resumeResult.Result.Question, sess.Step)
			} else {
				sendMessage(ctx, "حدث خطأ أثناء محاولة إكمال اللعبة.")
				activeAkiSessions.Lock()
				sess.IsActive = false
				activeAkiSessions.Unlock()
			}
			return true
		} else {
			sendMessage(ctx, "يرجى الإجابة بـ (نعم / 1) إذا كانت الشخصية صحيحة، أو (لا / 2) إذا كانت خاطئة.")
			return true
		}
	}

	ansMap := map[string]string{
		"1": "0", "١": "0", "نعم": "0",
		"2": "1", "٢": "1", "لا": "1",
		"3": "2", "٣": "2", "انا لا اعلم": "2", "لا اعلم": "2",
		"4": "3", "٤": "3", "من الممكن": "3", "ممكن": "3",
		"5": "4", "٥": "4", "الظاهر لا": "4",
	}

	answerCode, ok := ansMap[strings.TrimSpace(ctx.Text)]
	if !ok {
		return false
	}

	activeAkiSessions.Lock()
	defer activeAkiSessions.Unlock()

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("queue_prio", "0")
	_ = writer.WriteField("step", strconv.Itoa(sess.Step))
	_ = writer.WriteField("answer", answerCode)
	writer.Close()

	url := fmt.Sprintf("https://srv17.akinator.com:9641/game/sessions/%s/answer-question", sess.SessionID)
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "okhttp/5.4.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		sendMessage(ctx, "انقطع الاتصال، حاول الإجابة مرة أخرى.")
		return true
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Result struct {
			SessionId       string `json:"sessionId"`
			HasQuestion     int    `json:"hasQuestion"`
			Question        string `json:"question"`
			Progression     string `json:"progression"`
			Trouvitude      float64 `json:"trouvitude"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Aki Answer Body:", string(body))
		sendMessage(ctx, "حدث خطأ غير متوقع في سيرفر أكيناتور.")
		sess.IsActive = true
				sess.Guessing = true
		return true
	}

	// Check if we should guess (trouvitude > 85%)
	if result.Result.Trouvitude > 85.0 {
		guessPayload := &bytes.Buffer{}
		guessWriter := multipart.NewWriter(guessPayload)
		_ = guessWriter.WriteField("queue_prio", "0")
		_ = guessWriter.WriteField("step", strconv.Itoa(sess.Step+1))
		_ = guessWriter.WriteField("size", "1")
		_ = guessWriter.WriteField("add_question_mode", "0")
		_ = guessWriter.WriteField("trappable_user", "0")
		guessWriter.Close()

		guessUrl := fmt.Sprintf("https://srv17.akinator.com:9641/game/sessions/%s/set-and-get-toplist", sess.SessionID)
		guessReq, _ := http.NewRequest("POST", guessUrl, guessPayload)
		guessReq.Header.Set("Content-Type", guessWriter.FormDataContentType())
		guessReq.Header.Set("Accept", "application/json")
		guessReq.Header.Set("User-Agent", "okhttp/5.4.0")

		guessResp, gErr := client.Do(guessReq)
		if gErr == nil {
			defer guessResp.Body.Close()
			gBody, _ := io.ReadAll(guessResp.Body)
			var gResult struct {
				Result struct {
					Objects []struct {
						Name string `json:"name"`
						Description string `json:"description"`
						AbsolutePicturePath string `json:"absolutePicturePath"`
					} `json:"objects"`
				} `json:"result"`
			}
			if json.Unmarshal(gBody, &gResult) == nil && len(gResult.Result.Objects) > 0 {
				char := gResult.Result.Objects[0]
				txt := fmt.Sprintf("أعتقد أني عرفت الشخصية!\n\n*%s*\n_%s_\n\nهل إجابتي صحيحة؟\n1. نعم\n2. لا", char.Name, char.Description)
				if char.AbsolutePicturePath != "" {
					sendAkiImage(ctx, char.AbsolutePicturePath, txt)
				} else {
					sendMessage(ctx, txt)
				}
				sess.Guessing = true
				return true
			} else {
				fmt.Println("TOPLIST RESPONSE:", string(gBody))
			}
		}
	}

	if result.Result.Question != "" {
		sess.Step++
		sendAkiQuestion(ctx, result.Result.Question, sess.Step)
	} else {
		fmt.Println("Akinator strange response:", string(body))
	}
	return true
}
