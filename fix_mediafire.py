import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

helper = r"""
func resolveMediaFire(u string) string {
	resp, err := http.Get(u)
	if err != nil {
		return u
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	
	idx := strings.Index(html, "href=\"https://download")
	if idx != -1 {
		start := idx + 6
		end := strings.Index(html[start:], "\"")
		if end != -1 {
			return html[start : start+end]
		}
	}
	return u
}
"""

if '"io"' not in content:
    content = content.replace('"net/http"', '"net/http"\n\t"io"')

content = content.replace("func downloadAnslayerEpisode", helper + "\nfunc downloadAnslayerEpisode")

old_target = """	// Choose first fembed or ok.ru link, or just first link
	targetLink := links[0]"""

new_target = """	// Choose first fembed or ok.ru link, or just first link
	targetLink := links[0]
	
	if strings.Contains(targetLink, "mediafire.com") {
		// Replace file_premium with file
		targetLink = strings.Replace(targetLink, "file_premium", "file", 1)
		targetLink = resolveMediaFire(targetLink)
	}"""
content = content.replace(old_target, new_target)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
    f.write(content)
