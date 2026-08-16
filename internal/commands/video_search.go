package commands

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"whatsapp-bot/internal/youtube"
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

func fetchShortVideoLinks(query string, count int) ([]string, error) {
	// TikTok search APIs are heavily blocked by Cloudflare now.
	// As a robust alternative, we use YouTube Shorts which provides the exact same short video format.
	videoIDs, _, err := youtube.SearchVideos(query+" shorts", count+2, "")
	if err != nil {
		return nil, err
	}
	
	uniqueLinks := make([]string, 0)
	for _, id := range videoIDs {
		uniqueLinks = append(uniqueLinks, "https://youtu.be/"+id)
	}
	return uniqueLinks, nil
}

func downloadShortVideoDirect(ctx *BotContext, link string) {
	// If it's a youtube link, download natively
	if strings.Contains(link, "youtu.be") {
		videoID := strings.TrimPrefix(link, "https://youtu.be/")
		mediaData, err := youtube.DownloadMedia(videoID, false)
		if err == nil {
			sendDirectVideo(ctx, mediaData)
		}
		return
	}

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

	links, err := fetchShortVideoLinks(query, count)
	if err != nil || len(links) == 0 {
		sendMessage(ctx, "ما قدرت ألقى نتائج! (جرب كلمة ثانية)")
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
			downloadShortVideoDirect(ctx, link)
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
			downloadShortVideoDirect(ctx, l)
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
