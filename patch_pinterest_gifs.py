import re

with open("internal/pinterest/pinterest.go", "r") as f:
    content = f.read()

target = """	if len(results) > 0 && aspect != "matching" {"""

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
			req, _ := http.NewRequest("GET", "https://api.pinterest.com/v3/pins/"+id+"/?fields=story_pin_data,images,image_large_url,video_data", nil)
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
				var gifUrl string

				if vd, ok := data["video_data"].(map[string]interface{}); ok {
					if vidList, ok := vd["video_list"].(map[string]interface{}); ok {
						for _, key := range []string{"V_720P", "V_480P", "V_HLSV4_MAC", "V_EXP7"} {
							if vidObj, ok := vidList[key].(map[string]interface{}); ok {
								if u, ok := vidObj["url"].(string); ok {
									if strings.HasSuffix(strings.ToLower(u), ".mp4") || strings.HasSuffix(strings.ToLower(u), ".m3u8") {
										gifUrl = u
										break
									}
								}
							}
						}
					}
				}

				if gifUrl == "" {
					if spd, ok := data["story_pin_data"].(map[string]interface{}); ok {
						if pages, ok := spd["pages"].([]interface{}); ok && len(pages) > 0 {
							for _, pageIface := range pages {
								if page, ok := pageIface.(map[string]interface{}); ok {
									if blocks, ok := page["blocks"].([]interface{}); ok && len(blocks) > 0 {
										if block, ok := blocks[0].(map[string]interface{}); ok {
											if video, ok := block["video"].(map[string]interface{}); ok {
												if vidList, ok := video["video_list"].(map[string]interface{}); ok {
													for _, key := range []string{"V_720P", "V_480P", "V_EXP7"} {
														if vidObj, ok := vidList[key].(map[string]interface{}); ok {
															if u, ok := vidObj["url"].(string); ok {
																gifUrl = u
																break
															}
														}
													}
												}
											}
										}
									}
								}
								if gifUrl != "" {
									break
								}
							}
						}
					}
				}

				if gifUrl == "" {
					if images, ok := data["images"].(map[string]interface{}); ok {
						for _, key := range []string{"originals", "orig"} {
							if imgData, ok := images[key].(map[string]interface{}); ok {
								if u, ok := imgData["url"].(string); ok {
									if strings.HasSuffix(strings.ToLower(u), ".gif") {
										gifUrl = u
										break
									}
								}
							}
						}
					}
				}

				if gifUrl != "" {
					mu.Lock()
					if len(results) < count {
						results = append(results, PinResult{ID: id, Title: "GIF", URL: gifUrl})
					}
					mu.Unlock()
				}
			}
		}(p.ID)
	}
	wg.Wait()
	return results
}

func parsePinterestData("""

content = content.replace("func parsePinterestData(", new_code)

with open("internal/pinterest/pinterest.go", "w") as f:
    f.write(content)

