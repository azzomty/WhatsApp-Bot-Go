package commands

import (
	"context"
	"net"
	"sync"
	"os"
	"os/exec"
	"time"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type StardimaVideo struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	URL          string `json:"url"`
	PosterURL    string `json:"poster_url"`
	VideoQuality string `json:"video_quality"`
	IsSeries     bool   `json:"is_series"`
}

type StardimaSearchResponse struct {
	Videos []StardimaVideo `json:"videos"`
}

func SearchStardima(query string) ([]StardimaVideo, error) {
	reqURL := "https://stardima-37.cartoon.com.im/search?query=" + url.QueryEscape(query)
	req, _ := http.NewRequest("GET", reqURL, nil)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result StardimaSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Videos, nil
}

type StardimaSeason struct {
	ID   string
	Name string
}

func GetStardimaSeasons(showURL string) ([]StardimaSeason, error) {
	resp, err := http.Get(showURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	
	var body2 []byte
	if strings.Contains(showURL, "/play/") {
		body2 = body
	} else {
		playRe := regexp.MustCompile(`play/(\d+)`)
		m := playRe.FindStringSubmatch(string(body))
		if len(m) == 0 {
			return nil, fmt.Errorf("no play link found on show page")
		}

		playURL := showURL + "/play/" + m[1]
		resp2, err := http.Get(playURL)
		if err != nil {
			return nil, err
		}
		defer resp2.Body.Close()
		body2, _ = io.ReadAll(resp2.Body)
	}

	re := regexp.MustCompile(`data-season-id="(\d+)"[^>]*data-season-number="([^"]+)"`)
	matches := re.FindAllStringSubmatch(string(body2), -1)

	var seasons []StardimaSeason
	for _, match := range matches {
		seasons = append(seasons, StardimaSeason{ID: match[1], Name: match[2]})
	}
	return seasons, nil
}

type StardimaEpisode struct {
	ID            int    `json:"id"`
	EpisodeNumber int    `json:"episode_number"`
	Title         string `json:"title"`
	WatchURL      string `json:"watch_url"`
}

func GetStardimaEpisodes(seasonID string) ([]StardimaEpisode, error) {
	req, _ := http.NewRequest("GET", "https://stardima-37.cartoon.com.im/series/season/"+seasonID, nil)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Episodes []StardimaEpisode `json:"episodes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data.Episodes, nil
}

func GetStardimaHyperwatchingURL(movieURL string) (string, error) {
	resp, err := http.Get(movieURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	playRe := regexp.MustCompile(`(https://stardima-37\.cartoon\.com\.im/play/video-\d+)`)
	m := playRe.FindStringSubmatch(string(body))
	if len(m) > 1 {
		resp2, err := http.Get(m[1])
		if err == nil {
			defer resp2.Body.Close()
			body2, _ := io.ReadAll(resp2.Body)
			body = body2
		}
	}

	re := regexp.MustCompile(`https://v2\.hyperwatching\.com/watch/([a-zA-Z0-9]+)`)
	m2 := re.FindStringSubmatch(string(body))
	if len(m2) > 1 {
		return m2[0], nil
	}
	return "", fmt.Errorf("hyperwatching url not found")
}


func GetBestM3U8(hyperURL string) (string, error) {
	resp, err := http.Get(hyperURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	re := regexp.MustCompile(`data-page="([^"]+)"`)
	m := re.FindStringSubmatch(string(body))
	if len(m) < 2 {
		return "", fmt.Errorf("data-page not found")
	}

	jsonStr := strings.ReplaceAll(m[1], "&quot;", "\"")

	var pageData struct {
		Props struct {
			Video struct {
				Hashid  string `json:"hashid"`
				Servers []struct {
					ID       int    `json:"id"`
					ServerID int    `json:"server_id"`
					Name     string `json:"name"`
				} `json:"servers"`
			} `json:"video"`
		} `json:"props"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &pageData); err != nil {
		return "", err
	}

	// Try all servers that are available (ID != 0)
	for _, srv := range pageData.Props.Video.Servers {
		if srv.ID == 0 {
			continue
		}
		
		apiURL := fmt.Sprintf("https://v2.hyperwatching.com/embed/%s/server/%d/url", pageData.Props.Video.Hashid, srv.ID)
		resp2, err := http.Get(apiURL)
		if err != nil {
			continue
		}
		
		var apiData struct {
			WatchURL string `json:"watch_url"`
		}
		json.NewDecoder(resp2.Body).Decode(&apiData)
		resp2.Body.Close()
		
		parsed, err := url.Parse(apiData.WatchURL)
		if err != nil {
			continue
		}
		
		idParam := parsed.Query().Get("id")
		if idParam == "" {
			continue
		}
		
		// Try to extract M3U8 from this embed URL
		m3u8, err := ExtractM3U8(idParam)
		if err == nil && m3u8 != "" {
			return m3u8, nil
		}
	}

	return "", fmt.Errorf("no working server found")
}

func ExtractM3U8(embedURL string) (string, error) {
	// Custom dialer to bypass DNS issues for Lulustream/Luluvdo
	dialer := &net.Dialer{}
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				if strings.HasPrefix(addr, "luluvdo.com:") || strings.HasPrefix(addr, "lulustream.com:") {
					return dialer.DialContext(ctx, network, "104.26.6.79:443")
				}
				return dialer.DialContext(ctx, network, addr)
			},
		},
	}
	
	req, _ := http.NewRequest("GET", embedURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	re := regexp.MustCompile(`eval\(function\(p,a,c,k,e,d\).*?return p}\('(.*?)',(\d+),(\d+),'(.*?)'\.split`)
	match := re.FindStringSubmatch(html)
	if len(match) == 0 {
		return "", fmt.Errorf("eval block not found")
	}

	p := match[1]
	a, _ := strconv.Atoi(match[2])
	c, _ := strconv.Atoi(match[3])
	k := strings.Split(match[4], "|")

	baseN := func(num int, b int) string {
		if num == 0 {
			return "0"
		}
		res := ""
		for num > 0 {
			remainder := num % b
			if remainder > 35 {
				res = string(rune(remainder + 29)) + res
			} else if remainder > 9 {
				res = string(rune(remainder + 87)) + res
			} else {
				res = strconv.Itoa(remainder) + res
			}
			num /= b
		}
		return res
	}

	for i := c - 1; i >= 0; i-- {
		if k[i] != "" {
			escapedK := strings.ReplaceAll(k[i], "\\", "\\\\")
			p = regexp.MustCompile(`\b`+baseN(i, a)+`\b`).ReplaceAllString(p, escapedK)
		}
	}

	m3u8Re := regexp.MustCompile(`file\s*:\s*["'](https?://[^"']+\.m3u8[^"']*)["']`)
	m3match := m3u8Re.FindStringSubmatch(p)
	if len(m3match) > 1 {
		return m3match[1], nil
	}
	return "", fmt.Errorf("m3u8 not found in unpacked JS")
}

func GetUqloadEmbedURL(hyperURL string) (string, error) {
	resp, err := http.Get(hyperURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	re := regexp.MustCompile(`data-page="([^"]+)"`)
	m := re.FindStringSubmatch(string(body))
	if len(m) < 2 {
		return "", fmt.Errorf("data-page not found")
	}

	jsonStr := strings.ReplaceAll(m[1], "&quot;", "\"")

	var pageData struct {
		Props struct {
			Video struct {
				Hashid  string `json:"hashid"`
				Servers []struct {
					ID       int    `json:"id"`
					ServerID int    `json:"server_id"`
					Name     string `json:"name"`
				} `json:"servers"`
			} `json:"video"`
		} `json:"props"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &pageData); err != nil {
		return "", err
	}

	var uqloadServerID int
	for _, srv := range pageData.Props.Video.Servers {
		if strings.ToLower(srv.Name) == "uqload" {
			uqloadServerID = srv.ID
			break
		}
	}

	if uqloadServerID == 0 {
		return "", fmt.Errorf("uqload server not found")
	}

	apiURL := fmt.Sprintf("https://v2.hyperwatching.com/embed/%s/server/%d/url", pageData.Props.Video.Hashid, uqloadServerID)
	resp2, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp2.Body.Close()

	var apiData struct {
		WatchURL string `json:"watch_url"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&apiData); err != nil {
		return "", err
	}

	parsed, err := url.Parse(apiData.WatchURL)
	if err != nil {
		return "", err
	}

	idParam := parsed.Query().Get("id")
	if idParam != "" {
		return idParam, nil
	}

	return "", fmt.Errorf("could not extract uqload embed URL")
}

func GetUqloadM3U8(uqloadEmbedURL string) (string, error) {
	resp, err := http.Get(uqloadEmbedURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	re := regexp.MustCompile(`eval\(function\(p,a,c,k,e,d\).*?return p}\('(.*?)',(\d+),(\d+),'(.*?)'\.split`)
	match := re.FindStringSubmatch(html)
	if len(match) == 0 {
		return "", fmt.Errorf("uqload eval block not found")
	}

	p := match[1]
	a, _ := strconv.Atoi(match[2])
	c, _ := strconv.Atoi(match[3])
	k := strings.Split(match[4], "|")

	baseN := func(num int, b int) string {
		if num == 0 {
			return "0"
		}
		res := ""
		for num > 0 {
			remainder := num % b
			if remainder > 35 {
				res = string(rune(remainder + 29)) + res
			} else if remainder > 9 {
				res = string(rune(remainder + 87)) + res
			} else {
				res = strconv.Itoa(remainder) + res
			}
			num /= b
		}
		return res
	}

	for i := c - 1; i >= 0; i-- {
		if k[i] != "" {
			escapedK := strings.ReplaceAll(k[i], "\\", "\\\\")
			p = regexp.MustCompile(`\b`+baseN(i, a)+`\b`).ReplaceAllString(p, escapedK)
		}
	}

	m3u8Re := regexp.MustCompile(`file\s*:\s*["'](https?://[^"']+\.m3u8[^"']*)["']`)
	m3match := m3u8Re.FindStringSubmatch(p)
	if len(m3match) > 1 {
		return m3match[1], nil
	}
	return "", fmt.Errorf("m3u8 not found in unpacked JS")
}

func DownloadM3U8(m3u8URL string) ([]byte, error) {
	tmpFile := "/tmp/stardima_" + strconv.FormatInt(time.Now().UnixNano(), 10) + ".mp4"
	defer os.Remove(tmpFile)

	// Use yt-dlp to download m3u8
	cmd := exec.Command("yt-dlp", m3u8URL, "-o", tmpFile)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("yt-dlp failed: %v", err)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read output file: %v", err)
	}
	return data, nil
}

func GetStardimaFullList(category string) ([]string, error) {
	// First fetch page 1 to get total pages
	reqURL := fmt.Sprintf("https://stardima-37.cartoon.com.im/%s?page=1", category)
	req, _ := http.NewRequest("GET", reqURL, nil)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	
	var data struct {
		Pagination struct {
			LastPage int `json:"last_page"`
		} `json:"pagination"`
		Videos []StardimaVideo `json:"videos"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&data)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	totalPages := data.Pagination.LastPage
	if totalPages > 50 {
		totalPages = 50 // Cap at 50 pages (750 items) to avoid too much spam/WhatsApp limit
	}

	var allTitles []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// We already have page 1
	for _, v := range data.Videos {
		allTitles = append(allTitles, v.Title)
	}

	for i := 2; i <= totalPages; i++ {
		wg.Add(1)
		go func(page int) {
			defer wg.Done()
			r, _ := http.NewRequest("GET", fmt.Sprintf("https://stardima-37.cartoon.com.im/%s?page=%d", category, page), nil)
			r.Header.Set("X-Requested-With", "XMLHttpRequest")
			res, e := http.DefaultClient.Do(r)
			if e == nil {
				defer res.Body.Close()
				var d struct {
					Videos []StardimaVideo `json:"videos"`
				}
				if json.NewDecoder(res.Body).Decode(&d) == nil {
					mu.Lock()
					for _, v := range d.Videos {
						allTitles = append(allTitles, v.Title)
					}
					mu.Unlock()
				}
			}
		}(i)
	}
	wg.Wait()

	return allTitles, nil
}
