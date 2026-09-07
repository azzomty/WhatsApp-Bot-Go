import re

with open('internal/commands/media.go', 'r') as f:
    content = f.read()

# Replace the state session
if 'cartoonListSessions = make(map[string][]MediaResult)' not in content:
    content = content.replace(
        'cartoonSearchSessions = make(map[string]string)',
        'cartoonListSessions = make(map[string][]MediaResult)'
    )

content = content.replace(
    'cartoonSearchSessions[ctx.Sender.User] = query',
    'cartoonListSessions[ctx.Sender.User] = results'
)

# Replace the example text
content = re.sub(
    r'msg \+= fmt\.Sprintf\("\\nمثال:\\n`\.الجزء %s`", strings\.TrimPrefix\(results\[0\]\.Title, cartoonSearchSessions\[ctx\.Sender\.User\]\+" "\)\)',
    'msg += fmt.Sprintf("\\nمثال:\\n`.الجزء الأول`")',
    content
)

# Rewrite HandlePartCommand
new_handle_part = """func HandlePartCommand(ctx *BotContext) {
	partQuery := strings.TrimSpace(strings.TrimPrefix(ctx.Text, ".الجزء"))
	if partQuery == "" {
		sendMessage(ctx, "يرجى تحديد الجزء، مثال: .الجزء الثاني")
		return
	}
	
	cartoonMutex.Lock()
	results, ok := cartoonListSessions[ctx.Sender.User]
	cartoonMutex.Unlock()
	
	if !ok || len(results) == 0 {
		sendMessage(ctx, "يرجى البحث عن الكرتون أولاً باستخدام أمر .كرتون")
		return
	}
	
	var selected MediaResult
	found := false
	for _, r := range results {
		if strings.Contains(r.Title, partQuery) {
			selected = r
			found = true
			break
		}
	}
	
	if !found {
		sendMessage(ctx, "لم أتمكن من العثور على هذا الجزء في نتائج بحثك السابقة! تأكد من كتابة الاسم الصحيح كما ظهر في القائمة.")
		return
	}
	
	sendMediaResult(ctx, selected, ".كرتون")
}"""

content = re.sub(r'func HandlePartCommand\(ctx \*BotContext\) \{.*?\n\}\n', new_handle_part + "\n", content, flags=re.DOTALL)

with open('internal/commands/media.go', 'w') as f:
    f.write(content)
