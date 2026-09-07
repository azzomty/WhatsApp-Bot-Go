import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'r') as f:
    content = f.read()

new_func = """
func DownloadM3U8WithQuality(m3u8URL, quality string) ([]byte, error) {
	tmpFile := "/tmp/stardima_" + strconv.FormatInt(time.Now().UnixNano(), 10) + ".mp4"
	defer os.Remove(tmpFile)

	cmd := exec.Command("yt-dlp", "-N", "16", "--no-check-certificate", "-f", quality, m3u8URL, "-o", tmpFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp failed: %v\\nOutput: %s", err, string(out))
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read output file: %v", err)
	}
	return data, nil
}
"""
content = content + "\n" + new_func

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'w') as f:
    f.write(content)
