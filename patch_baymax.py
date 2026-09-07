import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

# 1. Add .baymax sticker handler to baymax(ctx) or to main message switch
# It's better to modify baymax(ctx) to handle both .baymax and .baymax sticker!
# Wait, baymax(ctx) is called when text == ".baymax"
# Let's search for `case ".baymax", ".buymax":`
# We can change it to check if text starts with .baymax

baymax_func_old = """func baymax(ctx *BotContext) {
	out := store.GetCustomOutput(getLID(ctx, ctx.Sender), ".baymax", "")
	if out == "" {
		out = store.GetCustomOutput(getLID(ctx, ctx.Sender), ".buymax", "")
	}
	if out != "" {
		sendMessage(ctx, out)
		return
	}

	name := store.GetBaymaxName(getLID(ctx, ctx.Sender))
	if name != "" {
		sendMessage(ctx, name)
	} else if ctx.Sender.User == "966508364121" || ctx.Event.Info.IsFromMe {
		sendMessage(ctx, "هاي هبهب 🤍")
	} else {
		sendMessage(ctx, "هاي عزام سينباي 🤍")
	}
}"""

# Wait, there are emojis in baymax_func_old? I should not hardcode emojis if they want no emojis. But wait, they only asked to remove emojis from .اغنية and .تحميل.
# But I can just do a regex replace for func baymax

baymax_pattern = re.compile(r'func baymax\(ctx \*BotContext\) \{.*?\n\}', re.DOTALL)

baymax_func_new = """func baymax(ctx *BotContext) {
	text := strings.TrimSpace(ctx.Text)
	if text == ".baymax sticker" || text == ".buymax sticker" {
		uMsg := UnwrapMessage(ctx.Event.Message)
		if uMsg != nil {
			quotedMsg := ctx.Event.Message.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage()
			if quotedMsg != nil && quotedMsg.GetStickerMessage() != nil {
				data, err := ctx.Client.Download(context.Background(), quotedMsg.GetStickerMessage())
				if err == nil {
					ioutil.WriteFile("baymax_sticker.webp", data, 0644)
					sendMessage(ctx, "تم حفظ ملصق بايماكس بنجاح!")
					return
				}
			}
		}
		sendMessage(ctx, "يرجى الرد على ملصق لكتابة الأمر: .baymax sticker")
		return
	}

	out := store.GetCustomOutput(getLID(ctx, ctx.Sender), ".baymax", "")
	if out == "" {
		out = store.GetCustomOutput(getLID(ctx, ctx.Sender), ".buymax", "")
	}
	if out != "" {
		sendMessage(ctx, out)
	} else {
		name := store.GetBaymaxName(getLID(ctx, ctx.Sender))
		if name != "" {
			sendMessage(ctx, name)
		} else if ctx.Sender.User == "966508364121" || ctx.Event.Info.IsFromMe {
			sendMessage(ctx, "هاي هبهب")
		} else {
			sendMessage(ctx, "هاي عزام سينباي")
		}
	}
	
	// Send sticker if exists
	webpData, err := ioutil.ReadFile("baymax_sticker.webp")
	if err == nil && len(webpData) > 0 {
		resp, err := ctx.Client.Upload(context.Background(), webpData, whatsmeow.MediaImage)
		if err == nil {
			stickerMsg := &waProto.StickerMessage{
				URL:           proto.String(resp.URL),
				DirectPath:    proto.String(resp.DirectPath),
				MediaKey:      resp.MediaKey,
				Mimetype:      proto.String("image/webp"),
				FileEncSHA256: resp.FileEncSHA256,
				FileSHA256:    resp.FileSHA256,
				FileLength:    proto.Uint64(uint64(len(webpData))),
				IsAnimated:    proto.Bool(false),
			}
			ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
				StickerMessage: stickerMsg,
			})
		}
	}
}"""

content = baymax_pattern.sub(baymax_func_new, content)

# Also we need to make sure `.baymax sticker` falls into this function.
# Let's check `ProcessMessage` switch statement:
# case ".baymax", ".buymax": baymax(ctx)
# Wait, if they type ".baymax sticker", it's a two-word command!
# `twoWordCmd == ".baymax sticker"` is NOT in the switch block. I should just add a prefix check for baymax before the switch!

baymax_hook = """	if strings.HasPrefix(strings.TrimSpace(ctx.Text), ".baymax") || strings.HasPrefix(strings.TrimSpace(ctx.Text), ".buymax") {
		baymax(ctx)
		return
	}"""

# I will put it right after HandleFontsCommand
insert_point2 = '	if HandleFontsCommand(ctx) {\n\t\treturn\n\t}'
replacement2 = """	if HandleFontsCommand(ctx) {
		return
	}
	if strings.HasPrefix(strings.TrimSpace(ctx.Text), ".baymax") || strings.HasPrefix(strings.TrimSpace(ctx.Text), ".buymax") {
		baymax(ctx)
		return
	}"""

if 'strings.HasPrefix(strings.TrimSpace(ctx.Text), ".baymax")' not in content:
    content = content.replace(insert_point2, replacement2)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
print("Added baymax sticker logic")
