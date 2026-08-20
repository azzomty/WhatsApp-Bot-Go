package pinterest

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"regexp"
	"net/textproto"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type PinResult struct {
	ID string

	Title string
	URL   string
}

type PendingRequest struct {
	Query       string
	Count       int
	IsVisual    bool
	Base64Image string
	Bookmark    string
}

var (
	PendingRequests = make(map[string]PendingRequest)
	pendingMutex    sync.RWMutex

	LastSearches = make(map[string]SearchState)
	searchMutex  sync.RWMutex

	// Store WhatsApp Message ID -> Pinterest Pin ID mapping for reactions
	MessagePinMap = make(map[string]string)
	mapMutex      sync.Mutex
)

// SearchState holds the state for continuing a search
type SearchState struct {
	Query       string
	Aspect      string
	Count       int
	IsVisual    bool
	Base64Image string
	Bookmark    string
}

func SetLastSearch(chatID, query, aspect string, count int, isVisual bool, base64Image string, bookmark string) {
	searchMutex.Lock()
	defer searchMutex.Unlock()
	LastSearches[chatID] = SearchState{Query: query, Aspect: aspect, Count: count, IsVisual: isVisual, Base64Image: base64Image, Bookmark: bookmark}
}

func GetLastSearch(chatID string) (SearchState, bool) {
	searchMutex.RLock()
	defer searchMutex.RUnlock()
	req, ok := LastSearches[chatID]
	return req, ok
}

func ClearPending(chatID string) {
	pendingMutex.Lock()
	defer pendingMutex.Unlock()
	delete(PendingRequests, chatID)
}

func SetPending(chatID, query string, count int, isVisual bool, base64Image string, bookmark string) {
	pendingMutex.Lock()
	defer pendingMutex.Unlock()
	PendingRequests[chatID] = PendingRequest{Query: query, Count: count, IsVisual: isVisual, Base64Image: base64Image, Bookmark: bookmark}
}

func GetPending(chatID string) (PendingRequest, bool) {
	pendingMutex.RLock()
	defer pendingMutex.RUnlock()
	req, ok := PendingRequests[chatID]
	return req, ok
}

func SaveMessagePin(msgID string, pinID string) {
	mapMutex.Lock()
	defer mapMutex.Unlock()
	MessagePinMap[msgID] = pinID
}

func GetMessagePin(msgID string) (string, bool) {
	mapMutex.Lock()
	defer mapMutex.Unlock()
	val, ok := MessagePinMap[msgID]
	return val, ok
}

func setPinterestHeaders(req *http.Request) {
	rawToken := os.Getenv("PINTEREST_TOKEN")
	// Clean the token robustly
	cleanToken := ""
	if strings.Contains(rawToken, "pina_") {
		parts := strings.Split(rawToken, "pina_")
		cleaned := strings.ReplaceAll(strings.ReplaceAll(parts[1], "\n", ""), "\r", "")
		cleaned = strings.ReplaceAll(cleaned, " ", "")
		cleanToken = "Bearer pina_" + cleaned
	} else {
		cleanToken = strings.TrimSpace(rawToken)
	}

	cookie := strings.ReplaceAll(strings.ReplaceAll(os.Getenv("PINTEREST_COOKIE"), "\n", ""), "\r", "")
	cookie = strings.ReplaceAll(cookie, "  ", "")

	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("Authorization", cleanToken)
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Host", "api.pinterest.com")
	req.Header.Set("User-Agent", "Pinterest for Android Tablet/14.23.2 (Nexus 10; 11)")
	req.Header.Set("X-Pinterest-AppState", "active")
	req.Header.Set("X-Pinterest-Device", "Nexus 10")
	req.Header.Set("X-Pinterest-Device-Manufacturer", "Genymobile")
	req.Header.Set("X-Pinterest-InstallId", "29ac4b08d4c84efebbb95ac02cdd308")
	req.Header.Set("X-Pinterest-WebView-Supported", "false")
}

func extractMediaByExtension(data interface{}, ext string) string {
	switch v := data.(type) {
	case map[string]interface{}:
		if u, ok := v["url"].(string); ok && strings.HasSuffix(strings.ToLower(u), ext) {
			return u
		}
		for _, val := range v {
			if res := extractMediaByExtension(val, ext); res != "" {
				return res
			}
		}
	case []interface{}:
		for _, val := range v {
			if res := extractMediaByExtension(val, ext); res != "" {
				return res
			}
		}
	}
	return ""
}

