package games

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
	"whatsapp-bot/internal/commands"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	LoadGames()
}

type Player struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Hand  []Card `json:"hand,omitempty"`
	Mark  string `json:"mark,omitempty"` // For XO, Connect4 ('1', '2' etc)
	MarkI int    `json:"mark_i,omitempty"`
}

type Card struct {
	Color string `json:"color"`
	Type  string `json:"type"`
	Name  string `json:"name"`
}

type GameState struct {
	Game         string   `json:"game"`
	Status       string   `json:"status"`
	Players      []Player `json:"players"`
	TurnIndex    int      `json:"turnIndex"`
	Board        []string `json:"board,omitempty"` // For XO and Connect4
	BoardMsgKey  string   `json:"boardMsgKey,omitempty"`
	Size         int      `json:"size,omitempty"`
	Target       int      `json:"target,omitempty"` // Bomb
	Word         string   `json:"word,omitempty"`   // Hangman
	Guessed      []string `json:"guessed,omitempty"`
	Lives        int      `json:"lives,omitempty"`
	Deck         []Card   `json:"deck,omitempty"` // Uno
	Discard      []Card   `json:"discard,omitempty"`
	CardsCount   int      `json:"cardsCount,omitempty"`
	CurrentColor string   `json:"currentColor,omitempty"`
}

var (
	gamesFile = "games_save.json"
	gamesMu   sync.Mutex
	gamesMap  = make(map[string]*GameState)
)

func LoadGames() {
	gamesMu.Lock()
	defer gamesMu.Unlock()
	b, err := os.ReadFile(gamesFile)
	if err == nil {
		json.Unmarshal(b, &gamesMap)
	}
}

func SaveGames() {
	gamesMu.Lock()
	defer gamesMu.Unlock()
	b, _ := json.Marshal(gamesMap)
	os.WriteFile(gamesFile, b, 0644)
}

func getGame(chatID string) *GameState {
	gamesMu.Lock()
	defer gamesMu.Unlock()
	return gamesMap[chatID]
}

func setGame(chatID string, g *GameState) {
	gamesMu.Lock()
	defer gamesMu.Unlock()
	gamesMap[chatID] = g
	go SaveGames()
}

func deleteGame(chatID string) {
	gamesMu.Lock()
	defer gamesMu.Unlock()
	delete(gamesMap, chatID)
	go SaveGames()
}

func safeSend(ctx *commands.BotContext, text string) {
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

func safeSendMentions(ctx *commands.BotContext, text string, mentions []string) {
	msg := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(text),
			ContextInfo: &waProto.ContextInfo{
				MentionedJID:  mentions,
				StanzaID:      proto.String(ctx.Event.Info.ID),
				Participant:   proto.String(ctx.Event.Info.Sender.String()),
				QuotedMessage: ctx.Event.Message,
			},
		},
	}
	ctx.Client.SendMessage(context.Background(), ctx.ChatID, msg)
}

