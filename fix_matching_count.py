with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

import re

# Add count definition
content = content.replace('''	query := strings.TrimSpace(strings.Replace(ctx.Text, ".تطقيم", "", 1))
	
	sendMessage(ctx, "جاري البحث عن تطقيمات")''', '''	query := strings.TrimSpace(strings.Replace(ctx.Text, ".تطقيم", "", 1))
	count := 20
	sendMessage(ctx, "جاري البحث عن تطقيمات")''')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
