import re

with open('/home/lennox/Desktop/اهها/Go_Bot/main.go', 'r') as f:
    content = f.read()

# Replace return with continue or remove return so it doesn't block other commands
old_syrian = """		if strings.HasPrefix(myNumber, "963") {
			commands.HandleSyrian(ctx)
			return
		}"""

new_syrian = """		if strings.HasPrefix(myNumber, "963") {
			commands.HandleSyrian(ctx)
		}"""

content = content.replace(old_syrian, new_syrian)

with open('/home/lennox/Desktop/اهها/Go_Bot/main.go', 'w') as f:
    f.write(content)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content2 = f.read()

# The anime slayer search issue - let's check it
