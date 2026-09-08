with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/mdm.go', 'r') as f:
    c = f.read()

old = '''func fetchCard(ctx *BotContext, query string) {
	apiURL := "https://db.ygoprodeck.com/api/v7/cardinfo.php?fname=" + url.QueryEscape(query)'''

new = '''func fetchCard(ctx *BotContext, query string) {
	bestQuery := GetClosestCardName(query)
	apiURL := "https://db.ygoprodeck.com/api/v7/cardinfo.php?name=" + url.QueryEscape(bestQuery)'''

c = c.replace(old, new)

# Also fallback to fname if name doesn't match?
# Actually fname= is fine but name= is exact match. Since we use closest name, name= is better!
# Wait, if GetClosestCardName returns a fallback that isn't exact, fname= might be better.
# Let's use fname= just in case, but bestQuery will be exact if it was found!

new2 = '''func fetchCard(ctx *BotContext, query string) {
	bestQuery := GetClosestCardName(query)
	apiURL := "https://db.ygoprodeck.com/api/v7/cardinfo.php?fname=" + url.QueryEscape(bestQuery)'''

c = c.replace(new, new2)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/mdm.go', 'w') as f:
    f.write(c)
