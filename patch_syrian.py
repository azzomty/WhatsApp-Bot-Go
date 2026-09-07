import re

with open("internal/commands/commands.go", "r") as f:
    c = f.read()

target = """var AutoJoinGroups bool

func HandleSyrian(ctx *BotContext) {
	if ctx.Text == ".دخلني قروبات" {
		AutoJoinGroups = !AutoJoinGroups
		if AutoJoinGroups {
			sendMessage(ctx, "تم تفعيل الانضمام التلقائي (الرقم السوري سيقوم بالانضمام لأي قروب يرسل رابطه).")
		} else {
			sendMessage(ctx, "تم إيقاف الانضمام التلقائي.")
		}
		return
	}

	if AutoJoinGroups && strings.Contains(ctx.Text, "chat.whatsapp.com/") {
		parts := strings.Split(ctx.Text, "chat.whatsapp.com/")
		if len(parts) > 1 {
			code := strings.FieldsFunc(parts[1], func(r rune) bool {
				return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_')
			})
			if len(code) > 0 {
				go func() {
					ctx.Client.JoinGroupWithLink(context.Background(), code[0])
				}()
			}
		}
	}
	HandleExchangeMessage(ctx)
}"""

new_target = """var (
	AutoJoinGroups bool
	SentGroupLinks = make(map[string]bool)
)

func HandleSyrian(ctx *BotContext) {
	if ctx.Text == ".دخلني قروبات" {
		AutoJoinGroups = !AutoJoinGroups
		if AutoJoinGroups {
			sendMessage(ctx, "تم تفعيل نظام استخراج القروبات (سيقوم البوت بإرسال أي رابط قروب غير منضم له إلى الخاص).")
		} else {
			sendMessage(ctx, "تم إيقاف نظام استخراج القروبات.")
		}
		return
	}

	if AutoJoinGroups && strings.Contains(ctx.Text, "chat.whatsapp.com/") {
		parts := strings.Split(ctx.Text, "chat.whatsapp.com/")
		if len(parts) > 1 {
			code := strings.FieldsFunc(parts[1], func(r rune) bool {
				return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_')
			})
			if len(code) > 0 {
				groupCode := code[0]
				go func() {
					if SentGroupLinks[groupCode] {
						return // Already sent this link
					}
					
					// Get group info to check if we are in it
					info, err := ctx.Client.GetGroupInfoFromLink(context.Background(), groupCode)
					if err != nil {
						return
					}
					
					// Check if we are already in the group
					joinedGroups, err := ctx.Client.GetJoinedGroups()
					if err == nil {
						for _, g := range joinedGroups {
							if g.JID == info.JID {
								return // We are already in this group
							}
						}
					}
					
					SentGroupLinks[groupCode] = true
					
					ownJID := ctx.Client.Store.ID.ToNonAD()
					msg := fmt.Sprintf("قروب جديد مو داخله:\nhttps://chat.whatsapp.com/%s\nالاسم: %s", groupCode, info.Name)
					ctx.Client.SendMessage(context.Background(), ownJID, &waProto.Message{
						ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String(msg)},
					})
				}()
			}
		}
	}
	HandleExchangeMessage(ctx)
}"""

c = c.replace(target, new_target)
with open("internal/commands/commands.go", "w") as f:
    f.write(c)

