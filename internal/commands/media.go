package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

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
	mediaSessions = make(map[string][]MediaResult)
	mediaMutex    sync.Mutex
	tmdbAPIKey    = "15d2ea6d0dc1d476efbca3eba2b9bbfb"
)

func HandleMediaCommand(ctx *BotContext, cmd string) {
	query := strings.TrimSpace(strings.TrimPrefix(ctx.Text, cmd))
	if query == "" {
		sendMessage(ctx, fmt.Sprintf("يرجى كتابة اسم العمل بعد الأمر، مثال:\n%s باتمان", cmd))
		return
	}

	sendMessage(ctx, "جاري البحث... ⏳")

	var results []MediaResult

	switch cmd {
	case ".فلم", ".فيلم":
		results = searchTMDB(query, "movie")
	case ".مسلسل":
		results = searchTMDB(query, "tv")
	case ".انمي", ".أنمي":
		results = searchJikan(query, "anime")
	case ".مانجا", ".مانهاوا":
		results = searchJikan(query, "manga")
	}

	if len(results) == 0 {
		sendMessage(ctx, "للأسف ما لقيت أي نتيجة لطلبك!")
		return
	}

	// Save to session for .new
	mediaMutex.Lock()
	mediaSessions[ctx.ChatID.String()] = results[1:] // save the rest
	mediaMutex.Unlock()

	sendMediaResult(ctx, results[0])
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

	sendMediaResult(ctx, res)
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

func sendMediaResult(ctx *BotContext, res MediaResult) {
	// Fetch extra details if it's TMDB
	fetchTMDBDetails(&res)

	msg := fmt.Sprintf("🎬 *%s*\n\n", res.Title)
	if res.Rating != "" && res.Rating != "0.0/10" && res.Rating != "0.00/10" {
		msg += fmt.Sprintf("⭐️ *التقييم:* %s\n", res.Rating)
	}
	if res.Year != "" {
		msg += fmt.Sprintf("📅 *السنة:* %s\n", res.Year)
	}
	if res.Episodes != "" {
		msg += fmt.Sprintf("🔢 *عدد الحلقات:* %s\n", res.Episodes)
	}
	if res.Duration != "" {
		msg += fmt.Sprintf("⏱️ *المدة:* %s\n", res.Duration)
	}
	if res.Status != "" {
		msg += fmt.Sprintf("✅ *الحالة:* %s\n", res.Status)
	}

	msg += fmt.Sprintf("\n📝 *الوصف:*\n%s\n\n", res.Description)
	msg += "💡 للمزيد من النتائج لنفس البحث، ارسل `.new`"

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
				return
			}
		}
	}

	// Fallback to text
	sendMessage(ctx, msg)
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
		apiURL = fmt.Sprintf("https://api.themoviedb.org/3/search/%s?api_key=%s&query=%s&language=ar-SA", searchType, tmdbAPIKey, url.QueryEscape(query))
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
		apiURL = fmt.Sprintf("https://api.jikan.moe/v4/%s?q=%s", searchType, url.QueryEscape(query))
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