func HandleGameCommand(ctx *commands.BotContext) bool {
	text := strings.TrimSpace(ctx.Text)
	if text == "" {
		return false
	}

	chatID := ctx.ChatID.ToNonAD().String()
	senderID := ctx.Sender.ToNonAD().String()

	// Remove lowerText var
	if text == ".menu" || text == ".القائمة" {
		menuText := "القائمة\n\n1. Uno\n2. Tic-Tac-Toe\n3. Connect 4\n4. Hangman\n5. Bomb Defusal\n6. Rock Paper Scissors\n\nللبدء اكتب: .1"
		safeSend(ctx, menuText)
		return true
	}

	g := getGame(chatID)

	if text == ".اوامر" {
		if g == nil {
			safeSend(ctx, "مافي لعبة شغالة.")
			return true
		}
		t := "الاوامر:\n"
		if g.Game == "uno" {
			t += ".دخول\n.بدء\n.اوراق\n.لعب [رقم]\n.سحب\n.لون [اللون]\n.save\n.load"
		}
		if g.Game == "xo" {
			t += ".دخول\n.بدء\n.[رقم المربع]"
		}
		if g.Game == "connect4" {
			t += ".دخول\n.بدء\n.[رقم العمود]"
		}
		if g.Game == "hangman" {
			t += ".دخول\n.بدء\n.حرف [الحرف]"
		}
		if g.Game == "bomb" {
			t += ".خمن [الرقم]"
		}
		if g.Game == "rps" {
			t += ".دخول\n.بدء"
		}
		t += "\n.انهاء"
		safeSend(ctx, t)
		return true
	}

	if text == ".انهاء" {
		if g != nil {
			deleteGame(chatID)
			safeSend(ctx, "تم انهاء اللعبة")
		}
		return true
	}

	if text == ".save" {
		if g != nil {
			SaveGames()
			safeSend(ctx, "تم حفظ اللعبة الحالية.")
		}
		return true
	}

	if text == ".load" {
		LoadGames()
		if getGame(chatID) != nil {
			safeSend(ctx, "تم استعادة اللعبة المحفوظة.")
		} else {
			safeSend(ctx, "لا توجد لعبة محفوظة.")
		}
		return true
	}

	if strings.HasPrefix(text, ".1") {
		if g != nil {
			safeSend(ctx, "في لعبة شغالة اكتب .انهاء")
			return true
		}
		cardsCount := 7
		parts := strings.Split(text, " ")
		if len(parts) > 1 {
			if c, err := strconv.Atoi(parts[1]); err == nil {
				cardsCount = c
			}
		}
		setGame(chatID, &GameState{Game: "uno", Status: "waiting", Deck: createDeck(), CardsCount: cardsCount, TurnIndex: 0})
		safeSend(ctx, "Uno: اكتب .دخول ثم .بدء")
		return true
	}
	if strings.HasPrefix(text, ".2") {
		if g != nil {
			safeSend(ctx, "في لعبة شغالة اكتب .انهاء")
			return true
		}
		setGame(chatID, &GameState{Game: "xo", Status: "waiting", Size: 3, Board: make([]string, 9), TurnIndex: 0})
		safeSend(ctx, "Tic-Tac-Toe: اكتب .دخول ثم .بدء")
		return true
	}
	if strings.HasPrefix(text, ".3") {
		if g != nil {
			safeSend(ctx, "في لعبة شغالة اكتب .انهاء")
			return true
		}
		board := make([]string, 42)
		for i := range board {
			board[i] = "0"
		}
		setGame(chatID, &GameState{Game: "connect4", Status: "waiting", Board: board, TurnIndex: 0})
		safeSend(ctx, "Connect 4: اكتب .دخول ثم .بدء")
		return true
	}
	if strings.HasPrefix(text, ".4") {
		if g != nil {
			safeSend(ctx, "في لعبة شغالة اكتب .انهاء")
			return true
		}
		parts := strings.Split(text, " ")
		lives := 5
		if len(parts) > 1 {
			if l, err := strconv.Atoi(parts[1]); err == nil {
				lives = l
			}
		}
		wordLen := 0
		if len(parts) > 2 {
			if w, err := strconv.Atoi(parts[2]); err == nil {
				wordLen = w
			}
		}
		words := []string{"تفاحة", "بطيخ", "سيارة", "كتاب", "قمر", "شمس"} // Default words
		if wordLen > 0 {
			var filtered []string
			for _, w := range words {
				if len([]rune(w)) == wordLen {
					filtered = append(filtered, w)
				}
			}
			if len(filtered) > 0 {
				words = filtered
			}
		}
		word := words[rand.Intn(len(words))]
		setGame(chatID, &GameState{Game: "hangman", Status: "waiting", Word: word, Lives: lives, TurnIndex: 0})
		safeSend(ctx, "Hangman: اكتب .دخول ثم .بدء")
		return true
	}
	if strings.HasPrefix(text, ".5") {
		if g != nil {
			safeSend(ctx, "في لعبة شغالة اكتب .انهاء")
			return true
		}
		setGame(chatID, &GameState{Game: "bomb", Status: "playing", Target: rand.Intn(1000) + 1})
		safeSend(ctx, "Bomb Defusal: تم اخفاء الرقم من 1 الى 1000. اكتب .خمن [رقم]")
		return true
	}
	if strings.HasPrefix(text, ".6") {
		if g != nil {
			safeSend(ctx, "في لعبة شغالة اكتب .انهاء")
			return true
		}
		setGame(chatID, &GameState{Game: "rps", Status: "waiting"})
		safeSend(ctx, "Rock Paper Scissors (حظ عشوائي):\nعشان تدخل اكتب .دخول\nعشان تبدء اللعبة اكتب .بدء")
		return true
	}

	if g == nil {
		return false // No active game, not a game command
	}

	if text == ".دخول" {
		if g.Status != "waiting" {
			return true
		}
		for _, p := range g.Players {
			if p.ID == senderID {
				safeSend(ctx, "انت داخل اللعبة")
				return true
			}
		}
		if (g.Game == "xo" || g.Game == "connect4" || g.Game == "rps") && len(g.Players) >= 2 {
			safeSend(ctx, "اكتمل العدد")
			return true
		}
		name := ctx.Event.Info.PushName
		if name == "" {
			name = "لاعب"
		}
		g.Players = append(g.Players, Player{ID: senderID, Name: name})
		SaveGames()
		safeSend(ctx, fmt.Sprintf("تم دخول %s\nالعدد الان: %d", name, len(g.Players)))
		return true
	}

	if strings.HasPrefix(text, ".يبدا ") {
		if g.Status != "waiting" {
			return true
		}
		pIndex := -1
		parts := strings.Split(text, " ")
		if len(parts) > 1 {
			if num, err := strconv.Atoi(parts[1]); err == nil && num >= 1 && num <= len(g.Players) {
				pIndex = num - 1
			}
		}
		if pIndex == -1 {
			safeSend(ctx, "اللاعب غير موجود. اكتب رقمه (مثال: .يبدا 2)")
			return true
		}
		g.TurnIndex = pIndex
		SaveGames()
		safeSendMentions(ctx, fmt.Sprintf("تم تحديد البداية للاعب @%s", strings.Split(g.Players[pIndex].ID, "@")[0]), []string{g.Players[pIndex].ID})
		return true
	}

	if text == ".بدء" {
		if g.Status != "waiting" {
			return true
		}

		if g.Game == "xo" {
			if len(g.Players) != 2 {
				safeSend(ctx, "تحتاج لاعبين 2")
				return true
			}
			g.Status = "playing"
			g.Players[0].Mark = "X"
			g.Players[1].Mark = "O"
			safeSend(ctx, getXoBoard(g))
			SaveGames()
			return true
		}
		if g.Game == "connect4" {
			if len(g.Players) != 2 {
				safeSend(ctx, "تحتاج لاعبين 2")
				return true
			}
			g.Status = "playing"
			g.Players[0].MarkI = 1
			g.Players[1].MarkI = 2
			safeSend(ctx, getConnect4Board(g))
			SaveGames()
			return true
		}
		if g.Game == "hangman" {
			if len(g.Players) < 1 {
				safeSend(ctx, "تحتاج لاعبين")
				return true
			}
			g.Status = "playing"
			turnID := g.Players[g.TurnIndex].ID
			disp := ""
			for range []rune(g.Word) {
				disp += "_ "
			}
			txt := fmt.Sprintf("الكلمة: %s\nالقلوب: %s\nدور: @%s", disp, strings.Repeat("❤️", g.Lives), strings.Split(turnID, "@")[0])
			safeSendMentions(ctx, txt, []string{turnID})
			SaveGames()
			return true
		}
		if g.Game == "rps" {
			if len(g.Players) != 2 {
				safeSend(ctx, "تحتاج لاعبين 2")
				return true
			}
			choices := []string{"حجر ✊", "ورقة ✋", "مقص ✌️"}
			c1 := choices[rand.Intn(len(choices))]
			c2 := choices[rand.Intn(len(choices))]
			p1, p2 := g.Players[0], g.Players[1]
			res := ""
			if c1 == c2 {
				res = "تعادل!"
			} else if (strings.Contains(c1, "حجر") && strings.Contains(c2, "مقص")) || (strings.Contains(c1, "ورقة") && strings.Contains(c2, "حجر")) || (strings.Contains(c1, "مقص") && strings.Contains(c2, "ورقة")) {
				res = "فاز " + p1.Name + "! 🎉"
			} else {
				res = "فاز " + p2.Name + "! 🎉"
			}
			safeSend(ctx, fmt.Sprintf("%s طلع: %s\n%s طلع: %s\n\nالنتيجة: %s", p1.Name, c1, p2.Name, c2, res))
			deleteGame(chatID)
			return true
		}
		if g.Game == "uno" {
			if len(g.Players) < 2 {
				safeSend(ctx, "تحتاج لاعبين على الاقل")
				return true
			}
			g.Status = "playing"
			maxCards := 100 / len(g.Players)
			if g.CardsCount > maxCards {
				g.CardsCount = maxCards
			}
			for i := range g.Players {
				for j := 0; j < g.CardsCount; j++ {
					g.Players[i].Hand = append(g.Players[i].Hand, popCard(&g.Deck))
				}
			}
			firstCard := popCard(&g.Deck)
			for firstCard.Color == "اسود" {
				g.Deck = append([]Card{firstCard}, g.Deck...) // unshift
				firstCard = popCard(&g.Deck)
			}
			g.Discard = append(g.Discard, firstCard)
			g.CurrentColor = firstCard.Color
			turnID := g.Players[g.TurnIndex].ID
			safeSendMentions(ctx, fmt.Sprintf("تم البدا\nالورقة الي بالساحة: %s\nدور الاعب: @%s", firstCard.Name, strings.Split(turnID, "@")[0]), []string{turnID})
			for _, p := range g.Players {
				sendHand(ctx, &p)
			}
			SaveGames()
			return true
		}
	}

	if g.Game == "xo" && g.Status == "playing" && strings.HasPrefix(text, ".") {
		move, err := strconv.Atoi(text[1:])
		if err == nil && move >= 1 && move <= 9 {
			turnID := g.Players[g.TurnIndex].ID
			if senderID != turnID {
				safeSend(ctx, "مو دورك")
				return true
			}
			idx := move - 1
			if g.Board[idx] != "" {
				safeSend(ctx, "المكان محجوز")
				return true
			}
			g.Board[idx] = g.Players[g.TurnIndex].Mark
			winner := checkXoWin(g.Board)
			isDraw := true
			for _, c := range g.Board {
				if c == "" {
					isDraw = false
					break
				}
			}
			if winner != "" || isDraw {
				wText := "تعادل!"
				if winner != "" {
					wText = "فاز " + g.Players[g.TurnIndex].Name + "!"
				}
				safeSend(ctx, getXoBoard(g)+"\n"+wText)
				deleteGame(chatID)
				return true
			}
			g.TurnIndex = 1 - g.TurnIndex
			safeSend(ctx, getXoBoard(g))
			SaveGames()
			return true
		}
	}

	if g.Game == "connect4" && g.Status == "playing" && strings.HasPrefix(text, ".") {
		col, err := strconv.Atoi(text[1:])
		if err == nil && col >= 1 && col <= 7 {
			turnID := g.Players[g.TurnIndex].ID
			if senderID != turnID {
				safeSend(ctx, "مو دورك")
				return true
			}
			col = col - 1
			placed := false
			for r := 5; r >= 0; r-- {
				if g.Board[r*7+col] == "0" {
					g.Board[r*7+col] = strconv.Itoa(g.Players[g.TurnIndex].MarkI)
					placed = true
					break
				}
			}
			if !placed {
				safeSend(ctx, "العمود ممتلئ")
				return true
			}
			winner := checkConnect4Win(g.Board)
			isDraw := true
			for _, c := range g.Board {
				if c == "0" {
					isDraw = false
					break
				}
			}
			if winner != "" || isDraw {
				wText := "تعادل!"
				if winner != "" {
					wText = "فاز " + g.Players[g.TurnIndex].Name + "!"
				}
				safeSend(ctx, getConnect4Board(g)+"\n"+wText)
				deleteGame(chatID)
				return true
			}
			g.TurnIndex = 1 - g.TurnIndex
			safeSend(ctx, getConnect4Board(g))
			SaveGames()
			return true
		}
	}

	if g.Game == "hangman" && g.Status == "playing" && strings.HasPrefix(text, ".حرف ") {
		turnID := g.Players[g.TurnIndex].ID
		if senderID != turnID {
			safeSend(ctx, "مو دورك")
			return true
		}
		letter := strings.TrimSpace(strings.TrimPrefix(text, ".حرف "))
		if len([]rune(letter)) == 0 {
			return true
		}
		letter = string([]rune(letter)[0])

		for _, gues := range g.Guessed {
			if gues == letter {
				safeSend(ctx, "تم تخمينه سابقا")
				return true
			}
		}
		g.Guessed = append(g.Guessed, letter)

		correct := strings.Contains(g.Word, letter)
		if !correct {
			g.Lives--
		}

		disp := ""
		won := true
		for _, c := range []rune(g.Word) {
			char := string(c)
			found := false
			for _, gues := range g.Guessed {
				if gues == char {
					found = true
					break
				}
			}
			if found {
				disp += char + " "
			} else {
				disp += "_ "
				won = false
			}
		}

		lost := g.Lives <= 0
		g.TurnIndex = (g.TurnIndex + 1) % len(g.Players)
		nextID := g.Players[g.TurnIndex].ID

		boardTxt := "الكلمة: " + disp + "\nالقلوب: " + strings.Repeat("❤️", max(g.Lives, 0))
		if won {
			boardTxt += "\n\nفزت! الكلمة هي " + g.Word
			safeSend(ctx, boardTxt)
			deleteGame(chatID)
		} else if lost {
			boardTxt += "\n\nخسرت! الكلمة كانت " + g.Word
			safeSend(ctx, boardTxt)
			deleteGame(chatID)
		} else {
			boardTxt += fmt.Sprintf("\nدور: @%s", strings.Split(nextID, "@")[0])
			safeSendMentions(ctx, boardTxt, []string{nextID})
			SaveGames()
		}
		return true
	}

	if g.Game == "bomb" && strings.HasPrefix(text, ".خمن ") {
		num, err := strconv.Atoi(strings.TrimPrefix(text, ".خمن "))
		if err != nil {
			return true
		}
		if num == g.Target {
			safeSend(ctx, fmt.Sprintf("فزت، الرقم السري كان %d", g.Target))
			deleteGame(chatID)
		} else if num < g.Target {
			safeSend(ctx, "اكبر")
		} else {
			safeSend(ctx, "اصغر")
		}
		return true
	}

	if g.Game == "uno" && g.Status == "playing" {
		if text == ".اوراق" {
			for _, p := range g.Players {
				if p.ID == senderID {
					sendHand(ctx, &p)
					break
				}
			}
			return true
		}
		if text == ".سحب" {
			turnID := g.Players[g.TurnIndex].ID
			if senderID != turnID {
				safeSend(ctx, "مو دورك")
				return true
			}
			p := &g.Players[g.TurnIndex]
			if len(g.Deck) == 0 {
				top := g.Discard[len(g.Discard)-1]
				g.Deck = g.Discard[:len(g.Discard)-1]
				rand.Shuffle(len(g.Deck), func(i, j int) { g.Deck[i], g.Deck[j] = g.Deck[j], g.Deck[i] })
				g.Discard = []Card{top}
			}
			p.Hand = append(p.Hand, popCard(&g.Deck))
			safeSend(ctx, fmt.Sprintf("سحب %s ورقة", p.Name))
			nextTurn(g)
			nextID := g.Players[g.TurnIndex].ID
			safeSendMentions(ctx, fmt.Sprintf("الورقة الي بالساحة: %s\nدور الاعب: @%s", g.Discard[len(g.Discard)-1].Name, strings.Split(nextID, "@")[0]), []string{nextID})
			sendHand(ctx, p)
			SaveGames()
			return true
		}
		if strings.HasPrefix(text, ".لعب ") {
			turnID := g.Players[g.TurnIndex].ID
			if senderID != turnID {
				safeSend(ctx, "مو دورك")
				return true
			}
			p := &g.Players[g.TurnIndex]
			parts := strings.Split(text, " ")
			if len(parts) < 2 {
				return true
			}
			cardIdx, err := strconv.Atoi(parts[1])
			cardIdx--
			if err != nil || cardIdx < 0 || cardIdx >= len(p.Hand) {
				safeSend(ctx, "رقم الورقة غلط")
				return true
			}

			cardToPlay := p.Hand[cardIdx]
			topCard := g.Discard[len(g.Discard)-1]
			isValid := false
			if cardToPlay.Color == "اسود" {
				isValid = true
			} else if cardToPlay.Color == g.CurrentColor {
				isValid = true
			} else if cardToPlay.Type == topCard.Type && cardToPlay.Type != "تغيير لون" && cardToPlay.Type != "+4" {
				isValid = true
			}

			if !isValid {
				safeSend(ctx, "ما تقدر تلعبها فوق "+topCard.Name)
				return true
			}

			// Remove card from hand
			p.Hand = append(p.Hand[:cardIdx], p.Hand[cardIdx+1:]...)
			g.Discard = append(g.Discard, cardToPlay)
			if cardToPlay.Color != "اسود" {
				g.CurrentColor = cardToPlay.Color
			}

			if len(p.Hand) == 0 {
				safeSend(ctx, "فاز "+p.Name+" باللعبة")
				deleteGame(chatID)
				return true
			}

			if cardToPlay.Type == "عكس" {
				if len(g.Players) == 2 {
					nextTurn(g)
				} else {
					// Reverse players array
					for i, j := 0, len(g.Players)-1; i < j; i, j = i+1, j-1 {
						g.Players[i], g.Players[j] = g.Players[j], g.Players[i]
					}
					// Find new turn index
					for i, pl := range g.Players {
						if pl.ID == senderID {
							g.TurnIndex = i
							break
						}
					}
				}
			}
			if cardToPlay.Type == "تخطي" {
				nextTurn(g)
			}
			if cardToPlay.Type == "+2" {
				nextTurn(g)
				t := &g.Players[g.TurnIndex]
				t.Hand = append(t.Hand, popCard(&g.Deck), popCard(&g.Deck))
				safeSend(ctx, fmt.Sprintf("سحب %s ورقتين", t.Name))
				sendHand(ctx, t)
			}
			if cardToPlay.Color == "اسود" {
				g.Status = "choosing_color"
				SaveGames()
				safeSend(ctx, "اختار لون اكتب: .لون احمر / الخ")
				return true
			}
			if cardToPlay.Type == "+4" {
				nextTurn(g)
				t := &g.Players[g.TurnIndex]
				for i := 0; i < 4; i++ {
					t.Hand = append(t.Hand, popCard(&g.Deck))
				}
				safeSend(ctx, fmt.Sprintf("سحب %s 4 اوراق", t.Name))
				sendHand(ctx, t)
				g.Status = "choosing_color"
				for i, pl := range g.Players {
					if pl.ID == senderID {
						g.TurnIndex = i
						break
					}
				}
				SaveGames()
				safeSend(ctx, "اختار لون اكتب: .لون احمر / الخ")
				return true
			}

			nextTurn(g)
			nextID := g.Players[g.TurnIndex].ID
			safeSendMentions(ctx, fmt.Sprintf("لعب %s: %s\nدور الاعب: @%s", p.Name, cardToPlay.Name, strings.Split(nextID, "@")[0]), []string{nextID})
			if len(p.Hand) == 1 {
				safeSend(ctx, "اونو من "+p.Name)
			}
			SaveGames()
			return true
		}
	}

	if g.Game == "uno" && g.Status == "choosing_color" && strings.HasPrefix(text, ".لون ") {
		turnID := g.Players[g.TurnIndex].ID
		if senderID != turnID {
			safeSend(ctx, "مو دورك")
			return true
		}
		col := strings.TrimSpace(strings.Split(text, " ")[1])
		if strings.Contains(col, "احمر") || strings.Contains(col, "أحمر") {
			g.CurrentColor = "احمر"
		} else if strings.Contains(col, "ازرق") || strings.Contains(col, "أزرق") {
			g.CurrentColor = "ازرق"
		} else if strings.Contains(col, "اخضر") || strings.Contains(col, "أخضر") {
			g.CurrentColor = "اخضر"
		} else if strings.Contains(col, "اصفر") || strings.Contains(col, "أصفر") {
			g.CurrentColor = "اصفر"
		} else {
			safeSend(ctx, "لون غير معروف")
			return true
		}
		g.Status = "playing"
		nextTurn(g)
		nextID := g.Players[g.TurnIndex].ID
		safeSendMentions(ctx, fmt.Sprintf("اللون صار %s\nدور الاعب: @%s", g.CurrentColor, strings.Split(nextID, "@")[0]), []string{nextID})
		SaveGames()
		return true
	}

	// Commands starting with . are often bot commands, so if it's not handled, return false
	// But actually we just return false if it's not a game command so normal bot can handle it
	return false
}

