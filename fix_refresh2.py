import re

with open("internal/commands/commands.go", "r") as f:
    c = f.read()

target = """				results, newBm := pinterest.SearchPinterest(last.Query, last.Aspect, last.Count, last.Bookmark)
				last.Bookmark = newBm
				pinterest.SetLastSearch(ctx.ChatID.String(), last.Query, last.Aspect, last.Count, last.IsVisual, last.Base64Image, newBm)
			}"""

new_target = """				var newBm string
				results, newBm = pinterest.SearchPinterest(last.Query, last.Aspect, last.Count, last.Bookmark)
				pinterest.SetLastSearch(ctx.ChatID.String(), last.Query, last.Aspect, last.Count, last.IsVisual, last.Base64Image, newBm)
			}"""

c = c.replace(target, new_target)
with open("internal/commands/commands.go", "w") as f:
    f.write(c)
