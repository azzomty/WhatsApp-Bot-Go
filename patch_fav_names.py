with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

old_logic = """		if strings.HasPrefix(afterSlayer, "مفضلة") {"""

new_logic = """		if strings.HasPrefix(afterSlayer, "مفضلة") || strings.HasPrefix(afterSlayer, "مفضله") || strings.HasPrefix(afterSlayer, "المفضلة") || strings.HasPrefix(afterSlayer, "المفضله") {"""

if old_logic in content:
    content = content.replace(old_logic, new_logic)
    with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
        f.write(content)
    print("Patched names!")
