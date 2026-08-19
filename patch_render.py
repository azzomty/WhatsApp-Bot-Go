import re

with open("main.go", "r") as f:
    content = f.read()

target = """func startRenderServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Go Bot is Alive! 🚀")
	})
	fmt.Printf("[Render Mode] Server listening on port %s\\n", port)
	http.ListenAndServe(":"+port, nil)
}"""

new_code = """
var (
	container *sqlstore.Container
)

func startRenderServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Go Bot is Alive! 🚀\\n\\nTo pair a new number, go to /pair?phone=YOURNUMBER")
	})
	http.HandleFunc("/pair", func(w http.ResponseWriter, r *http.Request) {
		phone := r.URL.Query().Get("phone")
		if phone == "" {
			fmt.Fprintf(w, "Error: Missing phone parameter. Usage: /pair?phone=966...")
			return
		}
		
		deviceStore := container.NewDevice()
		clientLog := waLog.Stdout("Client", "INFO", true)
		newClient := whatsmeow.NewClient(deviceStore, clientLog)
		
		err := newClient.Connect()
		if err != nil {
			fmt.Fprintf(w, "Connect error: %v", err)
			return
		}

		code, err := newClient.PairPhone(context.Background(), phone, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
		if err != nil {
			fmt.Fprintf(w, "Error pairing: %v", err)
			return
		}
		fmt.Fprintf(w, "Pairing code for %s: %s\\n\\nPlease enter this code on your phone.\\nAfter connecting, the bot will automatically start for this number on the server!", phone, code)
		
		go func() {
			for newClient.Store.ID == nil {
				// wait
			}
			startClient(deviceStore)
		}()
	})
	fmt.Printf("[Render Mode] Server listening on port %s\\n", port)
	http.ListenAndServe(":"+port, nil)
}"""

content = content.replace(target, new_code)

# We need to remove the local `container` var from main() since we made it global.
target2 = """	container, err := sqlstore.New(context.Background(), "sqlite3", "file:whatsapp_v3.db?_foreign_keys=on", dbLog)"""
new_code2 = """	var err error
	container, err = sqlstore.New(context.Background(), "sqlite3", "file:whatsapp_v3.db?_foreign_keys=on", dbLog)"""
content = content.replace(target2, new_code2)

with open("main.go", "w") as f:
    f.write(content)
