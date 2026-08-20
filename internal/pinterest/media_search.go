package pinterest

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"sync"
	"time"
)

func SearchPinterestMedia(query string, ext string, count int) []PinResult {
	// Add "gif" or "video" to the search query to bias the results
	searchQuery := query
	if ext == ".gif" {
		searchQuery += " gif"
	} else if ext == ".mp4" {
		searchQuery += " video"
	}

	pins, _ := SearchPinterest(searchQuery, "all", count, "")
	if len(pins) == 0 {
		return nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	var results []PinResult
	var wg sync.WaitGroup
	var mu sync.Mutex

	max := 15
	if len(pins) < max {
		max = len(pins)
	}

	// Regex to always find .mp4 media URLs (because WhatsApp requires mp4 for auto-playing GIFs)
	regex := regexp.MustCompile(`https://[^"']+\.mp4`)

	for i := 0; i < max; i++ {
		p := pins[i]
		if p.ID == "" {
			continue
		}

		wg.Add(1)
		go func(id string, origTitle string) {
			defer wg.Done()
			req, _ := http.NewRequest("GET", "https://www.pinterest.com/pin/"+id+"/", nil)
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

			resp, err := client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			htmlStr := string(bodyBytes)

			match := regex.FindString(htmlStr)
			if match != "" {
				// Avoid returning thumbnail or weird formats if possible
				mu.Lock()
				results = append(results, PinResult{
					ID:    id,
					Title: origTitle,
					URL:   match,
				})
				mu.Unlock()
			}
		}(p.ID, p.Title)
	}
	wg.Wait()
	return results
}
