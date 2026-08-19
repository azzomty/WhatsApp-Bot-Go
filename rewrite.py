import re

with open("/home/lennox/Desktop/اهها/Go_Bot/main.go", "r") as f:
    content = f.read()

# 1. We need to move text extraction to the top of case *events.Message:
# Find: case *events.Message:\n
new_top = """	case *events.Message:
		if v.Info.Timestamp.Before(startupTime) {
			return
		}

		isViewOnce := false
		unwrap := func(m *waProto.Message) *waProto.Message {
			for m != nil {
				if m.EphemeralMessage != nil && m.EphemeralMessage.Message != nil {
					m = m.EphemeralMessage.Message
					continue
				}
				if m.ViewOnceMessage != nil && m.ViewOnceMessage.Message != nil {
					isViewOnce = true
					m = m.ViewOnceMessage.Message
					continue
				}
				if m.ViewOnceMessageV2 != nil && m.ViewOnceMessageV2.Message != nil {
					isViewOnce = true
					m = m.ViewOnceMessageV2.Message
					continue
				}
				if m.ViewOnceMessageV2Extension != nil && m.ViewOnceMessageV2Extension.Message != nil {
					isViewOnce = true
					m = m.ViewOnceMessageV2Extension.Message
					continue
				}
				break
			}
			return m
		}

		uMsg := unwrap(v.Message)
		text := ""
		if uMsg != nil {
			if uMsg.GetExtendedTextMessage() != nil {
				text = uMsg.GetExtendedTextMessage().GetText()
			} else if uMsg.GetConversation() != "" {
				text = uMsg.GetConversation()
			} else if uMsg.GetImageMessage() != nil {
				text = uMsg.GetImageMessage().GetCaption()
			} else if uMsg.GetVideoMessage() != nil {
				text = uMsg.GetVideoMessage().GetCaption()
			}
		}
		text = strings.TrimSpace(text)
		
		senderLID := getLID(client, v.Info.Sender)
		senderID := senderLID // for consistency

		ctx := &commands.BotContext{
			Client: client,
			Event:  v,
			ChatID: v.Info.Chat,
			Sender: v.Info.Sender,
			Text:   text,
		}

		// --- NUMBER SPLIT & ALWAYS-ON LOGIC ---
		myNumber := client.Store.ID.User

		// AutoJoin logic (runs everywhere, ignores .bot off and .baymax)
		if commands.AutoJoinGroups && strings.Contains(text, "chat.whatsapp.com/") {
			parts := strings.Split(text, "chat.whatsapp.com/")
			if len(parts) > 1 {
				code := strings.FieldsFunc(parts[1], func(r rune) bool {
					return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_')
				})
				if len(code) > 0 {
					go func() {
						client.JoinGroupWithLink(context.Background(), code[0])
						// Silently join, no message sent as requested
					}()
				}
			}
		}

		// Handle Exchange (also ignores .bot off and .baymax)
		if commands.HandleExchangeMessage(ctx) {
			return
		}

		// Syrian Number Logic (ONLY runs AutoJoin and Exchange, which is done above)
		if strings.HasPrefix(myNumber, "963") {
			return
		}
		// ----------------------------------------
"""

content = content.replace("	case *events.Message:", new_top)

# Now we need to remove the OLD text extraction logic to avoid redeclaration.
old_extraction = """		text := ""

		isViewOnce := false

		unwrap := func(m *waProto.Message) *waProto.Message {
			for m != nil {
				if m.EphemeralMessage != nil && m.EphemeralMessage.Message != nil {
					m = m.EphemeralMessage.Message
					continue
				}
				if m.ViewOnceMessage != nil && m.ViewOnceMessage.Message != nil {
					isViewOnce = true
					m = m.ViewOnceMessage.Message
					continue
				}
				if m.ViewOnceMessageV2 != nil && m.ViewOnceMessageV2.Message != nil {
					isViewOnce = true
					m = m.ViewOnceMessageV2.Message
					continue
				}
				if m.ViewOnceMessageV2Extension != nil && m.ViewOnceMessageV2Extension.Message != nil {
					isViewOnce = true
					m = m.ViewOnceMessageV2Extension.Message
					continue
				}
				break
			}
			return m
		}

		uMsg := unwrap(v.Message)

		text = ""
		if uMsg.GetExtendedTextMessage() != nil {
			text = uMsg.GetExtendedTextMessage().GetText()
		} else if uMsg.GetConversation() != "" {
			text = uMsg.GetConversation()
		} else if uMsg.GetImageMessage() != nil {
			text = uMsg.GetImageMessage().GetCaption()
		} else if uMsg.GetVideoMessage() != nil {
			text = uMsg.GetVideoMessage().GetCaption()
		}

		text = strings.TrimSpace(text)
		senderLID := getLID(client, v.Info.Sender)
		senderID := senderLID // for consistency"""

content = content.replace(old_extraction, "")

# We also need to remove: 
# 		if v.Info.Timestamp.Before(startupTime) {
#			return
#		}
# Since we moved it to the top.
old_timestamp_check = """		if v.Info.Timestamp.Before(startupTime) {
			return
		}"""
content = content.replace(old_timestamp_check, "", 1)

# Remove old ctx creation
old_ctx = """		ctx := &commands.BotContext{
			Client: client,
			Event:  v,
			ChatID: v.Info.Chat,
			Sender: v.Info.Sender,
			Text:   text,
		}"""
content = content.replace(old_ctx, "")


# Fix swear words array
swear_search = 'swearWords := []string{"قحبة", "قحبه", "قحبتي", "قحباني", "خنزير", "خنزيري", "خنزيرتي", "خنزيرة", "خنزيره", "شرموط", "شرموطة", "شرموطه", "زاني", "زانية", "زانيه", "كس", "كسمك", "زب", "زبي", "معرص", "عرص", "منيوك", "منيوكة", "منيوكه"}'
swear_replace = 'swearWords := []string{"قحبة", "قحبه", "قحبتي", "قحباني", "خنزير", "خنزيري", "خنزيرتي", "خنزيرة", "خنزيره", "شرموط", "شرموطة", "شرموطه", "زاني", "زانية", "زانيه", "كس", "كسمك", "زب", "زبي", "معرص", "عرص", "منيوك", "منيوكة", "منيوكه", "بضان", "بضاني", "يبضاني", "بضانك", "زرق", "زرقها", "زرقيها", "طيزك", "طيزكي", "كسك", "كسكي", "شرموطتي", "شرموطي", "تعرص", "يلعن", "منيوكتي", "منيوكي", "يا ابن", "يبن", "قحبوني"}'
content = content.replace(swear_search, swear_replace)

# Save
with open("/home/lennox/Desktop/اهها/Go_Bot/main.go", "w") as f:
    f.write(content)
