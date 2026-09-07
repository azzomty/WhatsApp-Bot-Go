import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

# Add Mode to Session
content = content.replace("type AnslayerSession struct {\n\tState         string", "type AnslayerSession struct {\n\tMode          string\n\tState         string")

# Modify HandleAnslayerCommand
old_handle = """func HandleAnslayerCommand(ctx *BotContext) {
	parts := strings.SplitN(ctx.Text, " ", 3)
	
	if len(parts) >= 3 && parts[1] == "نشر" {
		anslayerMutex.Lock()
		anslayerReplyMsg = parts[2]
		anslayerMutex.Unlock()
		sendMessage(ctx, "✅ تم حفظ قالب الرد بنجاح!\\nالرسالة:\\n"+anslayerReplyMsg)
		return
	}
	
	if len(parts) < 2 {
		sendMessage(ctx, "يرجى كتابة اسم الأنمي بعد الأمر، مثلا:\\n.انمي سلاير ون بيس\\nأو لحفظ رسالة النشر:\\n.انمي سلاير نشر رسالتي هنا")
		return
	}
	
	query := strings.Join(parts[1:], " ")"""

new_handle = """func HandleAnslayerCommand(ctx *BotContext, mode string) {
	parts := strings.SplitN(ctx.Text, " ", 3)
	
	if mode == "marketing" && len(parts) >= 3 && parts[1] == "نشر" {
		anslayerMutex.Lock()
		anslayerReplyMsg = parts[2]
		anslayerMutex.Unlock()
		sendMessage(ctx, "✅ تم حفظ قالب الرد بنجاح!\\nالرسالة:\\n"+anslayerReplyMsg)
		return
	}
	
	var query string
	if mode == "marketing" {
		if len(parts) < 2 {
			sendMessage(ctx, "يرجى كتابة اسم الأنمي بعد الأمر، مثلا:\\n.انمي سلاير ون بيس\\nأو لحفظ رسالة النشر:\\n.انمي سلاير نشر رسالتي هنا")
			return
		}
		query = strings.Join(parts[1:], " ")
	} else {
		if len(parts) < 2 {
			sendMessage(ctx, "يرجى كتابة اسم الأنمي بعد الأمر، مثلا:\\n.انمي ون بيس")
			return
		}
		query = strings.Join(parts[1:], " ")
	}"""

content = content.replace(old_handle, new_handle)

# Fix Session Creation
content = content.replace("session := &AnslayerSession{\n\t\tState: \"select_anime\",\n\t\tAnimes: res.Response.Data,\n\t}", "session := &AnslayerSession{\n\t\tMode: mode,\n\t\tState: \"select_anime\",\n\t\tAnimes: res.Response.Data,\n\t}")

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
    f.write(content)
