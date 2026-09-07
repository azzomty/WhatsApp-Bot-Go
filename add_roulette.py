import re
with open("internal/commands/commands.go", "r") as f:
    c = f.read()

target1 = """	case ".مشرفين القروبات", ".المشرفين":
		HandleAdminCover(ctx)
		return"""
new1 = """	case ".مشرفين القروبات", ".المشرفين":
		HandleAdminCover(ctx)
		return
	case ".روليت":
		HandleRoulette(ctx)
		return"""
c = c.replace(target1, new1)

c = c.replace('twoWordCmd == ".تعديل رد"', 'twoWordCmd == ".تعديل رد" || twoWordCmd == ".روليت"')

with open("internal/commands/commands.go", "w") as f:
    f.write(c)

