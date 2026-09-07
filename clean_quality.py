import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

content = re.sub(r'	case "\.جودة":\n.*?sendMessage\(ctx, "يرجى تحديد الرقم، مثال: \.جودة 1"\)\n\t\t}', '', content, flags=re.DOTALL)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
