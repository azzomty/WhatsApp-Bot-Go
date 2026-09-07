import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

content = content.replace("""		HandleSearchCommand(ctx)
	} else if strings.HasPrefix(ctx.Text, ".رقم ") {
		parts := strings.Split(ctx.Text, " ")
		if len(parts) > 1 {
			idx, _ := strconv.Atoi(parts[1])
			HandleNumberSelect(ctx, idx)
		} else {
			sendMessage(ctx, "يرجى تحديد الرقم، مثال: .رقم 1")
		}
	} else if strings.HasPrefix(ctx.Text, ".حلقة ") {
		HandleEpisodeCommand(ctx)
	}""", """		if strings.HasPrefix(ctx.Text, ".انمي سلاير ") {
			HandleAnslayerCommand(ctx)
		} else {
			HandleSearchCommand(ctx)
		}
	} else if strings.HasPrefix(ctx.Text, ".رقم ") {
		parts := strings.Split(ctx.Text, " ")
		if len(parts) > 1 {
			idx, _ := strconv.Atoi(parts[1])
			if !HandleAnslayerNumberSelect(ctx, idx) {
				HandleNumberSelect(ctx, idx)
			}
		} else {
			sendMessage(ctx, "يرجى تحديد الرقم، مثال: .رقم 1")
		}
	} else if strings.HasPrefix(ctx.Text, ".حلقة ") {
		parts := strings.Split(ctx.Text, " ")
		if len(parts) > 1 {
			idx, _ := strconv.Atoi(parts[1])
			if !HandleAnslayerEpisodeSelect(ctx, idx) {
				HandleEpisodeCommand(ctx)
			}
		} else {
			sendMessage(ctx, "يرجى تحديد رقم الحلقة، مثال: .حلقة 1")
		}
	}""")

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
