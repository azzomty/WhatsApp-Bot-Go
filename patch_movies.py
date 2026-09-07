import re

with open('internal/commands/media.go', 'r') as f:
    content = f.read()

new_list = """func HandleCartoonList(ctx *BotContext) {
	sendMessage(ctx, "جاري جلب القائمة... ")
	
	reqURL1 := "https://wwmdrwjkrzdkqjqddfta.supabase.co/rest/v1/series?select=title&order=title.asc"
	req1, _ := http.NewRequest("GET", reqURL1, nil)
	req1.Header.Set("apikey", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Ind3bWRyd2prcnpka3FqcWRkZnRhIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA4MjAxNzUsImV4cCI6MjA5NjM5NjE3NX0.v3-gjEYfuJ4DE17OAHidvd38lCHUTU4ldb2SHLphU8s")
	
	reqURL2 := "https://wwmdrwjkrzdkqjqddfta.supabase.co/rest/v1/movies?select=title&order=title.asc"
	req2, _ := http.NewRequest("GET", reqURL2, nil)
	req2.Header.Set("apikey", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Ind3bWRyd2prcnpka3FqcWRkZnRhIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA4MjAxNzUsImV4cCI6MjA5NjM5NjE3NX0.v3-gjEYfuJ4DE17OAHidvd38lCHUTU4ldb2SHLphU8s")

	client := &http.Client{Timeout: 10 * time.Second}
	
	resp1, err1 := client.Do(req1)
	resp2, err2 := client.Do(req2)
	
	var shows []string
	uniqueShows := make(map[string]bool)
	
	if err1 == nil {
		defer resp1.Body.Close()
		var data []struct{ Title string `json:"title"` }
		json.NewDecoder(resp1.Body).Decode(&data)
		for _, item := range data {
			baseName := strings.Split(item.Title, " الجزء ")[0]
			baseName = strings.Split(baseName, " الموسم ")[0]
			if !uniqueShows[baseName] {
				uniqueShows[baseName] = true
				shows = append(shows, baseName)
			}
		}
	}
	
	if err2 == nil {
		defer resp2.Body.Close()
		var data []struct{ Title string `json:"title"` }
		json.NewDecoder(resp2.Body).Decode(&data)
		for _, item := range data {
			baseName := item.Title
			if !uniqueShows[baseName] {
				uniqueShows[baseName] = true
				shows = append(shows, baseName)
			}
		}
	}
	
	msg := "*قائمة الكراتين والأفلام المتوفرة:*\\n\\n"
	for _, show := range shows {
		msg += "- " + show + "\\n"
	}
	msg += "\\n*للبحث عن كرتون اكتب:* `.كرتون اسم_الكرتون`\\n*للبحث عن فلم اكتب:* `.فلم اسم_الفلم`"
	
	sendMessage(ctx, msg)
}
"""

content = re.sub(r'func HandleCartoonList\(ctx \*BotContext\) \{.*', new_list, content, flags=re.DOTALL)

# Add SearchArabicMovies function at the end
new_search = """

func SearchArabicMovies(query string) []MediaResult {
	q := strings.ReplaceAll(query, "ي", "_")
	q = strings.ReplaceAll(q, "ى", "_")
	q = strings.ReplaceAll(q, "أ", "_")
	q = strings.ReplaceAll(q, "إ", "_")
	q = strings.ReplaceAll(q, "آ", "_")
	q = strings.ReplaceAll(q, "ا", "_")
	q = strings.ReplaceAll(q, "ة", "_")
	q = strings.ReplaceAll(q, "ه", "_")
	
	escapedQ := strings.ReplaceAll(url.QueryEscape(q), "+", "%20")
	reqURL := fmt.Sprintf("https://wwmdrwjkrzdkqjqddfta.supabase.co/rest/v1/movies?select=id,title,story,poster_url&title=ilike.*%%25%s%%25*", escapedQ)
	
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
		ID          string `json:"id"`
		Title       string `json:"title"`
		Story       string `json:"story"`
		PosterURL   string `json:"poster_url"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}
	
	var results []MediaResult
	for _, item := range data {
		// Fetch video servers for this movie
		embeds := FetchMovieEmbeds(item.ID)
		
		desc := item.Story + "\\n\\n*روابط المشاهدة المباشرة:*\\n"
		for _, e := range embeds {
			desc += fmt.Sprintf("- السيرفر %d: %s\\n", e.ServerNumber, e.EmbedURL)
		}
		if len(embeds) == 0 {
			desc += "لا توجد روابط حالياً."
		}
		
		results = append(results, MediaResult{
			Title:       item.Title,
			Description: desc,
			PosterURL:   item.PosterURL,
			MediaType:   "movie",
		})
	}
	return results
}

func FetchMovieEmbeds(movieID string) []struct {
	ServerNumber int    `json:"server_number"`
	EmbedURL     string `json:"embed_url"`
} {
	reqURL := fmt.Sprintf("https://wwmdrwjkrzdkqjqddfta.supabase.co/rest/v1/video_servers?movie_id=eq.%s", movieID)
	req, _ := http.NewRequest("GET", reqURL, nil)
	req.Header.Set("apikey", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Ind3bWRyd2prcnpka3FqcWRkZnRhIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA4MjAxNzUsImV4cCI6MjA5NjM5NjE3NX0.v3-gjEYfuJ4DE17OAHidvd38lCHUTU4ldb2SHLphU8s")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	
	var embeds []struct {
		ServerNumber int    `json:"server_number"`
		EmbedURL     string `json:"embed_url"`
	}
	json.NewDecoder(resp.Body).Decode(&embeds)
	return embeds
}
"""
content += new_search

# Also update the `.فلم` handler
content = content.replace('results = searchTMDB(query, "movie")', 'results = SearchArabicMovies(query)')

with open('internal/commands/media.go', 'w') as f:
    f.write(content)

