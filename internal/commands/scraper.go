package commands

import (
	"encoding/json"
	"fmt"
	"regexp"
	"io"
	"net/http"
	"net/url"
	"strings"
	"bytes"
	"github.com/PuerkitoBio/goquery"
)

// SearchAlgolia searches Anime Witcher Algolia
func SearchAlgolia(query string) ([]map[string]interface{}, error) {
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
}

// ScrapeWitanimeEpisode tries to find an mp4upload or embed link for an episode
// ScrapeWitanimeEpisode finds an mp4upload/uqload link for an episode
func ScrapeWitanimeEpisode(animeName string, epNum int) (string, error) {
	// 1. Search Witanime
	searchURL := fmt.Sprintf("https://4h.b9p2m6c.shop/?search_param=animes&s=%s", url.QueryEscape(animeName))
	
	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("search failed: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	// 2. Get first anime link
	var animeLink string
	doc.Find("h3 a, a.overlay").Each(func(i int, s *goquery.Selection) {
		if animeLink == "" {
			href, exists := s.Attr("href")
			if exists && strings.Contains(href, "/anime/") {
				animeLink = href
			}
		}
	})

	if animeLink == "" {
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
	}

	// 3. Fetch anime page
	req, _ = http.NewRequest("GET", animeLink, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err = client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	// 4. Find the episode link
	var epLink string
	// Search for something like "الحلقة 1" or check the URL for "-1-مترجم"
	// Look for exact match or exact word match to avoid matching 1210 when looking for 1
	targetSuffix1 := fmt.Sprintf("-%d-", epNum)
	targetSuffix2 := fmt.Sprintf("-%d/", epNum)
	targetText := fmt.Sprintf("الحلقة %d ", epNum) // trailing space to avoid matching 12
	
	doc.Find("div.episodes-card-title a, h3 a, a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && strings.Contains(href, "/episode/") {
			text := strings.TrimSpace(s.Text()) + " "
			// Exact episode match
			if epLink == "" && (strings.Contains(href, targetSuffix1) || strings.Contains(href, targetSuffix2) || strings.HasPrefix(text, targetText)) {
				epLink = href
			}
		}
	})

	if epLink == "" {
		// Sometimes Witanime lists episodes in an external list or they use ajax, 
		// but usually they are on the page. Let's just guess the URL if not found.
		slug := strings.Trim(strings.ReplaceAll(animeLink, "https://4h.b9p2m6c.shop/anime/", ""), "/")
		epLink = fmt.Sprintf("https://4h.b9p2m6c.shop/episode/انمي-%s-الحلقة-%d-مترجمة/", slug, epNum)
	}

	// 5. Fetch episode page
	req, _ = http.NewRequest("GET", epLink, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err = client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	var links []string
	var internalIframe string
	
	// Witanime changed to li data-watch instead of a data-ep-url
	doc.Find("ul#episode-servers li").Each(func(i int, s *goquery.Selection) {
		link, _ := s.Attr("data-watch")
		if link == "" {
			noscript := s.Find("noscript iframe").First()
			link, _ = noscript.Attr("src")
		}
		if link != "" {
			links = append(links, link)
			// Prioritize internal players like anime4up or megamax because we can extract the m3u8 directly!
			lowerLink := strings.ToLower(link)
			if strings.Contains(lowerLink, "anime4up") || strings.Contains(lowerLink, "share4max") {
				internalIframe = link
			}
		}
	})
	
	// If we found an internal iframe, extract the m3u8 streamUrl!
	if internalIframe != "" {
		reqIframe, _ := http.NewRequest("GET", internalIframe, nil)
		reqIframe.Header.Set("User-Agent", "Mozilla/5.0")
		client := &http.Client{}
		respIframe, err := client.Do(reqIframe)
		if err == nil {
			defer respIframe.Body.Close()
			bodyBytes, _ := io.ReadAll(respIframe.Body)
			bodyStr := string(bodyBytes)
			
			// Regex to find streamUrl = "..."
			re := regexp.MustCompile(`streamUrl\s*=\s*['"]([^'"]+)['"]`)
			matches := re.FindStringSubmatch(bodyStr)
			if len(matches) > 1 {
				m3u8Link := matches[1]
				// Return the m3u8 link as the first option, followed by other embeds
				links = append([]string{m3u8Link}, links...)
			}
		}
	}
	
	if len(links) > 0 {
		return strings.Join(links, ","), nil
	}
	
	return "", fmt.Errorf("no supported embed found")
}
func getEnglishName(query string) string {
	apiURL := fmt.Sprintf("https://api.themoviedb.org/3/search/multi?api_key=15d2ea6d0dc1d476efbca3eba2b9bbfb&query=%s", url.QueryEscape(query))
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
