import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

old_unjoined = """	case ".القروبات الي مب داخلها":
		var codes []string
		linksMutex.Lock()
		for code := range allSavedLinks {
			codes = append(codes, code)
		}
		linksMutex.Unlock()
		if len(codes) > 0 {
			go filterAndSendLinks(ctx, codes)
		} else {
			sendMessage(ctx, "لا توجد روابط محفوظة حالياً.")
		}
	case ".فحص الروابط":
		interactiveChecking[ctx.Sender.User] = true
		interactiveSessionLinks[ctx.Sender.User] = []string{}
		sendMessage(ctx, "قم بإرسال الروابط الآن، وعند الانتهاء اكتب `.انتهيت`")"""

new_unjoined = """	case ".القروبات":
		if strings.Contains(ctx.Text, "مب داخلها") {
			var codes []string
			linksMutex.Lock()
			for code := range allSavedLinks {
				codes = append(codes, code)
			}
			linksMutex.Unlock()
			if len(codes) > 0 {
				go filterAndSendLinks(ctx, codes)
			} else {
				sendMessage(ctx, "لا توجد روابط محفوظة حالياً.")
			}
		}
	case ".فحص":
		if strings.Contains(ctx.Text, "الروابط") {
			interactiveChecking[ctx.Sender.User] = true
			interactiveSessionLinks[ctx.Sender.User] = []string{}
			sendMessage(ctx, "قم بإرسال الروابط الآن، وعند الانتهاء اكتب `.انتهيت`")
		}"""

content = content.replace(old_unjoined, new_unjoined)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
