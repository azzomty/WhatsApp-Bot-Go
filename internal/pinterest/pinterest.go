package pinterest

import (
	"bytes"
	"encoding/json"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type PendingRequest struct {
	Query       string
	Count       int
	IsVisual    bool
	Base64Image string
}

var (
	PendingRequests = make(map[string]PendingRequest)
	LastSearches    = make(map[string]LastSearch)
	pendingMutex    sync.RWMutex
	lastSearchMutex sync.RWMutex
)

type LastSearch struct {
	Query       string
	Aspect      string
	Count       int
	IsVisual    bool
	Base64Image string
}

func SetLastSearch(chatID, query, aspect string, count int, isVisual bool, base64Image string) {
	lastSearchMutex.Lock()
	defer lastSearchMutex.Unlock()
	LastSearches[chatID] = LastSearch{Query: query, Aspect: aspect, Count: count, IsVisual: isVisual, Base64Image: base64Image}
}

func GetLastSearch(chatID string) (LastSearch, bool) {
	lastSearchMutex.RLock()
	defer lastSearchMutex.RUnlock()
	req, ok := LastSearches[chatID]
	return req, ok
}

func SetPending(chatID, query string, count int, isVisual bool, base64Image string) {
	pendingMutex.Lock()
	defer pendingMutex.Unlock()
	PendingRequests[chatID] = PendingRequest{Query: query, Count: count, IsVisual: isVisual, Base64Image: base64Image}
}

func GetPending(chatID string) (PendingRequest, bool) {
	pendingMutex.RLock()
	defer pendingMutex.RUnlock()
	req, ok := PendingRequests[chatID]
	return req, ok
}

func ClearPending(chatID string) {
	pendingMutex.Lock()
	delete(PendingRequests, chatID)
	pendingMutex.Unlock()
}

func DownloadImage(url string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

type PinResult struct {
	URL    string
	Title  string
	PinURL string
}

type DDGResponse struct {
	Results []struct {
		Image  string `json:"image"`
		Title  string `json:"title"`
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"results"`
}

func SearchPinterest(query string, aspect string) []PinResult {
	query = url.QueryEscape(query + " site:pinterest.com")
	var vqd string
	client := &http.Client{}
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("GET", "https://duckduckgo.com/?q="+query, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		bodyStr := string(body)

		re := regexp.MustCompile(`vqd=["']?([a-zA-Z0-9_-]+)["']?`)
		matches := re.FindStringSubmatch(bodyStr)
		if len(matches) >= 2 {
			vqd = matches[1]
			break
		}
		time.Sleep(1 * time.Second)
	}

	if vqd == "" {
		return nil
	}

	req2, _ := http.NewRequest("GET", "https://duckduckgo.com/i.js?l=us-en&o=json&q="+query+"&vqd="+vqd+"&f=,,,&p=1", nil)
	req2.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req2.Header.Set("Referer", "https://duckduckgo.com/")
	resp2, err := client.Do(req2)
	if err != nil {
		return nil
	}
	defer resp2.Body.Close()

	var ddgResp DDGResponse
	if err := json.NewDecoder(resp2.Body).Decode(&ddgResp); err != nil {
		return nil
	}

	var results []PinResult
	for _, item := range ddgResp.Results {
		if strings.Contains(item.Image, "pinimg.com") {
			w := float64(item.Width)
			h := float64(item.Height)
			if w == 0 {
				w = 1
			}
			if h == 0 {
				h = 1
			}
			ratio := w / h

			keep := true
			if aspect == "icon" && (ratio < 0.7 || ratio > 1.3) {
				keep = false
			}
			if aspect == "wallpaper" && ratio > 0.9 {
				keep = false
			}
			if aspect == "banner" && ratio < 1.1 {
				keep = false
			}

			if keep {
				results = append(results, PinResult{
					Title: item.Title,
					URL:   item.Image,
				})
			}
		}
	}
	return results
}

func SearchPinterestLens(base64Image string, aspect string) []PinResult {
	pinToken := os.Getenv("PINTEREST_TOKEN")
	pinCookie := os.Getenv("PINTEREST_COOKIE")
	
	if pinToken == "" || pinCookie == "" {
		fmt.Println("PINTEREST_TOKEN or PINTEREST_COOKIE not set")
		return nil
	}

	imageBytes, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		fmt.Println("Base64 decode error:", err)
		return nil
	}

	// Step 1: Upload Image
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "image.jpg")
	if err == nil {
		part.Write(imageBytes)
	}
	writer.Close()

	req1, _ := http.NewRequest("PUT", "https://api.pinterest.com/v3/visual_search/lens/history/", body)
	req1.Header.Set("Content-Type", writer.FormDataContentType())
	req1.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req1.Header.Set("Accept-Language", "en-US")
	req1.Header.Set("Authorization", pinToken)
	req1.Header.Set("Cookie", pinCookie)
	req1.Header.Set("Host", "api.pinterest.com")
	req1.Header.Set("User-Agent", "Pinterest for Android Tablet/14.23.2 (Nexus 10; 11)")
	req1.Header.Set("X-Pinterest-AppState", "active")
	req1.Header.Set("X-Pinterest-Device", "Nexus 10")
	req1.Header.Set("X-Pinterest-Device-Manufacturer", "Genymobile")
	req1.Header.Set("X-Pinterest-InstallId", "29ac4b08d4c84efebbb95ac02cdd308")
	req1.Header.Set("X-Pinterest-WebView-Supported", "false")

	client := &http.Client{Timeout: 30 * time.Second}
	resp1, err := client.Do(req1)
	if err != nil {
		fmt.Println("Lens upload error:", err)
		return nil
	}
	defer resp1.Body.Close()

	resp1Bytes, _ := ioutil.ReadAll(resp1.Body)
	var resp1Json map[string]interface{}
	json.Unmarshal(resp1Bytes, &resp1Json)

	s3Url := ""
	if data, ok := resp1Json["data"].(map[string]interface{}); ok {
		if u, ok := data["image_url"].(string); ok && strings.HasPrefix(u, "s3://") {
			s3Url = u
		}
	}

	if s3Url == "" {
		fmt.Println("Could not extract s3_url from lens history")
		return nil
	}

	// Step 2: Search
	searchUrl := fmt.Sprintf("https://api.pinterest.com/v3/visual_search/lens/search/?camera_type=0&source_type=1&url=%s&page_size=24", s3Url)
	req2, _ := http.NewRequest("GET", searchUrl, nil)
	req2.Header = req1.Header // reuse headers
	req2.Header.Set("Content-Type", "") // remove content type

	resp2, err := client.Do(req2)
	if err != nil {
		fmt.Println("Lens search error:", err)
		return nil
	}
	defer resp2.Body.Close()

	resp2Bytes, _ := ioutil.ReadAll(resp2.Body)
	var resp2Json map[string]interface{}
	json.Unmarshal(resp2Bytes, &resp2Json)

	var results []PinResult
	data, ok := resp2Json["data"].([]interface{})
	if !ok {
		// some pinterest endpoints wrap it
		if d, ok := resp2Json["data"].(map[string]interface{}); ok {
			if res, ok := d["results"].([]interface{}); ok {
				data = res
			}
		}
	}

	for _, item := range data {
		pin, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var imgUrl string
		var w, h float64
		
		if images, ok := pin["images"].(map[string]interface{}); ok {
			for _, key := range []string{"originals", "orig", "736x", "474x"} {
				if imgData, ok := images[key].(map[string]interface{}); ok {
					if u, ok := imgData["url"].(string); ok {
						imgUrl = u
						if width, ok := imgData["width"].(float64); ok {
							w = width
						}
						if height, ok := imgData["height"].(float64); ok {
							h = height
						}
						break
					}
				}
			}
		}

		if imgUrl == "" {
			continue
		}
		
		if w == 0 {
			w = 1
		}
		if h == 0 {
			h = 1
		}
		ratio := w / h

		keep := true
		if aspect == "icon" && (ratio < 0.7 || ratio > 1.3) {
			keep = false
		}
		if aspect == "wallpaper" && ratio > 0.9 {
			keep = false
		}
		if aspect == "banner" && ratio < 1.1 {
			keep = false
		}

		if keep {
			title := ""
			if t, ok := pin["title"].(string); ok {
				title = t
			} else if desc, ok := pin["description"].(string); ok {
				title = desc
			}
			results = append(results, PinResult{
				Title: title,
				URL:   imgUrl,
			})
		}
	}
	
	return results
}

func GetMatchingPairs(results []PinResult, targetPairs int) []string {
	var pairs []string
	sentUrls := make(map[string]bool)
	pairsFound := 0

	// Strategy 1: Find Carousels from HTML
	for i := 0; i < len(results) && i < 15; i++ {
		if pairsFound >= targetPairs {
			break
		}
		pinUrl := results[i].PinURL
		if pinUrl == "" || !strings.Contains(pinUrl, "pinterest.com/pin/") {
			continue
		}

		req, _ := http.NewRequest("GET", pinUrl, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		re := regexp.MustCompile(`https://i\.pinimg\.com/originals/[a-zA-Z0-9/_-]+\.(?:jpg|png|jpeg)`)
		imgMatches := re.FindAllString(string(body), -1)

		if len(imgMatches) > 0 {
			var uniqueImages []string
			seen := make(map[string]bool)
			for _, img := range imgMatches {
				// Ignore Pinterest site assets that falsely trigger carousel logic
				if strings.Contains(img, "d53b014d86a6b6761bf649a0ed813c2b") {
					continue
				}
				if !seen[img] {
					seen[img] = true
					uniqueImages = append(uniqueImages, img)
				}
			}

			if len(uniqueImages) >= 2 {
				pairs = append(pairs, uniqueImages[0], uniqueImages[1])
				sentUrls[uniqueImages[0]] = true
				sentUrls[uniqueImages[1]] = true
				pairsFound++
			}
		}
	}

	// Strategy 2: Grouping by PinURL (this guarantees they are from the exact same post)
	if pairsFound < targetPairs && len(results) >= 2 {
		pinMap := make(map[string][]PinResult)

		for _, u := range results {
			if sentUrls[u.URL] {
				continue
			}

			// Group by the actual Pinterest Pin URL
			if u.PinURL != "" {
				exists := false
				for _, x := range pinMap[u.PinURL] {
					if x.URL == u.URL {
						exists = true
						break
					}
				}
				if !exists {
					pinMap[u.PinURL] = append(pinMap[u.PinURL], u)
				}
			}
		}

		// Send valid pairs from the same Pin
		for _, items := range pinMap {
			if pairsFound >= targetPairs {
				break
			}
			if len(items) >= 2 {
				pairs = append(pairs, items[0].URL, items[1].URL)
				sentUrls[items[0].URL] = true
				sentUrls[items[1].URL] = true
				pairsFound++
			}
		}
	}

	return pairs
}
