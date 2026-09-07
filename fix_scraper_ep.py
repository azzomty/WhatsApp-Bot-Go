import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/scraper.go', 'r') as f:
    code = f.read()

# Fix episode matching bug
old_match = '''	targetSuffix := fmt.Sprintf("-%d-", epNum)
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
	})'''

new_match = '''	// Look for exact match or exact word match to avoid matching 1210 when looking for 1
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
	})'''

code = code.replace(old_match, new_match)

# Make ScrapeWitanimeEpisode return ALL embed links so we can try them
old_servers = '''	var foundLink string
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

new_servers = '''	var links []string
	doc.Find("ul#episode-servers a").Each(func(i int, s *goquery.Selection) {
		link, _ := s.Attr("data-ep-url")
		if link == "" {
			noscript := s.Find("noscript iframe").First()
			link, _ = noscript.Attr("src")
		}
		
		if link != "" {
			links = append(links, link)
		}
	})
	
	if len(links) > 0 {
		return strings.Join(links, ","), nil
	}
	
	return "", fmt.Errorf("no supported embed found")
}'''

code = code.replace(old_servers, new_servers)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/scraper.go', 'w') as f:
    f.write(code)

print("Updated scraper logic")
