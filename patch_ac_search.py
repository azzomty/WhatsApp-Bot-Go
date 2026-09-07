import re

with open('internal/commands/media.go', 'r') as f:
    content = f.read()

ac_func = """
func searchArabicCartoon(query string) []MediaResult {
	reqURL := fmt.Sprintf("https://wwmdrwjkrzdkqjqddfta.supabase.co/rest/v1/series?select=*&title=ilike.*%25%s%25*", url.QueryEscape(query))
	
	req, _ := http.NewRequest("GET", reqURL, nil)
	req.Header.Set("apikey", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Ind3bWRyd2prcnpka3FqcWRkZnRhIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA4MjAxNzUsImV4cCI6MjA5NjM5NjE3NX0.v3-gjEYfuJ4DE17OAHidvd38lCHUTU4ldb2SHLphU8s")
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Ind3bWRyd2prcnpka3FqcWRkZnRhIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA4MjAxNzUsImV4cCI6MjA5NjM5NjE3NX0.v3-gjEYfuJ4DE17OAHidvd38lCHUTU4ldb2SHLphU8s")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	
	var data []struct {
		Title         string  `json:"title"`
		Description   string  `json:"description"`
		PosterURL     string  `json:"poster_url"`
		Rating        float64 `json:"rating"`
		YearStarted   int     `json:"year_started"`
		TotalEpisodes int     `json:"total_episodes"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}
	
	var results []MediaResult
	for _, item := range data {
		res := MediaResult{
			Title:       item.Title,
			Description: item.Description,
			Poster:      item.PosterURL,
		}
		if item.Rating > 0 {
			res.Rating = fmt.Sprintf("%.1f", item.Rating)
		}
		if item.YearStarted > 0 {
			res.Year = fmt.Sprintf("%d", item.YearStarted)
		}
		if item.TotalEpisodes > 0 {
			res.Episodes = fmt.Sprintf("%d", item.TotalEpisodes)
		}
		results = append(results, res)
	}
	return results
}
"""

content = re.sub(
    r'case "\.كرتون", "\.انمي_مدبلج":\n\s*results = searchTMDB\(query, "tv"\)',
    'case ".كرتون", ".انمي_مدبلج":\n\t\t\tresults = searchArabicCartoon(query)',
    content
)

content = content + "\n" + ac_func

with open('internal/commands/media.go', 'w') as f:
    f.write(content)
