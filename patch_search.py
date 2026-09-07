import re

with open('internal/commands/media.go', 'r') as f:
    content = f.read()

new_search = """func SearchArabicCartoon(query string) []MediaResult {
	q := strings.ReplaceAll(query, "ي", "_")
	q = strings.ReplaceAll(q, "ى", "_")
	q = strings.ReplaceAll(q, "أ", "_")
	q = strings.ReplaceAll(q, "إ", "_")
	q = strings.ReplaceAll(q, "آ", "_")
	q = strings.ReplaceAll(q, "ا", "_")
	q = strings.ReplaceAll(q, "ة", "_")
	q = strings.ReplaceAll(q, "ه", "_")
	
	// Because url.QueryEscape escapes '_' as well? No, '_' is not escaped.
	// We want to pass %25 for wildcard, and _ for single char.
	escapedQ := strings.ReplaceAll(url.QueryEscape(q), "+", "%20")
	
	reqURL := fmt.Sprintf("https://wwmdrwjkrzdkqjqddfta.supabase.co/rest/v1/series?select=*&title=ilike.*%%25%s%%25*", escapedQ)
"""

content = re.sub(
    r'func SearchArabicCartoon\(query string\) \[\]MediaResult \{\n\s*reqURL := fmt\.Sprintf\("https://wwmdrwjkrzdkqjqddfta\.supabase\.co/rest/v1/series\?select=\*&title=ilike\.\*%%25%s%%25\*", strings\.ReplaceAll\(url\.QueryEscape\(query\), "\+", "%20"\)\)',
    new_search,
    content
)

with open('internal/commands/media.go', 'w') as f:
    f.write(content)
