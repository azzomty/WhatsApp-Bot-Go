with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    c = f.read()

# Fix .اغنية search
old_search = '''			searchUrl := "https://www.youtube.com/results?search_query=" + url.QueryEscape(query)
			reqS, _ := http.NewRequest("GET", searchUrl, nil)
			reqS.Header.Set("User-Agent", "Mozilla/5.0")
			respS, errS := http.DefaultClient.Do(reqS)
			if errS != nil {
				sendMessage(ctx, "حدث خطأ أثناء البحث في يوتيوب.")
				return
			}
			defer respS.Body.Close()
			bodyS, _ := io.ReadAll(respS.Body)

			reSearch := regexp.MustCompile(`"videoId":"([^"]+)"`)
			matches := reSearch.FindStringSubmatch(string(bodyS))
			if len(matches) < 2 {
				sendMessage(ctx, "لم يتم العثور على نتائج.")
				return
			}
			videoID := matches[1]'''

new_search = '''			cmdSearch := exec.Command("./yt-dlp", "ytsearch1:"+query, "--get-id", "--extractor-args", "youtube:player_client=android,web")
			out, errS := cmdSearch.CombinedOutput()
			if errS != nil || len(out) < 5 {
				sendMessage(ctx, "لم يتم العثور على نتائج.")
				return
			}
			videoID := strings.TrimSpace(string(out))'''

c = c.replace(old_search, new_search)

# Fix yt-dlp arguments to bypass youtube blocks
old_video = '''args := []string{"--ffmpeg-location", "./ffmpeg", "-N", "4", "--no-check-certificate", "-f", "bestvideo[height<=720][ext=mp4]+bestaudio[ext=m4a]/best[height<=720][ext=mp4]/best", "--merge-output-format", "mp4", url, "-o", finalFile}'''
new_video = '''args := []string{"--ffmpeg-location", "./ffmpeg", "-N", "4", "--no-check-certificate", "--extractor-args", "youtube:player_client=android,web", "-f", "bestvideo[height<=720][ext=mp4]+bestaudio[ext=m4a]/best[height<=720][ext=mp4]/best", "--merge-output-format", "mp4", url, "-o", finalFile}'''

c = c.replace(old_video, new_video)

old_audio = '''args := []string{"--ffmpeg-location", "./ffmpeg", "-N", "4", "--no-check-certificate", "-f", "bestaudio/best", "-x", "--audio-format", "mp3", url, "-o", finalFile}'''
new_audio = '''args := []string{"--ffmpeg-location", "./ffmpeg", "-N", "4", "--no-check-certificate", "--extractor-args", "youtube:player_client=android,web", "-f", "bestaudio/best", "-x", "--audio-format", "mp3", url, "-o", finalFile}'''

c = c.replace(old_audio, new_audio)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(c)
