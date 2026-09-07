package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type SubDeck struct {
	Author string
	Info   string
	Cards  []string
}

type MDMSession struct {
	State      int
	Decks      []MDMDeck
	TargetDeck MDMDeck
	SubDecks   []SubDeck
	LastUpdate time.Time
}

type MDMDeck struct {
	Name string
	URL  string
}

var mdmSessions = make(map[string]*MDMSession)

func HandleMDMCommand(ctx *BotContext) bool {
	text := strings.TrimSpace(ctx.Text)
	chatID := ctx.ChatID.String()
	senderID := ctx.Sender.String()
	sessionKey := chatID + "_" + senderID

	if strings.HasPrefix(text, ".ميتا ماستر") || strings.HasPrefix(text, ".ميتا") {
		mdmSessions[sessionKey] = &MDMSession{
			State:      0,
			LastUpdate: time.Now(),
		}

		msg := "*مرحباً بك في قائمة Master Duel Meta!*\n\n" +
			"اختر أحد الأوامر التالية عبر كتابة رقمه:\n" +
			"1. Tier List (قائمة التير)\n" +
			"2. Top Decks (أفضل المجموعات الحالية)"

		sendMessage(ctx, msg)
		return true
	}

	if strings.HasPrefix(text, ".card") {
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			sendMessage(ctx, "يرجى كتابة اسم البطاقة.\nمثال: .card Ash Blossom")
			return true
		}
		query := parts[1]
		go fetchCard(ctx, query)
		return true
	}

	session, exists := mdmSessions[sessionKey]
	if !exists {
		return false
	}

	session.LastUpdate = time.Now()

	num, err := strconv.Atoi(text)
	if err != nil {
		return false
	}

	if session.State == 0 {
		if num == 1 {
			sendMessage(ctx, "جاري جلب Tier List...")
			go fetchTierList(ctx, sessionKey, false)
			return true
		} else if num == 2 {
			sendMessage(ctx, "جاري جلب Top Decks...")
			go fetchTierList(ctx, sessionKey, true)
			return true
		} else {
			sendMessage(ctx, "خيار غير صحيح. يرجى اختيار 1 أو 2.")
			return true
		}
	} else if session.State == 1 || session.State == 2 {
		if num < 1 || num > len(session.Decks) {
			sendMessage(ctx, "رقم المجموعة غير صحيح.")
			return true
		}
		session.TargetDeck = session.Decks[num-1]
		sendMessage(ctx, "جاري جلب المجموعات الفرعية لـ "+session.TargetDeck.Name+"...")
		go fetchDeckDetails(ctx, sessionKey)
		return true
	} else if session.State == 3 {
		if num < 1 || num > len(session.SubDecks) {
			sendMessage(ctx, "رقم المجموعة الفرعية غير صحيح.")
			return true
		}
		sd := session.SubDecks[num-1]
		msg := "*مجموعة " + sd.Author + " - " + session.TargetDeck.Name + "*\n"
		if sd.Info != "" {
			msg += "التصنيف: " + sd.Info + "\n"
		}
		msg += "\n"
		for _, c := range sd.Cards {
			msg += "- " + c + "\n"
		}
		sendMessage(ctx, msg)
		delete(mdmSessions, sessionKey)
		return true
	}

	return false
}

func fetchTierList(ctx *BotContext, sessionKey string, topDecksOnly bool) {
	req, _ := http.NewRequest("GET", "https://www.masterduelmeta.com/tier-list", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء الاتصال بالموقع.")
		return
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء قراءة البيانات.")
		return
	}

	session := mdmSessions[sessionKey]
	session.Decks = []MDMDeck{}

	msg := ""
	if topDecksOnly {
		msg = "*Top Decks (Trending)*\n\n"
	} else {
		msg = "*Tier List*\n\n"
	}

	counter := 1
	doc.Find(".tier-img-container").Each(func(i int, s *goquery.Selection) {
		tierImg := s.Find("img").AttrOr("alt", "Unknown Tier")

		isTrending := strings.Contains(strings.ToLower(tierImg), "trending")

		if topDecksOnly && !isTrending {
			return
		}
		if !topDecksOnly && isTrending {
			return
		}

		msg += fmt.Sprintf("==== *%s* ====\n", tierImg)

		columns := s.NextAllFiltered("div.columns").First()
		columns.Find("a.img-button").Each(func(j int, a *goquery.Selection) {
			href, _ := a.Attr("href")
			label := strings.TrimSpace(a.Find(".label").Text())
			if label == "" {
				label = a.Find("img").AttrOr("alt", "Unknown Deck")
			}

			deck := MDMDeck{
				Name: label,
				URL:  "https://www.masterduelmeta.com" + href,
			}
			session.Decks = append(session.Decks, deck)
			msg += fmt.Sprintf("%d. %s\n", counter, label)
			counter++
		})
		msg += "\n"
	})

	if len(session.Decks) == 0 {
		sendMessage(ctx, "لم يتم العثور على مجموعات.")
		delete(mdmSessions, sessionKey)
		return
	}

	msg += "أرسل رقم المجموعة لعرض تفاصيلها."
	if topDecksOnly {
		session.State = 2
	} else {
		session.State = 1
	}
	sendMessage(ctx, msg)
}

