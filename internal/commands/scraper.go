package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"bytes"
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
	// First, search for the series ID
	reqURL := fmt.Sprintf("https://wwmdrwjkrzdkqjqddfta.supabase.co/rest/v1/series?select=id,total_episodes&title=eq.%s", strings.ReplaceAll(url.QueryEscape(animeName), "+", "%20"))
	
	req, _ := http.NewRequest("GET", reqURL, nil)
	req.Header.Set("apikey", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Ind3bWRyd2prcnpka3FqcWRkZnRhIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA4MjAxNzUsImV4cCI6MjA5NjM5NjE3NX0.v3-gjEYfuJ4DE17OAHidvd38lCHUTU4ldb2SHLphU8s")
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Ind3bWRyd2prcnpka3FqcWRkZnRhIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA4MjAxNzUsImV4cCI6MjA5NjM5NjE3NX0.v3-gjEYfuJ4DE17OAHidvd38lCHUTU4ldb2SHLphU8s")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	var series []struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&series); err != nil || len(series) == 0 {
		return "", fmt.Errorf("series not found")
	}
	
	seriesID := series[0].ID
	
	// Now fetch the episode
	epURL := fmt.Sprintf("https://wwmdrwjkrzdkqjqddfta.supabase.co/rest/v1/episodes?select=watch_url&series_id=eq.%s&episode_number=eq.%d", seriesID, epNum)
	req2, _ := http.NewRequest("GET", epURL, nil)
	req2.Header.Set("apikey", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Ind3bWRyd2prcnpka3FqcWRkZnRhIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA4MjAxNzUsImV4cCI6MjA5NjM5NjE3NX0.v3-gjEYfuJ4DE17OAHidvd38lCHUTU4ldb2SHLphU8s")
	req2.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Ind3bWRyd2prcnpka3FqcWRkZnRhIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA4MjAxNzUsImV4cCI6MjA5NjM5NjE3NX0.v3-gjEYfuJ4DE17OAHidvd38lCHUTU4ldb2SHLphU8s")
	
	resp2, err := client.Do(req2)
	if err != nil {
		return "", err
	}
	defer resp2.Body.Close()
	
	var eps []struct {
		WatchURL string `json:"watch_url"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&eps); err != nil || len(eps) == 0 {
		return "", fmt.Errorf("episode not found")
	}
	
	return eps[0].WatchURL, nil
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