func nextTurn(g *GameState) {
	g.TurnIndex = (g.TurnIndex + 1) % len(g.Players)
}

func popCard(deck *[]Card) Card {
	if len(*deck) == 0 {
		return Card{}
	}
	card := (*deck)[len(*deck)-1]
	*deck = (*deck)[:len(*deck)-1]
	return card
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func sendHand(ctx *commands.BotContext, player *Player) {
	readMore := strings.Repeat(string(rune(8206)), 4001)
	msgText := fmt.Sprintf("اوراق الاعب @%s\n%s\n\n", strings.Split(player.ID, "@")[0], readMore)
	for i, c := range player.Hand {
		msgText += fmt.Sprintf("\u200F%d- %s\n", i+1, c.Name)
	}
	msgText += "\nلعب: .لعب رقم\nسحب: .سحب"
	safeSendMentions(ctx, msgText, []string{player.ID})
}

func createDeck() []Card {
	colors := []string{"احمر", "ازرق", "اخضر", "اصفر"}
	var deck []Card
	for _, c := range colors {
		deck = append(deck, Card{Color: c, Type: "0", Name: "0 " + c})
		for i := 1; i <= 9; i++ {
			deck = append(deck, Card{Color: c, Type: strconv.Itoa(i), Name: strconv.Itoa(i) + " " + c})
			deck = append(deck, Card{Color: c, Type: strconv.Itoa(i), Name: strconv.Itoa(i) + " " + c})
		}
		for i := 0; i < 2; i++ {
			deck = append(deck, Card{Color: c, Type: "تخطي", Name: "تخطي " + c})
			deck = append(deck, Card{Color: c, Type: "عكس", Name: "عكس " + c})
			deck = append(deck, Card{Color: c, Type: "+2", Name: "+2 " + c})
		}
	}
	for i := 0; i < 4; i++ {
		deck = append(deck, Card{Color: "اسود", Type: "تغيير لون", Name: "تغيير لون"})
		deck = append(deck, Card{Color: "اسود", Type: "+4", Name: "+4"})
	}
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	return deck
}

func getEmojiNumber(n int) string {
	emojis := []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣", "🔟"}
	if n >= 1 && n <= 10 {
		return emojis[n-1]
	}
	return fmt.Sprintf("[%d]", n)
}

func getXoBoard(g *GameState) string {
	txt := fmt.Sprintf("دور: @%s\n\n", strings.Split(g.Players[g.TurnIndex].ID, "@")[0])
	for i := 0; i < g.Size; i++ {
		var row []string
		for j := 0; j < g.Size; j++ {
			idx := i*g.Size + j
			if g.Board[idx] == "X" {
				row = append(row, "X")
			} else if g.Board[idx] == "O" {
				row = append(row, "O")
			} else {
				row = append(row, getEmojiNumber(idx+1))
			}
		}
		txt += strings.Join(row, " | ") + "\n"
	}
	return txt
}

func checkXoWin(board []string) string {
	size := 3
	for i := 0; i < size; i++ {
		first := board[i*size]
		if first != "" {
			win := true
			for j := 1; j < size; j++ {
				if board[i*size+j] != first {
					win = false
					break
				}
			}
			if win {
				return first
			}
		}
		first = board[i]
		if first != "" {
			win := true
			for j := 1; j < size; j++ {
				if board[j*size+i] != first {
					win = false
					break
				}
			}
			if win {
				return first
			}
		}
	}
	first := board[0]
	if first != "" {
		win := true
		for i := 1; i < size; i++ {
			if board[i*size+i] != first {
				win = false
				break
			}
		}
		if win {
			return first
		}
	}
	first = board[size-1]
	if first != "" {
		win := true
		for i := 1; i < size; i++ {
			if board[i*size+(size-1-i)] != first {
				win = false
				break
			}
		}
		if win {
			return first
		}
	}
	return ""
}

func getConnect4Board(g *GameState) string {
	txt := fmt.Sprintf("دور: @%s\n\n", strings.Split(g.Players[g.TurnIndex].ID, "@")[0])
	for r := 0; r < 6; r++ {
		rowStr := ""
		for c := 0; c < 7; c++ {
			val := g.Board[r*7+c]
			if val == "1" {
				rowStr += "🟣"
			} else if val == "2" {
				rowStr += "🔵"
			} else {
				rowStr += "⚫"
			}
		}
		txt += rowStr + "\n"
	}
	txt += "1️⃣2️⃣3️⃣4️⃣5️⃣6️⃣7️⃣"
	return txt
}

func checkConnect4Win(board []string) string {
	rows, cols := 6, 7
	check := func(r, c, dr, dc int) string {
		p := board[r*cols+c]
		if p == "0" {
			return ""
		}
		for i := 1; i < 4; i++ {
			nr, nc := r+dr*i, c+dc*i
			if nr < 0 || nr >= rows || nc < 0 || nc >= cols {
				return ""
			}
			if board[nr*cols+nc] != p {
				return ""
			}
		}
		return p
	}
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			win := check(r, c, 0, 1)
			if win == "" {
				win = check(r, c, 1, 0)
			}
			if win == "" {
				win = check(r, c, 1, 1)
			}
			if win == "" {
				win = check(r, c, 1, -1)
			}
			if win != "" {
				return win
			}
		}
	}
	return ""
}
