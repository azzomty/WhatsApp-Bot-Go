import re

with open("main.go", "r") as f:
    content = f.read()

target = """					} else if req.IsVisual && req.Base64Image != "" {
						results = pinterest.SearchPinterestLens(req.Base64Image, aspect, overrideCount)
					} else if aspect == "gif" {
						results = pinterest.SearchPinterest(req.Query+" gif", "gif", overrideCount)
					} else if aspect == "video" {"""

new_code = """					} else if req.IsVisual && req.Base64Image != "" {
						results = pinterest.SearchPinterestLens(req.Base64Image, aspect, overrideCount)
					} else if aspect == "gif" {
						results = pinterest.SearchPinterestGifs(req.Query, overrideCount)
					} else if aspect == "video" {"""

content = content.replace(target, new_code)

with open("main.go", "w") as f:
    f.write(content)

