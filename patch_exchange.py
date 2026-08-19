import re

with open("internal/commands/exchange.go", "r") as f:
    content = f.read()

target = """func StartExchangeBackgroundLoop(client *whatsmeow.Client) {
	for {
		time.Sleep(5 * time.Hour)
		
		if !AutoJoinGroups {
			continue // Only run if AutoJoinGroups is true
		}

		favs := store.GetFavorites()
		for _, favStr := range favs {
			if !AutoJoinGroups {
				break
			}
			favJid, err := types.ParseJID(favStr)
			if err != nil {
				continue
			}
			msg := "تبادل عشان نتبادل اكتب .تبادل وبعدها ارسل روابطك. ولما تنتهي من إرسال الروابط اكتب .انتهيت"
			client.SendMessage(context.Background(), favJid, &waProto.Message{
				Conversation: proto.String(msg),
			})
			time.Sleep(15 * time.Second) // Wait between sends to avoid spam/ban
		}
	}
}"""

new_code = """func BroadcastExchange(client *whatsmeow.Client) {
	go func() {
		favs := store.GetFavorites()
		for _, favStr := range favs {
			favJid, err := types.ParseJID(favStr)
			if err != nil {
				continue
			}
			msg := "تبادل عشان نتبادل اكتب .تبادل وبعدها ارسل روابطك. ولما تنتهي من إرسال الروابط اكتب .انتهيت"
			client.SendMessage(context.Background(), favJid, &waProto.Message{
				Conversation: proto.String(msg),
			})
			time.Sleep(15 * time.Second) // Wait between sends to avoid spam/ban
		}
	}()
}"""

content = content.replace(target, new_code)

target_handle = """	// Handle setting exchange group
	if ctx.Text == "!تبادل" && ctx.Event.Info.IsGroup {"""

new_handle = """	// Handle .نشر broadcast
	if ctx.Text == ".نشر" {
		sendMessage(ctx, "⏳ جاري إرسال رسالة التبادل لجميع الأرقام في المفضلة...")
		BroadcastExchange(ctx.Client)
		return true
	}

	// Handle setting exchange group
	if ctx.Text == "!تبادل" && ctx.Event.Info.IsGroup {"""

content = content.replace(target_handle, new_handle)

with open("internal/commands/exchange.go", "w") as f:
    f.write(content)
