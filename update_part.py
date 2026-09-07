import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'r') as f:
    content = f.read()

old_logic = """	sendMessage(ctx, fmt.Sprintf("تم اختيار الموسم: *%s*\nلتحميل حلقة اكتب: `.حلقة` متبوعاً بالرقم (مثال: `.حلقة 1`)", selSeason.Name))"""

new_logic = """	go func() {
		episodes, err := GetStardimaEpisodes(selSeason.ID)
		countText := ""
		if err == nil && len(episodes) > 0 {
			countText = fmt.Sprintf("\\nعدد حلقاته: *%d حلقة*", len(episodes))
		}
		
		sendMessage(ctx, fmt.Sprintf("تم اختيار الموسم: *%s*%s\\nلتحميل حلقة اكتب: `.حلقة` متبوعاً بالرقم (مثال: `.حلقة 1`)", selSeason.Name, countText))
	}()"""

if old_logic in content:
    content = content.replace(old_logic, new_logic)
    with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'w') as f:
        f.write(content)
    print("Replaced!")
else:
    print("Not found")
