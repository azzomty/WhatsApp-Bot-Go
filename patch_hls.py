import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    content = f.read()

# 1. Update the URL extraction to support HLS
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
content = content.replace(old_url_extract, new_url_extract)

# 2. Update the download logic to use ffmpeg if it's HLS
old_download_logic = '''				// 5. Download the final MP3
				respDL, errDL := http.Get(downloadURL)
				if errDL != nil {
					sendMessage(ctx, "حدث خطأ أثناء تحميل الملف الصوتي.")
					return
				}
				defer respDL.Body.Close()
				
				audioData, _ := io.ReadAll(respDL.Body)'''

new_download_logic = '''				// 5. Download the final MP3
				var audioData []byte
				if isHLS {
					tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("sc_%d.mp3", time.Now().UnixNano()))
					cmd := exec.Command("./ffmpeg", "-y", "-i", downloadURL, "-c:a", "libmp3lame", "-q:a", "2", tmpFile)
					err := cmd.Run()
					if err != nil {
						sendMessage(ctx, "حدث خطأ أثناء تحويل الملف الصوتي.")
						return
					}
					audioData, _ = ioutil.ReadFile(tmpFile)
					os.Remove(tmpFile)
				} else {
					respDL, errDL := http.Get(downloadURL)
					if errDL != nil {
						sendMessage(ctx, "حدث خطأ أثناء تحميل الملف الصوتي.")
						return
					}
					defer respDL.Body.Close()
					audioData, _ = io.ReadAll(respDL.Body)
				}'''
content = content.replace(old_download_logic, new_download_logic)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(content)
print("Patched downloader to support HLS")
