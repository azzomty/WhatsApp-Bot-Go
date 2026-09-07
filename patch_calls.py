import re

# Patch main.go
with open("main.go", "r") as f:
    main_code = f.read()

# We need to capture the bookmark from .new or .refresh
target_case_new = """					if choice == ".new" || choice == ".refresh" {
						if last, ok := pinterest.GetLastSearch(v.Info.Chat.String()); ok {
							req = pinterest.PendingRequest{Query: last.Query, Count: last.Count}
							suffix = ""
							aspect = last.Aspect
							overrideCount = last.Count
							goto RunSearch
						}
					}"""

new_case_new = """					if choice == ".new" || choice == ".refresh" {
						if last, ok := pinterest.GetLastSearch(v.Info.Chat.String()); ok {
							req = pinterest.PendingRequest{Query: last.Query, Count: last.Count}
							suffix = ""
							aspect = last.Aspect
							overrideCount = last.Count
							currentBookmark = last.Bookmark // We need to define this var above
							goto RunSearch
						}
					}"""
main_code = main_code.replace(target_case_new, new_case_new)

# Add currentBookmark = "" at the top of the switch
main_code = main_code.replace("""				overrideCount := req.Count

				switch choice {""", """				overrideCount := req.Count
				currentBookmark := ""

				switch choice {""")

# Update RunSearch
target_runsearch = """			RunSearch:
				pinterest.SetLastSearch(v.Info.Chat.String(), req.Query, aspect, overrideCount, req.IsVisual, req.Base64Image)
				pinterest.ClearPending(v.Info.Chat.String())"""

new_runsearch = """			RunSearch:
				pinterest.ClearPending(v.Info.Chat.String())"""
main_code = main_code.replace(target_runsearch, new_runsearch)

# Update the goroutine calls
target_go = """				go func() {
					var results []pinterest.PinResult
					if aspect == "foryou" {
						results = pinterest.ForYouPinterest("all")
					} else if aspect == "matching" {
						results = pinterest.SearchPinterestMatchingIcons(req.Query)
						overrideCount = 2 // Match pairs always return 2
					} else if req.IsVisual && req.Base64Image != "" {
						results = pinterest.SearchPinterestLens(req.Base64Image, aspect, overrideCount)
					} else if aspect == "gif" {
						results = pinterest.SearchTenorGifs(req.Query, overrideCount)
					} else if aspect == "video" {
						results = pinterest.SearchPinterestMedia(req.Query, ".mp4", overrideCount)
					} else {
						results = pinterest.SearchPinterest(req.Query+suffix, aspect, overrideCount)
					}"""

new_go = """				go func() {
					var results []pinterest.PinResult
					newBookmark := ""
					
					if aspect == "foryou" {
						results = pinterest.ForYouPinterest("all")
					} else if aspect == "matching" {
						results = pinterest.SearchPinterestMatchingIcons(req.Query)
						overrideCount = 2 // Match pairs always return 2
					} else if req.IsVisual && req.Base64Image != "" {
						results = pinterest.SearchPinterestLens(req.Base64Image, aspect, overrideCount)
					} else if aspect == "gif" {
						results = pinterest.SearchTenorGifs(req.Query, overrideCount)
					} else if aspect == "video" {
						results = pinterest.SearchPinterestMedia(req.Query, ".mp4", overrideCount)
					} else {
						results, newBookmark = pinterest.SearchPinterest(req.Query+suffix, aspect, overrideCount, currentBookmark)
					}
					
					pinterest.SetLastSearch(v.Info.Chat.String(), req.Query, aspect, overrideCount, req.IsVisual, req.Base64Image, newBookmark)"""

main_code = main_code.replace(target_go, new_go)

with open("main.go", "w") as f:
    f.write(main_code)

# Patch commands.go for .new
with open("internal/commands/commands.go", "r") as f:
    cmd_code = f.read()

target_cmd_new = """	if last, ok := pinterest.GetLastSearch(ctx.ChatID.String()); ok {
		pinterest.SetPending(ctx.ChatID.String(), last.Query, last.Count, last.IsVisual, last.Base64Image)"""
new_cmd_new = """	if last, ok := pinterest.GetLastSearch(ctx.ChatID.String()); ok {
		pinterest.SetPending(ctx.ChatID.String(), last.Query, last.Count, last.IsVisual, last.Base64Image)"""
# Wait, .new in commands.go simulates user input, it sets pending and then simulates "/1". No, wait!

