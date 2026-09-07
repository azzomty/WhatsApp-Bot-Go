import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

# 1. Update go monitorComments call
content = content.replace("go monitorComments(epIDStr, ch)", "go monitorComments(ctx, epIDStr, ch)")

# 2. Update monitorComments signature
old_sig = "func monitorComments(epID string, stopCh chan struct{}) {"
new_sig = "func monitorComments(ctx *BotContext, epID string, stopCh chan struct{}) {"
content = content.replace(old_sig, new_sig)

# 3. Add notification
old_notif = """				if !finishedNotified {
					// Need a ctx to send message! But monitorComments doesn't have ctx!
					// Let's pass chatID and use a global sender?
					// Actually, monitorComments is passed stopCh but NO ctx.
					// I can just print it to logs, but the user wants a notification!
				}"""
new_notif = """				if !finishedNotified {
					sendMessage(ctx, "✅ تم الرد على جميع التعليقات القديمة في هذه الحلقة! البوت الآن في وضع الاستعداد للتعليقات الجديدة فقط، يمكنك اختيار حلقة أخرى إذا أردت.")
					finishedNotified = true
				}"""
content = content.replace(old_notif, new_notif)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
    f.write(content)
