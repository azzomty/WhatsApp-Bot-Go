import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/scraper.go', 'r') as f:
    code = f.read()

old_servers = '''	var links []string
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
	
	return "", fmt.Errorf("no supported embed found")'''

new_servers = '''	var links []string
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
	
	return "", fmt.Errorf("no supported embed found")'''

# Need to add "regexp" to imports if not present
if '"regexp"' not in code:
    code = code.replace('"fmt"', '"fmt"\n\t"regexp"')

code = code.replace(old_servers, new_servers)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/scraper.go', 'w') as f:
    f.write(code)

print("Updated scraper to extract internal m3u8 stream!")
