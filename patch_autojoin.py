import re

with open("main.go", "r") as f:
    content = f.read()

target = """		// AutoJoin logic (runs everywhere, ignores .bot off and .baymax)
		if commands.AutoJoinGroups && strings.Contains(text, "chat.whatsapp.com/") {"""

new_code = """		// AutoJoin logic (ONLY for Syrian number, ignores .bot off and .baymax)
		if strings.HasPrefix(myNumber, "963") && commands.AutoJoinGroups && strings.Contains(text, "chat.whatsapp.com/") {"""

content = content.replace(target, new_code)

target2 = """		// Syrian Number Logic (ONLY runs AutoJoin and Exchange, which is done above)
		if strings.HasPrefix(myNumber, "963") {
			return
		}"""
new_code2 = """		// Syrian Number Logic (ONLY runs AutoJoin and Exchange, which is done above)
		if strings.HasPrefix(myNumber, "963") {
			return
		}

		// If it's the Saudi number, we disable AutoJoin entirely (since we already skipped it above).
		// We just let it continue to the rest of the bot logic.
"""
content = content.replace(target2, new_code2)

with open("main.go", "w") as f:
    f.write(content)
