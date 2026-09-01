package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	anslayerToken     = "a458a73d7d556e1f5474a8c50cf1fb17d54314a3"
	anslayerClientID  = "android-app2"
	anslayerClientSec = "7befba6263cc14c90d2f1d6da2c5cf9b251bfbbd"
	anslayerReplyMsg  = ""
	anslayerMonitored = "" // Current episode_id string
	anslayerStopChan  chan struct{}
	anslayerMutex     sync.Mutex

	anslayerUsersFile = "anslayer_users.json"
	anslayerUsers     = make(map[string]bool)
	ansUsersMutex     sync.Mutex
)

type AnslayerAnime struct {
	AnimeID   string `json:"anime_id"`
	AnimeName string `json:"anime_name"`
}

type AnslayerEpisode struct {
	EpisodeID     string `json:"episode_id"`
	EpisodeName   string `json:"episode_name"`
	EpisodeNumber string `json:"episode_number"`
}

type AnslayerSession struct {
	Mode          string
	State         string
	Animes        []AnslayerAnime
	SelectedAnime AnslayerAnime
	Episodes      []AnslayerEpisode
}

var ansSessions = make(map[string]*AnslayerSession)

func init() {
	loadAnslayerUsers()
}

func loadAnslayerUsers() {
	ansUsersMutex.Lock()
	defer ansUsersMutex.Unlock()
	data, err := os.ReadFile(anslayerUsersFile)
	if err == nil {
		json.Unmarshal(data, &anslayerUsers)
	}
}

func saveAnslayerUser(uid string) {
	ansUsersMutex.Lock()
	defer ansUsersMutex.Unlock()
	anslayerUsers[uid] = true
	data, _ := json.Marshal(anslayerUsers)
	os.WriteFile(anslayerUsersFile, data, 0644)
}

func hasAnslayerUser(uid string) bool {
	ansUsersMutex.Lock()
	defer ansUsersMutex.Unlock()
	return anslayerUsers[uid]
}

func reqHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+anslayerToken)
	req.Header.Set("Client-Id", anslayerClientID)
	req.Header.Set("Client-Secret", anslayerClientSec)
	req.Header.Set("User-Agent", "okhttp/3.12.13")
}

