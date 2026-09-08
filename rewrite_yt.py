import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    c = f.read()

# Replace the whole go func() block for .اغنية with the new logic
old_block_match = re.search(r'go func\(\) \{.*?\n\t\t\}\(\)\n\t\treturn true', c, re.DOTALL)
if not old_block_match:
    print("Could not find block!")
    exit(1)

old_block = old_block_match.group(0)

new_block = '''go func() {
			// Search and extract details using yt-dlp
			cmdSearch := exec.Command("./yt-dlp", "ytsearch1:"+query, "--dump-json", "--extractor-args", "youtube:player_client=android,web")
			out, errS := cmdSearch.CombinedOutput()
			if errS != nil || len(out) < 10 {
				sendMessage(ctx, "لم يتم العثور على الأغنية.")
				return
			}
			
			// Try to find the JSON line in the output (yt-dlp prints warnings)
			var jsonStr string
			for _, line := range strings.Split(string(out), "\\n") {
				if strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}") {
					jsonStr = line
					break
				}
			}
			
			if jsonStr == "" {
				sendMessage(ctx, "فشل جلب تفاصيل الأغنية.")
				return
			}

			var video struct {
				Title      string `json:"title"`
				Channel    string `json:"uploader"`
				Views      int    `json:"view_count"`
				Thumbnail  string `json:"thumbnail"`
				WebpageURL string `json:"webpage_url"`
			}
			
			if err := json.Unmarshal([]byte(jsonStr), &video); err != nil {
				sendMessage(ctx, "فشل قراءة تفاصيل الأغنية.")
				return
			}

			// Extract Details
			caption := fmt.Sprintf("*%s*\\n\\nالقناة: %s\\nالمشاهدات: %d", video.Title, video.Channel, video.Views)

			if video.Thumbnail != "" {
				sendImageFromURL(ctx, video.Thumbnail, caption)
			} else {
				sendMessage(ctx, caption)
			}

			// Download Audio
			processDownload(ctx, video.WebpageURL, "audio")
		}()
		return true'''

c = c.replace(old_block, new_block)

# Ensure "encoding/json" is imported
if '"encoding/json"' not in c:
    c = c.replace('"context"', '"context"\n\t"encoding/json"')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(c)
