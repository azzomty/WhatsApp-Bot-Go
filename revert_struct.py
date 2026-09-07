import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

# 1. Structs
old_struct = """type AnslayerEpisode struct {
	EpisodeID     int `json:"episode_id"`
	EpisodeName   string `json:"episode_name"`
	EpisodeNumber float64 `json:"episode_number"`
}"""
new_struct = """type AnslayerEpisode struct {
	EpisodeID     string `json:"episode_id"`
	EpisodeName   string `json:"episode_name"`
	EpisodeNumber string `json:"episode_number"`
}"""
content = content.replace(old_struct, new_struct)

# 2. HandleAnslayerEpisodeSelect
start_select = content.find("func HandleAnslayerEpisodeSelect(ctx *BotContext, epNumInt int) bool {")
end_select = content.find("if session.Mode == \"watch\"", start_select)
old_select = content[start_select:end_select]

new_select = """func HandleAnslayerEpisodeSelect(ctx *BotContext, epNumInt int) bool {
	session, ok := ansSessions[ctx.Sender.User]
	if !ok || session.State != "select_episode" {
		return false
	}
	
	epNumStr := strconv.Itoa(epNumInt)
	var epIDStr string
	for _, e := range session.Episodes {
		if e.EpisodeNumber == epNumStr {
			epIDStr = e.EpisodeID
			break
		}
	}
	
	if epIDStr == "" {
		if epNumInt > 0 && epNumInt <= len(session.Episodes) {
			epIDStr = session.Episodes[epNumInt-1].EpisodeID
		}
	}
	
	if epIDStr == "" {
		sendMessage(ctx, "لم يتم العثور على الحلقة.")
		return true
	}
	
	"""
content = content.replace(old_select, new_select)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
    f.write(content)
