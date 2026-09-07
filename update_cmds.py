import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

# Add to HandleCommands
route_code = """	} else if strings.HasPrefix(text, ".حلقة") {
		numStr := strings.TrimSpace(strings.TrimPrefix(text, ".حلقة"))
		if num, err := strconv.Atoi(numStr); err == nil {
			HandleStardimaEpisode(ctx, num)
		} else {
			sendMessage(ctx, "يرجى كتابة رقم الحلقة بشكل صحيح، مثال: .حلقة 1")
		}
	} else if strings.HasPrefix(text, ".جودة") {
		numStr := strings.TrimSpace(strings.TrimPrefix(text, ".جودة"))
		if num, err := strconv.Atoi(numStr); err == nil {
			HandleStardimaQuality(ctx, num)
		} else {
			sendMessage(ctx, "يرجى كتابة رقم الجودة بشكل صحيح، مثال: .جودة 1")
		}
	}"""
content = re.sub(r'\} else if strings\.HasPrefix\(text, "\.حلقة"\) \{.*?\n\t\}', route_code, content, flags=re.DOTALL)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
