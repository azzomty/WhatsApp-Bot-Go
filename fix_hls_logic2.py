import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    content = f.read()

# Replace the download section of .اغنية
pattern = re.compile(r'// 4\. Download & Send MP3\s+respDL, errDL := http\.Get\(dlUrl\)\s+if errDL != nil \{\s+sendMessage\(ctx, "فشل تحميل الملف\."\)\s+return\s+\}\s+defer respDL\.Body\.Close\(\)\s+audioData, _ := io\.ReadAll\(respDL\.Body\)')

new_logic = '''// 4. Download & Send MP3
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

content = pattern.sub(new_logic, content)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(content)
print("Patched download logic")
