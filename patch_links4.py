import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

old_syrian = """func HandleSyrian(ctx *BotContext) {
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

new_syrian = """func HandleSyrian(ctx *BotContext) {
	if ctx.Text == ".دخلني قروبات" {
		AutoJoinGroups = !AutoJoinGroups
		return
	}

	re := regexp.MustCompile(`chat\\.whatsapp\\.com/([A-Za-z0-9_-]+)`)
	matches := re.FindAllStringSubmatch(ctx.Text, -1)

	var foundCodes []string
	if len(matches) > 0 {
		linksMutex.Lock()
		for _, m := range matches {
			code := m[1]
			foundCodes = append(foundCodes, code)
			if !allSavedLinks[code] {
				allSavedLinks[code] = true
				appendLinkToFile(code)
			}
		}
		linksMutex.Unlock()
	}

	if interactiveChecking[ctx.Sender.User] {
		if ctx.Text == ".انتهيت" {
			interactiveChecking[ctx.Sender.User] = false
			codes := interactiveSessionLinks[ctx.Sender.User]
			go filterAndSendLinks(ctx, codes)
		} else if len(foundCodes) > 0 {
			interactiveSessionLinks[ctx.Sender.User] = append(interactiveSessionLinks[ctx.Sender.User], foundCodes...)
		}
		return
	}

	if AutoJoinGroups && len(foundCodes) > 0 {
		go func() {
			for _, linkCode := range foundCodes {
				seenLinksMutex.Lock()
				isSeen := seenGroupLinks[linkCode]
				if !isSeen {
					seenGroupLinks[linkCode] = true
				}
				seenLinksMutex.Unlock()

				if !isSeen {
					myJID := ctx.Client.Store.ID.ToNonAD()
					link := "https://chat.whatsapp.com/" + linkCode
					ctx.Client.SendMessage(context.Background(), myJID, &waProto.Message{
						Conversation: proto.String(link),
					})
					time.Sleep(2 * time.Second)
				}
			}
		}()
	}
	
	HandleExchangeMessage(ctx)
}"""

content = content.replace(old_syrian, new_syrian)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
