import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

# Fix the structs first
old_struct = """type AnslayerEpisode struct {
	EpisodeID     string `json:"episode_id"`
	EpisodeName   string `json:"episode_name"`
	EpisodeNumber string `json:"episode_number"`
}"""

new_struct = """type AnslayerEpisode struct {
	EpisodeID     int `json:"episode_id"`
	EpisodeName   string `json:"episode_name"`
	EpisodeNumber float64 `json:"episode_number"` // sometimes it's float? let's use float64 just in case
}"""

content = content.replace(old_struct, new_struct)

# Fix HandleAnslayerEpisodeSelect
old_select = """func HandleAnslayerEpisodeSelect(ctx *BotContext, epNumInt int) bool {
	session, ok := ansSessions[ctx.Sender.User]
	if !ok || session.State != "select_episode" {
		return false
	}
	
	epNumStr := strconv.Itoa(epNumInt)
	var epID string
	for _, e := range session.Episodes {
		if e.EpisodeNumber == epNumStr {
			epID = e.EpisodeID
			break
		}
	}
	
	if epID == "" {
		// fallback to index if episode number not exactly matching?
		// some movies are '1'
		if epNumInt > 0 && epNumInt <= len(session.Episodes) {
			epID = session.Episodes[epNumInt-1].EpisodeID
		}
	}"""

new_select = """func HandleAnslayerEpisodeSelect(ctx *BotContext, epNumInt int) bool {
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
		// fallback to index if episode number not exactly matching?
		// some movies are '1'
		if epNumInt > 0 && epNumInt <= len(session.Episodes) {
			epID = session.Episodes[epNumInt-1].EpisodeID
		}
	}"""

content = content.replace(old_select, new_select)

# Also fix the download call
old_download = """	if session.Mode == "watch" {
		sendMessage(ctx, "جاري جلب الحلقة وتحميلها...")
		go downloadAnslayerEpisode(ctx, session.SelectedAnime, epNumStr, epID)
		delete(ansSessions, ctx.Sender.User)
		return true
	}
	
	if session.Mode == "marketing" {
		sendMessage(ctx, "جاري الانطلاق في حملة التسويق على هذه الحلقة...")
		go runAnslayerMarketing(ctx, session.SelectedAnime, epNumStr, epID)
		delete(ansSessions, ctx.Sender.User)
		return true
	}"""

new_download = """	epNumStr := strconv.Itoa(epNumInt)
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
	}"""

content = content.replace(old_download, new_download)

# Now fix the HandleAnslayerCommand logic
old_handle = """// HandleAnslayerCommand processes .انمي سلاير ...
func HandleAnslayerCommand(ctx *BotContext, mode string) {
	parts := strings.SplitN(ctx.Text, " ", 3)
	
	if mode == "marketing" && len(parts) >= 3 && parts[1] == "نشر" {
		anslayerMutex.Lock()
		anslayerReplyMsg = parts[2]
		anslayerMutex.Unlock()
		sendMessage(ctx, "✅ تم حفظ قالب الرد بنجاح!\nالرسالة:\n"+anslayerReplyMsg)
		return
	}
	
	var query string
	if mode == "marketing" {
		if len(parts) < 3 {
			sendMessage(ctx, "يرجى كتابة اسم الأنمي بعد الأمر، مثلا:\n.انمي سلاير ون بيس\nأو لحفظ رسالة النشر:\n.انمي سلاير نشر رسالتي هنا")
			return
		}
		query = parts[2]
	} else {
		// mode == "watch"
		// text is `.انمي ون بيس`
		// parts is `[".انمي", "ون", "بيس"]` if we split normally, but here we used SplitN(ctx.Text, " ", 3)
		// Wait, SplitN(..., 3) on `.انمي ون بيس` gives `[".انمي", "ون", "بيس"]`
		// But query should be "ون بيس"!
		fullParts := strings.SplitN(ctx.Text, " ", 2)
		if len(fullParts) < 2 {
			sendMessage(ctx, "يرجى كتابة اسم الأنمي بعد الأمر، مثلا:\n.انمي ون بيس")
			return
		}
		query = fullParts[1]
	}"""

new_handle = """// HandleAnslayerCommand processes .انمي سلاير ...
func HandleAnslayerCommand(ctx *BotContext, mode string) {
	fullParts := strings.SplitN(ctx.Text, " ", 2)
	
	if mode == "marketing" {
		if len(fullParts) < 2 {
			sendMessage(ctx, "يرجى كتابة اسم الأنمي، مثلا:\n.انمي سلاير ون بيس\nأو لحفظ رسالة النشر:\n.انمي سلاير نشر رسالتي هنا")
			return
		}
		// The second part should start with "سلاير"
		afterAnmi := strings.TrimSpace(fullParts[1])
		if !strings.HasPrefix(afterAnmi, "سلاير") {
			return // Should not happen due to commands.go routing, but just in case
		}
		afterSlayer := strings.TrimSpace(strings.TrimPrefix(afterAnmi, "سلاير"))
		if afterSlayer == "" {
			sendMessage(ctx, "يرجى كتابة اسم الأنمي، مثلا:\n.انمي سلاير ون بيس")
			return
		}
		
		if strings.HasPrefix(afterSlayer, "نشر") {
			replyMsg := strings.TrimSpace(strings.TrimPrefix(afterSlayer, "نشر"))
			if replyMsg == "" {
				sendMessage(ctx, "يرجى كتابة الرسالة بعد كلمة نشر.")
				return
			}
			anslayerMutex.Lock()
			anslayerReplyMsg = replyMsg
			anslayerMutex.Unlock()
			sendMessage(ctx, "✅ تم حفظ قالب الرد بنجاح!\nالرسالة:\n"+anslayerReplyMsg)
			return
		}
		query = afterSlayer
	} else {
		if len(fullParts) < 2 {
			sendMessage(ctx, "يرجى كتابة اسم الأنمي، مثلا:\n.انمي ون بيس")
			return
		}
		query = fullParts[1]
	}"""

content = content.replace(old_handle, new_handle)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
    f.write(content)
