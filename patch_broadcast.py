import re

with open("internal/commands/exchange.go", "r") as f:
    c = f.read()

# Update BroadcastExchange
target_broadcast = """func BroadcastExchange(client *whatsmeow.Client) {
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
new_broadcast = """func BroadcastExchange(client *whatsmeow.Client) {
	go func() {
		favs := store.GetFavorites()
		for _, favStr := range favs {
			if store.GetStrike(favStr) >= 3 {
				continue // Skip users who ignored 3 times
			}
			
			favJid, err := types.ParseJID(favStr)
			if err != nil {
				continue
			}
			
			store.IncrementStrike(favStr)
			
			msg := "للتبادل اكتب .تبادل"
			client.SendMessage(context.Background(), favJid, &waProto.Message{
				Conversation: proto.String(msg),
			})
			time.Sleep(15 * time.Second) // Wait between sends to avoid spam/ban
		}
	}()
}"""
c = c.replace(target_broadcast, new_broadcast)

# Update HandleExchangeMessage to reset strikes when they reply with .تبادل
target_handle = """	// Handle starting exchange session
	if ctx.Text == ".تبادل" {
		exchangeMu.Lock()"""
new_handle = """	// Handle starting exchange session
	if ctx.Text == ".تبادل" {
		store.ResetStrike(chatStr)
		exchangeMu.Lock()"""
c = c.replace(target_handle, new_handle)

with open("internal/commands/exchange.go", "w") as f:
    f.write(c)

