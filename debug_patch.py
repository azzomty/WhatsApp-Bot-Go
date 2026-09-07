import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'r') as f:
    content = f.read()

old_func = """func GetBestM3U8(hyperURL string) (string, error) {"""
new_func = """func GetBestM3U8(hyperURL string) (string, error) {
	fmt.Println("DEBUG: GetBestM3U8 called with hyperURL:", hyperURL)
"""

content = content.replace(old_func, new_func)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'w') as f:
    f.write(content)
