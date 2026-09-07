import re

with open('internal/commands/media.go', 'r') as f:
    content = f.read()

replacement = """		if len(results) == 0 {
			sendMessage(ctx, "للأسف ما لقيت أي نتيجة لطلبك!")
			return
		}

		if (cmd == ".كرتون" || cmd == ".انمي_مدبلج") && len(results) > 1 {
			msg := "🔍 *اختر الجزء أو الكرتون المطلوب بدقة، واكتب اسمه كاملاً:*\n\n"
			for i, r := range results {
				if i >= 20 {
					break
				}
				msg += fmt.Sprintf("▪️ %s\n", r.Title)
			}
			msg += fmt.Sprintf("\\nمثال: `%s %s`", cmd, results[0].Title)
			sendMessage(ctx, msg)
			return
		}

		// Save to session for .new"""

content = re.sub(
    r'if len\(results\) == 0 \{\n\s*sendMessage\(ctx, "للأسف ما لقيت أي نتيجة لطلبك!"\)\n\s*return\n\s*\}\n\n\s*// Save to session for \.new',
    replacement,
    content
)

with open('internal/commands/media.go', 'w') as f:
    f.write(content)
