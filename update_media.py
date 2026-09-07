import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'r') as f:
    code = f.read()

# Replace the switch case for .كرتون
old_case = '''	case ".كرتون", ".انمي_مدبلج":
		results = searchTMDB(query, "tv")'''

new_case = '''	case ".كرتون", ".انمي_مدبلج":
		// Use Anime Witcher Algolia for exact Arabic results!
		algoliaRes, err := SearchAlgolia(query)
		if err == nil {
			title := ""
			if t, ok := algoliaRes["name"].(string); ok { title = t }
			poster := ""
			if p, ok := algoliaRes["poster_uri"].(string); ok { poster = p }
			desc := ""
			if details, ok := algoliaRes["details"].(map[string]interface{}); ok {
				if story, ok := algoliaRes["story"].(string); ok {
					desc = story
				}
			}
			
			res := MediaResult{
				Title:       title,
				PosterURL:   poster,
				Description: desc,
				Episodes:    "متوفرة في قاعدة بيانات ويت انمي",
			}
			results = []MediaResult{res}
		} else {
			results = searchTMDB(query, "tv")
		}'''

code = code.replace(old_case, new_case)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'w') as f:
    f.write(code)

print("Replaced successfully")
