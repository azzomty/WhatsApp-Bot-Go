import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    content = f.read()

# I will find the block that handles .اغنية
# It starts at: searchUrl := "https://www.youtube.com/results?search_query=" + url.QueryEscape(query)
# And ends before processDownload(ctx, videoURL, "audio") or processDownload(ctx, url, "audio")

start_str = '			searchUrl := "https://www.youtube.com/results?search_query="'
end_str = '			processDownload(ctx, videoURL, "audio")'

if start_str not in content:
    print("Could not find start_str")

start_idx = content.find(start_str)
end_idx = content.find(end_str) + len(end_str)

new_code = """			cmd := exec.Command("./yt-dlp", "scsearch1:" + query, "--no-warnings", "--print", "%(webpage_url)s|%(title)s|%(uploader)s|%(view_count)s|%(like_count)s|%(duration_string)s|%(upload_date)s|%(thumbnail)s")
			out, err := cmd.CombinedOutput()
			if err != nil {
				sendMessage(ctx, "حدث خطأ أثناء البحث في ساوند كلاود.\\nقد لا توجد نتائج.")
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
			
			scURL, title, uploader, views, likes, duration, date, thumb := d[0], d[1], d[2], d[3], d[4], d[5], d[6], d[7]
			
			caption := fmt.Sprintf("*%s*\\n\\nالفنان: %s\\nاستماعات: %s\\nإعجابات: %s\\nالمدة: %s\\nتاريخ الرفع: %s", title, uploader, views, likes, duration, date)
			
			errThumb := sendImageFromURL(ctx, thumb, caption)
			if errThumb != nil {
				sendMessage(ctx, caption)
			}
			
			processDownload(ctx, scURL, "audio")"""

content = content[:start_idx] + new_code + content[end_idx:]

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(content)
print("Patched for Soundcloud")
