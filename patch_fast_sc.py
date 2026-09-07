import re

# 1. Get the fast SoundCloud code
with open('/home/lennox/Desktop/اهها/Go_Bot/fast_soundcloud.go', 'r') as f:
    fast_content = f.read()

start_idx = fast_content.find('if twoWordCmd == ".اغنية" {')
end_idx = fast_content.find('\n\n\tif strings.HasPrefix(text, ".تحميل")', start_idx)
if end_idx == -1:
    end_idx = fast_content.find('\n\n\tsession, exists :=', start_idx)

fast_sc_logic = fast_content[start_idx:end_idx]

# 2. Rename it to .ساوند || .ساوند كلاود
fast_sc_logic = fast_sc_logic.replace('if twoWordCmd == ".اغنية" {', 'if strings.HasPrefix(text, ".ساوند") || strings.HasPrefix(text, ".ساوند كلاود") {\n\t\tparts := strings.SplitN(text, " ", 2)\n\t\tif len(parts) < 2 {\n\t\t\tsendMessage(ctx, "يرجى كتابة اسم الأغنية للبحث عنها.\\nمثال: .ساوند hello adele")\n\t\t\treturn true\n\t\t}\n\t\tquery := parts[1]')

# Wait, in fast_soundcloud.go, query was parsed using parts := strings.SplitN(text, " ", 3)
# Let's just fix the prefix parsing manually.
fast_sc_logic = re.sub(r'if twoWordCmd == "\.اغنية" \{.*?\n\t\tquery := [^\n]*', 'if strings.HasPrefix(text, ".ساوند") || strings.HasPrefix(text, ".ساوند كلاود") {\n\t\tparts := strings.SplitN(text, " ", 2)\n\t\tif len(parts) < 2 {\n\t\t\tsendMessage(ctx, "يرجى كتابة اسم الأغنية للبحث عنها.\\nمثال: .ساوند hello adele")\n\t\t\treturn true\n\t\t}\n\t\tquery := parts[1]', fast_sc_logic, flags=re.DOTALL)


# 3. Add HLS support to it
old_url_extract = '''				progUrl := ""
				if m := progRe.FindStringSubmatch(bodyStr); len(m) > 1 {
					progUrl = m[1]
				}'''

new_url_extract = '''				progUrl := ""
				isHLS := false
				if m := progRe.FindStringSubmatch(bodyStr); len(m) > 1 {
					progUrl = m[1]
				} else {
					hlsRe := regexp.MustCompile(`"url":"([^"]+)","preset":"[^"]+","duration":[0-9]+,"snipped":false,"format":{"protocol":"hls"`)
					if m := hlsRe.FindStringSubmatch(bodyStr); len(m) > 1 {
						progUrl = m[1]
						isHLS = true
					}
				}'''
fast_sc_logic = fast_sc_logic.replace(old_url_extract, new_url_extract)

old_download_logic = '''					// 4. Download & Send MP3
					respDL, errDL := http.Get(dlUrl)
					if errDL != nil {
						sendMessage(ctx, "فشل تحميل الملف.")
						return
					}
					defer respDL.Body.Close()
					audioData, _ := io.ReadAll(respDL.Body)'''

new_download_logic = '''					// 4. Download & Send MP3
					var audioData []byte
					if isHLS {
						tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("sc_%d.mp3", time.Now().UnixNano()))
						cmd := exec.Command("./ffmpeg", "-y", "-i", dlUrl, "-c:a", "libmp3lame", "-q:a", "2", tmpFile)
						err := cmd.Run()
						if err != nil {
							sendMessage(ctx, "حدث خطأ أثناء تحويل الملف الصوتي.")
							return
						}
						audioData, _ = ioutil.ReadFile(tmpFile)
						os.Remove(tmpFile)
					} else {
						respDL, errDL := http.Get(dlUrl)
						if errDL != nil {
							sendMessage(ctx, "فشل تحميل الملف.")
							return
						}
						defer respDL.Body.Close()
						audioData, _ = io.ReadAll(respDL.Body)
					}'''
fast_sc_logic = fast_sc_logic.replace(old_download_logic, new_download_logic)


# 4. Replace the old slow .ساوند block in downloader.go
with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    current_content = f.read()

cur_start_idx = current_content.find('if strings.HasPrefix(text, ".ساوند") || strings.HasPrefix(text, ".ساوند كلاود") {')
cur_end_idx = current_content.find('\n\n\tif strings.HasPrefix(text, ".اغنية")', cur_start_idx)
if cur_end_idx == -1:
    cur_end_idx = current_content.find('\n\n\t//', cur_start_idx) # fallback if needed
    
new_content = current_content[:cur_start_idx] + fast_sc_logic + current_content[cur_end_idx:]

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(new_content)
print("Patched downloader to restore fast SoundCloud logic for .ساوند")
