import re

with open("internal/pinterest/pinterest.go", "r") as f:
    content = f.read()

# 1. Update SearchState and SetLastSearch
target_state = """type SearchState struct {
	Query       string
	Aspect      string
	Count       int
	IsVisual    bool
	Base64Image string
}

func SetLastSearch(chatID, query, aspect string, count int, isVisual bool, base64Image string) {"""

new_state = """type SearchState struct {
	Query       string
	Aspect      string
	Count       int
	IsVisual    bool
	Base64Image string
	Bookmark    string
}

func SetLastSearch(chatID, query, aspect string, count int, isVisual bool, base64Image string, bookmark string) {"""
content = content.replace(target_state, new_state)

target_set = """	lastSearches[chatID] = SearchState{
		Query:       query,
		Aspect:      aspect,
		Count:       count,
		IsVisual:    isVisual,
		Base64Image: base64Image,
	}"""
new_set = """	lastSearches[chatID] = SearchState{
		Query:       query,
		Aspect:      aspect,
		Count:       count,
		IsVisual:    isVisual,
		Base64Image: base64Image,
		Bookmark:    bookmark,
	}"""
content = content.replace(target_set, new_set)

# 2. Update SearchPinterest to support loop and bookmark
old_search = """func SearchPinterest(query string, aspect string, count int) []PinResult {
	query = url.QueryEscape(query)
	pageSize := count + 10
	if pageSize > 100 {
		pageSize = 100
	}
	searchUrl := fmt.Sprintf("https://api.pinterest.com/v3/search/pins/?rs=typed&pinrep_img_width=474x&query=%s&page_size=%d", query, pageSize)
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
}"""

new_search = """func SearchPinterest(query string, aspect string, count int, startBookmark string) ([]PinResult, string) {
	query = url.QueryEscape(query)
	var allPins []PinResult
	currentBookmark := startBookmark

	for len(allPins) < count {
		pageSize := 100
		searchUrl := fmt.Sprintf("https://api.pinterest.com/v3/search/pins/?rs=typed&pinrep_img_width=474x&query=%s&page_size=%d", query, pageSize)
		if currentBookmark != "" {
			searchUrl += "&bookmark=" + currentBookmark
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
			break
		}

		if len(newPins) == 0 {
			break
		}
	}

	return allPins, currentBookmark
}"""

content = content.replace(old_search, new_search)

with open("internal/pinterest/pinterest.go", "w") as f:
    f.write(content)
