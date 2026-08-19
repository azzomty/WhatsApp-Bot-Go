import re

with open("internal/pinterest/pinterest.go", "r") as f:
    content = f.read()

# First, restore SearchPinterestGifs definition if it doesn't exist
if "func SearchPinterestGifs" not in content:
    target = """func parsePinterestData("""
    new_code = """func SearchPinterestGifs(query string, count int) []PinResult {
	pins := SearchPinterest(query+" gif", "all", count+20)
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

func parsePinterestData("""
    content = content.replace(target, new_code)

with open("internal/pinterest/pinterest.go", "w") as f:
    f.write(content)

with open("main.go", "r") as f:
    main_content = f.read()

main_content = main_content.replace('results = pinterest.SearchPinterest(req.Query+" gif", "gif", overrideCount)', 'results = pinterest.SearchPinterestGifs(req.Query, overrideCount)')

with open("main.go", "w") as f:
    f.write(main_content)

