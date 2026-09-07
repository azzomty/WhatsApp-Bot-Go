import re

with open("internal/pinterest/pinterest.go", "r") as f:
    p = f.read()
p = p.replace('func SetPending(chatID, query string, count int, isVisual bool, base64Image string) {', 'func SetPending(chatID, query string, count int, isVisual bool, base64Image string, bookmark string) {')
p = p.replace('PendingRequests[chatID] = PendingRequest{Query: query, Count: count, IsVisual: isVisual, Base64Image: base64Image}', 'PendingRequests[chatID] = PendingRequest{Query: query, Count: count, IsVisual: isVisual, Base64Image: base64Image, Bookmark: bookmark}')
with open("internal/pinterest/pinterest.go", "w") as f:
    f.write(p)

with open("internal/commands/commands.go", "r") as f:
    c = f.read()
c = c.replace('results = pinterest.SearchPinterest(query, "gif", 5)', 'results, _ = pinterest.SearchPinterest(query, "gif", 5, "")')
c = c.replace('pinterest.SetLastSearch(ctx.ChatID.String(), query, "gif", count, false, "")', 'pinterest.SetLastSearch(ctx.ChatID.String(), query, "gif", count, false, "", "")')
with open("internal/commands/commands.go", "w") as f:
    f.write(c)

