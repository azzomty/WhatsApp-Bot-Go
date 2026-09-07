import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/scraper.go', 'r') as f:
    code = f.read()

old_fallback = '''	if animeLink == "" {
		// Try fallback: maybe we got redirected to the anime page directly
		if strings.Contains(resp.Request.URL.String(), "/anime/") {
			animeLink = resp.Request.URL.String()
		} else {
			return "", fmt.Errorf("anime not found on witanime")
		}
	}'''

new_fallback = '''	if animeLink == "" {
		if strings.Contains(resp.Request.URL.String(), "/anime/") {
			animeLink = resp.Request.URL.String()
		} else {
			// Try to get English name from TMDB if the name was Arabic
			engName := getEnglishName(animeName)
			if engName != "" && engName != animeName {
				return ScrapeWitanimeEpisode(engName, epNum)
			}
			return "", fmt.Errorf("anime not found on witanime")
		}
	}'''

code = code.replace(old_fallback, new_fallback)

# Add getEnglishName function
eng_func = '''func getEnglishName(query string) string {
	apiURL := fmt.Sprintf("https://api.themoviedb.org/3/search/tv?api_key=15d2ea6d0dc1d476efbca3eba2b9bbfb&query=%s", url.QueryEscape(query))
	resp, err := http.Get(apiURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	
	if results, ok := result["results"].([]interface{}); ok && len(results) > 0 {
		if resMap, ok := results[0].(map[string]interface{}); ok {
			if name, ok := resMap["name"].(string); ok {
				return name
			}
		}
	}
	return ""
}
'''

code += '\n' + eng_func

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/scraper.go', 'w') as f:
    f.write(code)

print("Added TMDB English fallback to scraper")
