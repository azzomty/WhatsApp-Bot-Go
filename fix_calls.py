import re

# media_search.go
with open("internal/pinterest/media_search.go", "r") as f:
    ms = f.read()
ms = ms.replace('results := SearchPinterest(query, aspect, count)', 'results, _ := SearchPinterest(query, aspect, count, "")')
with open("internal/pinterest/media_search.go", "w") as f:
    f.write(ms)

# pinterest.go
with open("internal/pinterest/pinterest.go", "r") as f:
    p = f.read()
p = p.replace('pins := SearchPinterest("matching icons "+query, "all", 10)', 'pins, _ := SearchPinterest("matching icons "+query, "all", 10, "")')
p = p.replace('results := SearchPinterest(query, "gif", 5)', 'results, _ := SearchPinterest(query, "gif", 5, "")')
with open("internal/pinterest/pinterest.go", "w") as f:
    f.write(p)
