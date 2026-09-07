import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/scraper.go', 'r') as f:
    code = f.read()

old_find = '''	doc.Find("h3 a").Each(func(i int, s *goquery.Selection) {
		if animeLink == "" {
			animeLink, _ = s.Attr("href")
		}
	})'''

new_find = '''	doc.Find("h3 a").Each(func(i int, s *goquery.Selection) {
		if animeLink == "" {
			href, _ := s.Attr("href")
			if strings.Contains(href, "/anime/") {
				animeLink = href
			}
		}
	})'''

code = code.replace(old_find, new_find)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/scraper.go', 'w') as f:
    f.write(code)

print("Fixed scraper anime link grabbing")
