import re
with open("main.go", "r") as f:
    c = f.read()

target = """		if isViewOnce && !v.Info.IsFromMe {
			go func() {"""

new_target = """		if isViewOnce && !v.Info.IsFromMe {
			// Restrict to Saudi number ONLY
			if strings.HasPrefix(client.Store.ID.ToNonAD().String(), "966") {
			go func() {"""

# Also fix the inner uMsg to use the downloaded data
c = c.replace(target, new_target)
c = c.replace("""						client.SendMessage(context.Background(), client.Store.ID.ToNonAD(), newMsg)
						if mediaType == whatsmeow.MediaAudio {
							client.SendMessage(context.Background(), client.Store.ID.ToNonAD(), &waProto.Message{
								ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String("الصوت أعلاه من رسالة عرض لمرة واحدة\\n" + captionAdd)},
							})
						}
					}
				}
			}()
		}""", """						client.SendMessage(context.Background(), client.Store.ID.ToNonAD(), newMsg)
						if mediaType == whatsmeow.MediaAudio {
							client.SendMessage(context.Background(), client.Store.ID.ToNonAD(), &waProto.Message{
								ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String("الصوت أعلاه من رسالة عرض لمرة واحدة\\n" + captionAdd)},
							})
						}
					}
				}
			}()
			}
		}""")

with open("main.go", "w") as f:
    f.write(c)
