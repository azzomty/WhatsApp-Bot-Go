with open("internal/commands/commands.go", "r") as f:
    c = f.read()

c = c.replace('results = pinterest.SearchPinterest(last.Query, last.Aspect, last.Count)', 'results, _ = pinterest.SearchPinterest(last.Query, last.Aspect, last.Count, "")')
c = c.replace('pinterest.SetLastSearch(ctx.ChatID.String(), "", "foryou", count, false, "")', 'pinterest.SetLastSearch(ctx.ChatID.String(), "", "foryou", count, false, "", "")')

with open("internal/commands/commands.go", "w") as f:
    f.write(c)