// HandleAnslayerCommand processes .انمي سلاير ...
func HandleAnslayerCommand(ctx *BotContext, mode string) {
	fullParts := strings.SplitN(ctx.Text, " ", 2)
	
	var query string
	if mode == "marketing" {
		if len(fullParts) < 2 {
			sendMessage(ctx, "يرجى كتابة اسم الأنمي، مثلا:\n.انمي سلاير ون بيس\nأو لحفظ رسالة النشر:\n.انمي سلاير نشر رسالتي هنا")
			return
		}
		afterAnmi := strings.TrimSpace(fullParts[1])
		if !strings.HasPrefix(afterAnmi, "سلاير") {
			return
		}
		afterSlayer := strings.TrimSpace(strings.TrimPrefix(afterAnmi, "سلاير"))
		if afterSlayer == "" {
			sendMessage(ctx, "يرجى كتابة اسم الأنمي، مثلا:\n.انمي سلاير ون بيس")
			return
		}
		if strings.HasPrefix(afterSlayer, "مفضلة") {
			anslayerMutex.Lock()
			msg := anslayerReplyMsg
			anslayerMutex.Unlock()
			if msg == "" {
				sendMessage(ctx, "يرجى أولاً حفظ قالب الرد باستخدام أمر:\n.انمي سلاير نشر <رسالتك>")
				return
			}
			startFavMarketing(ctx, msg)
			return
		}

		if strings.HasPrefix(afterSlayer, "نشر") {
			replyMsg := strings.TrimSpace(strings.TrimPrefix(afterSlayer, "نشر"))
			if replyMsg == "" {
				sendMessage(ctx, "يرجى كتابة الرسالة بعد كلمة نشر.")
				return
			}
			anslayerMutex.Lock()
			anslayerReplyMsg = replyMsg
			anslayerMutex.Unlock()
			sendMessage(ctx, "✅ تم حفظ قالب الرد بنجاح!\nالرسالة:\n"+anslayerReplyMsg)
			return
		}
		query = afterSlayer
	} else {
		if len(fullParts) < 2 {
			sendMessage(ctx, "يرجى كتابة اسم الأنمي، مثلا:\n.انمي ون بيس")
			return
		}
		query = fullParts[1]
	}
	searchParams := map[string]interface{}{
		"_offset":   0,
		"_limit":    30,
		"_order_by": "latest_first",
		"list_type": "filter",
		"anime_name": query,
		"just_info": "Yes",
	}
	b, _ := json.Marshal(searchParams)
	u := "https://anslayer.com/anime/public/animes/get-published-animes?json=" + url.QueryEscape(string(b))
	
	req, _ := http.NewRequest("GET", u, nil)
	reqHeaders(req)
	
	sendMessage(ctx, "جاري البحث في انمي سلاير...")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		sendMessage(ctx, "حدث خطأ في الاتصال.")
		return
	}
	defer resp.Body.Close()
	
	var res struct {
		Response struct {
			Data []AnslayerAnime `json:"data"`
		} `json:"response"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	
	if len(res.Response.Data) == 0 {
		sendMessage(ctx, "لم يتم العثور على أنمي بهذا الاسم.")
		return
	}
	
	var msg strings.Builder
	msg.WriteString("*نتائج البحث في انمي سلاير:*\n\n")
	
	session := &AnslayerSession{
		Mode: mode,
		State: "select_anime",
		Animes: res.Response.Data,
	}
	ansSessions[ctx.Sender.User] = session
	
	for i, a := range res.Response.Data {
		msg.WriteString(fmt.Sprintf("%d. %s\n", i+1, a.AnimeName))
		if i == 9 { break } // show max 10
	}
	msg.WriteString("\n*للاختيار اكتب:* `.رقم` متبوعاً بالرقم (مثال: `.رقم 1`)")
	sendMessage(ctx, msg.String())
}

func HandleAnslayerNumberSelect(ctx *BotContext, number int) bool {
	session, ok := ansSessions[ctx.Sender.User]
	if !ok {
		return false
	}
	
	if session.State == "select_anime" {
		if number < 1 || number > len(session.Animes) {
			sendMessage(ctx, "رقم غير صحيح.")
			return true
		}
		
		selected := session.Animes[number-1]
		session.SelectedAnime = selected
		
		sendMessage(ctx, fmt.Sprintf("تم اختيار الأنمي: *%s*\nجاري جلب الحلقات...", selected.AnimeName))
		
		u := fmt.Sprintf("https://anslayer.com/anime/public/anime/get-anime-details?anime_id=%s&fetch_episodes=Yes&more_info=No", selected.AnimeID)
		req, _ := http.NewRequest("GET", u, nil)
		reqHeaders(req)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			sendMessage(ctx, "خطأ في جلب الحلقات.")
			return true
		}
		defer resp.Body.Close()
		
		var res struct {
			Response struct {
				Episodes struct {
					Data []AnslayerEpisode `json:"data"`
				} `json:"episodes"`
			} `json:"response"`
		}
		json.NewDecoder(resp.Body).Decode(&res)
		
		eps := res.Response.Episodes.Data
		if len(eps) == 0 {
			sendMessage(ctx, "لا توجد حلقات متاحة.")
			return true
		}
		
		// reverse to show newest first? actually just store them
		session.Episodes = eps
		session.State = "select_episode"
		ansSessions[ctx.Sender.User] = session
		
		sendMessage(ctx, fmt.Sprintf("يوجد %d حلقة متاحة.\n*للاختيار اكتب:* `.حلقة` متبوعاً بالرقم الفعلي للحلقة (مثال: `.حلقة 1`)", len(eps)))
		return true
	}
	return false
}

func HandleAnslayerEpisodeSelect(ctx *BotContext, epNumInt int) bool {
	session, ok := ansSessions[ctx.Sender.User]
	if !ok || session.State != "select_episode" {
		return false
	}
	
	epNumStr := strconv.Itoa(epNumInt)
	var epIDStr string
	for _, e := range session.Episodes {
		if e.EpisodeNumber == epNumStr {
			epIDStr = e.EpisodeID
			break
		}
	}
	
	if epIDStr == "" {
		if epNumInt > 0 && epNumInt <= len(session.Episodes) {
			epIDStr = session.Episodes[epNumInt-1].EpisodeID
		}
	}
	
	if epIDStr == "" {
		sendMessage(ctx, "لم يتم العثور على الحلقة.")
		return true
	}
	
	if session.Mode == "watch" {
		sendMessage(ctx, "جاري جلب الحلقة وتحميلها...")
		go downloadAnslayerEpisode(ctx, session.SelectedAnime, epNumStr, epIDStr)
		delete(ansSessions, ctx.Sender.User)
		return true
	}
	
	anslayerMutex.Lock()
	if anslayerReplyMsg == "" {
		anslayerMutex.Unlock()
		sendMessage(ctx, "⚠️ لم تقم بضبط رسالة النشر!\nيرجى كتابة:\n.انمي سلاير نشر رسالتي\nقبل بدء المراقبة.")
		return true
	}
	
	if anslayerStopChan != nil {
		close(anslayerStopChan)
	}
	anslayerStopChan = make(chan struct{})
	anslayerMonitored = epIDStr
	ch := anslayerStopChan
	anslayerMutex.Unlock()
	
	sendMessage(ctx, "✅ تم بدء مراقبة التعليقات للحلقة!\nسيقوم البوت بالرد فوراً على أي تعليق جديد (ولن يرد على شخص مرتين).\nلإيقاف المراقبة، اطلب حلقة أخرى أو أعد تشغيل البوت.")
	
	go monitorComments(ctx, epIDStr, ch)
	
	delete(ansSessions, ctx.Sender.User)
	return true
}
func monitorComments(ctx *BotContext, epID string, stopCh chan struct{}) {
	epIDFloat, _ := strconv.ParseFloat(epID, 64)
	oldOffset := 30
	finishedNotified := false

	for {
		select {
		case <-stopCh:
			return
		default:
			anslayerMutex.Lock()
			msg := anslayerReplyMsg
			anslayerMutex.Unlock()

			if msg == "" {
				time.Sleep(10 * time.Second)
				continue
			}

			// Check newest first
			replied, hasComments := checkAndReplyBatch(epIDFloat, msg, 0, 30)
			if replied {
				time.Sleep(65 * time.Second)
				continue
			}

			// If no new comments to reply to, check older comments
			replied, hasComments = checkAndReplyBatch(epIDFloat, msg, oldOffset, 30)
			if replied {
				oldOffset += 1 
				time.Sleep(65 * time.Second)
				continue
			}

			if !hasComments {
				if !finishedNotified {
					sendMessage(ctx, "✅ تم الرد على جميع التعليقات القديمة في هذه الحلقة! البوت الآن في وضع الاستعداد للتعليقات الجديدة فقط، يمكنك اختيار حلقة أخرى إذا أردت.")
					finishedNotified = true
				}
				// reset to 0 to only monitor new ones
				oldOffset = 0
			} else {
				// advance to older
				oldOffset += 30
			}
			time.Sleep(10 * time.Second)
		}
	}
}

func checkAndReplyBatch(epIDFloat float64, msg string, offset, limit int) (bool, bool) {
	params := map[string]interface{}{
		"_order_by": "latest_first",
		"hide_irrelevant": "Yes",
		"episode_id": epIDFloat,
		"_limit": limit,
		"myfirst": "Yes",
		"_offset": offset,
	}
	b, _ := json.Marshal(params)
	u := "https://anslayer.com/anime/public/episode-comments/get-episode-comments?json=" + url.QueryEscape(string(b))
	
	req, _ := http.NewRequest("GET", u, nil)
	reqHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, false
	}
	defer resp.Body.Close()
	
	var res struct {
		Response struct {
			Data []struct {
				CommentID string `json:"episode_comment_id"`
				UserID    string `json:"user_id"`
			} `json:"data"`
		} `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return false, false
	}
	
	if len(res.Response.Data) == 0 {
		return false, false // No comments in this batch
	}
	
	for _, c := range res.Response.Data {
		if c.UserID == "" || c.CommentID == "" { continue }
		if hasAnslayerUser(c.UserID) { continue }
		
		// Found one to reply to
		payload := url.Values{}
		payload.Set("episode_comment_id", c.CommentID)
		payload.Set("reply_text", msg)
		payload.Set("spoiler", "No")
		payload.Set("recipient_id", "")
		payload.Set("notification_type", "reply")
		
		reqR, _ := http.NewRequest("POST", "https://anslayer.com/anime/public/episode-comments/create-episode-comment-reply", strings.NewReader(payload.Encode()))
		reqHeaders(reqR)
		reqR.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		
		respR, errR := http.DefaultClient.Do(reqR)
		if errR == nil {
			respR.Body.Close()
			if respR.StatusCode == 200 {
				saveAnslayerUser(c.UserID)
				fmt.Println("Replied to user:", c.UserID)
				return true, true
			}
		}
	}
	return false, true
}

