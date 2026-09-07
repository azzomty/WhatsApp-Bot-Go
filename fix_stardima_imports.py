with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'r') as f:
    content = f.read()

# Remove the injected block
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

content = content.replace("package commands\n" + inject, "package commands\n")

# Re-inject it AFTER the imports block
# Find the end of imports block
import_end = content.find(")\n")
if import_end != -1:
    content = content[:import_end+2] + inject + content[import_end+2:]

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'w') as f:
    f.write(content)
