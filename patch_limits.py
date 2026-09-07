with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

# Replace the > 100 limit
content = content.replace('''	if count > 100 {
		count = 100
	}''', '''	if count > 500 {
		count = 500
	}''')

# Replace the > 20 limit for foryou
content = content.replace('''	if count > 20 {
		count = 20
	}''', '''	if count > 200 {
		count = 200
	}''')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
print("Patched limits")
