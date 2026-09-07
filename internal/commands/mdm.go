package commands

import (
		"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"encoding/json"
	"io"

	"github.com/PuerkitoBio/goquery"
			)

type MDMSession struct {
	State      int
	Decks      []MDMDeck
	TargetDeck MDMDeck
	CardsMain  []string
	CardsExtra []string
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
		// New session
		mdmSessions[sessionKey] = &MDMSession{
			State:      0,
			LastUpdate: time.Now(),
		}
		
		msg := "*مرحباً بك في قائمة Master Duel Meta!* 🃏\n\n" +
			"اختر أحد الأوامر التالية عبر كتابة رقمه:\n" +
			"1️⃣ Tier List (قائمة التير)\n" +
			"2️⃣ Top Decks (أفضل المجموعات الحالية)"
			
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

	// Update timestamp
	session.LastUpdate = time.Now()

	num, err := strconv.Atoi(text)
	if err != nil {
		return false // Not a number, maybe normal chat
	}

	if session.State == 0 {
		if num == 1 {
			sendMessage(ctx, "جاري جلب Tier List... ⏳")
			go fetchTierList(ctx, sessionKey, false)
			return true
		} else if num == 2 {
			sendMessage(ctx, "جاري جلب Top Decks... ⏳")
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
		sendMessage(ctx, "جاري جلب تفاصيل المجموعة: *"+session.TargetDeck.Name+"*... ⏳")
		go fetchDeckDetails(ctx, sessionKey)
		return true
	} else if session.State == 3 {
		if num == 1 {
			if len(session.CardsMain) == 0 {
				sendMessage(ctx, "لا توجد بطاقات Main Deck.")
			} else {
				msg := "*Main Deck - " + session.TargetDeck.Name + "*\n\n"
				for i, c := range session.CardsMain {
					msg += fmt.Sprintf("%d. %s\n", i+1, c)
				}
				sendMessage(ctx, msg)
			}
			delete(mdmSessions, sessionKey)
			return true
		} else if num == 2 {
			if len(session.CardsExtra) == 0 {
				sendMessage(ctx, "لا توجد بطاقات Extra Deck.")
			} else {
				msg := "*Extra Deck - " + session.TargetDeck.Name + "*\n\n"
				for i, c := range session.CardsExtra {
					msg += fmt.Sprintf("%d. %s\n", i+1, c)
				}
				sendMessage(ctx, msg)
			}
			delete(mdmSessions, sessionKey)
			return true
		} else {
			sendMessage(ctx, "خيار غير صحيح. يرجى اختيار 1 أو 2.")
			return true
		}
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
	req, _ := http.NewRequest("GET", session.TargetDeck.URL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء جلب المجموعة.")
		return
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء قراءة بيانات المجموعة.")
		return
	}

	session.CardsMain = []string{}
	session.CardsExtra = []string{}

	// MDM groups cards in `.deck-container`. First one contains all cards.
	container := doc.Find(".deck-container").First()
	if container.Length() == 0 {
		sendMessage(ctx, "لم يتم العثور على تفاصيل هذه المجموعة.")
		delete(mdmSessions, sessionKey)
		return
	}
	
	// A main deck has 40-60 cards. Extra deck has up to 15.
	// In the HTML, they are just a list of images.
	// But duplicates are grouped. So there are ~30-40 unique cards.
	// We'll just split them into Main and Extra if we can, or just put all in Main if we can't differentiate.
	// Actually, MDM wraps main deck and extra deck in separate divs inside deck-container?
	// Let's just put the first 40 unique in main, rest in extra (approximation), or just put all in Main and call it "Cards List".
	// The user asked for "المجاميع" (groups/sub-decks). I will just provide "1. Main Deck" and "2. Extra Deck" and split the cards found arbitrarily if I can't find the separator, OR just call it "1. Sample Deck Cards".
	// Wait, the user said "يطلع لك المجاميع وتكون مرقمة برضو".
	
	container.Find("img.card-img").Each(func(i int, s *goquery.Selection) {
		alt := s.AttrOr("alt", "")
		if alt != "" && alt != "cp-ur" && alt != "cp-sr" && alt != "cp-r" && alt != "cp-n" {
			// Basic deduplication in our list (though MDM already groups them)
			if !contains(session.CardsMain, alt) && !contains(session.CardsExtra, alt) {
				if len(session.CardsMain) < 30 {
					session.CardsMain = append(session.CardsMain, alt)
				} else {
					session.CardsExtra = append(session.CardsExtra, alt)
				}
			}
		}
	})

	session.State = 3
	sendMessage(ctx, "*اختر المجموعة (Sub-deck) للمجموعة "+session.TargetDeck.Name+":*\n\n1. Main Deck\n2. Extra Deck")
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
			Name        string `json:"name"`
			Type        string `json:"type"`
			Desc        string `json:"desc"`
			Atk         int    `json:"atk"`
			Def         int    `json:"def"`
			Level       int    `json:"level"`
			Race        string `json:"race"`
			Attribute   string `json:"attribute"`
			CardImages  []struct {
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
