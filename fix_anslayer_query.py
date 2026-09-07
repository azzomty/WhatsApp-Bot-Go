import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

old_logic = """	var query string
	if mode == "marketing" {
		if len(parts) < 2 {
			sendMessage(ctx, "يرجى كتابة اسم الأنمي بعد الأمر، مثلا:\n.انمي سلاير ون بيس\nأو لحفظ رسالة النشر:\n.انمي سلاير نشر رسالتي هنا")
			return
		}
		query = strings.Join(parts[1:], " ")
	} else {
		if len(parts) < 2 {
			sendMessage(ctx, "يرجى كتابة اسم الأنمي بعد الأمر، مثلا:\n.انمي ون بيس")
			return
		}
		query = strings.Join(parts[1:], " ")
	}"""

new_logic = """	var query string
	if mode == "marketing" {
		if len(parts) < 3 {
			sendMessage(ctx, "يرجى كتابة اسم الأنمي بعد الأمر، مثلا:\n.انمي سلاير ون بيس\nأو لحفظ رسالة النشر:\n.انمي سلاير نشر رسالتي هنا")
			return
		}
		query = parts[2]
	} else {
		// mode == "watch"
		// text is `.انمي ون بيس`
		// parts is `[".انمي", "ون", "بيس"]` if we split normally, but here we used SplitN(ctx.Text, " ", 3)
		// Wait, SplitN(..., 3) on `.انمي ون بيس` gives `[".انمي", "ون", "بيس"]`
		// But query should be "ون بيس"!
		fullParts := strings.SplitN(ctx.Text, " ", 2)
		if len(fullParts) < 2 {
			sendMessage(ctx, "يرجى كتابة اسم الأنمي بعد الأمر، مثلا:\n.انمي ون بيس")
			return
		}
		query = fullParts[1]
	}"""

content = content.replace(old_logic, new_logic)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
    f.write(content)
