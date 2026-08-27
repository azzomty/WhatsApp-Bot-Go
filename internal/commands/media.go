package commands

import (
	"os"
	"strconv"

	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"os/exec"
	"runtime"

	"whatsapp-bot/internal/youtube"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

type MediaResult struct {
	Title       string
	Description string
	Rating      string
	Year        string
	PosterURL   string
	Episodes    string
	Duration    string
	Status      string
	IsTMDB      bool
	TMDBID      int
	MediaType   string // "movie" or "tv"
}

var (
	mediaSessions   = make(map[string][]MediaResult)
	mediaMutex      sync.Mutex
	cartoonSessions = make(map[string]string)
	cartoonListSessions = make(map[string][]MediaResult)
	cartoonMutex    sync.Mutex
	tmdbAPIKey      = "15d2ea6d0dc1d476efbca3eba2b9bbfb"
)

func HandleMediaCommand(ctx *BotContext, cmd string) {
	query := strings.TrimSpace(strings.TrimPrefix(ctx.Text, cmd))
	if query == "" {
		sendMessage(ctx, fmt.Sprintf("يرجى كتابة اسم العمل بعد الأمر، مثال:\n%s باتمان", cmd))
		return
	}

	sendMessage(ctx, "جاري البحث... ")

	var results []MediaResult

	switch cmd {
	case ".فلم", ".فيلم":
		results = searchTMDB(query, "movie")
	case ".مسلسل":
		results = searchTMDB(query, "tv")
	case ".كرتون", ".انمي_مدبلج":
			results = SearchArabicCartoon(query)
		cartoonMutex.Lock()
		cartoonListSessions[ctx.Sender.User] = results
		cartoonMutex.Unlock()
	case ".انمي", ".أنمي":
		results = searchJikan(query, "anime")
	case ".مانجا", ".مانهاوا":
		results = searchJikan(query, "manga")
		results = searchJikan(query, "manga")
	}

		if len(results) == 0 {
			sendMessage(ctx, "للأسف ما لقيت أي نتيجة لطلبك!")
			return
		}

		if (cmd == ".كرتون" || cmd == ".انمي_مدبلج") && len(results) > 1 {
			msg := "*اختر الجزء أو الكرتون المطلوب بدقة، واكتب اسمه كاملاً:*\n\n"
			for i, r := range results {
				if i >= 20 {
					break
				}
				msg += fmt.Sprintf("- %s\n", r.Title)
			}
			msg += fmt.Sprintf("\nمثال:\n`.الجزء الأول`")
			sendMessage(ctx, msg)
			return
		}
		// Save to session for .new
	mediaMutex.Lock()
	mediaSessions[ctx.ChatID.String()] = results[1:] // save the rest
	mediaMutex.Unlock()

	sendMediaResult(ctx, results[0], cmd)
}

func HandleMediaNew(ctx *BotContext) bool {
	mediaMutex.Lock()
	results, ok := mediaSessions[ctx.ChatID.String()]
	if !ok || len(results) == 0 {
		mediaMutex.Unlock()
		return false
	}
	res := results[0]
	mediaSessions[ctx.ChatID.String()] = results[1:]
	mediaMutex.Unlock()

	sendMediaResult(ctx, res, "")
	
	// If it's a TV show, update cartoon session so .حلقة works on the new result!
	if res.Episodes != "" || res.MediaType == "tv" {
		cartoonMutex.Lock()
		cartoonSessions[ctx.Sender.User] = res.Title
		cartoonMutex.Unlock()
	}
	
	return true
}

func fetchTMDBDetails(res *MediaResult) {
	if !res.IsTMDB || res.TMDBID == 0 {
		return
	}
	apiURL := fmt.Sprintf("https://api.themoviedb.org/3/%s/%d?api_key=%s&language=ar-SA", res.MediaType, res.TMDBID, tmdbAPIKey)
	resp, err := http.Get(apiURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var data struct {
		Status           string `json:"status"`
		Runtime          int    `json:"runtime"`          // for movie
		EpisodeRunTime   []int  `json:"episode_run_time"` // for tv
		NumberOfEpisodes int    `json:"number_of_episodes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
		// Translate Status
		statusMap := map[string]string{
			"Rumored": "إشاعة", "Planned": "مُخطط له", "In Production": "قيد الإنتاج",
			"Post Production": "ما بعد الإنتاج", "Released": "تم العرض", "Canceled": "ملغي",
			"Ended": "منتهي", "Returning Series": "مستمر",
		}
		if s, ok := statusMap[data.Status]; ok {
			res.Status = s
		} else {
			res.Status = data.Status
		}

		if res.MediaType == "tv" {
			if data.NumberOfEpisodes > 0 {
				res.Episodes = fmt.Sprintf("%d", data.NumberOfEpisodes)
			}
			if len(data.EpisodeRunTime) > 0 {
				res.Duration = fmt.Sprintf("%d دقيقة", data.EpisodeRunTime[0])
			}
		} else {
			if data.Runtime > 0 {
				res.Duration = fmt.Sprintf("%d دقيقة", data.Runtime)
			}
		}
	}
}

func sendMediaResult(ctx *BotContext, res MediaResult, cmd string) {
	// Fetch extra details if it's TMDB
	fetchTMDBDetails(&res)

	msg := fmt.Sprintf("*%s*\n\n", res.Title)
	if res.Rating != "" && res.Rating != "0.0/10" && res.Rating != "0.00/10" {
		msg += fmt.Sprintf("*التقييم:* %s\n", res.Rating)
	}
	if res.Year != "" {
		msg += fmt.Sprintf("*السنة:* %s\n", res.Year)
	}
	if res.Episodes != "" && cmd != ".كرتون" && cmd != ".انمي_مدبلج" {
		msg += fmt.Sprintf("*عدد الحلقات:* %s\n", res.Episodes)
	}
	if res.Duration != "" {
		msg += fmt.Sprintf("*المدة:* %s\n", res.Duration)
	}
	if res.Status != "" {
		msg += fmt.Sprintf("*الحالة:* %s\n", res.Status)
	}

	msg += fmt.Sprintf("\n*الوصف:*\n%s\n\n", res.Description)
	msg += "للمزيد من النتائج لنفس البحث، ارسل `.new`"

	var sent bool
	if res.PosterURL != "" {
		data, err := downloadImage(res.PosterURL)
		if err == nil {
			resp, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaImage)
			if err == nil {
				imgMsg := &waProto.ImageMessage{
					URL:           proto.String(resp.URL),
					DirectPath:    proto.String(resp.DirectPath),
					MediaKey:      resp.MediaKey,
					Mimetype:      proto.String("image/jpeg"),
					FileEncSHA256: resp.FileEncSHA256,
					FileSHA256:    resp.FileSHA256,
					FileLength:    proto.Uint64(uint64(len(data))),
					Caption:       proto.String(msg),
				}
				ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
					ImageMessage: imgMsg,
				})
				sent = true
			}
		}
	}

	if !sent {
		sendMessage(ctx, msg)
	}

	if cmd == ".كرتون" || cmd == ".انمي_مدبلج" {
		cartoonMutex.Lock()
		cartoonSessions[ctx.Sender.User] = res.Title
		cartoonMutex.Unlock()
		
		epCount := res.Episodes
		if epCount == "" || epCount == "غير معروف" {
			epCount = "غير معروف (ابحث بالحلقة)"
		}
		sendMessage(ctx, fmt.Sprintf("عدد الحلقات المتوفرة: %s\n\nلتحميل حلقة، ارسل `.حلقة رقم_الحلقة` (مثال: `.حلقة 40`)", epCount))
	}
}

func getTMDBGenreID(query, mediaType string) string {
	q := strings.ReplaceAll(query, "أ", "ا")
	q = strings.ReplaceAll(q, "إ", "ا")
	switch q {
	case "اكشن", "حركة":
		if mediaType == "tv" {
			return "10759"
		}
		return "28"
	case "مغامرة", "مغامرات":
		if mediaType == "tv" {
			return "10759"
		}
		return "12"
	case "انيميشن", "رسوم متحركة", "كرتون":
		return "16"
	case "كوميديا", "كوميدي", "مضحك":
		return "35"
	case "جريمة", "عصابات":
		return "80"
	case "وثائقي", "حقيقي":
		return "99"
	case "دراما", "حزين":
		return "18"
	case "عائلي", "عائلة", "اطفال":
		if mediaType == "tv" {
			return "10762"
		}
		return "10751"
	case "فانتازيا", "خيال", "سحر":
		if mediaType == "tv" {
			return "10765"
		}
		return "14"
	case "تاريخ", "تاريخي":
		return "36"
	case "رعب", "مخيف":
		return "27"
	case "موسيقى", "موسيقي":
		return "10402"
	case "غموض", "لغز":
		return "9648"
	case "رومنسي", "رومانسي", "رومانسية", "حب":
		return "10749"
	case "خيال علمي", "فضاء":
		if mediaType == "tv" {
			return "10765"
		}
		return "878"
	case "اثارة", "تشويق":
		return "53"
	case "حرب", "حربي":
		if mediaType == "tv" {
			return "10768"
		}
		return "10752"
	case "غربي", "كاوبوي":
		return "37"
	default:
		return ""
	}
}

func getJikanGenreID(query string) string {
	q := strings.ReplaceAll(query, "أ", "ا")
	switch q {
	case "اكشن", "حركة": return "1"
	case "مغامرة", "مغامرات": return "2"
	case "سيارات": return "3"
	case "كوميديا", "كوميدي", "مضحك": return "4"
	case "خرف", "خيال": return "10"
	case "شياطين", "شيطان": return "6"
	case "غموض", "لغز": return "7"
	case "دراما", "حزين": return "8"
	case "ايتشي": return "9"
	case "فانتازيا", "سحر": return "16"
	case "رعب", "مخيف": return "14"
	case "اطفال", "عائلي": return "15"
	case "موسيقى", "موسيقي": return "19"
	case "شريحة من الحياة", "حياة": return "36"
	case "رياضة", "رياضي": return "30"
	case "رومنسي", "رومانسية", "رومانسي", "حب": return "22"
	case "تاريخ", "تاريخي": return "13"
	case "خيال علمي", "فضاء": return "24"
	case "شونين": return "27"
	case "شوجو": return "25"
	case "سينين": return "42"
	case "مدرسي", "مدرسة": return "23"
	case "العاب", "لعبة": return "11"
	case "نفسي": return "40"
	case "ايسيكاي", "عالم اخر": return "62"
	case "مصاص دماء", "مصاصين دماء": return "32"
	case "عسكري": return "38"
	case "بوليسي", "شرطة": return "39"
	case "ساموراي": return "21"
	default: return ""
	}
}

func searchTMDB(query, searchType string) []MediaResult {
	var apiURL string
	genreID := getTMDBGenreID(query, searchType)

	if genreID != "" {
		apiURL = fmt.Sprintf("https://api.themoviedb.org/3/discover/%s?api_key=%s&with_genres=%s&language=ar-SA&sort_by=popularity.desc", searchType, tmdbAPIKey, genreID)
	} else {
		apiURL = fmt.Sprintf("https://api.themoviedb.org/3/search/%s?api_key=%s&query=%s&language=ar-SA", searchType, tmdbAPIKey, strings.ReplaceAll(url.QueryEscape(query), "+", "%20"))
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var data struct {
		Results []struct {
			ID           int     `json:"id"`
			Title        string  `json:"title"`
			Name         string  `json:"name"`
			Overview     string  `json:"overview"`
			VoteAverage  float64 `json:"vote_average"`
			ReleaseDate  string  `json:"release_date"`
			FirstAirDate string  `json:"first_air_date"`
			PosterPath   string  `json:"poster_path"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}

	var results []MediaResult
	for _, item := range data.Results {
		title := item.Title
		if title == "" {
			title = item.Name
		}
		year := item.ReleaseDate
		if year == "" {
			year = item.FirstAirDate
		}
		if len(year) > 4 {
			year = year[:4]
		}
		poster := ""
		if item.PosterPath != "" {
			poster = "https://image.tmdb.org/t/p/w500" + item.PosterPath
		}

		desc := item.Overview
		if desc == "" {
			desc = "لا توجد قصة بالعربية لهذا العمل."
		}

		results = append(results, MediaResult{
			Title:       title,
			Description: desc,
			Rating:      fmt.Sprintf("%.1f/10", item.VoteAverage),
			Year:        year,
			PosterURL:   poster,
			IsTMDB:      true,
			TMDBID:      item.ID,
			MediaType:   searchType,
		})
	}
	return results
}

func searchJikan(query, searchType string) []MediaResult {
	var apiURL string
	genreID := getJikanGenreID(query)

	if genreID != "" {
		apiURL = fmt.Sprintf("https://api.jikan.moe/v4/%s?genres=%s&order_by=score&sort=desc", searchType, genreID)
	} else {
		apiURL = fmt.Sprintf("https://api.jikan.moe/v4/%s?q=%s", searchType, strings.ReplaceAll(url.QueryEscape(query), "+", "%20"))
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var data struct {
		Data []struct {
			Title    string  `json:"title"`
			Synopsis string  `json:"synopsis"`
			Score    float64 `json:"score"`
			Year     int     `json:"year"`
			Episodes int     `json:"episodes"`
			Chapters int     `json:"chapters"` // For manga
			Status   string  `json:"status"`
			Duration string  `json:"duration"`
			Images   struct {
				Jpg struct {
					ImageURL string `json:"image_url"`
				} `json:"jpg"`
			} `json:"images"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}

	var results []MediaResult
	for _, item := range data.Data {
		desc := item.Synopsis
		if desc == "" {
			desc = "لا توجد قصة متوفرة."
		} else if len(desc) > 800 {
			desc = desc[:800] + "..."
		}

		yearStr := ""
		if item.Year > 0 {
			yearStr = fmt.Sprintf("%d", item.Year)
		}

		epStr := ""
		if searchType == "anime" && item.Episodes > 0 {
			epStr = fmt.Sprintf("%d", item.Episodes)
		} else if searchType == "manga" && item.Chapters > 0 {
			epStr = fmt.Sprintf("%d فصول", item.Chapters)
		}

		statusStr := item.Status
		if statusStr == "Finished Airing" || statusStr == "Finished" {
			statusStr = "منتهي"
		} else if statusStr == "Currently Airing" || statusStr == "Publishing" {
			statusStr = "مستمر"
		}

		durStr := item.Duration
		if durStr == "Unknown" {
			durStr = ""
		}

		results = append(results, MediaResult{
			Title:       item.Title,
			Description: desc,
			Rating:      fmt.Sprintf("%.2f/10", item.Score),
			Year:        yearStr,
			PosterURL:   item.Images.Jpg.ImageURL,
			Episodes:    epStr,
			Duration:    durStr,
			Status:      statusStr,
			IsTMDB:      false,
		})
	}
	return results
}

func downloadImage(url string) ([]byte, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var buf []byte
	buffer := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			buf = append(buf, buffer[:n]...)
		}
		if err != nil {
			break
		}
	}
	return buf, nil
}


func HandleEpisodeCommand(ctx *BotContext) {
	parts := strings.Split(ctx.Text, " ")
	if len(parts) < 2 {
		sendMessage(ctx, "يرجى كتابة رقم الحلقة، مثال: .حلقة 40")
		return
	}
	epNum := parts[1]

	cartoonMutex.Lock()
	showName, ok := cartoonSessions[ctx.Sender.User]
	cartoonMutex.Unlock()

	if !ok || showName == "" {
		sendMessage(ctx, "لم تقم بالبحث عن أي كرتون مسبقاً. ابحث أولاً باستخدام أمر .كرتون (مثال: .كرتون سبونج بوب)")
		return
	}

	sendMessage(ctx, fmt.Sprintf("جاري جلب الحلقة %s من %s (من سيرفر مباشر)... ", epNum, showName))

	go func() {
		epNumInt, _ := strconv.Atoi(epNum)
		embedLink, err := ScrapeWitanimeEpisode(showName, epNumInt)
		
		if err != nil || embedLink == "" {
			sendMessage(ctx, "لم أتمكن من إيجاد سيرفر مباشر، سأحاول جلبها من يوتيوب... ")
			searchQuery := fmt.Sprintf("%s حلقة %s مدبلج بالعربي", showName, epNum)
			videoID, err := youtube.SearchVideo(searchQuery)
			if err != nil {
				sendMessage(ctx, "عذراً، الحلقة غير متوفرة.")
				return
			}
			
			data, err := youtube.DownloadMedia(videoID, false)
			if err != nil {
				sendMessage(ctx, fmt.Sprintf("حدث خطأ أثناء تحميل الحلقة: %v", err))
				return
			}
			sendVideoData(ctx, data, showName, epNum)
			return
		}

		// embedLink is a comma-separated list of links
		links := strings.Split(embedLink, ",")
		var outPath string
		var dlErr error
		
		// Try downloading from each link until one works
		for _, link := range links {
			if strings.TrimSpace(link) == "" {
				continue
			}
			sendMessage(ctx, "جاري تجربة السيرفر المباشر... ")
			outPath, dlErr = youtube.DownloadDirectURL(link)
			if dlErr == nil && outPath != "" {
				break
			}
		}

		if dlErr != nil || outPath == "" {
			sendMessage(ctx, "حدث خطأ: السيرفرات المباشرة لا تستجيب، سأحاول جلب الحلقة من يوتيوب...")
			searchQuery := fmt.Sprintf("%s حلقة %s مدبلج بالعربي", showName, epNum)
			videoID, err := youtube.SearchVideo(searchQuery)
			if err != nil {
				sendMessage(ctx, "عذراً، الحلقة غير متوفرة.")
				return
			}
			
			data, err := youtube.DownloadMedia(videoID, false)
			if err != nil {
				sendMessage(ctx, fmt.Sprintf("حدث خطأ أثناء تحميل الحلقة: %v", err))
				return
			}
			sendVideoData(ctx, data, showName, epNum)
			return
		}
		defer os.Remove(outPath)
		
		data, err := os.ReadFile(outPath)
		if err != nil {
			sendMessage(ctx, "حدث خطأ أثناء قراءة الملف.")
			return
		}
		sendVideoData(ctx, data, showName, epNum)
	}()
}



func sendVideoData(ctx *BotContext, data []byte, animeName, epNum string) {
	if len(data) <= 64*1024*1024 {
		// Send as a single video
		resp, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaVideo)
		if err != nil {
			sendMessage(ctx, "حدث خطأ أثناء رفع الحلقة للواتساب.")
			return
		}

		vidMsg := &waProto.VideoMessage{
			URL:           proto.String(resp.URL),
			DirectPath:    proto.String(resp.DirectPath),
			MediaKey:      resp.MediaKey,
			Mimetype:      proto.String("video/mp4"),
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
			Caption:       proto.String(fmt.Sprintf("*%s* - الحلقة %s", animeName, epNum)),
		}

		ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
			VideoMessage: vidMsg,
		})
		return
	}
	
	// File is > 64MB, split it!
	sendMessage(ctx, "حجم الحلقة كبير جداً (أكثر من 64 ميجا)، جاري تقسيمها إلى أجزاء وإرسالها... ")
	
	tempDir, err := os.MkdirTemp("", "video_split")
	if err != nil {
		sendMessage(ctx, "فشل في إنشاء مجلد مؤقت للتقسيم.")
		return
	}
	defer os.RemoveAll(tempDir)
	
	inputPath := tempDir + "/input.mp4"
	if err := os.WriteFile(inputPath, data, 0644); err != nil {
		sendMessage(ctx, "فشل في كتابة الملف المؤقت.")
		return
	}
	
	outPattern := tempDir + "/part_%03d.mp4"
	
	ffmpegPath := "node_modules/ffmpeg-static/ffmpeg"
	if runtime.GOOS == "windows" {
		if _, err := os.Stat("node_modules/ffmpeg-static/ffmpeg.exe"); err == nil {
			ffmpegPath = "node_modules/ffmpeg-static/ffmpeg.exe"
		}
	}
	
	cmd := exec.Command(ffmpegPath, "-i", inputPath, "-c", "copy", "-f", "segment", "-segment_time", "600", "-reset_timestamps", "1", outPattern)
	if err := cmd.Run(); err != nil {
		sendMessage(ctx, "فشل في تقسيم الفيديو. قد تكون الحلقة بصيغة غير مدعومة للتقسيم السريع.")
		return
	}
	
	files, err := os.ReadDir(tempDir)
	if err != nil {
		return
	}
	
	var parts []string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "part_") {
			parts = append(parts, tempDir+"/"+f.Name())
		}
	}
	
	// Send each part
	for i, partPath := range parts {
		partData, err := os.ReadFile(partPath)
		if err != nil {
			continue
		}
		
		resp, err := ctx.Client.Upload(context.Background(), partData, whatsmeow.MediaVideo)
		if err != nil {
			continue
		}
		
		caption := fmt.Sprintf("*%s* - الحلقة %s\n(الجزء %d من %d)", animeName, epNum, i+1, len(parts))
		
		vidMsg := &waProto.VideoMessage{
			URL:           proto.String(resp.URL),
			DirectPath:    proto.String(resp.DirectPath),
			MediaKey:      resp.MediaKey,
			Mimetype:      proto.String("video/mp4"),
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(partData))),
			Caption:       proto.String(caption),
		}
		
		ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
			VideoMessage: vidMsg,
		})
	}
}


func SearchArabicCartoon(query string) []MediaResult {
	q := strings.ReplaceAll(query, "ي", "_")
	q = strings.ReplaceAll(q, "ى", "_")
	q = strings.ReplaceAll(q, "أ", "_")
	q = strings.ReplaceAll(q, "إ", "_")
	q = strings.ReplaceAll(q, "آ", "_")
	q = strings.ReplaceAll(q, "ا", "_")
	q = strings.ReplaceAll(q, "ة", "_")
	q = strings.ReplaceAll(q, "ه", "_")
	
	// Because url.QueryEscape escapes '_' as well? No, '_' is not escaped.
	// We want to pass %25 for wildcard, and _ for single char.
	escapedQ := strings.ReplaceAll(url.QueryEscape(q), "+", "%20")
	
	reqURL := fmt.Sprintf("https://wwmdrwjkrzdkqjqddfta.supabase.co/rest/v1/series?select=*&title=ilike.*%%25%s%%25*", escapedQ)

	
	req, _ := http.NewRequest("GET", reqURL, nil)
	req.Header.Set("apikey", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Ind3bWRyd2prcnpka3FqcWRkZnRhIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA4MjAxNzUsImV4cCI6MjA5NjM5NjE3NX0.v3-gjEYfuJ4DE17OAHidvd38lCHUTU4ldb2SHLphU8s")
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Ind3bWRyd2prcnpka3FqcWRkZnRhIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA4MjAxNzUsImV4cCI6MjA5NjM5NjE3NX0.v3-gjEYfuJ4DE17OAHidvd38lCHUTU4ldb2SHLphU8s")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	
	var data []struct {
		Title         string  `json:"title"`
		Description   string  `json:"description"`
		PosterURL     string  `json:"poster_url"`
		Rating        float64 `json:"rating"`
		YearStarted   int     `json:"year_started"`
		TotalEpisodes int     `json:"total_episodes"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}
	
	var results []MediaResult
	for _, item := range data {
		res := MediaResult{
			Title:       item.Title,
			Description: item.Description,
			PosterURL: item.PosterURL,
		}
		if item.Rating > 0 {
			res.Rating = fmt.Sprintf("%.1f", item.Rating)
		}
		if item.YearStarted > 0 {
			res.Year = fmt.Sprintf("%d", item.YearStarted)
		}
		if item.TotalEpisodes > 0 {
			res.Episodes = fmt.Sprintf("%d", item.TotalEpisodes)
		}
		results = append(results, res)
	}
	return results
}


func HandlePartCommand(ctx *BotContext) {
	partQuery := strings.TrimSpace(strings.TrimPrefix(ctx.Text, ".الجزء"))
	if partQuery == "" {
		sendMessage(ctx, "يرجى تحديد الجزء، مثال: .الجزء الثاني")
		return
	}
	
	cartoonMutex.Lock()
	results, ok := cartoonListSessions[ctx.Sender.User]
	cartoonMutex.Unlock()
	
	if !ok || len(results) == 0 {
		sendMessage(ctx, "يرجى البحث عن الكرتون أولاً باستخدام أمر .كرتون")
		return
	}
	
	var selected MediaResult
	found := false
	for _, r := range results {
		if strings.Contains(r.Title, partQuery) {
			selected = r
			found = true
			break
		}
	}
	
	if !found {
		sendMessage(ctx, "لم أتمكن من العثور على هذا الجزء في نتائج بحثك السابقة! تأكد من كتابة الاسم الصحيح كما ظهر في القائمة.")
		return
	}
	
	sendMediaResult(ctx, selected, ".كرتون")
}

func HandleCartoonList(ctx *BotContext) {
	sendMessage(ctx, "جاري جلب القائمة... ")
	
	reqURL := "https://wwmdrwjkrzdkqjqddfta.supabase.co/rest/v1/series?select=title&order=title.asc"
	req, _ := http.NewRequest("GET", reqURL, nil)
	req.Header.Set("apikey", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Ind3bWRyd2prcnpka3FqcWRkZnRhIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA4MjAxNzUsImV4cCI6MjA5NjM5NjE3NX0.v3-gjEYfuJ4DE17OAHidvd38lCHUTU4ldb2SHLphU8s")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء جلب القائمة.")
		return
	}
	defer resp.Body.Close()
	
	var data []struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		sendMessage(ctx, "حدث خطأ أثناء جلب القائمة.")
		return
	}
	
	// Deduplicate base names (e.g. "المحقق كونان الجزء الأول" -> "المحقق كونان")
	uniqueShows := make(map[string]bool)
	var shows []string
	
	for _, item := range data {
		baseName := strings.Split(item.Title, " الجزء ")[0]
		baseName = strings.Split(baseName, " الموسم ")[0]
		if !uniqueShows[baseName] {
			uniqueShows[baseName] = true
			shows = append(shows, baseName)
		}
	}
	
	msg := "*قائمة الكراتين المتوفرة:*\n\n"
	for _, show := range shows {
		msg += "- " + show + "\n"
	}
	msg += "\n*للبحث عن أي كرتون، اكتب:* `.كرتون اسم_الكرتون`"
	
	sendMessage(ctx, msg)
}
