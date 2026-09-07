import re

with open("internal/commands/commands.go", "r") as f:
    c = f.read()

target_refresh = """				results, _ = pinterest.SearchPinterest(last.Query, last.Aspect, last.Count, "")
			}"""

new_refresh = """				results, newBm := pinterest.SearchPinterest(last.Query, last.Aspect, last.Count, last.Bookmark)
				last.Bookmark = newBm
				pinterest.SetLastSearch(ctx.ChatID.String(), last.Query, last.Aspect, last.Count, last.IsVisual, last.Base64Image, newBm)
			}"""

c = c.replace(target_refresh, new_refresh)

with open("internal/commands/commands.go", "w") as f:
    f.write(c)

