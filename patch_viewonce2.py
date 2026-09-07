import re
with open("main.go", "r") as f:
    c = f.read()

target = """				if err == nil && len(data) > 0 {
					resp, err := client.Upload(context.Background(), data, mediaType)"""
new_target = """				if err != nil {
					client.SendMessage(context.Background(), client.Store.ID.ToNonAD(), &waProto.Message{
						ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String("فشل تحميل الميديا من رسالة العرض لمرة واحدة: " + err.Error())},
					})
				}
				if err == nil && len(data) > 0 {
					resp, err := client.Upload(context.Background(), data, mediaType)
					if err != nil {
					    client.SendMessage(context.Background(), client.Store.ID.ToNonAD(), &waProto.Message{
						    ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String("فشل رفع الميديا لرسالة العرض لمرة واحدة: " + err.Error())},
					    })
					}"""
c = c.replace(target, new_target)
with open("main.go", "w") as f:
    f.write(c)
