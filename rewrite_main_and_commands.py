import re

# 1. CLEAN UP MAIN.GO
with open("main.go", "r") as f:
    main_content = f.read()

# We want to remove the logic block from main.go
start_marker = "// --- NUMBER SPLIT & ALWAYS-ON LOGIC ---"
end_marker = "// ----------------------------------------"

if start_marker in main_content and end_marker in main_content:
    before = main_content.split(start_marker)[0]
    after = main_content.split(end_marker)[1]
    main_content = before + after

with open("main.go", "w") as f:
    f.write(main_content)


# 2. UPDATE COMMANDS.GO
with open("internal/commands/commands.go", "r") as f:
    cmd_content = f.read()

# Find the start of Handle
handle_start = "func Handle(ctx *BotContext) {"
handle_body = """func Handle(ctx *BotContext) {
	myNumber := ctx.Client.Store.ID.User

	// ==========================================
	// MOROCCAN NUMBER (212)
	// ==========================================
	if strings.HasPrefix(myNumber, "212") {
		// ONLY runs "وش لقبك" / "وش لقبي"
		if ctx.Event.Message != nil && ctx.Event.Message.ExtendedTextMessage != nil && ctx.Event.Message.ExtendedTextMessage.ContextInfo != nil {
			qMsg := ctx.Event.Message.ExtendedTextMessage.ContextInfo.QuotedMessage
			if qMsg != nil {
				participant := ctx.Event.Message.ExtendedTextMessage.ContextInfo.GetParticipant()
				myJid := ctx.Client.Store.ID.ToNonAD().String()
				
				if participant == myJid {
					qText := ""
					if qMsg.ExtendedTextMessage != nil {
						qText = qMsg.ExtendedTextMessage.GetText()
					} else if qMsg.Conversation != nil {
						qText = qMsg.GetConversation()
					}
					qText = strings.TrimSpace(qText)
					if qText == "وش لقبك" || qText == "وش لقبي" {
						sendMessage(ctx, "New character unlock hibi🔓💫")
					}
				}
			}
		}
		return // Do not process anything else
	}

	// ==========================================
	// SYRIAN NUMBER (963)
	// ==========================================
	if strings.HasPrefix(myNumber, "963") {
		// AutoJoin Logic
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

		// Exchange Logic (.نشر, .مفضلة, .تبادل, الخ)
		HandleExchangeMessage(ctx)
		
		return // Do not process anything else
	}

	// ==========================================
	// SAUDI NUMBER (966) - FULL BOT
	// ==========================================
	// Runs everything EXCEPT AutoJoin

"""

# We need to replace the start of Handle, up to the `.bot off` check.
# Let's just find `func Handle(ctx *BotContext) {` and inject our logic.
# Then we remove the Hibi check that we added earlier since we want it for Saudi too.

# Let's do this safely by rewriting the first few lines of Handle.
target = """func Handle(ctx *BotContext) {
	if ctx.Text == ".bot off" {"""

cmd_content = cmd_content.replace(target, handle_body + """
	if ctx.Text == ".bot off" {""")

# We need to remove the old Hibi check that we injected earlier so we don't have it twice (we'll inject it for Saudi properly)
old_hibi = """	// Check for "وش لقبك" response
	if ctx.Event.Message != nil && ctx.Event.Message.ExtendedTextMessage != nil && ctx.Event.Message.ExtendedTextMessage.ContextInfo != nil {
		qMsg := ctx.Event.Message.ExtendedTextMessage.ContextInfo.QuotedMessage
		if qMsg != nil {
			participant := ctx.Event.Message.ExtendedTextMessage.ContextInfo.GetParticipant()
			myJid := ctx.Client.Store.ID.ToNonAD().String()
			
			// If they replied to me
			if participant == myJid {
				qText := ""
				if qMsg.ExtendedTextMessage != nil {
					qText = qMsg.ExtendedTextMessage.GetText()
				} else if qMsg.Conversation != nil {
					qText = qMsg.GetConversation()
				}
				
				if strings.TrimSpace(qText) == "وش لقبك" {
					sendMessage(ctx, "New character unlock hibi🔓💫")
					return
				}
			}
		}
	}"""
cmd_content = cmd_content.replace(old_hibi, "")

# Now add Hibi to Saudi number right after IsBotEnabled check
target_bot_enabled = """	if !IsBotEnabled {
		return
	}"""
new_bot_enabled = """	if !IsBotEnabled {
		return
	}

	// Hibi Check for Saudi Number
	if ctx.Event.Message != nil && ctx.Event.Message.ExtendedTextMessage != nil && ctx.Event.Message.ExtendedTextMessage.ContextInfo != nil {
		qMsg := ctx.Event.Message.ExtendedTextMessage.ContextInfo.QuotedMessage
		if qMsg != nil {
			participant := ctx.Event.Message.ExtendedTextMessage.ContextInfo.GetParticipant()
			myJid := ctx.Client.Store.ID.ToNonAD().String()
			
			if participant == myJid {
				qText := ""
				if qMsg.ExtendedTextMessage != nil {
					qText = qMsg.ExtendedTextMessage.GetText()
				} else if qMsg.Conversation != nil {
					qText = qMsg.GetConversation()
				}
				qText = strings.TrimSpace(qText)
				if qText == "وش لقبك" || qText == "وش لقبي" {
					sendMessage(ctx, "New character unlock hibi🔓💫")
					return
				}
			}
		}
	}"""
cmd_content = cmd_content.replace(target_bot_enabled, new_bot_enabled)

with open("internal/commands/commands.go", "w") as f:
    f.write(cmd_content)
