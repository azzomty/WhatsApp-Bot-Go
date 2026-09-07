import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/scraper.go', 'r') as f:
    code = f.read()

old_search = '''func SearchAlgolia(query string) (map[string]interface{}, error) {
	apiURL := "https://d8lh9i7zl7-dsn.algolia.net/1/indexes/series/query"
	payload := fmt.Sprintf(`{"params": "query=%s&hitsPerPage=1"}`, url.QueryEscape(query))

	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Algolia-API-Key", "b56c01ef52540ef334bcdbaa00ded9e4")
	req.Header.Set("X-Algolia-Application-Id", "D8LH9I7ZL7")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	json.Unmarshal(body, &data)

	if hits, ok := data["hits"].([]interface{}); ok && len(hits) > 0 {
		return hits[0].(map[string]interface{}), nil
	}
	return nil, fmt.Errorf("no results")
}'''

new_search = '''func SearchAlgolia(query string) ([]map[string]interface{}, error) {
	apiURL := "https://d8lh9i7zl7-dsn.algolia.net/1/indexes/series/query"
	payload := fmt.Sprintf(`{"params": "query=%s&hitsPerPage=10"}`, url.QueryEscape(query))

	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Algolia-API-Key", "b56c01ef52540ef334bcdbaa00ded9e4")
	req.Header.Set("X-Algolia-Application-Id", "D8LH9I7ZL7")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	json.Unmarshal(body, &data)

	if hits, ok := data["hits"].([]interface{}); ok && len(hits) > 0 {
		var res []map[string]interface{}
		for _, hit := range hits {
			res = append(res, hit.(map[string]interface{}))
		}
		return res, nil
	}
	return nil, fmt.Errorf("no results")
}'''

code = code.replace(old_search, new_search)
with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/scraper.go', 'w') as f:
    f.write(code)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'r') as f:
    media = f.read()

old_media_alg = '''		algoliaRes, err := SearchAlgolia(query)
		if err == nil {
			title := ""
			if t, ok := algoliaRes["name"].(string); ok { title = t }
			poster := ""
			if p, ok := algoliaRes["poster_uri"].(string); ok { poster = p }
			desc := ""
			if _, ok := algoliaRes["details"].(map[string]interface{}); ok {
				if story, ok := algoliaRes["story"].(string); ok {
					desc = story
				}
			}
			
			res := MediaResult{
				Title:       title,
				PosterURL:   poster,
				Description: desc,
				Episodes:    "متوفرة في قاعدة بيانات ويت انمي",
			}
			results = []MediaResult{res}
		} else {
			results = searchTMDB(query, "tv")
		}'''

new_media_alg = '''		algoliaResList, err := SearchAlgolia(query)
		if err == nil {
			for _, algoliaRes := range algoliaResList {
				title := ""
				if t, ok := algoliaRes["name"].(string); ok { title = t }
				poster := ""
				if p, ok := algoliaRes["poster_uri"].(string); ok { poster = p }
				desc := ""
				if _, ok := algoliaRes["details"].(map[string]interface{}); ok {
					if story, ok := algoliaRes["story"].(string); ok {
						desc = story
					}
				}
				
				res := MediaResult{
					Title:       title,
					PosterURL:   poster,
					Description: desc,
					Episodes:    "غير معروف", // We will extract if available
				}
				results = append(results, res)
			}
		} else {
			results = searchTMDB(query, "tv")
		}'''

media = media.replace(old_media_alg, new_media_alg)

# Fix epCount logic
epCountLogic = '''		epCount := res.Episodes
		if epCount == "" {
			epCount = "غير معروف"
		}'''
epCountFixed = '''		epCount := res.Episodes
		if epCount == "" || epCount == "غير معروف" {
			epCount = "غير معروف (ابحث بالحلقة)"
		}'''
media = media.replace(epCountLogic, epCountFixed)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'w') as f:
    f.write(media)

print("Fixed algolia multiple hits")
