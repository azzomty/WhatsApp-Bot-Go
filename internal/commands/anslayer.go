package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	AnimeID   int    `json:"anime_id"`
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
	parts := strings.SplitN(ctx.Text, " ", 3)
	
	if mode == "marketing" && len(parts) >= 3 && parts[1] == "نشر" {
		anslayerMutex.Lock()
		anslayerReplyMsg = parts[2]
		anslayerMutex.Unlock()
		sendMessage(ctx, "✅ تم حفظ قالب الرد بنجاح!\nالرسالة:\n"+anslayerReplyMsg)
		return
	}
	
	var query string
	if mode == "marketing" {
		if len(parts) < 2 {
			sendMessage(ctx, "يرجى كتابة اسم الأنمي بعد الأمر، مثلا:\n.انمي سلاير ون بيس\nأو لحفظ رسالة النشر:\n.انمي سلاير نشر رسالتي هنا")
			return
		}
		query = strings.Join(parts[1:], " ")
	} else {
		if len(parts) < 2 {
			sendMessage(ctx, "يرجى كتابة اسم الأنمي بعد الأمر، مثلا:\n.انمي ون بيس")
			return
		}
		query = strings.Join(parts[1:], " ")
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
		
		u := fmt.Sprintf("https://anslayer.com/anime/public/anime/get-anime-details?anime_id=%d&fetch_episodes=Yes&more_info=No", selected.AnimeID)
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
	var epID string
	for _, e := range session.Episodes {
		if e.EpisodeNumber == epNumStr {
			epID = e.EpisodeID
			break
		}
	}
	
	if epID == "" {
		// fallback to index if episode number not exactly matching?
		// some movies are '1'
		if epNumInt > 0 && epNumInt <= len(session.Episodes) {
			epID = session.Episodes[epNumInt-1].EpisodeID
		}
	}
	
	if epID == "" {
		sendMessage(ctx, "لم يتم العثور على الحلقة.")
		return true
	}
	
	if session.Mode == "watch" {
		sendMessage(ctx, "جاري جلب الحلقة وتحميلها...")
		go downloadAnslayerEpisode(ctx, session.SelectedAnime, epNumStr, epID)
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
	anslayerMonitored = epID
	ch := anslayerStopChan
	anslayerMutex.Unlock()
	
	sendMessage(ctx, "✅ تم بدء مراقبة التعليقات للحلقة!\nسيقوم البوت بالرد فوراً على أي تعليق جديد (ولن يرد على شخص مرتين).\nلإيقاف المراقبة، اطلب حلقة أخرى أو أعد تشغيل البوت.")
	
	go monitorComments(epID, ch)
	
	delete(ansSessions, ctx.Sender.User)
	return true
}

func monitorComments(epID string, stopCh chan struct{}) {
	epIDFloat, _ := strconv.ParseFloat(epID, 64)
	oldOffset := 30

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
			if replied := checkAndReplyBatch(epIDFloat, msg, 0, 30); replied {
				time.Sleep(65 * time.Second)
				continue
			}

			// If no new comments to reply to, check older comments
			if replied := checkAndReplyBatch(epIDFloat, msg, oldOffset, 30); replied {
				oldOffset += 1 // Next time we can fetch from the same offset or advance slightly. Since we replied to one, the list shifted. We'll just advance oldOffset when the whole page is exhausted.
				time.Sleep(65 * time.Second)
				continue
			}

			// If we didn't reply to anything in oldOffset, advance it
			oldOffset += 30
			time.Sleep(10 * time.Second)
		}
	}
}

func checkAndReplyBatch(epIDFloat float64, msg string, offset, limit int) bool {
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
		return false
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
		return false
	}
	
	if len(res.Response.Data) == 0 {
		return false // No comments in this batch
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
				return true
			}
		}
	}
	return false
}
func downloadAnslayerEpisode(ctx *BotContext, anime AnslayerAnime, epNum string, epID string) {
	// First get the episode details to find episode_urls
	u := fmt.Sprintf("https://anslayer.com/anime/public/anime/get-anime-details?anime_id=%d&fetch_episodes=Yes&more_info=No", anime.AnimeID)
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
	
	var vidUrl string
	for _, e := range res.Response.Episodes.Data {
		if e.EpisodeID == epID {
			for _, u := range e.EpisodeUrls {
				if u.ServerName == "muilt" {
					vidUrl = u.Url
					break
				}
			}
			break
		}
	}
	
	if vidUrl == "" {
		sendMessage(ctx, "عذرا، لم أجد سيرفر صالح لهذه الحلقة.")
		return
	}
	
	// Fetch muilt links
	req2, _ := http.NewRequest("GET", vidUrl, nil)
	reqHeaders(req2)
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		sendMessage(ctx, "فشل جلب سيرفرات المشاهدة.")
		return
	}
	defer resp2.Body.Close()
	
	var links []string
	json.NewDecoder(resp2.Body).Decode(&links)
	
	if len(links) == 0 {
		sendMessage(ctx, "لا توجد روابط لهذه الحلقة.")
		return
	}
	
	// Choose first fembed or ok.ru link, or just first link
	targetLink := links[0]
	
	// Download using yt-dlp via existing function DownloadM3U8WithQuality (works for any url yt-dlp supports)
	data, err := DownloadM3U8WithQuality(targetLink, "bestvideo[height<=720]+bestaudio/best[height<=720]")
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء التحميل: " + err.Error())
		return
	}
	sendVideoDataWithSplit(ctx, data, anime.AnimeName, epNum, false)
}
