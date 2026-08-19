import re

with open("internal/commands/commands.go", "r") as f:
    content = f.read()

target = """func Handle(ctx *BotContext) {
	if ctx.Text == ".bot off" {"""

new_code = """func Handle(ctx *BotContext) {
	fmt.Printf("[RADAR-CMD] Reached commands.Handle with text: '%s'\\n", ctx.Text)
	if ctx.Text == ".bot off" {"""

content = content.replace(target, new_code)

with open("internal/commands/commands.go", "w") as f:
    f.write(content)
