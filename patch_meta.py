import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    content = f.read()

old_code = """			args := []string{videoURL, "--no-warnings", "--print", "%(id)s|%(title)s|%(uploader)s|%(view_count)s|%(like_count)s|%(duration_string)s|%(upload_date)s|%(thumbnail)s", "--extractor-args", "youtube:player_client=android,ios"}
			if _, err := os.Stat("cookies.txt"); err == nil {
				args = append([]string{"--cookies", "cookies.txt"}, args...)
			}
			cmd := exec.Command("./yt-dlp", args...)
			out, err := cmd.CombinedOutput()
			if err != nil {
				sendMessage(ctx, "حدث خطأ أثناء جلب بيانات الأغنية:\\n" + string(out)[:min(len(string(out)), 300)])
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
			
			caption := fmt.Sprintf("*%s*\\n\\nالقناة: %s\\nالمشاهدات: %s\\nالإعجابات: %s\\nالمدة: %s\\nتاريخ الرفع: %s", title, uploader, views, likes, duration, date)"""

new_code = """			reqM, _ := http.NewRequest("GET", videoURL, nil)
			reqM.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
			respM, errM := http.DefaultClient.Do(reqM)
			if errM != nil {
				sendMessage(ctx, "حدث خطأ أثناء جلب بيانات الصفحة.")
				return
			}
			defer respM.Body.Close()
			bodyM, _ := io.ReadAll(respM.Body)
			bodyStr := string(bodyM)
			
			titleRe := regexp.MustCompile(`"title":"(.*?)"`)
			authorRe := regexp.MustCompile(`"author":"(.*?)"`)
			viewsRe := regexp.MustCompile(`"viewCount":"(.*?)"`)
			
			title := "غير معروف"
			if m := titleRe.FindStringSubmatch(bodyStr); len(m) > 1 { title = m[1] }
			uploader := "غير معروف"
			if m := authorRe.FindStringSubmatch(bodyStr); len(m) > 1 { uploader = m[1] }
			views := "غير معروف"
			if m := viewsRe.FindStringSubmatch(bodyStr); len(m) > 1 { views = m[1] }
			
			thumb := "https://i.ytimg.com/vi/" + videoID + "/hqdefault.jpg"
			
			caption := fmt.Sprintf("*%s*\\n\\nالقناة: %s\\nالمشاهدات: %s", title, uploader, views)
			id := videoID"""

if "args := []string{videoURL, \"--no-warnings\"" in content:
    content = content.replace(old_code, new_code)
    with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
        f.write(content)
    print("Patched meta successfully")
else:
    print("Could not find old_code")
