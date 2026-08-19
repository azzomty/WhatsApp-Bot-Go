import re

with open("main.go", "r") as f:
    content = f.read()

target = """		if v.Info.IsGroup && !store.IsGroupActivated(v.Info.Chat.String()) {
			// Allow ONLY .baymax to activate
			if v.Message != nil && v.Message.Conversation != nil {
				text := strings.TrimSpace(strings.ToLower(*v.Message.Conversation))
				if text == ".baymax" || text == ".buymax" {
					store.ActivateGroup(v.Info.Chat.String())
					store.SaveActivatedGroups(".")
					// Let it fall through to HandleMessage so baymax personalized message triggers
				}
			} else if v.Message != nil && v.Message.ExtendedTextMessage != nil && v.Message.ExtendedTextMessage.Text != nil {
				text := strings.TrimSpace(strings.ToLower(*v.Message.ExtendedTextMessage.Text))
				if text == ".baymax" || text == ".buymax" {
					store.ActivateGroup(v.Info.Chat.String())
					store.SaveActivatedGroups(".")
					// Let it fall through to HandleMessage so baymax personalized message triggers
				}
			}
			return
		}"""

new_code = """		if v.Info.IsGroup && !store.IsGroupActivated(v.Info.Chat.String()) {
			isActivating := false
			if v.Message != nil && v.Message.Conversation != nil {
				text := strings.TrimSpace(strings.ToLower(*v.Message.Conversation))
				if text == ".baymax" || text == ".buymax" {
					store.ActivateGroup(v.Info.Chat.String())
					store.SaveActivatedGroups(".")
					isActivating = true
				}
			} else if v.Message != nil && v.Message.ExtendedTextMessage != nil && v.Message.ExtendedTextMessage.Text != nil {
				text := strings.TrimSpace(strings.ToLower(*v.Message.ExtendedTextMessage.Text))
				if text == ".baymax" || text == ".buymax" {
					store.ActivateGroup(v.Info.Chat.String())
					store.SaveActivatedGroups(".")
					isActivating = true
				}
			}
			if !isActivating {
				return
			}
		}"""

content = content.replace(target, new_code)

with open("main.go", "w") as f:
    f.write(content)
