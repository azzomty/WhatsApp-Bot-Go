import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/scraper.go', 'r') as f:
    code = f.read()

new_scrape = '''// ScrapeWitanimeEpisode finds an mp4upload/uqload link for an episode
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
	doc.Find("div.anime-card-container a").Each(func(i int, s *goquery.Selection) {
		if animeLink == "" {
			href, exists := s.Attr("href")
			if exists && strings.Contains(href, "/anime/") {
				animeLink = href
			}
		}
	})

	if animeLink == "" {
		// Try fallback: maybe we got redirected to the anime page directly
		if strings.Contains(resp.Request.URL.String(), "/anime/") {
			animeLink = resp.Request.URL.String()
		} else {
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
	targetSuffix := fmt.Sprintf("-%d-", epNum)
	targetText := fmt.Sprintf("الحلقة %d", epNum)
	
	doc.Find("div.episodes-card-title a, h3 a, a").Each(func(i int, s *goquery.Selection) {
		if epLink == "" {
			href, exists := s.Attr("href")
			if exists && strings.Contains(href, "/episode/") {
				text := strings.TrimSpace(s.Text())
				if strings.Contains(href, targetSuffix) || strings.Contains(text, targetText) {
					epLink = href
				}
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

	var foundLink string
	doc.Find("ul#episode-servers a").Each(func(i int, s *goquery.Selection) {
		text := strings.ToLower(s.Text())
		link, _ := s.Attr("data-ep-url")
		if link == "" {
			noscript := s.Find("noscript iframe").First()
			link, _ = noscript.Attr("src")
		}
		
		if link != "" {
			// prioritize mp4upload or uqload or vk
			if strings.Contains(text, "mp4upload") || strings.Contains(link, "mp4upload") {
				foundLink = link
			} else if foundLink == "" && (strings.Contains(text, "uqload") || strings.Contains(text, "vk") || strings.Contains(text, "dood")) {
				foundLink = link
			}
		}
	})
	
	if foundLink != "" {
		return foundLink, nil
	}
	
	return "", fmt.Errorf("no supported embed found")
}'''

# Replace ScrapeWitanimeEpisode
# Find the start and end of the old function
start_idx = code.find('func ScrapeWitanimeEpisode')
if start_idx != -1:
    code = code[:start_idx] + new_scrape
    with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/scraper.go', 'w') as f:
        f.write(code)
    print("Scraper updated")
else:
    print("Function not found")
