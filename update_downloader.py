import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    content = f.read()

# 1. Remove emojis from HandleDownloadCommand messages
content = content.replace('هل تريد تحميله كـ صوت 🎵 أم كـ فيديو 🎬؟\\n\\nأرسل:\\n.صوت\\nأو\\n.فيديو', 'هل تريد تحميله كصوت أم كفيديو؟\\nأرسل:\\n.صوت\\nأو\\n.فيديو')
content = content.replace('⏳ جاري تحميل الصوت، يرجى الانتظار...', 'جاري تحميل الصوت، يرجى الانتظار...')
content = content.replace('⏳ جاري تحميل الفيديو، يرجى الانتظار...', 'جاري تحميل الفيديو، يرجى الانتظار...')

# 2. Remove emojis from processDownload messages
content = content.replace('❌ الرابط غير مدعوم أو غير صحيح.', 'الرابط غير مدعوم أو غير صحيح.')
content = content.replace('❌ تم حظر تحميل هذا الرابط مؤقتاً من قبل الموقع (حماية ضد البوتات). جرب رابط آخر أو موقع آخر.', 'تم حظر تحميل هذا الرابط مؤقتاً من قبل الموقع. جرب رابط آخر.')
content = content.replace('❌ حدث خطأ أثناء التحميل:\\n', 'حدث خطأ أثناء التحميل:\\n')
content = content.replace('❌ فشل في قراءة الملف بعد التحميل.', 'فشل في قراءة الملف بعد التحميل.')
content = content.replace('❌ فشل إرسال الملف (قد يكون حجمه كبيراً جداً للواتساب). الخطأ: %v', 'فشل إرسال الملف. الخطأ: %v')


# 3. Add .اغنية command logic in HandleDownloadCommand
song_logic = """	if strings.HasPrefix(text, ".اغنية") || strings.HasPrefix(text, ".أغنية") {
		parts := strings.SplitN(text, " ", 2)
		if len(parts) < 2 {
			sendMessage(ctx, "يرجى كتابة اسم الأغنية للبحث عنها.\\nمثال: .اغنية hello adele")
			return true
		}
		query := parts[1]
		sendMessage(ctx, "جاري البحث...")
		
		go func() {
			cmd := exec.Command("./yt-dlp", "ytsearch1:" + query, "--no-warnings", "--print", "%(id)s|%(title)s|%(uploader)s|%(view_count)s|%(like_count)s|%(duration_string)s|%(upload_date)s|%(thumbnail)s")
			out, err := cmd.CombinedOutput()
			if err != nil {
				sendMessage(ctx, "حدث خطأ أثناء البحث.")
				return
			}
			
			lines := strings.Split(string(out), "\\n")
			var dataLine string
			for _, line := range lines {
				if strings.Contains(line, "|") && !strings.Contains(line, "Deprecated") && !strings.Contains(line, "WARNING") {
					dataLine = line
					break
				}
			}
			
			if dataLine == "" {
				sendMessage(ctx, "لم يتم العثور على نتائج.")
				return
			}
			
			d := strings.Split(dataLine, "|")
			if len(d) < 8 {
				sendMessage(ctx, "خطأ في قراءة بيانات الأغنية.")
				return
			}
			
			id, title, uploader, views, likes, duration, date, thumb := d[0], d[1], d[2], d[3], d[4], d[5], d[6], d[7]
			
			caption := fmt.Sprintf("*%s*\\n\\nالقناة: %s\\nالمشاهدات: %s\\nالإعجابات: %s\\nالمدة: %s\\nتاريخ الرفع: %s", title, uploader, views, likes, duration, date)
			
			// Send thumbnail
			errThumb := sendImageFromURL(ctx, thumb, caption)
			if errThumb != nil {
				sendMessage(ctx, caption) // Fallback if image fails
			}
			
			// Now download audio
			url := "https://www.youtube.com/watch?v=" + id
			processDownload(ctx, url, "audio")
		}()
		
		return true
	}
"""

if "strings.HasPrefix(text, \".اغنية\")" not in content:
    content = content.replace('if strings.HasPrefix(text, ".تحميل") {', song_logic + '\n\tif strings.HasPrefix(text, ".تحميل") {')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(content)
print("Updated downloader.go")
