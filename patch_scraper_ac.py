import re

with open('internal/commands/scraper.go', 'r') as f:
    content = f.read()

new_func = """
func ScrapeWitanimeEpisode(animeName string, epNum int) (string, error) {
	// First, search for the series ID
	reqURL := fmt.Sprintf("https://wwmdrwjkrzdkqjqddfta.supabase.co/rest/v1/series?select=id,total_episodes&title=eq.%s", url.QueryEscape(animeName))
	
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
"""

content = re.sub(r'func ScrapeWitanimeEpisode\(animeName string, epNum int\) \(string, error\) \{.*?\n\}\n', new_func, content, flags=re.DOTALL)

with open('internal/commands/scraper.go', 'w') as f:
    f.write(content)
