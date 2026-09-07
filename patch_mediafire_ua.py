with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

old_func = r"""func resolveMediaFire(u string) string {
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
}"""

new_func = r"""func resolveMediaFire(u string) string {
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	
	resp, err := http.DefaultClient.Do(req)
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
}"""

if old_func in content:
    content = content.replace(old_func, new_func)
    with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
        f.write(content)
    print("Patched Mediafire UA!")
else:
    print("Could not find old func")
