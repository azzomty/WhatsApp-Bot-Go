import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

# Add the new commands to the list of handled commands in ProcessMessage
if 'HandleFontsCommand(ctx)' not in content:
    # Find a good place to insert HandleFontsCommand
    insert_point = '	if HandleDownloadCommand(ctx) {\n\t\treturn\n\t}'
    replacement = """	if HandleDownloadCommand(ctx) {
		return
	}
	if HandleFontsCommand(ctx) {
		return
	}"""
    content = content.replace(insert_point, replacement)

# Add HandleFontsCommand function at the end
new_func = """
func HandleFontsCommand(ctx *BotContext) bool {
	text := strings.TrimSpace(ctx.Text)
	if text == ".قائمة الخطوط" || text == ".fonts list" {
		sampleTextEn := "Abc 123"
		sampleTextAr := "تجربة"
		
		var msgs []string
		var currentMsg strings.Builder
		currentMsg.WriteString("📜 *قائمة الخطوط المتوفرة:*\n\n")
		
		for i, font := range FontStyles {
			// Apply font to sample
			sampleEn := ApplyFont(sampleTextEn, i)
			sampleAr := ApplyFont(sampleTextAr, i)
			
			line := fmt.Sprintf("*%d.* %s | %s - %s\n", i+1, sampleEn, sampleAr, font.Name)
			if currentMsg.Len() + len(line) > 3000 {
				msgs = append(msgs, currentMsg.String())
				currentMsg.Reset()
			}
			currentMsg.WriteString(line)
		}
		if currentMsg.Len() > 0 {
			msgs = append(msgs, currentMsg.String())
		}
		
		for _, m := range msgs {
			sendMessage(ctx, m)
			time.Sleep(500 * time.Millisecond) // avoid spam
		}
		sendMessage(ctx, "✍️ للاستخدام: اكتب `.خط رقم_الخط` ثم انزل سطر واكتب نصك.\nمثال:\n.خط 5\nمرحبا بك")
		return true
	}
	
	if strings.HasPrefix(text, ".خط ") {
		parts := strings.SplitN(text, "\\n", 2)
		if len(parts) < 2 {
			sendMessage(ctx, "يرجى كتابة رقم الخط ثم النزول سطر جديد وكتابة النص.\nمثال:\n.خط 5\nhello")
			return true
		}
		
		// Parse the first line to get the number
		firstLine := strings.TrimSpace(parts[0])
		numStr := strings.TrimSpace(strings.TrimPrefix(firstLine, ".خط"))
		numStr = strings.TrimSpace(strings.TrimPrefix(numStr, "رقم")) // In case they write .خط رقم 50
		numStr = strings.TrimSpace(numStr)
		
		num, err := strconv.Atoi(numStr)
		if err != nil || num < 1 || num > len(FontStyles) {
			sendMessage(ctx, fmt.Sprintf("رقم الخط غير صحيح. يرجى اختيار رقم من 1 إلى %d.", len(FontStyles)))
			return true
		}
		
		// 0-indexed internally
		fontIdx := num - 1
		textToFormat := parts[1]
		
		formatted := ApplyFont(textToFormat, fontIdx)
		sendMessage(ctx, formatted)
		return true
	}
	
	return false
}
"""

if 'func HandleFontsCommand' not in content:
    content += new_func

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
print("Added Font commands")
