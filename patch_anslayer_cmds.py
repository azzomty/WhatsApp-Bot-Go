import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

cmds_code = """
		if strings.HasPrefix(afterSlayer, "تسجيل") {
			parts := strings.Fields(afterSlayer)
			if len(parts) < 3 {
				sendMessage(ctx, "يرجى كتابة الإيميل والباسورد.\nمثال: .انمي سلاير تسجيل email@gmail.com 123456")
				return
			}
			email := parts[1]
			password := parts[2]
			
			sendMessage(ctx, "جاري محاولة تسجيل الدخول للحساب: " + email)
			
			payload := url.Values{}
			payload.Set("username", email)
			payload.Set("password", password)
			payload.Set("device_id", "48c6467c5039bf74")
			
			reqL, _ := http.NewRequest("POST", "https://anslayer.com/anime/public/oauth/login", strings.NewReader(payload.Encode()))
			reqHeaders(reqL, "")
			reqL.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			
			respL, errL := http.DefaultClient.Do(reqL)
			if errL != nil {
				sendMessage(ctx, "حدث خطأ أثناء الاتصال بسيرفر أنمي سلاير.")
				return
			}
			defer respL.Body.Close()
			
			if respL.StatusCode != 200 {
				body, _ := io.ReadAll(respL.Body)
				if strings.Contains(string(body), "غير صحيحة") {
					sendMessage(ctx, "الإيميل أو كلمة المرور غير صحيحة!")
				} else {
					sendMessage(ctx, "فشل تسجيل الدخول. الكود: " + strconv.Itoa(respL.StatusCode) + "\n" + string(body)[:min(len(body), 500)])
				}
				return
			}
			
			// Success! Parse the token.
			var result map[string]interface{}
			json.NewDecoder(respL.Body).Decode(&result)
			
			token := ""
			userID := ""
			
			// Recursive search for token and user_id
			var search func(m map[string]interface{})
			search = func(m map[string]interface{}) {
				for k, v := range m {
					if k == "access_token" || k == "token" {
						if s, ok := v.(string); ok {
							token = s
						}
					}
					if k == "user_id" || k == "id" {
						if s, ok := v.(string); ok {
							userID = s
						} else if f, ok := v.(float64); ok {
							userID = strconv.FormatFloat(f, 'f', -1, 64)
						}
					}
					if child, ok := v.(map[string]interface{}); ok {
						search(child)
					}
				}
			}
			search(result)
			
			if token == "" {
				body, _ := json.Marshal(result)
				sendMessage(ctx, "تم تسجيل الدخول لكن لم أتمكن من استخراج التوكن!\n" + string(body)[:min(len(body), 500)])
				return
			}
			
			ansAccountsMutex.Lock()
			ansAccounts = append(ansAccounts, AnslayerAccount{
				Email: email,
				Token: token,
				UserID: userID,
			})
			ansAccountsMutex.Unlock()
			saveAnslayerAccounts()
			
			sendMessage(ctx, "✅ تم تسجيل الدخول وحفظ الحساب بنجاح!\nالإيميل: " + email + "\nID: " + userID)
			return
		}
		
		if strings.HasPrefix(afterSlayer, "نسيان") {
			parts := strings.Fields(afterSlayer)
			if len(parts) < 2 {
				sendMessage(ctx, "يرجى كتابة الإيميل.\nمثال: .انمي سلاير نسيان email@gmail.com")
				return
			}
			email := parts[1]
			
			payload := url.Values{}
			payload.Set("email", email)
			
			reqL, _ := http.NewRequest("POST", "https://anslayer.com/anime/public/oauth/forgot-password", strings.NewReader(payload.Encode()))
			reqHeaders(reqL, "")
			reqL.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			
			respL, errL := http.DefaultClient.Do(reqL)
			if errL != nil {
				sendMessage(ctx, "حدث خطأ.")
				return
			}
			defer respL.Body.Close()
			if respL.StatusCode == 200 {
				sendMessage(ctx, "✅ تم إرسال رابط استعادة كلمة المرور إلى إيميلك: " + email)
			} else {
				sendMessage(ctx, "❌ فشل. ربما الإيميل غير مسجل.")
			}
			return
		}

		if strings.HasPrefix(afterSlayer, "حساباتي") {
			ansAccountsMutex.RLock()
			defer ansAccountsMutex.RUnlock()
			if len(ansAccounts) == 0 {
				sendMessage(ctx, "لا يوجد أي حسابات مسجلة.")
				return
			}
			var msg strings.Builder
			msg.WriteString("📋 *حسابات أنمي سلاير المسجلة:*\n\n")
			for i, acc := range ansAccounts {
				msg.WriteString(fmt.Sprintf("%d. %s (ID: %s)\n", i+1, acc.Email, acc.UserID))
			}
			msg.WriteString("\nلحذف حساب: .انمي سلاير حذف <رقم>")
			sendMessage(ctx, msg.String())
			return
		}

		if strings.HasPrefix(afterSlayer, "حذف") {
			parts := strings.Fields(afterSlayer)
			if len(parts) < 2 {
				sendMessage(ctx, "يرجى تحديد رقم الحساب للحذف.")
				return
			}
			idx, _ := strconv.Atoi(parts[1])
			idx -= 1
			ansAccountsMutex.Lock()
			if idx >= 0 && idx < len(ansAccounts) {
				email := ansAccounts[idx].Email
				ansAccounts = append(ansAccounts[:idx], ansAccounts[idx+1:]...)
				saveAnslayerAccounts()
				ansAccountsMutex.Unlock()
				sendMessage(ctx, "✅ تم حذف الحساب: " + email)
			} else {
				ansAccountsMutex.Unlock()
				sendMessage(ctx, "رقم الحساب غير صحيح.")
			}
			return
		}
"""

def min_func():
    return """
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
"""

if "func min(" not in content:
    content += min_func()

if "strings.HasPrefix(afterSlayer, \"تسجيل\")" not in content:
    content = content.replace('if strings.HasPrefix(afterSlayer, "مفضلة") {', cmds_code + '\n\t\tif strings.HasPrefix(afterSlayer, "مفضلة") {')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
    f.write(content)
print("Patched cmds")
