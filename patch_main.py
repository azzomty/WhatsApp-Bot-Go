import re
with open("main.go", "r") as f:
    c = f.read()

target = """	if client.Store.ID != nil {
		fmt.Println("Already logged in.")
		// Wait for socket to be completely connected before joining groups
		go func() {
			time.Sleep(5 * time.Second)"""
new_target = """	if client.Store.ID != nil {
		fmt.Println("Already logged in.")
		// Start background group clearer
		commands.StartGroupClearer(client)
		
		// Wait for socket to be completely connected before joining groups
		go func() {
			time.Sleep(5 * time.Second)"""
c = c.replace(target, new_target)

with open("main.go", "w") as f:
    f.write(c)

