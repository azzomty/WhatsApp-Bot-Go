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
	"strings"
	"sync"
	"time"
)

type PinResult struct {
	Title string
	URL   string
}

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
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("Authorization", os.Getenv("PINTEREST_TOKEN"))
	req.Header.Set("Cookie", os.Getenv("PINTEREST_COOKIE"))
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

func SearchPinterest(query string, aspect string) []PinResult {
	query = url.QueryEscape(query)
	searchUrl := fmt.Sprintf("https://api.pinterest.com/v3/search/pins/?rs=typed&pinrep_img_width=474x&query=%s", query)
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
	query = url.QueryEscape("matching icons " + query)
	searchUrl := fmt.Sprintf("https://api.pinterest.com/v3/search/pins/?rs=typed&pinrep_img_width=474x&query=%s", query)
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
	
	var results []PinResult
	for _, item := range data {
		pin, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		
		// Look for carousel_data
		if cd, ok := pin["carousel_data"].(map[string]interface{}); ok {
			if slots, ok := cd["carousel_slots"].([]interface{}); ok && len(slots) >= 2 {
				// extract the two images
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
					results = append(results, PinResult{Title: "Matching Left", URL: urls[0]})
					results = append(results, PinResult{Title: "Matching Right", URL: urls[1]})
					return results // only return the first valid pair!
				}
			}
		} else if images, ok := pin["images"].([]interface{}); ok && len(images) >= 2 {
			// Some endpoints return a flat array for carousels
			var urls []string
			for i := 0; i < 2; i++ {
				if imgsMap, ok := images[i].(map[string]interface{}); ok {
					for _, key := range []string{"originals", "orig", "736x", "474x"} {
						if imgData, ok := imgsMap[key].(map[string]interface{}); ok {
							if u, ok := imgData["url"].(string); ok {
								urls = append(urls, u)
								break
							}
						}
					}
				}
			}
			if len(urls) >= 2 {
				results = append(results, PinResult{Title: "Matching Left", URL: urls[0]})
				results = append(results, PinResult{Title: "Matching Right", URL: urls[1]})
				return results
			}
		}
	}
	return results
}

func SearchPinterestLens(base64Image string, aspect string) []PinResult {
	imageBytes, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		fmt.Println("Base64 decode error:", err)
		return nil
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "image.jpg")
	if err == nil {
		part.Write(imageBytes)
	}
	writer.Close()

	req1, _ := http.NewRequest("PUT", "https://api.pinterest.com/v3/visual_search/lens/history/", body)
	req1.Header.Set("Content-Type", writer.FormDataContentType())
	setPinterestHeaders(req1)

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

	searchUrl := fmt.Sprintf("https://api.pinterest.com/v3/visual_search/lens/search/?camera_type=0&source_type=1&url=%s&page_size=24", url.QueryEscape(s3Url))
	req2, _ := http.NewRequest("GET", searchUrl, nil)
	setPinterestHeaders(req2)

	resp2, err := client.Do(req2)
	if err != nil {
		fmt.Println("Lens search error:", err)
		return nil
	}
	defer resp2.Body.Close()

	resp2Bytes, _ := ioutil.ReadAll(resp2.Body)
	data := extractDataFromJSON(resp2Bytes)
	
	return parsePinterestData(data, aspect)
}

func GetMatchingPairs(results []PinResult, targetPairs int) []string {
	// Not really needed anymore if SearchPinterestMatchingIcons handles it directly, but keeping it so main.go doesn't break
	return nil
}

func DownloadImage(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
