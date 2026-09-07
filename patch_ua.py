import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'r') as f:
    content = f.read()

old_func = """func GetBestM3U8(hyperURL string) (string, error) {
	resp, err := http.Get(hyperURL)
	if err != nil {
		return "", err
	}"""

new_func = """func GetBestM3U8(hyperURL string) (string, error) {
	req, _ := http.NewRequest("GET", hyperURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}"""

content = content.replace(old_func, new_func)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'w') as f:
    f.write(content)
