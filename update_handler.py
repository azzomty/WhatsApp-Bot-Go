import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'r') as f:
    content = f.read()

old_str = """	if query == "" {
		sendMessage(ctx, "يرجى كتابة اسم الكرتون بعد الأمر، مثال: .ستارديما داني الشبح")
		return
	}"""

new_str = """	if query == "" {
		sendMessage(ctx, "يرجى كتابة اسم الكرتون بعد الأمر، مثال: .ستارديما داني الشبح")
		return
	}
	
	if query == "قائمة الافلام" || query == "قائمة الأفلام" {
		HandleStardimaList(ctx, "aflam")
		return
	}
	if query == "قائمة الكراتين" {
		HandleStardimaList(ctx, "mosalsalat")
		return
	}"""

if old_str in content:
    content = content.replace(old_str, new_str)
    with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'w') as f:
        f.write(content)
    print("Replaced!")
else:
    print("Not found")
