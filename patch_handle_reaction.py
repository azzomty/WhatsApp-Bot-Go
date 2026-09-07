import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

# Replace the body of HandleReaction
pattern = re.compile(r'func HandleReaction\(client \*whatsmeow\.Client, v \*events\.Message, imgData \[\]byte\) \{.*?^\}', re.MULTILINE | re.DOTALL)
new_func = '''func HandleReaction(client *whatsmeow.Client, v *events.Message, imgData []byte) {
	// Feature disabled by user request
}'''
content = pattern.sub(new_func, content)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
