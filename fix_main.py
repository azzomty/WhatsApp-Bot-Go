import re

with open("main.go", "r") as f:
    content = f.read()

# Remove radar 1
content = content.replace('fmt.Printf("[RADAR] Received message from %s in chat %s\\n", v.Info.Sender.String(), v.Info.Chat.String())\n\n\t\tisViewOnce := false', 'isViewOnce := false')

# Remove radar 2
content = content.replace('fmt.Printf("[RADAR] Ignored old message from %s (Message Time: %v, Startup Time: %v)\\n", v.Info.Sender.String(), v.Info.Timestamp, startupTime)', '')

# Remove radar 3
content = content.replace('fmt.Println("[RADAR] Bot is disabled, ignoring.")', '')

# Fix /pair API
target_api = """		go func() {
			for newClient.Store.ID == nil {
				// wait
			}
			startClient(deviceStore)
		}()"""

fixed_api = """		go func() {
			for newClient.Store.ID == nil {
				// wait
			}
			newClient.AddEventHandler(func(evt interface{}) { eventHandler(newClient, evt) })
			fmt.Printf("تم تسجيل الدخول بنجاح للرقم %s عبر ريندر! البوت جاهز.\\n", newClient.Store.ID)
		}()"""

content = content.replace(target_api, fixed_api)

with open("main.go", "w") as f:
    f.write(content)

with open("internal/commands/commands.go", "r") as f:
    cmd_content = f.read()

cmd_content = cmd_content.replace('fmt.Printf("[RADAR-CMD] Reached commands.Handle with text: \'%s\'\\n", ctx.Text)', '')

with open("internal/commands/commands.go", "w") as f:
    f.write(cmd_content)

