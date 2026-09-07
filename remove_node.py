with open('/home/lennox/Desktop/اهها/Go_Bot/main.go', 'r') as f:
    content = f.read()

node_code = """	go func() {
		cmd := exec.Command("node", "sticker_server.js")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}()"""

if node_code in content:
    content = content.replace(node_code, "")
    with open('/home/lennox/Desktop/اهها/Go_Bot/main.go', 'w') as f:
        f.write(content)
    print("Removed node server from main!")
