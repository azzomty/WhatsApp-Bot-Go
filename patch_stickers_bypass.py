import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/stickers/stickers.go', 'r') as f:
    content = f.read()

new_logic = """func GenerateSticker(inputData []byte, isVideo bool, pack string, author string) ([]byte, error) {
	// Check if input is already a WEBP
	isWebp := false
	if len(inputData) > 12 && string(inputData[0:4]) == "RIFF" && string(inputData[8:12]) == "WEBP" {
		isWebp = true
	}

	tmpDir := os.TempDir()
	timestamp := time.Now().UnixNano()
	
	inputExt := ".jpg"
	if isVideo {
		inputExt = ".mp4"
	}
	if isWebp {
		inputExt = ".webp"
	}
	
	inputPath := filepath.Join(tmpDir, fmt.Sprintf("in_%d%s", timestamp, inputExt))
	outputPath := filepath.Join(tmpDir, fmt.Sprintf("out_%d.webp", timestamp))
	exifPath := filepath.Join(tmpDir, fmt.Sprintf("exif_%d.exif", timestamp))
	
	err := ioutil.WriteFile(inputPath, inputData, 0644)
	if err != nil {
		return nil, err
	}
	
	// Create EXIF
	exifBytes := createExif(pack, author)
	ioutil.WriteFile(exifPath, exifBytes, 0644)
	
	defer os.Remove(inputPath)
	defer os.Remove(outputPath)
	defer os.Remove(exifPath)

	if isWebp {
		// Just inject EXIF into the existing webp!
		err = exec.Command("./webpmux", "-set", "exif", exifPath, inputPath, "-o", outputPath).Run()
		if err != nil {
			return nil, fmt.Errorf("webpmux failed: %v", err)
		}
	} else {
		var cmd *exec.Cmd
		if isVideo {
			cmd = exec.Command("./ffmpeg", "-y", "-i", inputPath,
				"-vcodec", "libwebp",
				"-vf", "scale=512:512:force_original_aspect_ratio=decrease,format=rgba,pad=512:512:-1:-1:color=#00000000",
				"-r", "10",
				"-lossless", "0",
				"-compression_level", "6",
				"-qscale", "10",
				"-loop", "0",
				"-preset", "picture",
				"-threads", "4",
				"-an",
				"-t", "00:00:05.000",
				outputPath,
			)
		} else {
			cmd = exec.Command("./ffmpeg", "-y", "-i", inputPath,
				"-vcodec", "libwebp",
				"-vf", "scale=512:512:force_original_aspect_ratio=decrease,format=rgba,pad=512:512:-1:-1:color=#00000000",
				outputPath,
			)
		}

		err = cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("ffmpeg failed: %v", err)
		}

		// Inject EXIF using webpmux
		exec.Command("./webpmux", "-set", "exif", exifPath, outputPath, "-o", outputPath).Run()
	}

	webpData, err := ioutil.ReadFile(outputPath)
	if err != nil {
		return nil, err
	}

	return webpData, nil
}"""

pattern = re.compile(r'func GenerateSticker\(inputData \[\]byte, isVideo bool, pack string, author string\) \(\[\]byte, error\) \{.*?\n\}', re.DOTALL)
content = pattern.sub(new_logic, content)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/stickers/stickers.go', 'w') as f:
    f.write(content)
print("Patched GenerateSticker for WebP bypass")
