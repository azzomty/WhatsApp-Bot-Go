import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

# Remove the duplicate case I added
content = content.replace("""	case ".انمي":
		if len(parts) > 1 && parts[1] == "سلاير" {
			HandleAnslayerCommand(ctx)
		}
""", "")

old_search = """	case ".فلم", ".فيلم", ".مسلسل", ".انمي", ".أنمي", ".مانجا", ".مانهاوا", ".كرتون", ".انمي_مدبلج":
		HandleSearchCommand(ctx)"""

new_search = """	case ".فلم", ".فيلم", ".مسلسل", ".انمي", ".أنمي", ".مانجا", ".مانهاوا", ".كرتون", ".انمي_مدبلج":
		if len(parts) > 1 && parts[1] == "سلاير" {
			HandleAnslayerCommand(ctx)
		} else {
			HandleSearchCommand(ctx)
		}"""

content = content.replace(old_search, new_search)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
