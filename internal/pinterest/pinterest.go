package pinterest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

type PendingRequest struct {
	Query string
	Count int
}

var (
	PendingRequests = make(map[string]PendingRequest)
	pendingMutex    sync.RWMutex
)

func SetPending(chatID, query string, count int) {
	pendingMutex.Lock()
	defer pendingMutex.Unlock()
	PendingRequests[chatID] = PendingRequest{Query: query, Count: count}
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
	req, _ := http.NewRequest("GET", "https://duckduckgo.com/?q="+query, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	bodyStr := string(body)

	re := regexp.MustCompile(`vqd=([a-zA-Z0-9_-]+)`)
	matches := re.FindStringSubmatch(bodyStr)
	if len(matches) < 2 {
		return nil
	}
	vqd := matches[1]

	req2, _ := http.NewRequest("GET", "https://duckduckgo.com/i.js?l=us-en&o=json&q="+query+"&vqd="+vqd+"&f=,,,&p=1", nil)
	req2.Header.Set("User-Agent", "Mozilla/5.0")
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
					URL:    item.Image,
					Title:  item.Title,
					PinURL: item.URL,
				})
			}
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
