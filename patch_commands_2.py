import re

with open("internal/commands/commands.go", "r") as f:
    content = f.read()

target = """func Handle(ctx *BotContext) {
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


	if ctx.Text == ".bot off" {"""

new_code = """func Handle(ctx *BotContext) {
	if ctx.Text == ".bot off" {"""

content = content.replace(target, new_code)

with open("internal/commands/commands.go", "w") as f:
    f.write(content)
