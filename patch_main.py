import re

with open("main.go", "r") as f:
    content = f.read()

# First, clean any old injected code in main.go
target_clean = """		// --- NUMBER SPLIT & ALWAYS-ON LOGIC ---
		myNumber := client.Store.ID.User

		// AutoJoin logic (ONLY for Syrian number, ignores .bot off and .baymax)
		if strings.HasPrefix(myNumber, "963") && commands.AutoJoinGroups && strings.Contains(text, "chat.whatsapp.com/") {
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

		// If it's the Saudi number, we disable AutoJoin entirely (since we already skipped it above).
		// We just let it continue to the rest of the bot logic.

		// ----------------------------------------"""
content = content.replace(target_clean, "")

target_inject = """		ctx := &commands.BotContext{
			Client: client,
			Event:  v,
			ChatID: v.Info.Chat,
			Sender: v.Info.Sender,
			Text:   text,
		}"""

new_inject = """		ctx := &commands.BotContext{
			Client: client,
			Event:  v,
			ChatID: v.Info.Chat,
			Sender: v.Info.Sender,
			Text:   text,
		}

		myNumber := client.Store.ID.User
		if strings.HasPrefix(myNumber, "212") {
			commands.HandleMoroccan(ctx)
			return
		}
		if strings.HasPrefix(myNumber, "963") {
			commands.HandleSyrian(ctx)
			return
		}

		// Saudi logic continues down below
"""

content = content.replace(target_inject, new_inject)

with open("main.go", "w") as f:
    f.write(content)
