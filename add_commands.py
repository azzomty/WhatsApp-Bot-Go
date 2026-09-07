import re
with open("internal/commands/commands.go", "r") as f:
    c = f.read()

target1 = """	case ".مواعيد صلاة", ".مواعيد الصلاة":
		HandlePrayerTimes(ctx, query)
		return"""
new1 = """	case ".مواعيد صلاة", ".مواعيد الصلاة":
		HandlePrayerTimes(ctx, query)
		return
	case ".مشرفين القروبات", ".المشرفين":
		HandleAdminCover(ctx)
		return"""
c = c.replace(target1, new1)

target2 = """		if twoWordCmd == ".مواعيد صلاة" || twoWordCmd == ".مواعيد الصلاة" || twoWordCmd == ".فك ميوت" """
new2 = """		if twoWordCmd == ".مواعيد صلاة" || twoWordCmd == ".مشرفين القروبات" || twoWordCmd == ".مواعيد الصلاة" || twoWordCmd == ".فك ميوت" """
c = c.replace(target2, new2)
c = c.replace(target2.replace('		if ', '				if '), new2.replace('		if ', '				if '))

target3 = """		if twoWordCmd == ".مواعيد صلاة" || twoWordCmd == ".مشرفين القروبات" || twoWordCmd == ".مواعيد الصلاة" || twoWordCmd == ".فك ميوت" """
new3 = """		if twoWordCmd == ".مواعيد صلاة" || twoWordCmd == ".مشرفين القروبات" || twoWordCmd == ".مواعيد الصلاة" || twoWordCmd == ".فك ميوت" """
# wait, the first replace already did it.

target4 = """.توقيت [المدينة] (لمعرفة التوقيت الحالي)
.طقس [المدينة] (لمعرفة حالة الطقس)"""
new4 = """.توقيت [المدينة] (لمعرفة التوقيت الحالي)
.طقس [المدينة] (لمعرفة حالة الطقس)
.مشرفين القروبات (لاستخراج أرقام أقل عدد مشرفين يغطون كل قروباتك)
.تنظيف القروبات (لحذف رسائل كل القروبات يدوياً - البوت يحذفها تلقائياً كل ساعتين)"""
c = c.replace(target4, new4)

target_manual_clear = """	case ".مشرفين القروبات", ".المشرفين":
		HandleAdminCover(ctx)
		return"""
new_manual_clear = """	case ".مشرفين القروبات", ".المشرفين":
		HandleAdminCover(ctx)
		return
	case ".تنظيف القروبات":
		sendMessage(ctx, "جاري حذف رسائل جميع القروبات التي أنت بها...")
		ClearAllGroups(ctx.Client)
		sendMessage(ctx, "تم حذف رسائل جميع القروبات بنجاح! ✅")
		return"""
c = c.replace(target_manual_clear, new_manual_clear)

with open("internal/commands/commands.go", "w") as f:
    f.write(c)

