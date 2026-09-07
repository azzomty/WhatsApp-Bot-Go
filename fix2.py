import re

with open("internal/pinterest/media_search.go", "r") as f:
    ms = f.read()
ms = ms.replace('pins := SearchPinterest(searchQuery, "all", count)', 'pins, _ := SearchPinterest(searchQuery, "all", count, "")')
with open("internal/pinterest/media_search.go", "w") as f:
    f.write(ms)

with open("internal/pinterest/pinterest.go", "r") as f:
    p = f.read()
p = p.replace('pins := SearchPinterest(query+" gif", "all", count+20)', 'pins, _ := SearchPinterest(query+" gif", "all", count+20, "")')
with open("internal/pinterest/pinterest.go", "w") as f:
    f.write(p)
