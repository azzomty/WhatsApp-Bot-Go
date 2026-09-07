import re

with open('internal/commands/media.go', 'r') as f:
    content = f.read()

# 1. Update the `.كرتون` list msg to remove emojis
content = content.replace(
    'msg := "🔍 *اختر الجزء أو الكرتون المطلوب بدقة، واكتب اسمه كاملاً:*\\n\\n"',
    'msg := "*اختر الجزء أو الكرتون المطلوب بدقة، واكتب اسمه كاملاً:*\\n\\n"'
).replace(
    'msg += fmt.Sprintf("▪️ %s\\n", r.Title)',
    'msg += fmt.Sprintf("- %s\\n", r.Title)'
).replace(
    'msg += fmt.Sprintf("\\nمثال: `%s %s`", cmd, results[0].Title)',
    'msg += fmt.Sprintf("\\nمثال:\\n`.الجزء %s`", strings.TrimPrefix(results[0].Title, cartoonSearchSessions[ctx.Sender.User]+" "))'
)

# 2. Add cartoonSearchSessions tracking
if 'cartoonSearchSessions = make(map[string]string)' not in content:
    content = content.replace(
        'cartoonSessions = make(map[string]string)',
        'cartoonSessions = make(map[string]string)\n\tcartoonSearchSessions = make(map[string]string)'
    )

content = content.replace(
    'results = SearchArabicCartoon(query)',
    'results = SearchArabicCartoon(query)\n\t\tcartoonMutex.Lock()\n\t\tcartoonSearchSessions[ctx.Sender.User] = query\n\t\tcartoonMutex.Unlock()'
)


# 3. Add the two new commands logic at the end of media.go
new_funcs = """

func HandlePartCommand(ctx *BotContext) {
	partQuery := strings.TrimSpace(strings.TrimPrefix(ctx.Text, ".الجزء"))
	
	cartoonMutex.Lock()
	baseSearch, ok := cartoonSearchSessions[ctx.Sender.User]
	cartoonMutex.Unlock()
	
	if !ok || baseSearch == "" {
		sendMessage(ctx, "يرجى البحث عن الكرتون أولاً باستخدام أمر .كرتون")
		return
	}
	
	// Reconstruct the full title (e.g. "كونان الجزء الثاني")
	fullTitle := baseSearch + " الجزء " + partQuery
	if partQuery == "" {
		sendMessage(ctx, "يرجى تحديد الجزء، مثال: .الجزء الثاني")
		return
	}
	
	results := SearchArabicCartoon(fullTitle)
	if len(results) == 0 {
		sendMessage(ctx, "لم أتمكن من العثور على هذا الجزء!")
		return
	}
	
	sendMediaResult(ctx, results[0], ".كرتون")
}

func HandleCartoonList(ctx *BotContext) {
	sendMessage(ctx, "جاري جلب القائمة... ⏳")
	
	reqURL := "https://wwmdrwjkrzdkqjqddfta.supabase.co/rest/v1/series?select=title&order=title.asc"
	req, _ := http.NewRequest("GET", reqURL, nil)
	req.Header.Set("apikey", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Ind3bWRyd2prcnpka3FqcWRkZnRhIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODA4MjAxNzUsImV4cCI6MjA5NjM5NjE3NX0.v3-gjEYfuJ4DE17OAHidvd38lCHUTU4ldb2SHLphU8s")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء جلب القائمة.")
		return
	}
	defer resp.Body.Close()
	
	var data []struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		sendMessage(ctx, "حدث خطأ أثناء جلب القائمة.")
		return
	}
	
	// Deduplicate base names (e.g. "المحقق كونان الجزء الأول" -> "المحقق كونان")
	uniqueShows := make(map[string]bool)
	var shows []string
	
	for _, item := range data {
		baseName := strings.Split(item.Title, " الجزء ")[0]
		baseName = strings.Split(baseName, " الموسم ")[0]
		if !uniqueShows[baseName] {
			uniqueShows[baseName] = true
			shows = append(shows, baseName)
		}
	}
	
	msg := "*قائمة الكراتين المتوفرة:*\n\n"
	for _, show := range shows {
		msg += "- " + show + "\n"
	}
	msg += "\n*للبحث عن أي كرتون، اكتب:* `.كرتون اسم_الكرتون`"
	
	sendMessage(ctx, msg)
}
"""

if "HandlePartCommand" not in content:
    content += new_funcs

with open('internal/commands/media.go', 'w') as f:
    f.write(content)

# 4. Also register these in commands.go
with open('internal/commands/commands.go', 'r') as f:
    cmd_content = f.read()

if "case \".الجزء\":" not in cmd_content:
    cmd_content = cmd_content.replace(
        'case ".حلقة":',
        'case ".الجزء":\n\t\tHandlePartCommand(ctx)\n\tcase ".قائمة الكراتين":\n\t\tHandleCartoonList(ctx)\n\tcase ".حلقة":'
    )

with open('internal/commands/commands.go', 'w') as f:
    f.write(cmd_content)

