import re

with open('internal/commands/media.go', 'r') as f:
    content = f.read()

bad_block = r'''			if \(cmd == "\.كرتون" \|\| cmd == "\.انمي_مدبلج"\) && len\(results\) > 1 \{
				msg := "🔍 \*اختر الجزء أو الكرتون المطلوب بدقة، واكتب اسمه كاملاً:\*

"
				for i, r := range results \{
					if i >= 20 \{
						break
					\}
					msg \+= fmt\.Sprintf\("▪️ %s
", r\.Title\)
				\}
				msg \+= fmt\.Sprintf\("
مثال: `%s %s`", cmd, results\[0\]\.Title\)
				sendMessage\(ctx, msg\)
				return
			\}'''

good_block = """		if (cmd == ".كرتون" || cmd == ".انمي_مدبلج") && len(results) > 1 {
			msg := "🔍 *اختر الجزء أو الكرتون المطلوب بدقة، واكتب اسمه كاملاً:*\\n\\n"
			for i, r := range results {
				if i >= 20 {
					break
				}
				msg += fmt.Sprintf("▪️ %s\\n", r.Title)
			}
			msg += fmt.Sprintf("\\nمثال: `%s %s`", cmd, results[0].Title)
			sendMessage(ctx, msg)
			return
		}"""

content = re.sub(bad_block, good_block, content)

with open('internal/commands/media.go', 'w') as f:
    f.write(content)
