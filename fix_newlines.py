with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

import re

# We will just replace the bad HandleFontsCommand with a clean one
pattern = re.compile(r'func HandleFontsCommand\(ctx \*BotContext\) bool \{.*\}', re.DOTALL)
clean_func = """func HandleFontsCommand(ctx *BotContext) bool {
	text := strings.TrimSpace(ctx.Text)
	if text == ".قائمة الخطوط" || text == ".fonts list" {
		sampleTextEn := "Abc 123"
		sampleTextAr := "تجربة"
		
		var msgs []string
		var currentMsg strings.Builder
		currentMsg.WriteString("📜 *قائمة الخطوط المتوفرة:*\\n\\n")
		
		for i, font := range FontStyles {
			sampleEn := ApplyFont(sampleTextEn, i)
			sampleAr := ApplyFont(sampleTextAr, i)
			
			line := fmt.Sprintf("*%d.* %s | %s - %s\\n", i+1, sampleEn, sampleAr, font.Name)
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
		sendMessage(ctx, "✍️ للاستخدام: اكتب `.خط رقم_الخط` ثم انزل سطر واكتب نصك.\\nمثال:\\n.خط 5\\nمرحبا بك")
		return true
	}
	
	if strings.HasPrefix(text, ".خط ") {
		parts := strings.SplitN(text, "\\n", 2)
		if len(parts) < 2 {
			sendMessage(ctx, "يرجى كتابة رقم الخط ثم النزول سطر جديد وكتابة النص.\\nمثال:\\n.خط 5\\nhello")
			return true
		}
		
		firstLine := strings.TrimSpace(parts[0])
		numStr := strings.TrimSpace(strings.TrimPrefix(firstLine, ".خط"))
		numStr = strings.TrimSpace(strings.TrimPrefix(numStr, "رقم"))
		numStr = strings.TrimSpace(numStr)
		
		num, err := strconv.Atoi(numStr)
		if err != nil || num < 1 || num > len(FontStyles) {
			sendMessage(ctx, fmt.Sprintf("رقم الخط غير صحيح. يرجى اختيار رقم من 1 إلى %d.", len(FontStyles)))
			return true
		}
		
		fontIdx := num - 1
		textToFormat := parts[1]
		
		formatted := ApplyFont(textToFormat, fontIdx)
		sendMessage(ctx, formatted)
		return true
	}
	
	return false
}"""

content = pattern.sub(clean_func, content)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
print("Fixed newlines in fonts command")
