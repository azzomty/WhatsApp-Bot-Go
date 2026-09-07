import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'r') as f:
    content = f.read()

old_func = """func GetStardimaHyperwatchingURL(movieURL string) (string, error) {
	resp, err := http.Get(movieURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	re := regexp.MustCompile(`data-page="([^"]+)"`)
	m := re.FindStringSubmatch(string(body))
	if len(m) < 2 {
		return "", fmt.Errorf("data-page not found")
	}

	jsonStr := strings.ReplaceAll(m[1], "&quot;", "\"")

	var pageData struct {
		Props struct {
			Video struct {
				Hashid  string `json:"hashid"`
				Servers []struct {
					ID       int    `json:"id"`
					ServerID int    `json:"server_id"`
					Name     string `json:"name"`
				} `json:"servers"`
			} `json:"video"`
		} `json:"props"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &pageData); err != nil {
		return "", err
	}

	var uqloadServerID int
	for _, srv := range pageData.Props.Video.Servers {
		if strings.ToLower(srv.Name) == "uqload" {
			uqloadServerID = srv.ID
			break
		}
	}

	if uqloadServerID == 0 {
		return "", fmt.Errorf("uqload server not found")
	}

	// GET https://stardima-37.cartoon.com.im/api/video/{hashid}/{server_id}
	apiURL := fmt.Sprintf(GetStardimaBaseURL()+"/api/video/%s/%d", pageData.Props.Video.Hashid, uqloadServerID)
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	respAPI, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer respAPI.Body.Close()

	var apiData struct {
		WatchURL string `json:"watch_url"`
	}
	if err := json.NewDecoder(respAPI.Body).Decode(&apiData); err != nil {
		return "", err
	}

	return apiData.WatchURL, nil
}"""

new_func = """func GetStardimaHyperwatchingURL(movieURL string) (string, error) {
	playURL := strings.Replace(movieURL, "/movie/", "/play/", 1)
	req, _ := http.NewRequest("GET", playURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	re := regexp.MustCompile(`<iframe[^>]*src="([^"]*v2\.hyperwatching\.com[^"]+)"`)
	m := re.FindStringSubmatch(string(body))
	if len(m) > 1 {
		return m[1], nil
	}
	
	// Fallback to og:video
	reOG := regexp.MustCompile(`<meta[^>]*property="og:video"[^>]*content="([^"]+)"`)
	mOG := reOG.FindStringSubmatch(string(body))
	if len(mOG) > 1 {
		return mOG[1], nil
	}

	return "", fmt.Errorf("hyperwatching URL not found in /play/ page")
}"""

content = content.replace(old_func, new_func)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'w') as f:
    f.write(content)
