import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

# Replace HasPrefix("مفضلة") block
old_fav = """		if strings.HasPrefix(afterSlayer, "مفضلة") || strings.HasPrefix(afterSlayer, "مفضله") || strings.HasPrefix(afterSlayer, "المفضلة") || strings.HasPrefix(afterSlayer, "المفضله") {
			anslayerMutex.Lock()
			msg := anslayerReplyMsg
			anslayerMutex.Unlock()
			if msg == "" {
				sendMessage(ctx, "يرجى أولاً حفظ قالب الرد باستخدام أمر:\\n.انمي سلاير نشر <رسالتك>")
				return
			}
			startFavMarketing(ctx, msg)
			return
		}"""

new_fav = """		if strings.HasPrefix(afterSlayer, "مفضلة") || strings.HasPrefix(afterSlayer, "مفضله") || strings.HasPrefix(afterSlayer, "المفضلة") || strings.HasPrefix(afterSlayer, "المفضله") {
			anslayerMutex.Lock()
			msg := anslayerReplyMsg
			anslayerMutex.Unlock()
			if msg == "" {
				sendMessage(ctx, "يرجى أولاً حفظ قالب الرد باستخدام أمر:\\n.انمي سلاير نشر <رسالتك>")
				return
			}
			
			ansAccountsMutex.RLock()
			count := len(ansAccounts)
			ansAccountsMutex.RUnlock()
			
			if count == 0 {
				sendMessage(ctx, "لا توجد حسابات مسجلة! يرجى إضافة حساب أولاً:\\n.انمي سلاير تسجيل <الايميل> <الباسورد>")
				return
			}
			
			var m strings.Builder
			m.WriteString("اختر الحساب الذي تريد بدء المراقبة به:\\n")
			ansAccountsMutex.RLock()
			for i, acc := range ansAccounts {
				m.WriteString(fmt.Sprintf("%d. حساب (%s)\\n", i+1, acc.Email))
			}
			ansAccountsMutex.RUnlock()
			m.WriteString("0. 🚀 تشغيل كل الحسابات معاً!\\n\\nللاختيار أرسل .رقم")
			
			session := &AnslayerSession{
				Mode: "marketing",
				State: "select_account",
			}
			ansSessions[ctx.Sender.User] = session
			
			sendMessage(ctx, m.String())
			return
		}"""

if old_fav in content:
    content = content.replace(old_fav, new_fav)

# Now Handle sessions for select_account
old_session = """	if session.Mode == "marketing" && session.State == "select_anime" {"""
new_session = """	if session.Mode == "marketing" && session.State == "select_account" {
		num, err := strconv.Atoi(strings.TrimPrefix(ctx.Text, "."))
		if err != nil {
			sendMessage(ctx, "يرجى إرسال رقم صحيح.")
			return
		}
		
		ansAccountsMutex.RLock()
		var selected []AnslayerAccount
		if num == 0 {
			selected = append(selected, ansAccounts...)
		} else if num > 0 && num <= len(ansAccounts) {
			selected = append(selected, ansAccounts[num-1])
		}
		ansAccountsMutex.RUnlock()
		
		if len(selected) == 0 {
			sendMessage(ctx, "رقم الحساب غير صحيح.")
			return
		}
		
		delete(ansSessions, ctx.Sender.User)
		anslayerMutex.Lock()
		msg := anslayerReplyMsg
		anslayerMutex.Unlock()
		
		sendMessage(ctx, fmt.Sprintf("✅ تم بدء المراقبة لـ %d حساب(ات)!", len(selected)))
		startFavMarketing(ctx, msg, selected)
		return
	}
	
	if session.Mode == "marketing" && session.State == "select_anime" {"""

if 'State == "select_account"' not in content:
    content = content.replace(old_session, new_session)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
    f.write(content)
print("Patched fav flow")
