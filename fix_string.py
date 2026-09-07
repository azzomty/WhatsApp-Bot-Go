import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'r') as f:
    content = f.read()

content = re.sub(r'msg := "يرجى اختيار الجودة المطلوبة:.*?\)"', 'msg := "يرجى اختيار الجودة المطلوبة:\\n1. جودة 1080p (الأعلى - سيتم تقسيمها لأجزاء لو حجمها كبير)\\n2. جودة 720p (عالية - فيديو واحد)\\n3. جودة 480p (متوسطة وسريعة - فيديو واحد)\\n\\nللاختيار اكتب: `.جودة` متبوعاً بالرقم (مثال: `.جودة 1`)"', content, flags=re.DOTALL)
content = content.replace('\\n(الجزء', '\\n(الجزء')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'w') as f:
    f.write(content)
