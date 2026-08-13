package pinterest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

func SearchPinterestMedia(query string, ext string) []PinResult {
	// Add "gif" or "video" to the search query to bias the results
	searchQuery := query
	if ext == ".gif" {
		searchQuery += " gif"
	} else if ext == ".mp4" {
		searchQuery += " video"
	}

	pins := SearchPinterest(searchQuery, "all")
	if len(pins) == 0 {
		return nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	var results []PinResult
	var wg sync.WaitGroup
	var mu sync.Mutex

	max := 20
	if len(pins) < max {
		max = len(pins)
	}

	for i := 0; i < max; i++ {
		p := pins[i]
		if p.ID == "" {
			continue
		}
		
		wg.Add(1)
		go func(id string, origTitle string) {
			defer wg.Done()
			req, _ := http.NewRequest("GET", "https://api.pinterest.com/v3/pins/"+id+"/?fields=carousel_data,story_pin_data,images,videos,image_large_url,image_medium_url", nil)
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
				url := extractMediaByExtension(data, ext)
				if url != "" {
					mu.Lock()
					results = append(results, PinResult{
						ID:    id,
						Title: origTitle,
						URL:   url,
					})
					mu.Unlock()
				}
			}
		}(p.ID, p.Title)
	}
	wg.Wait()
	return results
}