func resolveMediaFire(u string) string {
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return u
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	
	idx := strings.Index(html, "href=\"https://download")
	if idx != -1 {
		start := idx + 6
		end := strings.Index(html[start:], "\"")
		if end != -1 {
			return html[start : start+end]
		}
	}
	return u
}

func downloadAnslayerEpisode(ctx *BotContext, anime AnslayerAnime, epNum string, epID string) {
	// First get the episode details to find episode_urls
	u := fmt.Sprintf("https://anslayer.com/anime/public/anime/get-anime-details?anime_id=%s&fetch_episodes=Yes&more_info=No", anime.AnimeID)
	req, _ := http.NewRequest("GET", u, nil)
	reqHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		sendMessage(ctx, "خطأ في الاتصال.")
		return
	}
	
	var res struct {
		Response struct {
			Episodes struct {
				Data []struct {
					EpisodeID string `json:"episode_id"`
					EpisodeUrls []struct {
						ServerName string `json:"episode_server_name"`
						Url        string `json:"episode_url"`
					} `json:"episode_urls"`
				} `json:"data"`
			} `json:"episodes"`
		} `json:"response"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	resp.Body.Close()
	
	var links []string
	
	for _, e := range res.Response.Episodes.Data {
		if e.EpisodeID == epID {
			for _, u := range e.EpisodeUrls {
				if u.ServerName == "muilt" {
					// Fetch muilt links
					req2, _ := http.NewRequest("GET", u.Url, nil)
					reqHeaders(req2)
					resp2, err := http.DefaultClient.Do(req2)
					if err == nil {
						var mLinks []string
						json.NewDecoder(resp2.Body).Decode(&mLinks)
						links = append(links, mLinks...)
						resp2.Body.Close()
					}
				} else {
					// Add backup servers
					links = append(links, u.Url)
				}
			}
			break
		}
	}
	
	if len(links) == 0 {
		sendMessage(ctx, "لا توجد روابط لهذه الحلقة.")
		return
	}
	
	var data []byte
	var success bool

	for i, targetLink := range links {
		if strings.Contains(targetLink, "mediafire.com") {
			targetLink = strings.Replace(targetLink, "file_premium", "file", 1)
			targetLink = resolveMediaFire(targetLink)
		}
		
		if i > 0 {
			sendMessage(ctx, fmt.Sprintf("السيرفر السابق محذوف، جاري تجربة سيرفر بديل (%d/%d)...", i+1, len(links)))
		}
		
		data, err = DownloadM3U8WithQuality(targetLink, "bestvideo[height<=720]+bestaudio/best[height<=720]")
		if err == nil {
			success = true
			break
		}
	}
	
	if !success {
		sendMessage(ctx, "فشلت جميع السيرفرات في التحميل. (ربما تم حذف الحلقة من جميع المصادر بسبب حقوق النشر)")
		return
	}
	
	sendVideoDataWithSplit(ctx, data, anime.AnimeName, epNum, false)
}


var favStopChan chan struct{}

func startFavMarketing(ctx *BotContext, msg string) {
	if favStopChan != nil {
		close(favStopChan) // Stop previous
	}
	favStopChan = make(chan struct{})
	
	// Fetch favorites
	u := "https://anslayer.com/anime/public/animes/get-published-animes?json=%7B%22_offset%22%3A0%2C%22_limit%22%3A100%2C%22_order_by%22%3A%22latest_first%22%2C%22list_type%22%3A%22favorites%22%2C%22just_info%22%3A%22Yes%22%2C%22user_id%22%3A9174886%7D"
	req, _ := http.NewRequest("GET", u, nil)
	reqHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		sendMessage(ctx, "فشل جلب المفضلة.")
		return
	}
	defer resp.Body.Close()
	
	var res struct {
		Response struct {
			Data []AnslayerAnime `json:"data"`
		} `json:"response"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	
	if len(res.Response.Data) == 0 {
		sendMessage(ctx, "قائمة المفضلة فارغة.")
		return
	}
	
	animeIDs := make([]string, 0)
	for _, a := range res.Response.Data {
		animeIDs = append(animeIDs, a.AnimeID)
	}
	
	sendMessage(ctx, fmt.Sprintf("✅ تم جلب %d أنمي من المفضلة، سيبدأ البوت الآن بنشر التعليقات عليها جميعاً بشكل دوري!", len(animeIDs)))
	
	go monitorFavComments(ctx, animeIDs, msg, favStopChan)
}

