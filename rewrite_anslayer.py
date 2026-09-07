import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

# 1. Structs
old_struct = """type AnslayerEpisode struct {
	EpisodeID     string `json:"episode_id"`
	EpisodeName   string `json:"episode_name"`
	EpisodeNumber string `json:"episode_number"`
}"""
new_struct = """type AnslayerEpisode struct {
	EpisodeID     int `json:"episode_id"`
	EpisodeName   string `json:"episode_name"`
	EpisodeNumber float64 `json:"episode_number"`
}"""
content = content.replace(old_struct, new_struct)

# 2. HandleAnslayerCommand
start_handle = content.find("func HandleAnslayerCommand(ctx *BotContext, mode string) {")
end_handle = content.find("searchParams := map[string]interface{}{", start_handle)
old_handle = content[start_handle:end_handle]

new_handle = """func HandleAnslayerCommand(ctx *BotContext, mode string) {
	fullParts := strings.SplitN(ctx.Text, " ", 2)
	
	var query string
	if mode == "marketing" {
		if len(fullParts) < 2 {
			sendMessage(ctx, "يرجى كتابة اسم الأنمي، مثلا:\\n.انمي سلاير ون بيس\\nأو لحفظ رسالة النشر:\\n.انمي سلاير نشر رسالتي هنا")
			return
		}
		afterAnmi := strings.TrimSpace(fullParts[1])
		if !strings.HasPrefix(afterAnmi, "سلاير") {
			return
		}
		afterSlayer := strings.TrimSpace(strings.TrimPrefix(afterAnmi, "سلاير"))
		if afterSlayer == "" {
			sendMessage(ctx, "يرجى كتابة اسم الأنمي، مثلا:\\n.انمي سلاير ون بيس")
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
			sendMessage(ctx, "✅ تم حفظ قالب الرد بنجاح!\\nالرسالة:\\n"+anslayerReplyMsg)
			return
		}
		query = afterSlayer
	} else {
		if len(fullParts) < 2 {
			sendMessage(ctx, "يرجى كتابة اسم الأنمي، مثلا:\\n.انمي ون بيس")
			return
		}
		query = fullParts[1]
	}
	"""
content = content.replace(old_handle, new_handle)

# 3. HandleAnslayerEpisodeSelect
start_select = content.find("func HandleAnslayerEpisodeSelect(ctx *BotContext, epNumInt int) bool {")
end_select = content.find("return true\n}\n\nfunc monitorComments", start_select) + 14
old_select = content[start_select:end_select]

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
	
	anslayerMutex.Lock()
	if anslayerReplyMsg == "" {
		anslayerMutex.Unlock()
		sendMessage(ctx, "⚠️ لم تقم بضبط رسالة النشر!\\nيرجى كتابة:\\n.انمي سلاير نشر رسالتي\\nقبل بدء المراقبة.")
		return true
	}
	
	if anslayerStopChan != nil {
		close(anslayerStopChan)
	}
	anslayerStopChan = make(chan struct{})
	anslayerMonitored = epIDStr
	ch := anslayerStopChan
	anslayerMutex.Unlock()
	
	sendMessage(ctx, "✅ تم بدء مراقبة التعليقات للحلقة!\\nسيقوم البوت بالرد فوراً على أي تعليق جديد (ولن يرد على شخص مرتين).\\nلإيقاف المراقبة، اطلب حلقة أخرى أو أعد تشغيل البوت.")
	
	go monitorComments(epIDStr, ch)
	
	delete(ansSessions, ctx.Sender.User)
	return true
}"""

content = content.replace(old_select, new_select)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
    f.write(content)
