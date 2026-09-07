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

content = content.replace("	AutoJoinGroups = false", "	AutoJoinGroups = false\n" + add_globals)

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

# 3. Rewrite HandleSyrian using regex matching
content = re.sub(r'func HandleSyrian\(ctx \*BotContext\) \{.*?\}\n\}', """func HandleSyrian(ctx *BotContext) {
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
					}
					time.Sleep(2 * time.Second) // Delay between joins
				}
			}
		}()
	}
}""", content, flags=re.DOTALL)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