func fetchDeckDetails(ctx *BotContext, sessionKey string) {
	session := mdmSessions[sessionKey]

	// Fetch from Top Decks API (last 30 days)
	apiURL := "https://www.masterduelmeta.com/api/v1/top-decks?deck=" + url.QueryEscape(session.TargetDeck.Name)
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	res, err := http.DefaultClient.Do(req)

	session.SubDecks = []SubDeck{}

	if err == nil {
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)

		var data []struct {
			Author struct {
				Username string `json:"username"`
			} `json:"author"`
			DeckType struct {
				Name string `json:"name"`
			} `json:"deckType"`
			RankedType struct {
				ShortName string `json:"shortName"`
			} `json:"rankedType"`
			TournamentNumber string `json:"tournamentNumber"`
			Main             []struct {
				Card struct {
					Name string `json:"name"`
				} `json:"card"`
				Amount int `json:"amount"`
			} `json:"main"`
			Extra []struct {
				Card struct {
					Name string `json:"name"`
				} `json:"card"`
				Amount int `json:"amount"`
			} `json:"extra"`
		}

		if json.Unmarshal(body, &data) == nil {
			for _, d := range data {
				if strings.EqualFold(d.DeckType.Name, session.TargetDeck.Name) {
					sd := SubDeck{Author: d.Author.Username}
					if d.TournamentNumber != "" {
						sd.Info = "Tournament " + d.TournamentNumber
					} else if d.RankedType.ShortName != "" {
						sd.Info = d.RankedType.ShortName
					}

					for _, c := range d.Main {
						sd.Cards = append(sd.Cards, fmt.Sprintf("%dx %s", c.Amount, c.Card.Name))
					}
					if len(d.Extra) > 0 {
						sd.Cards = append(sd.Cards, "--- Extra Deck ---")
						for _, c := range d.Extra {
							sd.Cards = append(sd.Cards, fmt.Sprintf("%dx %s", c.Amount, c.Card.Name))
						}
					}

					session.SubDecks = append(session.SubDecks, sd)
					if len(session.SubDecks) >= 15 {
						break
					}
				}
			}
		}
	}

	// Fallback to web scraping if API returns nothing or fails
	if len(session.SubDecks) == 0 {
		req2, _ := http.NewRequest("GET", session.TargetDeck.URL, nil)
		req2.Header.Set("User-Agent", "Mozilla/5.0")
		res2, err2 := http.DefaultClient.Do(req2)
		if err2 == nil {
			defer res2.Body.Close()
			doc, err3 := goquery.NewDocumentFromReader(res2.Body)
			if err3 == nil {
				container := doc.Find(".deck-container").First()
				if container.Length() > 0 {
					sd := SubDeck{Author: "Sample Deck", Info: "Main page list"}
					var uniqueCards []string
					container.Find("img.card-img").Each(func(i int, s *goquery.Selection) {
						alt := s.AttrOr("alt", "")
						if alt != "" && alt != "cp-ur" && alt != "cp-sr" && alt != "cp-r" && alt != "cp-n" {
							if !contains(uniqueCards, alt) {
								uniqueCards = append(uniqueCards, alt)
								sd.Cards = append(sd.Cards, "1x "+alt) // Assuming 1x since HTML grouping loses amounts easily without deeper parsing
							}
						}
					})
					if len(sd.Cards) > 0 {
						session.SubDecks = append(session.SubDecks, sd)
					}
				}
			}
		}
	}

	if len(session.SubDecks) == 0 {
		sendMessage(ctx, "لم يتم العثور على تفاصيل هذه المجموعة.")
		delete(mdmSessions, sessionKey)
		return
	}

	msg := "*اختر المجموعة الفرعية لـ " + session.TargetDeck.Name + ":*\n\n"
	for i, sd := range session.SubDecks {
		infoStr := ""
		if sd.Info != "" {
			infoStr = " (" + sd.Info + ")"
		}
		msg += fmt.Sprintf("%d. %s%s\n", i+1, sd.Author, infoStr)
	}
	session.State = 3
	sendMessage(ctx, msg)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func fetchCard(ctx *BotContext, query string) {
	apiURL := "https://db.ygoprodeck.com/api/v7/cardinfo.php?fname=" + url.QueryEscape(query)
	res, err := http.Get(apiURL)
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء البحث عن البطاقة.")
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		sendMessage(ctx, "لم يتم العثور على البطاقة.")
		return
	}

	body, _ := io.ReadAll(res.Body)
	var data struct {
		Data []struct {
			Name       string `json:"name"`
			Type       string `json:"type"`
			Desc       string `json:"desc"`
			Atk        int    `json:"atk"`
			Def        int    `json:"def"`
			Level      int    `json:"level"`
			Race       string `json:"race"`
			Attribute  string `json:"attribute"`
			CardImages []struct {
				ImageURL string `json:"image_url"`
			} `json:"card_images"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &data); err != nil || len(data.Data) == 0 {
		sendMessage(ctx, "خطأ في قراءة بيانات البطاقة.")
		return
	}

	card := data.Data[0]
	caption := fmt.Sprintf("*%s*\n\n", card.Name)
	caption += fmt.Sprintf("Type: %s\n", card.Type)
	if card.Attribute != "" {
		caption += fmt.Sprintf("Attribute: %s | Race: %s\n", card.Attribute, card.Race)
	}
	if card.Level > 0 {
		caption += fmt.Sprintf("Level/Rank: %d\n", card.Level)
	}
	if strings.Contains(card.Type, "Monster") {
		caption += fmt.Sprintf("ATK: %d / DEF: %d\n", card.Atk, card.Def)
	}
	caption += fmt.Sprintf("\nDescription:\n%s", card.Desc)

	if len(card.CardImages) > 0 {
		imgURL := card.CardImages[0].ImageURL
		err = sendImageFromURL(ctx, imgURL, caption)
		if err != nil {
			sendMessage(ctx, caption)
		}
	} else {
		sendMessage(ctx, caption)
	}
}
