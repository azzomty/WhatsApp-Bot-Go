import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

# Add to Handle function
# Before HandleEpisodeCommand
add = """		if strings.HasPrefix(ctx.Text, ".انمي سلاير ") {
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
		}"""

content = re.sub(r'\t\tHandleSearchCommand\(ctx\)\n\t} else if strings\.HasPrefix\(ctx\.Text, "\.رقم "\) \{\n\t\tparts := strings\.Split\(ctx\.Text, " "\)\n\t\tif len\(parts\) > 1 \{\n\t\t\tidx, _ := strconv\.Atoi\(parts\[1\]\)\n\t\t\tHandleNumberSelect\(ctx, idx\)\n\t\t\} else \{\n\t\t\tsendMessage\(ctx, "يرجى تحديد الرقم، مثال: \.رقم 1"\)\n\t\t\}\n\t\} else if strings\.HasPrefix\(ctx\.Text, "\.حلقة "\) \{\n\t\tHandleEpisodeCommand\(ctx\)\n\t\}', add, content)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
