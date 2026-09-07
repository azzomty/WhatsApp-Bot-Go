import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

old_syrian = """func HandleSyrian(ctx *BotContext) {
	if ctx.Text == ".دخلني قروبات" {
		AutoJoinGroups = !AutoJoinGroups
		return
	}

	re := regexp.MustCompile(`chat\\.whatsapp\\.com/([A-Za-z0-9_-]+)`)"""

new_syrian = """func HandleSyrian(ctx *BotContext) {
	if ctx.Text == ".دخلني قروبات" {
		AutoJoinGroups = !AutoJoinGroups
		return
	}
	if strings.Contains(ctx.Text, ".القروبات") && strings.Contains(ctx.Text, "مب داخلها") {
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
		return
	}
	if strings.Contains(ctx.Text, ".فحص") && strings.Contains(ctx.Text, "الروابط") {
		interactiveChecking[ctx.Sender.User] = true
		interactiveSessionLinks[ctx.Sender.User] = []string{}
		sendMessage(ctx, "قم بإرسال الروابط الآن، وعند الانتهاء اكتب `.انتهيت`")
		return
	}

	re := regexp.MustCompile(`chat\\.whatsapp\\.com/([A-Za-z0-9_-]+)`)"""

content = content.replace(old_syrian, new_syrian)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
