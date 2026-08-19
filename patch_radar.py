import re

with open("main.go", "r") as f:
    content = f.read()

target = """	case *events.Message:


		isViewOnce := false"""

new_code = """	case *events.Message:
		fmt.Printf("[RADAR] Received message from %s in chat %s\\n", v.Info.Sender.String(), v.Info.Chat.String())

		isViewOnce := false"""

content = content.replace(target, new_code)

target2 = """		if v.Info.Timestamp.Before(startupTime) {
			return
		}"""

new_code2 = """		if v.Info.Timestamp.Before(startupTime) {
			fmt.Printf("[RADAR] Ignored old message from %s (Message Time: %v, Startup Time: %v)\\n", v.Info.Sender.String(), v.Info.Timestamp, startupTime)
			return
		}"""

content = content.replace(target2, new_code2)

target3 = """		if !store.IsBotEnabled() {
			return
		}"""

new_code3 = """		if !store.IsBotEnabled() {
			fmt.Println("[RADAR] Bot is disabled, ignoring.")
			return
		}"""

content = content.replace(target3, new_code3)

with open("main.go", "w") as f:
    f.write(content)
