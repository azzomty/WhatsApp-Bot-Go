import re

with open('/home/lennox/Desktop/اهها/Go_Bot/main.go', 'r') as f:
    code = f.read()

pattern = r'(devices, err := container\.GetAllDevices\(context\.Background\(\)\)\n\s+if err != nil {\n\s+panic\(err\)\n\s+}\n)(.*?)(for _, deviceStore := range devices {)'
replacement = r'''\1
	if len(devices) == 0 {
		deviceStore := container.NewDevice()
		devices = append(devices, deviceStore)
	}

	\3'''

new_code = re.sub(pattern, replacement, code, flags=re.DOTALL)

# Add QR logic to startClient
pattern_start = r'(err := client\.Connect\(\)\n\s+if err != nil {\n\s+fmt\.Println\("Error connecting:", err\)\n\s+return\n\s+})'
replacement_start = r'''if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err := client.Connect()
		if err != nil {
			fmt.Println("Error connecting:", err)
			return
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				fmt.Println("Please scan this QR code:")
				// We can't import qrterminal easily without modifying go.mod, so we just print the code.
				// Wait, qrterminal was probably already imported! Let's just use it if it's there.
			} else {
				fmt.Println("QR channel event:", evt.Event)
			}
		}
	} else {
		err := client.Connect()
		if err != nil {
			fmt.Println("Error connecting:", err)
			return
		}
	}'''

# Actually, let's check if qrterminal is in imports.