func SearchTenorGifs(query string, count int) []PinResult {
	query = url.QueryEscape(query)
	req, _ := http.NewRequest("GET", "https://tenor.com/search/"+query+"-gifs", nil)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	
	body, _ := ioutil.ReadAll(resp.Body)
	mp4Regex := regexp.MustCompile(`https://media.tenor.com/[^"]*\.mp4`)
	matches := mp4Regex.FindAllString(string(body), -1)
	
	var results []PinResult
	unique := make(map[string]bool)
	for _, m := range matches {
		if !unique[m] {
			unique[m] = true
			results = append(results, PinResult{Title: "GIF", URL: m})
			if len(results) >= count {
				break
			}
		}
	}
	return results
}

func SearchPinterestGifs(query string, count int) []PinResult {
	pins, _ := SearchPinterest(query+" gif", "all", count+20, "")
	if len(pins) == 0 {
		return nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	var results []PinResult
	var wg sync.WaitGroup
	var mu sync.Mutex

	max := count + 10
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

			var respJson interface{}
			json.Unmarshal(bodyBytes, &respJson)

			mp4Url := extractMediaByExtension(respJson, ".mp4")

			if mp4Url != "" {
				mu.Lock()
				if len(results) < count {
					results = append(results, PinResult{ID: id, Title: "GIF", URL: mp4Url})
				}
				mu.Unlock()
			}
		}(p.ID)
	}
	wg.Wait()
	return results
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

		var imgUrl string
		var w, h float64

		if aspect == "video" {
			imgUrl = extractMediaByExtension(pin, ".mp4")
		} else if aspect == "gif" {
			imgUrl = extractMediaByExtension(pin, ".gif")
			// Sometimes GIFs are just labeled .gif in images.originals.url, which is also caught
		}

		if imgUrl == "" {
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
		}

		// Fallback for search/pins/ which returns flat image_large_url
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
				if w == 0 {
					if sizeData, ok := pin["image_large_size_points"].(map[string]interface{}); ok {
						if width, ok := sizeData["width"].(float64); ok {
							w = width
						}
						if height, ok := sizeData["height"].(float64); ok {
							h = height
						}
					}
				}
			} else if u, ok := pin["image_medium_url"].(string); ok {
				imgUrl = strings.Replace(u, "474x", "736x", 1) // Try to get higher res
				if sizeData, ok := pin["image_medium_size_pixels"].(map[string]interface{}); ok {
					if width, ok := sizeData["width"].(float64); ok {
						w = width
					}
					if height, ok := sizeData["height"].(float64); ok {
						h = height
					}
				}
				if w == 0 {
					if sizeData, ok := pin["image_medium_size_points"].(map[string]interface{}); ok {
						if width, ok := sizeData["width"].(float64); ok {
							w = width
						}
						if height, ok := sizeData["height"].(float64); ok {
							h = height
						}
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
	// For ForYou feed, sometimes it's under data[0] depending on API v3
	return data
}

func SearchPinterest(query string, aspect string, count int, startBookmark string) ([]PinResult, string) {
	query = url.QueryEscape(query)
	var allPins []PinResult
	currentBookmark := startBookmark

	// We need to fetch enough pages to satisfy `count` (up to 100 max for safety)
	for len(allPins) < count {
		searchUrl := fmt.Sprintf("https://api.pinterest.com/v3/search/pins/?rs=typed&pinrep_img_width=474x&query=%s&page_size=250", query)
		if currentBookmark != "" {
			searchUrl += "&bookmark=" + url.QueryEscape(currentBookmark)
		}

		req, _ := http.NewRequest("GET", searchUrl, nil)
		setPinterestHeaders(req)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			break
		}

		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

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

		newPins := parsePinterestData(data, aspect)
		allPins = append(allPins, newPins...)

		if bm, ok := respJson["bookmark"].(string); ok && bm != "" && bm != currentBookmark {
			currentBookmark = bm
		} else {
			break // No more pages available
		}

		if len(newPins) == 0 {
			break // No more results
		}
		
		// Stop if we fetched too many to prevent infinite loops
		if len(allPins) > 300 {
			break
		}
	}

	return allPins, currentBookmark
}

func GetRelatedPins(pinID string) []PinResult {
	searchUrl := fmt.Sprintf("https://api.pinterest.com/v3/pins/%s/related/pins/?pinrep_img_width=474x", pinID)
	req, _ := http.NewRequest("GET", searchUrl, nil)
	setPinterestHeaders(req)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	// The API returns the pins inside data
	return parsePinterestData(extractDataFromJSON(bodyBytes), "all")
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
	// First get search results normally
	pins, _ := SearchPinterest("matching icons "+query, "all", 10, "")
	if len(pins) == 0 {
		return nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
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
			req, _ := http.NewRequest("GET", "https://api.pinterest.com/v3/pins/"+id+"/?fields=carousel_data,story_pin_data,images,image_large_url,image_medium_url", nil)
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
				var urls []string

				if cd, ok := data["carousel_data"].(map[string]interface{}); ok {
					if slots, ok := cd["carousel_slots"].([]interface{}); ok && len(slots) >= 2 {
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
					}
				}

				if len(urls) < 2 {
					urls = nil // reset if we didn't find 2
					if spd, ok := data["story_pin_data"].(map[string]interface{}); ok {
						if pages, ok := spd["pages"].([]interface{}); ok && len(pages) >= 2 {
							for i := 0; i < 2; i++ {
								if page, ok := pages[i].(map[string]interface{}); ok {
									if blocks, ok := page["blocks"].([]interface{}); ok && len(blocks) > 0 {
										if block, ok := blocks[0].(map[string]interface{}); ok {
											if image, ok := block["image"].(map[string]interface{}); ok {
												for _, key := range []string{"originals", "orig", "1200x", "736x", "474x"} {
													if imgData, ok := image[key].(map[string]interface{}); ok {
														if u, ok := imgData["url"].(string); ok {
															urls = append(urls, u)
															break
														}
													}
												}
											}
										}
									}
									if len(urls) <= i { // if not found in blocks, try directly on page
										if image, ok := page["image"].(map[string]interface{}); ok {
											for _, key := range []string{"originals", "orig", "1200x", "736x", "474x"} {
												if imgData, ok := image[key].(map[string]interface{}); ok {
													if u, ok := imgData["url"].(string); ok {
														urls = append(urls, u)
														break
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}

				if len(urls) < 2 {
					var singleUrl string
					if images, ok := data["images"].(map[string]interface{}); ok {
						for _, key := range []string{"originals", "orig", "1200x", "736x", "474x"} {
							if imgData, ok := images[key].(map[string]interface{}); ok {
								if u, ok := imgData["url"].(string); ok {
									singleUrl = u
									break
								}
							}
						}
					}
					if singleUrl == "" {
						if u, ok := data["image_large_url"].(string); ok {
							singleUrl = u
						} else if u, ok := data["image_medium_url"].(string); ok {
							singleUrl = strings.Replace(u, "474x", "736x", 1)
						}
					}
					if singleUrl != "" {
						urls = []string{singleUrl, "https://i.imgur.com/RoC2Bv6.png"} // White image fallback
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
		}(p.ID)
	}
	wg.Wait()
	return results
}

func SearchPinterestLens(base64Image string, aspect string, count int) []PinResult {
	imageBytes, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		fmt.Println("Base64 decode error:", err)
		return nil
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("camera_type", "0")
	writer.WriteField("source_type", "1")
	
	// Request slightly more than count to have enough valid results after filtering
	pageSize := count + 10
	if pageSize > 100 {
		pageSize = 100
	}
	writer.WriteField("page_size", fmt.Sprintf("%d", pageSize))

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="image"; filename="null"`)
	h.Set("Content-Type", "application/octet-stream")
	part, err := writer.CreatePart(h)
	if err == nil {
		part.Write(imageBytes)
	}
	writer.Close()

	req, _ := http.NewRequest("POST", "https://api.pinterest.com/v3/visual_search/lens/search/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	setPinterestHeaders(req)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Lens search error:", err)
		return nil
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	data := extractDataFromJSON(bodyBytes)

	return parsePinterestData(data, aspect)
}

func GetMatchingPairs(results []PinResult, targetPairs int) []string {
	var urls []string
	for _, p := range results {
		urls = append(urls, p.URL)
		if len(urls) >= targetPairs*2 {
			break
		}
	}
	return urls
}

func DownloadImage(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}
