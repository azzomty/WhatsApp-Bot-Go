import re

with open('/home/lennox/Desktop/اهها/Go_Bot/main.go', 'r') as f:
    content = f.read()

cookie_logic = """	// Generate cookies.txt from Render Environment Variable
	ytCookies := os.Getenv("YOUTUBE_COOKIES")
	if ytCookies != "" {
		err := os.WriteFile("cookies.txt", []byte(ytCookies), 0644)
		if err == nil {
			fmt.Println("YouTube cookies generated from Environment Variable!")
		}
	}
"""

if 'os.Getenv("YOUTUBE_COOKIES")' not in content:
    content = content.replace("func main() {\n\tgo initDeps()", "func main() {\n" + cookie_logic + "\n\tgo initDeps()")
    with open('/home/lennox/Desktop/اهها/Go_Bot/main.go', 'w') as f:
        f.write(content)
    print("Patched main for cookies")
