import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'r') as f:
    content = f.read()

old_err = """	resp, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaVideo)
	if err != nil {
		sendMessage(ctx, "فشل في رفع الحلقة للواتساب.")
		return
	}"""

new_err = """	resp, err := ctx.Client.Upload(context.Background(), data, whatsmeow.MediaVideo)
	if err != nil {
		fmt.Println("UPLOAD ERROR:", err)
		sendMessage(ctx, "فشل في رفع الحلقة للواتساب: حجمها كبير جداً للواتساب.")
		return
	}"""

content = content.replace(old_err, new_err)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'w') as f:
    f.write(content)
