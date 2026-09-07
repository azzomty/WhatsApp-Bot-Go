import re

with open("main.go", "r") as f:
    c = f.read()

target = """					var urlsToSend []string
					for i, res := range results {
						if i >= overrideCount {
							break
						}
						urlsToSend = append(urlsToSend, res.URL)
					}

					count := 0"""

new_target = """					var urlsToSend []string
					for i, res := range results {
						if i >= overrideCount {
							break
						}
						urlsToSend = append(urlsToSend, res.URL)
					}

					fmt.Printf("DEBUG: Found %d results. urlsToSend has %d urls.\\n", len(results), len(urlsToSend))

					count := 0"""
c = c.replace(target, new_target)

with open("main.go", "w") as f:
    f.write(c)

