import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'r') as f:
    content = f.read()

old_logic = """	playRe := regexp.MustCompile(`play/(\d+)`)
	m := playRe.FindStringSubmatch(string(body))
	if len(m) == 0 {
		return nil, fmt.Errorf("no play link found on show page")
	}

	playURL := showURL + "/play/" + m[1]
	resp2, err := http.Get(playURL)
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()
	body2, _ := io.ReadAll(resp2.Body)"""

new_logic = """	var body2 []byte
	if strings.Contains(showURL, "/play/") {
		body2 = body
	} else {
		playRe := regexp.MustCompile(`play/(\d+)`)
		m := playRe.FindStringSubmatch(string(body))
		if len(m) == 0 {
			return nil, fmt.Errorf("no play link found on show page")
		}

		playURL := showURL + "/play/" + m[1]
		resp2, err := http.Get(playURL)
		if err != nil {
			return nil, err
		}
		defer resp2.Body.Close()
		body2, _ = io.ReadAll(resp2.Body)
	}"""

if old_logic in content:
    content = content.replace(old_logic, new_logic)
    with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'w') as f:
        f.write(content)
    print("Replaced!")
else:
    print("Not found")
