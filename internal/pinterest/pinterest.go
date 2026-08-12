package pinterest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type PinResult struct {
	ID     string
	Title  string
	URL    string
	PinURL string
}

type PendingRequest struct {
	Query       string
	Count       int
	IsVisual    bool
	Base64Image string
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

var UserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
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

func ClearPending(chatID string) {
	pendingMutex.Lock()
	defer pendingMutex.Unlock()
	delete(PendingRequests, chatID)
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

func setPinterestHeaders(req *http.Request) {
	token := os.Getenv("PINTEREST_TOKEN")
	if token == "" {
		token = "Bearer pina" + "_AEATFWAVAAPDYAQAGBAGGD6W3DSPDHYBABHO2LFZZGGZJ4ODDM46P5VVRHTQLEQMNJUIUZ6N4LWYFSV3HCGCNBRMWQYOJMAA"
	}
	cookie := os.Getenv("PINTEREST_COOKIE")
	if cookie == "" {
		cookie = "_b=AZehPVTHje5FSKPWa+hL4qmM/XEDJuxk13yIX8h3VBWeJwNgD6CaB3qWfEhPQT8YcaY=; _pinterest_ct=TWc9PSZnZWpBakE1TFQzdkViSURTRTN5VkNqRjZtMUdjeDU1SEpONzNZU0dVc0w2S2ZXVGZTeFNqNVJOSkF4UTFFMUVwaXcrWUZyczl3UmJrdEdSeHMrcHcyc0NuTEQ4NXBPdkdKemVGcG1hVm43OD0mM240em12M0R4eEJNZ2d4YjZLaVZtUHpIWXI0PQ==; _ir=0"
	}

	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("Authorization", token)
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Host", "api.pinterest.com")
	req.Header.Set("User-Agent", "Pinterest for Android Tablet/14.23.2 (Nexus 10; 11)")
	req.Header.Set("X-Pinterest-AppState", "active")
	req.Header.Set("X-Pinterest-Device", "Nexus 10")
	req.Header.Set("X-Pinterest-Device-Manufacturer", "Genymobile")
	req.Header.Set("X-Pinterest-InstallId", "29ac4b08d4c84efebbb95ac02cdd308")
	req.Header.Set("X-Pinterest-WebView-Supported", "false")
}

func parsePinterestData(data []interface{}, aspect string) []PinResult {
	var results []PinResult
	for _, item := range data {
		pin, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		if promoted, ok := pin["is_promoted"].(bool); ok && promoted {
			continue
		}
		if typeStr, ok := pin["type"].(string); ok && typeStr == "ad" {
			continue
		}
		if _, hasPromoter := pin["promoter"]; hasPromoter {
			continue
		}
		if _, hasAdMatch := pin["ad_match_reason"]; hasAdMatch {
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
			if u, ok := pin["image_large_url"].(string); ok {
				imgUrl = u
				if sizeData, ok := pin["image_large_size_pixels"].(map[string]interface{}); ok {
					if width, ok := sizeData["width"].(float64); ok {
						w = width
					}
					if height, ok := sizeData["height"].(float64); ok {
						h = height
					}
				}
			} else if u, ok := pin["image_medium_url"].(string); ok {
				imgUrl = strings.Replace(u, "474x", "736x", 1) 
				if sizeData, ok := pin["image_medium_size_pixels"].(map[string]interface{}); ok {
					if width, ok := sizeData["width"].(float64); ok {
						w = width
					}
					if height, ok := sizeData["height"].(float64); ok {
						h = height
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
			
			id := ""
			if idStr, ok := pin["id"].(string); ok {
				id = idStr
			}

			results = append(results, PinResult{
				ID:    id,
				Title: title,
				URL:   imgUrl,
			})
		}
	}
	return results
}

func extractDataFromJSON(bodyBytes []byte) []interface{} {
	var respJson map[string]interface{}
	json.Unmarshal(bodyBytes, &respJson)

	var data []interface{}
	if d, ok := respJson["data"].([]interface{}); ok {
		data = d
	} else if d, ok := respJson["data"].(map[string]interface{}); ok {
		if res, ok := d["results"].([]interface{}); ok {
			data = res
		}
	}
	return data
}

func SearchPinterest(query string, aspect string) []PinResult {
	// V3 Native API Method
	q := url.QueryEscape(query)
	searchUrl := fmt.Sprintf("https://api.pinterest.com/v3/search/pins/?rs=typed&pinrep_img_width=474x&query=%s", q)
	req, _ := http.NewRequest("GET", searchUrl, nil)
	setPinterestHeaders(req)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		data := extractDataFromJSON(bodyBytes)
		results := parsePinterestData(data, aspect)
		if len(results) > 0 {
			return results
		}
	}

	q = url.QueryEscape(query + " site:pinterest.com")
	req, _ = http.NewRequest("GET", "https://duckduckgo.com/?q="+q, nil)
	req.Header.Set("User-Agent", UserAgents[rand.Intn(len(UserAgents))])

	client = &http.Client{Timeout: 10 * time.Second}
	resp, err = client.Do(req)
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

	req2, _ := http.NewRequest("GET", "https://duckduckgo.com/i.js?l=us-en&o=json&q="+q+"&vqd="+vqd+"&f=,,,&p=1", nil)
	req2.Header.Set("User-Agent", UserAgents[rand.Intn(len(UserAgents))])
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
			if w == 0 { w = 1 }
			if h == 0 { h = 1 }
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

func ForYouPinterest(aspect string) []PinResult {
	searchUrl := "https://api.pinterest.com/v3/feeds/home/?item_count=0&pinrep_img_width=474x"
	req, _ := http.NewRequest("GET", searchUrl, nil)
	setPinterestHeaders(req)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	data := extractDataFromJSON(bodyBytes)
	
	return parsePinterestData(data, aspect)
}

func SearchPinterestMatchingIcons(query string) []PinResult {
	// V3 Native API Method
	q := url.QueryEscape("matching icons " + query)
	searchUrl := fmt.Sprintf("https://api.pinterest.com/v3/search/pins/?rs=typed&pinrep_img_width=474x&query=%s", q)
	req, _ := http.NewRequest("GET", searchUrl, nil)
	setPinterestHeaders(req)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		data := extractDataFromJSON(bodyBytes)
		pins := parsePinterestData(data, "all")
		
		var results []PinResult
		var wg sync.WaitGroup
		var mu sync.Mutex

		max := 10
		if len(pins) < max {
			max = len(pins)
		}

			for i := 0; i < max; i++ {
				p := pins[i]
				if p.ID == "" {
					continue
				}
				
				wg.Add(1)
				go func(id string) {
					defer wg.Done()
					req, _ := http.NewRequest("GET", "https://api.pinterest.com/v3/pins/"+id+"/", nil)
					setPinterestHeaders(req)
					
					resp, err := client.Do(req)
					if err != nil {
						return
					}
					
					bodyBytes, _ := ioutil.ReadAll(resp.Body)
					resp.Body.Close()
					
					var respJson map[string]interface{}
					json.Unmarshal(bodyBytes, &respJson)
					
					if data, ok := respJson["data"].(map[string]interface{}); ok {
						if cd, ok := data["carousel_data"].(map[string]interface{}); ok {
							if slots, ok := cd["carousel_slots"].([]interface{}); ok && len(slots) >= 2 {
								var urls []string
								for i := 0; i < 2; i++ {
									if slot, ok := slots[i].(map[string]interface{}); ok {
										if images, ok := slot["images"].(map[string]interface{}); ok {
											for _, key := range []string{"originals", "orig", "736x", "474x"} {
												if imgData, ok := images[key].(map[string]interface{}); ok {
													if u, ok := imgData["url"].(string); ok {
														urls = append(urls, u)
														break
													}
												}
											}
										}
									}
								}
								if len(urls) >= 2 {
									mu.Lock()
									if len(results) == 0 {
										results = append(results, PinResult{Title: "Matching Left", URL: urls[0]})
										results = append(results, PinResult{Title: "Matching Right", URL: urls[1]})
									}
									mu.Unlock()
								}
							}
						}
					}
				}(p.ID)
			}
			wg.Wait()
		if len(results) > 0 {
			return results
		}
	}
	
	// Fallback to DuckDuckGo Method
	pins := SearchPinterest("matching icons "+query, "all")
	if len(pins) == 0 {
		return nil
	}
	
	pairs := GetMatchingPairs(pins, 1)
	var results []PinResult
	for _, url := range pairs {
		results = append(results, PinResult{URL: url})
	}
	return results
}

func SearchPinterestLens(base64Image string, aspect string) []PinResult {
	apiKey := os.Getenv("RAPIDAPI_KEY")
	if apiKey == "" {
		fmt.Println("RAPIDAPI_KEY not set")
		return nil
	}

	apiUrl := "https://pinterest-lens-reverse-image-search-api.p.rapidapi.com/search"
	
	payloadData := map[string]string{
		"image_base64": base64Image,
	}
	payloadBytes, _ := json.Marshal(payloadData)

	req, _ := http.NewRequest("POST", apiUrl, bytes.NewBuffer(payloadBytes))
	req.Header.Add("x-rapidapi-key", apiKey)
	req.Header.Add("x-rapidapi-host", "pinterest-lens-reverse-image-search-api.p.rapidapi.com")
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("RapidAPI Request Error:", err)
		return nil
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyStr := string(bodyBytes)

	// Extract any pinimg.com URLs from the raw JSON response
	re := regexp.MustCompile(`https?://i\.pinimg\.com/[a-zA-Z0-9_/-]+\.(jpg|jpeg|png)`)
	matches := re.FindAllString(bodyStr, -1)
	
	// Remove duplicates
	uniqueURLs := make(map[string]bool)
	var results []PinResult
	
	for _, match := range matches {
		// Try to only keep originals or 736x for high quality
		if strings.Contains(match, "originals") || strings.Contains(match, "736x") {
			if !uniqueURLs[match] {
				uniqueURLs[match] = true
				results = append(results, PinResult{
					Title: "Pinterest Visual Search",
					URL:   match,
				})
			}
		}
	}
	
	if len(results) == 0 {
		for _, match := range matches {
			if !uniqueURLs[match] {
				uniqueURLs[match] = true
				results = append(results, PinResult{
					Title: "Pinterest Visual Search",
					URL:   match,
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

	// Strategy 2: Grouping by PinURL
	if pairsFound < targetPairs && len(results) >= 2 {
		pinMap := make(map[string][]PinResult)

		for _, u := range results {
			if sentUrls[u.URL] {
				continue
			}

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

func DownloadImage(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
