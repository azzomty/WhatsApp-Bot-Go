import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'r') as f:
    content = f.read()

old_loop = """		parsed, err := url.Parse(apiData.WatchURL)
		if err != nil {
			continue
		}
		
		idParam := parsed.Query().Get("id")
		if idParam == "" {
			continue
		}
		
		// Try to extract M3U8 from this embed URL
		m3u8, err := ExtractM3U8(idParam)"""

new_loop = """		// Some servers wrap it in an iframe, some don't.
		// If there is an 'id' param, it might be the real embed URL.
		var embedURL string
		parsed, err := url.Parse(apiData.WatchURL)
		if err == nil {
			idParam := parsed.Query().Get("id")
			if idParam != "" {
				embedURL = idParam
			} else {
				embedURL = apiData.WatchURL
			}
		} else {
			embedURL = apiData.WatchURL
		}
		
		// Try to extract M3U8 from this embed URL
		m3u8, err := ExtractM3U8(embedURL)"""

content = content.replace(old_loop, new_loop)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'w') as f:
    f.write(content)
