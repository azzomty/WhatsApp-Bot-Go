import re

with open("internal/commands/commands.go", "r") as f:
    content = f.read()

# Find the start of the Handle func and inject the Hibi character unlock check
target = """	if !IsBotEnabled {
		return
	}"""
    
new_code = """	if !IsBotEnabled {
		return
	}

	// Check for "وش لقبك" response
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

content = content.replace(target, new_code)

with open("internal/commands/commands.go", "w") as f:
    f.write(content)
