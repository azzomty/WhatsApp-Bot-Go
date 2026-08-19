import re

with open("internal/commands/commands.go", "r") as f:
    content = f.read()

target = """func HandleSyrian(ctx *BotContext) {
	if AutoJoinGroups && strings.Contains(ctx.Text, "chat.whatsapp.com/") {"""

new_code = """func HandleSyrian(ctx *BotContext) {
	if ctx.Text == ".دخلني قروبات" {
		AutoJoinGroups = !AutoJoinGroups
		if AutoJoinGroups {
			sendMessage(ctx, "تم تفعيل الانضمام التلقائي (البوت سيقوم بالانضمام لأي رابط قروب يرسل).")
		} else {
			sendMessage(ctx, "تم إيقاف الانضمام التلقائي.")
		}
		return
	}

	if AutoJoinGroups && strings.Contains(ctx.Text, "chat.whatsapp.com/") {"""

content = content.replace(target, new_code)

# Remove the toggle from Handle (Saudi) so it only works on Syrian
target_saudi = """func Handle(ctx *BotContext) {
	if ctx.Text == ".bot off" {"""

# We need to find the old toggle inside Handle (if it's there)
target_toggle = """	if ctx.Text == ".دخلني قروبات" {
		AutoJoinGroups = !AutoJoinGroups
		if AutoJoinGroups {
			sendMessage(ctx, "تم تفعيل الانضمام التلقائي (البوت سيقوم بالانضمام لأي رابط قروب يرسل).")
		} else {
			sendMessage(ctx, "تم إيقاف الانضمام التلقائي.")
		}
		return
	}"""
content = content.replace(target_toggle, "")

with open("internal/commands/commands.go", "w") as f:
    f.write(content)
