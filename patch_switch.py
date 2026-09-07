import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

add = """	case ".انمي":
		if len(parts) > 1 && parts[1] == "سلاير" {
			HandleAnslayerCommand(ctx)
		}
"""

content = content.replace('	case ".ستارديما":\n		HandleStardimaCommand(ctx)\n', '	case ".ستارديما":\n		HandleStardimaCommand(ctx)\n' + add)

old_number = """	case ".رقم":
		HandleNumberSelect(ctx)"""

new_number = """	case ".رقم":
		if len(parts) > 1 {
			idx, _ := strconv.Atoi(parts[1])
			if HandleAnslayerNumberSelect(ctx, idx) {
				return
			}
		}
		HandleNumberSelect(ctx)"""

content = content.replace(old_number, new_number)

old_episode = """	case ".حلقة":
		if activeSource[ctx.Sender.User] == "stardima" {"""

new_episode = """	case ".حلقة":
		if len(parts) > 1 {
			idx, _ := strconv.Atoi(parts[1])
			if HandleAnslayerEpisodeSelect(ctx, idx) {
				return
			}
		}
		if activeSource[ctx.Sender.User] == "stardima" {"""

content = content.replace(old_episode, new_episode)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
