import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

# Replace the whole HandleAnslayerEpisodeSelect function
start = content.find("func HandleAnslayerEpisodeSelect")
end = content.find("return false\n}", start) + 14

new_func = """func HandleAnslayerEpisodeSelect(ctx *BotContext, epNumInt int) bool {
	session, ok := ansSessions[ctx.Sender.User]
	if !ok || session.State != "select_episode" {
		return false
	}
	
	var epID int
	for _, e := range session.Episodes {
		if int(e.EpisodeNumber) == epNumInt {
			epID = e.EpisodeID
			break
		}
	}
	
	if epID == 0 {
		if epNumInt > 0 && epNumInt <= len(session.Episodes) {
			epID = session.Episodes[epNumInt-1].EpisodeID
		}
	}
	
	if epID == 0 {
		sendMessage(ctx, "لم يتم العثور على الحلقة.")
		return true
	}
	
	epNumStr := strconv.Itoa(epNumInt)
	epIDStr := strconv.Itoa(epID)
	
	if session.Mode == "watch" {
		sendMessage(ctx, "جاري جلب الحلقة وتحميلها...")
		go downloadAnslayerEpisode(ctx, session.SelectedAnime, epNumStr, epIDStr)
		delete(ansSessions, ctx.Sender.User)
		return true
	}
	
	if session.Mode == "marketing" {
		sendMessage(ctx, "جاري الانطلاق في حملة التسويق على هذه الحلقة...")
		go runAnslayerMarketing(ctx, session.SelectedAnime, epNumStr, epIDStr)
		delete(ansSessions, ctx.Sender.User)
		return true
	}
	
	return false
}"""

content = content[:start] + new_func + content[end:]

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
    f.write(content)
