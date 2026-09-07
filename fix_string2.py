import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'r') as f:
    content = f.read()

def replace_with_backticks(match):
    return 'msg := `يرجى اختيار الجودة المطلوبة:\n1. جودة 1080p (الأعلى - سيتم تقسيمها لأجزاء لو حجمها كبير)\n2. جودة 720p (عالية - فيديو واحد)\n3. جودة 480p (متوسطة وسريعة - فيديو واحد)\n\nللاختيار اكتب: .جودة متبوعاً بالرقم (مثال: .جودة 1)`'

content = re.sub(r'msg := "يرجى اختيار الجودة المطلوبة:.*?\)"', replace_with_backticks, content, flags=re.DOTALL)
# Also fix the caption string that has \n(الجزء
content = re.sub(r'caption := fmt\.Sprintf\("\*%s\* - الحلقة %s\\n\(الجزء.*?"\)', 'caption := fmt.Sprintf("*%s* - الحلقة %s\\n(الجزء %d من %d)", animeName, epNum, i+1, len(parts))', content)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'w') as f:
    f.write(content)
