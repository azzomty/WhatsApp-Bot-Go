import re

with open("main.go", "r") as f:
    content = f.read()

target = """	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err := client.Connect()
		if err != nil {
			panic(err)
		}

		for evt := range qrChan {
			if evt.Event == "code" {
				fmt.Println("===========================================")
				fmt.Println("امسح رمز الـ QR التالي باستخدام تطبيق واتساب:")
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println("===========================================")
			} else {
				fmt.Println("QR Event:", evt.Event)
			}
		}
	} else {
		err := client.Connect()
		if err != nil {
			panic(err)
		}
		fmt.Printf("تم تسجيل الدخول بنجاح للرقم %s! البوت جاهز.\\n", client.Store.ID)
	}"""

new_code = """	err := client.Connect()
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	fmt.Printf("تم تسجيل الدخول بنجاح للرقم %s! البوت جاهز.\\n", client.Store.ID)"""

content = content.replace(target, new_code)

with open("main.go", "w") as f:
    f.write(content)
