import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'r') as f:
    content = f.read()

old_str = """	sendMessage(ctx, fmt.Sprintf("تم اختيار الموسم: *%s*\\nلتحميل حلقة اكتب: `.حلقة` متبوعاً بالرقم (مثال: `.حلقة 1`)", selSeason.Name))
}"""

new_str = """	sendMessage(ctx, fmt.Sprintf("تم اختيار الموسم: *%s*\\nجاري حساب عدد الحلقات...", selSeason.Name))
	
	go func() {
		episodes, err := GetStardimaEpisodes(selSeason.ID)
		if err == nil && len(episodes) > 0 {
			sendMessage(ctx, fmt.Sprintf("الموسم يحتوي على *%d حلقة*\\nلتحميل حلقة اكتب: `.حلقة` متبوعاً بالرقم (مثال: `.حلقة 1`)", len(episodes)))
		} else {
			sendMessage(ctx, "لتحميل حلقة اكتب: `.حلقة` متبوعاً بالرقم (مثال: `.حلقة 1`)")
		}
	}()
}"""

if old_str in content:
    content = content.replace(old_str, new_str)
    with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'w') as f:
        f.write(content)
    print("Replaced!")
else:
    print("Not found!")