func checkAndReplyAnimeBatch(animeIDFloat float64, msg string, offset, limit int) (bool, bool) {
	params := map[string]interface{}{
		"_order_by": "latest_first",
		"hide_irrelevant": "Yes",
		"anime_id": animeIDFloat,
		"_limit": limit,
		"myfirst": "Yes",
		"_offset": offset,
	}
	b, _ := json.Marshal(params)
	u := "https://anslayer.com/anime/public/anime-comments/get-anime-comments?json=" + url.QueryEscape(string(b))
	
	req, _ := http.NewRequest("GET", u, nil)
	reqHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, false
	}
	defer resp.Body.Close()
	
	var res struct {
		Response struct {
			Data []struct {
				CommentID string `json:"anime_comment_id"`
				UserID    string `json:"user_id"`
			} `json:"data"`
		} `json:"response"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	
	if len(res.Response.Data) == 0 {
		return false, false
	}
	
	for _, c := range res.Response.Data {
		if c.UserID == "" || c.CommentID == "" { continue }
		if hasAnslayerUser(c.UserID) { continue }
		
		payload := url.Values{}
		payload.Set("anime_comment_id", c.CommentID)
		payload.Set("reply_text", msg)
		payload.Set("spoiler", "No")
		payload.Set("recipient_id", "")
		payload.Set("notification_type", "reply")
		
		reqR, _ := http.NewRequest("POST", "https://anslayer.com/anime/public/anime-comments/create-anime-comment-reply", strings.NewReader(payload.Encode()))
		reqHeaders(reqR)
		reqR.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		
		respR, errR := http.DefaultClient.Do(reqR)
		if errR == nil {
			respR.Body.Close()
			if respR.StatusCode == 200 {
				ansUsersMutex.Lock()
				anslayerUsers[c.UserID] = true
				ansUsersMutex.Unlock()
				
				f, _ := os.OpenFile("anslayer_users.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				f.WriteString(c.UserID + "\n")
				f.Close()
				
				fmt.Println("Replied to user:", c.UserID, "on anime comment:", c.CommentID)
				return true, true
			}
		}
	}
	return false, true
}

func getLatestEpisodeID(animeID string) string {
	u := fmt.Sprintf("https://anslayer.com/anime/public/anime/get-anime-details?anime_id=%s&fetch_episodes=Yes&more_info=No", animeID)
	req, _ := http.NewRequest("GET", u, nil)
	reqHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	
	var res struct {
		Response struct {
			Episodes struct {
				Data []AnslayerEpisode `json:"data"`
			} `json:"episodes"`
		} `json:"response"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	
	if len(res.Response.Episodes.Data) > 0 {
		// Usually the first item is the newest episode if ordered by latest
		return res.Response.Episodes.Data[0].EpisodeID
	}
	return ""
}

