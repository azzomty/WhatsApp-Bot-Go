import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    content = f.read()

old_search = """		go func() {
			cmd := exec.Command("./yt-dlp", "ytsearch1:" + query, "--no-warnings", "--print", "%(id)s|%(title)s|%(uploader)s|%(view_count)s|%(like_count)s|%(duration_string)s|%(upload_date)s|%(thumbnail)s")
			out, err := cmd.CombinedOutput()
			if err != nil {
				sendMessage(ctx, "حدث خطأ أثناء البحث:\\n" + string(out))
				return
			}"""

new_search = """		go func() {
			searchUrl := "https://www.youtube.com/results?search_query=" + url.QueryEscape(query)
			reqS, _ := http.NewRequest("GET", searchUrl, nil)
			reqS.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
			respS, errS := http.DefaultClient.Do(reqS)
			if errS != nil {
				sendMessage(ctx, "حدث خطأ أثناء البحث في يوتيوب.")
				return
			}
			defer respS.Body.Close()
			bodyS, _ := io.ReadAll(respS.Body)
			
			re := regexp.MustCompile(`"videoId":"([^"]+)"`)
			matches := re.FindStringSubmatch(string(bodyS))
			if len(matches) < 2 {
				sendMessage(ctx, "لم يتم العثور على نتائج.")
				return
			}
			videoID := matches[1]
			videoURL := "https://www.youtube.com/watch?v=" + videoID
			
			cmd := exec.Command("./yt-dlp", videoURL, "--no-warnings", "--print", "%(id)s|%(title)s|%(uploader)s|%(view_count)s|%(like_count)s|%(duration_string)s|%(upload_date)s|%(thumbnail)s")
			out, err := cmd.CombinedOutput()
			if err != nil {
				sendMessage(ctx, "حدث خطأ أثناء جلب بيانات الأغنية:\\n" + string(out)[:min(len(string(out)), 300)])
				return
			}"""

if old_search in content:
    content = content.replace(old_search, new_search)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(content)
print("Patched youtube search")
