import re

with open("main.go", "r") as f:
    c = f.read()

target = """					count := 0
					for _, u := range urlsToSend {
						data, err := pinterest.DownloadImage(u)
						if err == nil && len(data) > 100 {"""

new_target = """					count := 0
					for _, u := range urlsToSend {
						data, err := pinterest.DownloadImage(u)
						if err != nil {
							fmt.Printf("DEBUG: DownloadImage failed for %s: %v\\n", u, err)
						} else if len(data) <= 100 {
							fmt.Printf("DEBUG: DownloadImage data too small for %s: %d bytes\\n", u, len(data))
						}
						if err == nil && len(data) > 100 {"""
c = c.replace(target, new_target)

with open("main.go", "w") as f:
    f.write(c)

