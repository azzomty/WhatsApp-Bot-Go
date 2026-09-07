import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

old_case = """	case ".فلم", ".فيلم", ".مسلسل", ".انمي", ".أنمي", ".مانجا", ".مانهاوا", ".كرتون", ".انمي_مدبلج":
		HandleMediaCommand(ctx, cmdName)"""

new_case = """	case ".فلم", ".فيلم", ".مسلسل", ".انمي", ".أنمي", ".مانجا", ".مانهاوا", ".كرتون", ".انمي_مدبلج":
		if len(parts) > 1 && parts[1] == "سلاير" {
			HandleAnslayerCommand(ctx, "marketing")
		} else if firstWord == ".انمي" || firstWord == ".أنمي" {
			HandleAnslayerCommand(ctx, "watch")
		} else {
			HandleMediaCommand(ctx, cmdName)
		}"""

content = content.replace(old_case, new_case)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
