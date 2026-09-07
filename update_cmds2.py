import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

new_case = """	case ".جودة":
		parts := strings.Split(ctx.Text, " ")
		if len(parts) > 1 {
			idx, _ := strconv.Atoi(parts[1])
			HandleStardimaQuality(ctx, idx)
		} else {
			sendMessage(ctx, "يرجى تحديد الرقم، مثال: .جودة 1")
		}
	case ".قفل":"""

content = content.replace('	case ".قفل":', new_case)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
