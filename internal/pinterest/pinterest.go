package pinterest

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type PinResult struct {
	ID    string

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

		if promoted, ok := pin["is_promoted"].(bool); ok && promoted {
			continue
		}
		if typeStr, ok := pin["type"].(string); ok && typeStr == "ad" {
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
	// First get search results normally
	pins := SearchPinterest("matching icons "+query, "all")
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

func SearchPinterestLens(base64Image string, aspect string) []PinResult {
	imageBytes, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		fmt.Println("Base64 decode error:", err)
		return nil
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	writer.WriteField("camera_type", "0")
	writer.WriteField("source_type", "1")
	writer.WriteField("page_size", "24")
	
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
