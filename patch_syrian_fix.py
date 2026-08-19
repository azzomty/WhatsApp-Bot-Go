import re

with open("internal/commands/commands.go", "r") as f:
    content = f.read()

target = """func HandleSyrian(ctx *BotContext) {


	if AutoJoinGroups && strings.Contains(ctx.Text, "chat.whatsapp.com/") {"""

new_code = """func HandleSyrian(ctx *BotContext) {
	if ctx.Text == ".دخلني قروبات" {
		AutoJoinGroups = !AutoJoinGroups
		if AutoJoinGroups {
			sendMessage(ctx, "✅ تم تفعيل الانضمام التلقائي (الرقم السوري سيقوم بالانضمام لأي قروب يرسل رابطه).")
		} else {
			sendMessage(ctx, "❌ تم إيقاف الانضمام التلقائي.")
		}
		return
	}

	if AutoJoinGroups && strings.Contains(ctx.Text, "chat.whatsapp.com/") {"""

content = content.replace(target, new_code)

with open("internal/commands/commands.go", "w") as f:
    f.write(content)
