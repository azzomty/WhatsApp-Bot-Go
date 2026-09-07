import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

# 1. Add globals
add_globals = """
var (
	allSavedLinks = make(map[string]bool)
	linksMutex sync.Mutex
	interactiveChecking = make(map[string]bool)
	interactiveSessionLinks = make(map[string][]string)
)

func init() {
	data, err := os.ReadFile("all_group_links.txt")
	if err == nil {
		lines := strings.Split(string(data), "\\n")
		for _, l := range lines {
			if l != "" {
				allSavedLinks[l] = true
			}
		}
	}
}

func appendLinkToFile(code string) {
	f, err := os.OpenFile("all_group_links.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		f.WriteString(code + "\\n")
		f.Close()
	}
}

func filterAndSendLinks(ctx *BotContext, codes []string) {
	sendMessage(ctx, "جاري فحص الروابط... (قد يستغرق بعض الوقت)")
	
	joinedGroups, err := ctx.Client.GetJoinedGroups(context.Background())
	if err != nil {
		sendMessage(ctx, "فشل في جلب قائمة القروبات الحالية.")
		return
	}
	
	joinedMap := make(map[string]bool)
	for _, g := range joinedGroups {
		joinedMap[g.JID.String()] = true
	}
	
	var unjoined []string
	for _, code := range codes {
		info, err := ctx.Client.GetGroupInfoFromLink(context.Background(), code)
		if err == nil {
			if !joinedMap[info.JID.String()] {
				unjoined = append(unjoined, "https://chat.whatsapp.com/"+code)
			}
		}
		time.Sleep(500 * time.Millisecond) // limit API calls
	}
	
	if len(unjoined) == 0 {
		sendMessage(ctx, "أنت متواجد في جميع هذه القروبات بالفعل (أو الروابط غير صالحة).")
		return
	}
	
	// Send in batches of 50
	batchSize := 50
	for i := 0; i < len(unjoined); i += batchSize {
		end := i + batchSize
		if end > len(unjoined) {
			end = len(unjoined)
		}
		msg := "*القروبات التي لست فيها (*" + strconv.Itoa(len(unjoined)) + "*):*\\n\\n" + strings.Join(unjoined[i:end], "\\n\\n")
		sendMessage(ctx, msg)
		time.Sleep(1 * time.Second)
	}
}
"""

content = content.replace("var AutoJoinGroups bool = false", add_globals + "\nvar AutoJoinGroups bool = false")

# 2. Add commands to Handle
add_cmds = """	case ".القروبات الي مب داخلها":
		linksMutex.Lock()
		var codes []string
		for c := range allSavedLinks {
			codes = append(codes, c)
		}
		linksMutex.Unlock()
		if len(codes) == 0 {
			sendMessage(ctx, "لا توجد روابط محفوظة حالياً.")
			return
		}
		go filterAndSendLinks(ctx, codes)
	case ".فحص الروابط":
		interactiveChecking[ctx.Sender.User] = true
		interactiveSessionLinks[ctx.Sender.User] = []string{}
		sendMessage(ctx, "أرسل لي روابط القروبات الآن (في رسالة أو عدة رسائل)، وعندما تنتهي أرسل كلمة: .انتهيت")
"""
content = content.replace('\tcase ".دخلني قروبات":\n\t\tHandleSyrian(ctx)', '\tcase ".دخلني قروبات":\n\t\tHandleSyrian(ctx)\n' + add_cmds)

# 3. Rewrite HandleSyrian
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
						if strings.HasPrefix(myJID.String(), "963") {
							// For Syrian numbers, skip adding +1
						} else {
							// Original logic
						}
						_, err := ctx.Client.JoinGroupWithLink(linkCode)
						if err == nil {
							fmt.Println("Joined group:", linkCode)
						} else {
							fmt.Println("Failed to join group:", err)
						}
					}()
				}
			}
		}
	}
}"""

new_syrian = """func HandleSyrian(ctx *BotContext) {
	if ctx.Text == ".دخلني قروبات" {
		AutoJoinGroups = !AutoJoinGroups
		return
	}

	re := regexp.MustCompile(`chat\.whatsapp\.com/([A-Za-z0-9_-]+)`)
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
		// If in interactive mode, do NOT auto-join even if auto-join is true
		return
	}

	if AutoJoinGroups && len(foundCodes) > 0 {
		go func() {
			for _, code := range foundCodes {
				seenLinksMutex.Lock()
				isSeen := seenGroupLinks[code]
				if !isSeen {
					seenGroupLinks[code] = true
				}
				seenLinksMutex.Unlock()

				if !isSeen {
					_, err := ctx.Client.JoinGroupWithLink(code)
					if err == nil {
						fmt.Println("Joined group:", code)
					} else {
						fmt.Println("Failed to join group:", err)
					}
					time.Sleep(2 * time.Second) // Delay between joins to avoid rate limit
				}
			}
		}()
	}
}"""

content = content.replace(old_syrian, new_syrian)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