func monitorFavComments(ctx *BotContext, animeIDs []string, msg string, stopCh chan struct{}) {
	oldOffsets := make(map[string]int)
	for _, id := range animeIDs {
		oldOffsets[id] = 30 // Start looking back from 30 for episode comments
	}
	
	for {
		select {
		case <-stopCh:
			return
		default:
			repliedInThisLoop := false
			
			for _, animeID := range animeIDs {
				// Fetch the latest episode ID for this anime dynamically
				latestEpID := getLatestEpisodeID(animeID)
				if latestEpID == "" {
					continue
				}
				
				epIDFloat, _ := strconv.ParseFloat(latestEpID, 64)
				
				// 1. Check Newest first (offset 0)
				replied, _ := checkAndReplyBatch(epIDFloat, msg, 0, 30) // Use existing episode batch func
				if replied {
					repliedInThisLoop = true
					time.Sleep(65 * time.Second)
					break
				}
				
				// 2. If no new comment, check older comments
				offset := oldOffsets[animeID]
				replied, hasComments := checkAndReplyBatch(epIDFloat, msg, offset, 30)
				if replied {
					oldOffsets[animeID] += 1
					repliedInThisLoop = true
					time.Sleep(65 * time.Second)
					break
				} else {
					if hasComments {
						oldOffsets[animeID] += 30
					}
				}
			}
			
			if !repliedInThisLoop {
				time.Sleep(10 * time.Second)
			}
		}
	}
}
