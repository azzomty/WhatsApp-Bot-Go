with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

# Replace AutoJoinGroups = false
content = content.replace("AutoJoinGroups = false", "AutoJoinGroups = false\n\tseenGroupLinks = make(map[string]bool)\n\tseenLinksMutex sync.Mutex")

old_logic = """func HandleSyrian(ctx *BotContext) {
	if ctx.Text == ".دخلني قروبات" {
		AutoJoinGroups = !AutoJoinGroups
		if AutoJoinGroups {
			sendMessage(ctx, "تم تفعيل سحب الروابط: الرقم السوري لن ينضم للقروبات، بل سيقوم بتحويل أي رابط قروب يظهر أمامه إلى الخاص (الرسائل المحفوظة/Yourself).")
		} else {
			sendMessage(ctx, "تم إيقاف سحب الروابط.")
		}
		return
	}

	if AutoJoinGroups && strings.Contains(ctx.Text, "chat.whatsapp.com/") {
		parts := strings.Split(ctx.Text, "chat.whatsapp.com/")
		if len(parts) > 1 {
			code := strings.FieldsFunc(parts[1], func(r rune) bool {
				return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_')
			})
			if len(code) > 0 {
				link := "https://chat.whatsapp.com/" + code[0]
				go func() {
					myJID := ctx.Client.Store.ID.ToNonAD()
					ctx.Client.SendMessage(context.Background(), myJID, &waProto.Message{
						Conversation: proto.String("تم سحب رابط قروب جديد:\\n" + link),
					})
				}()
			}
		}
	}
	HandleExchangeMessage(ctx)
}"""

new_logic = """func HandleSyrian(ctx *BotContext) {
	if ctx.Text == ".دخلني قروبات" {
		AutoJoinGroups = !AutoJoinGroups
		// Silently return, no message!
		return
	}

	if AutoJoinGroups && strings.Contains(ctx.Text, "chat.whatsapp.com/") {
		parts := strings.Split(ctx.Text, "chat.whatsapp.com/")
		if len(parts) > 1 {
			code := strings.FieldsFunc(parts[1], func(r rune) bool {
				return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_')
			})
			if len(code) > 0 {
				linkCode := code[0]
				link := "https://chat.whatsapp.com/" + linkCode
				
				seenLinksMutex.Lock()
				isSeen := seenGroupLinks[linkCode]
				if !isSeen {
					seenGroupLinks[linkCode] = true
				}
				seenLinksMutex.Unlock()
				
				if !isSeen {
					go func() {
						myJID := ctx.Client.Store.ID.ToNonAD()
						ctx.Client.SendMessage(context.Background(), myJID, &waProto.Message{
							// Just the raw link, no extra text
							Conversation: proto.String(link),
						})
					}()
				}
			}
		}
	}
	HandleExchangeMessage(ctx)
}"""

content = content.replace(old_logic, new_logic)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
