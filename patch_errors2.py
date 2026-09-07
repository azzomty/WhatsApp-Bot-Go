import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

# Fix getLatestEpisodeID
old_latest = """func getLatestEpisodeID(animeID string, token string) string {
	u := fmt.Sprintf("https://anslayer.com/anime/public/anime/get-anime-details?anime_id=%s&fetch_episodes=Yes&more_info=No", animeID)
	req, _ := http.NewRequest("GET", u, nil)
	reqHeaders(req, "")"""
new_latest = """func getLatestEpisodeID(animeID string, token string) string {
	u := fmt.Sprintf("https://anslayer.com/anime/public/anime/get-anime-details?anime_id=%s&fetch_episodes=Yes&more_info=No", animeID)
	req, _ := http.NewRequest("GET", u, nil)
	reqHeaders(req, token)"""
content = content.replace(old_latest, new_latest)

# Fix checkAndReplyBatch
old_check = """func checkAndReplyBatch(epIDFloat float64, msg string, offset, limit int, acc AnslayerAccount) (bool, bool) {
	u := fmt.Sprintf("https://anslayer.com/anime/public/episode-comments/get-episode-comments?episode_id=%v&_offset=%d&_limit=%d&_order_by=latest_first", epIDFloat, offset, limit)
	req, _ := http.NewRequest("GET", u, nil)
	reqHeaders(req, "")"""
new_check = """func checkAndReplyBatch(epIDFloat float64, msg string, offset, limit int, acc AnslayerAccount) (bool, bool) {
	u := fmt.Sprintf("https://anslayer.com/anime/public/episode-comments/get-episode-comments?episode_id=%v&_offset=%d&_limit=%d&_order_by=latest_first", epIDFloat, offset, limit)
	req, _ := http.NewRequest("GET", u, nil)
	reqHeaders(req, acc.Token)"""
content = content.replace(old_check, new_check)

old_check2 = """		reqR, _ := http.NewRequest("POST", "https://anslayer.com/anime/public/episode-comments/create-episode-comment-reply", strings.NewReader(payload.Encode()))
		reqHeaders(reqR, "")"""
new_check2 = """		reqR, _ := http.NewRequest("POST", "https://anslayer.com/anime/public/episode-comments/create-episode-comment-reply", strings.NewReader(payload.Encode()))
		reqHeaders(reqR, acc.Token)"""
content = content.replace(old_check2, new_check2)

# Fix startFavMarketingSingle
old_start = """func startFavMarketingSingle(ctx *BotContext, msg string, acc AnslayerAccount) {
	u := "https://anslayer.com/anime/public/animes/get-published-animes?json=%7B%22_offset%22%3A0%2C%22_limit%22%3A100%2C%22_order_by%22%3A%22latest_first%22%2C%22list_type%22%3A%22favorites%22%2C%22user_id%22%3A\" + acc.UserID + \"%7D"
	req, _ := http.NewRequest("GET", u, nil)
	reqHeaders(req, "")"""
new_start = """func startFavMarketingSingle(ctx *BotContext, msg string, acc AnslayerAccount) {
	u := "https://anslayer.com/anime/public/animes/get-published-animes?json=%7B%22_offset%22%3A0%2C%22_limit%22%3A100%2C%22_order_by%22%3A%22latest_first%22%2C%22list_type%22%3A%22favorites%22%2C%22user_id%22%3A\" + acc.UserID + \"%7D"
	req, _ := http.NewRequest("GET", u, nil)
	reqHeaders(req, acc.Token)"""
content = content.replace(old_start, new_start)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
    f.write(content)
print("Patched errors 2")
