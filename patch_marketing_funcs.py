import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

# startFavMarketing
old_start = """func startFavMarketing(ctx *BotContext, msg string) {"""
new_start = """func startFavMarketing(ctx *BotContext, msg string, accounts []AnslayerAccount) {
	for _, acc := range accounts {
		go startFavMarketingSingle(ctx, msg, acc)
	}
}

func startFavMarketingSingle(ctx *BotContext, msg string, acc AnslayerAccount) {"""
if old_start in content:
    content = content.replace(old_start, new_start)

content = content.replace("user_id%22%3A9174886", "user_id%22%3A\" + acc.UserID + \"")
content = content.replace("reqHeaders(req)", "reqHeaders(req, acc.Token)")

# getLatestEpisodeID
old_latest = """func getLatestEpisodeID(animeID string) string {"""
new_latest = """func getLatestEpisodeID(animeID string, token string) string {"""
if old_latest in content:
    content = content.replace(old_latest, new_latest)

content = content.replace(
"""	req, _ := http.NewRequest("GET", u, nil)
	reqHeaders(req)""",
"""	req, _ := http.NewRequest("GET", u, nil)
	reqHeaders(req, token)""")

# monitorFavComments
old_monitor = """func monitorFavComments(ctx *BotContext, animeIDs []string, msg string, stopCh chan struct{}) {"""
new_monitor = """func monitorFavComments(ctx *BotContext, animeIDs []string, msg string, stopCh chan struct{}, acc AnslayerAccount) {"""
if old_monitor in content:
    content = content.replace(old_monitor, new_monitor)

content = content.replace(
"""				latestEpID := getLatestEpisodeID(animeID)""",
"""				latestEpID := getLatestEpisodeID(animeID, acc.Token)""")

content = content.replace(
"""				replied, _ := checkAndReplyBatch(epIDFloat, msg, 0, 30)""",
"""				replied, _ := checkAndReplyBatch(epIDFloat, msg, 0, 30, acc)""")
content = content.replace(
"""				replied, hasComments := checkAndReplyBatch(epIDFloat, msg, offset, 30)""",
"""				replied, hasComments := checkAndReplyBatch(epIDFloat, msg, offset, 30, acc)""")

# checkAndReplyBatch
old_check = """func checkAndReplyBatch(epIDFloat float64, msg string, offset, limit int) (bool, bool) {"""
new_check = """func checkAndReplyBatch(epIDFloat float64, msg string, offset, limit int, acc AnslayerAccount) (bool, bool) {"""
if old_check in content:
    content = content.replace(old_check, new_check)

content = content.replace(
"""	req, _ := http.NewRequest("GET", u, nil)
	reqHeaders(req)""",
"""	req, _ := http.NewRequest("GET", u, nil)
	reqHeaders(req, acc.Token)""")

content = content.replace(
"""		reqR, _ := http.NewRequest("POST", "https://anslayer.com/anime/public/episode-comments/create-episode-comment-reply", strings.NewReader(payload.Encode()))
		reqHeaders(reqR)""",
"""		reqR, _ := http.NewRequest("POST", "https://anslayer.com/anime/public/episode-comments/create-episode-comment-reply", strings.NewReader(payload.Encode()))
		reqHeaders(reqR, acc.Token)""")

# Inside startFavMarketingSingle (formerly startFavMarketing):
content = content.replace(
"""	go monitorFavComments(ctx, animeIDs, msg, favStopChannels[ctx.Sender.User])""",
"""	go monitorFavComments(ctx, animeIDs, msg, favStopChannels[ctx.Sender.User], acc)""")

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
    f.write(content)
print("Patched funcs")
