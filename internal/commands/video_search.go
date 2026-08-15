package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

type MultiVideoSession struct {
	Query  string
	Count  int
	Offset int
	Links  []string
}

var (
	multiVideoSessions   = make(map[string]*MultiVideoSession)
	multiVideoSessionMux sync.Mutex
)

func setMultiVideoSession(key string, session *MultiVideoSession) {
	multiVideoSessionMux.Lock()
	defer multiVideoSessionMux.Unlock()
	multiVideoSessions[key] = session
}

func getMultiVideoSession(key string) *MultiVideoSession {
	multiVideoSessionMux.Lock()
	defer multiVideoSessionMux.Unlock()
	return multiVideoSessions[key]
}

func fetchTikTokLinksAPI(query string) ([]string, error) {
	// Try multiple APIs recursively via regex to find any tiktok url
	apis := []string{
		"https://aemt.me/tiktoksearch?text=" + url.QueryEscape(query),
		"https://api.vreden.web.id/api/tiktoksearch?query=" + url.QueryEscape(query),
	}

	client := &http.Client{Timeout: 15 * time.Second}
	var allBody string

	for _, apiURL := range apis {
		resp, err := client.Get(apiURL)
		if err == nil && resp.StatusCode == 200 {
			body, _ := io.ReadAll(resp.Body)
			allBody += string(body) + " "
			resp.Body.Close()
		}
	}

	if allBody == "" {
		return nil, fmt.Errorf("all apis failed")
	}

	// Some APIs return "https://www.tiktok.com/@user/video/123", others return direct mp4 links
	reTikTok := regexp.MustCompile(`(?i)tiktok\.com/@[^/]+/video/\d+`)
	matches := reTikTok.FindAllString(allBody, -1)
	
	uniqueLinks := make([]string, 0)
	seen := make(map[string]bool)
	
	for _, match := range matches {
		link := "https://www." + match
		if !seen[link] {
			seen[link] = true
			uniqueLinks = append(uniqueLinks, link)
		}
	}

	// If no standard tiktok links found, let's look for any generic play urls (tikwm style) if the API directly returns MP4s
	if len(uniqueLinks) == 0 {
		reMP4 := regexp.MustCompile(`(?i)https?://[^\s"']+\.mp4[^\s"']*`)
		mp4Matches := reMP4.FindAllString(allBody, -1)
		for _, match := range mp4Matches {
			if !seen[match] {
				seen[match] = true
				uniqueLinks = append(uniqueLinks, match)
			}
		}
	}

	return uniqueLinks, nil
}

func downloadTikTokDirect(ctx *BotContext, link string) {
	// If it's already a direct mp4
	if strings.Contains(link, ".mp4") {
		mediaResp, err := http.Get(link)
		if err == nil && mediaResp.StatusCode == 200 {
			defer mediaResp.Body.Close()
			mediaData, _ := io.ReadAll(mediaResp.Body)
			sendDirectVideo(ctx, mediaData)
			return
		}
	}

	// Otherwise, it's a tiktok.com URL, so use tikwm to download it
	apiURL := "https://www.tikwm.com/api/?url=" + url.QueryEscape(link)
	resp, err := http.Get(apiURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var result struct {
		Code int `json:"code"`
		Data struct {
			Play string `json:"play"`
		} `json:"data"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return
	}

	if result.Code != 0 || result.Data.Play == "" {
		return
	}

	mediaResp, err := http.Get(result.Data.Play)
	if err != nil || mediaResp.StatusCode != 200 {
		return
	}
	defer mediaResp.Body.Close()

	mediaData, _ := io.ReadAll(mediaResp.Body)
	sendDirectVideo(ctx, mediaData)
}

func sendDirectVideo(ctx *BotContext, mediaData []byte) {
	uploadedMedia, err := ctx.Client.Upload(context.Background(), mediaData, whatsmeow.MediaVideo)
	if err != nil {
		return
	}

	finalMsg := &waProto.Message{
		VideoMessage: &waProto.VideoMessage{
			URL:           proto.String(uploadedMedia.URL),
			DirectPath:    proto.String(uploadedMedia.DirectPath),
			MediaKey:      uploadedMedia.MediaKey,
			Mimetype:      proto.String("video/mp4"),
			FileEncSHA256: uploadedMedia.FileEncSHA256,
			FileSHA256:    uploadedMedia.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(mediaData))),
		},
	}
	ctx.Client.SendMessage(context.Background(), ctx.ChatID, finalMsg)
}

func multiVideoSearch(ctx *BotContext) {
	text := strings.TrimSpace(strings.TrimPrefix(ctx.Text, strings.Split(ctx.Text, " ")[0]))
	if text == "" {
		sendMessage(ctx, "اكتب كلمة البحث! مثلاً:\n.فيديو قطط 3")
		return
	}

	parts := strings.Fields(text)
	count := 1
	query := text

	if len(parts) > 1 {
		if c, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
			count = c
			query = strings.Join(parts[:len(parts)-1], " ")
		}
	} else if len(parts) == 1 {
		if c, err := strconv.Atoi(parts[0]); err == nil {
			count = c
			query = ""
		}
	}

	if query == "" {
		sendMessage(ctx, "وين كلمة البحث؟")
		return
	}

	if count > 5 {
		count = 5
	}

	sessionKey := ctx.ChatID.String() + "_" + ctx.Sender.String()
	existingSession := getMultiVideoSession(sessionKey)
	if existingSession != nil && existingSession.Query == query {
		existingSession.Count = count
		setMultiVideoSession(sessionKey, existingSession)
		multiVideoSearchNew(ctx)
		return
	}

	links, err := fetchTikTokLinksAPI(query)
	if err != nil || len(links) == 0 {
		sendMessage(ctx, "ما قدرت ألقى نتائج! (تأكد من سيرفرك أو جرب كلمة ثانية)")
		return
	}

	setMultiVideoSession(sessionKey, &MultiVideoSession{
		Query:  query,
		Count:  count,
		Offset: count,
		Links:  links,
	})

	limit := count
	if limit > len(links) {
		limit = len(links)
	}

	var wg sync.WaitGroup
	for i := 0; i < limit; i++ {
		wg.Add(1)
		go func(link string) {
			defer wg.Done()
			downloadTikTokDirect(ctx, link)
		}(links[i])
	}
	wg.Wait()
}

func multiVideoSearchNew(ctx *BotContext) {
	sessionKey := ctx.ChatID.String() + "_" + ctx.Sender.String()
	session := getMultiVideoSession(sessionKey)

	if session == nil {
		sendMessage(ctx, "ما عندك بحث سابق عشان أجيب لك مقاطع جديدة! ابحث أول بـ .فيديو")
		return
	}

	if session.Offset >= len(session.Links) {
		sendMessage(ctx, "خلصت كل المقاطع المتاحة لهذا البحث!")
		return
	}

	limit := session.Count
	if session.Offset+limit > len(session.Links) {
		limit = len(session.Links) - session.Offset
	}

	var wg sync.WaitGroup
	for i := 0; i < limit; i++ {
		link := session.Links[session.Offset+i]
		wg.Add(1)
		go func(l string) {
			defer wg.Done()
			downloadTikTokDirect(ctx, l)
		}(link)
	}

	session.Offset += limit
	setMultiVideoSession(sessionKey, session)

	wg.Wait()
}

func handleNewCommand(ctx *BotContext) {
	sessionKey := ctx.ChatID.String() + "_" + ctx.Sender.String()
	
	if getMultiVideoSession(sessionKey) != nil {
		multiVideoSearchNew(ctx)
		return
	}
	
	refreshPinterest(ctx)
}
