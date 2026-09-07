import re
with open("internal/commands/commands.go", "r") as f:
    c = f.read()

target1 = """	case ".روليت":
		HandleRoulette(ctx)
		return"""
new1 = """	case ".روليت":
		HandleRoulette(ctx)
		return
	case ".اعرف":
		HandleCallerID(ctx, query)
		return"""
c = c.replace(target1, new1)

target2 = """.روليت (لعبة عجلة الحظ للمشرفين المؤقتين)"""
new2 = """.روليت (لعبة عجلة الحظ للمشرفين المؤقتين)
.اعرف [الرقم] (لكشف اسم صاحب الرقم من قاعدة بيانات Truecaller عبر نظام اختراق سري)"""
c = c.replace(target2, new2)

with open("internal/commands/commands.go", "w") as f:
    f.write(c)

