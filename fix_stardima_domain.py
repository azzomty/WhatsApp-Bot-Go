import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'r') as f:
    content = f.read()

# Add GetStardimaBaseURL if not exists
inject = """
var stardimaBaseURL string
var stardimaMutex sync.Mutex

func GetStardimaBaseURL() string {
	stardimaMutex.Lock()
	defer stardimaMutex.Unlock()
	if stardimaBaseURL != "" {
		return stardimaBaseURL
	}
	resp, err := http.Get("https://stardima.com")
	if err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		re := regexp.MustCompile(`https://(stardima[a-zA-Z0-9-]*\\.cartoon\\.com\\.im)`)
		m := re.FindStringSubmatch(string(body))
		if len(m) > 1 {
			stardimaBaseURL = "https://" + m[1]
			return stardimaBaseURL
		}
	}
	return "https://stardima-x3.cartoon.com.im"
}
"""

if "func GetStardimaBaseURL" not in content:
    content = content.replace("package commands\n", "package commands\n" + inject)

# Replace all occurrences of https://stardima-37.cartoon.com.im with GetStardimaBaseURL()
content = content.replace('"https://stardima-37.cartoon.com.im/search?query="', 'GetStardimaBaseURL() + "/search?query="')
content = content.replace('"https://stardima-37.cartoon.com.im/series/season/"', 'GetStardimaBaseURL() + "/series/season/"')
content = content.replace('`https://stardima-37\\.cartoon\\.com\\.im/play/video-\\d+`', '`https://stardima[a-zA-Z0-9-]*\\.cartoon\\.com\\.im/play/video-\\d+`')
content = content.replace('"https://stardima-37.cartoon.com.im/%s?page=1"', 'GetStardimaBaseURL() + "/%s?page=1"')
content = content.replace('"https://stardima-37.cartoon.com.im/%s?page=%d"', 'GetStardimaBaseURL() + "/%s?page=%d"')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'w') as f:
    f.write(content)
print("Replaced domain")
