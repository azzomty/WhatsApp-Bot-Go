package stickers

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func createExif(pack, author string) []byte {
	jsonStr := fmt.Sprintf(`{"sticker-pack-id":"com.snowcorp.stickerly.android.stickercontentprovider b5e7275f-f1de-4137-961f-57becfad34f2","sticker-pack-name":"%s","sticker-pack-publisher":"%s","emojis":["🤖"]}`, pack, author)
	exifData := []byte{
		0x49, 0x49, 0x2A, 0x00, 0x08, 0x00, 0x00, 0x00, 0x01, 0x00, 0x41, 0x57,
		0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0x16, 0x00, 0x00, 0x00,
	}
	
	length := len(jsonStr)
	exifData[14] = byte(length & 0xFF)
	exifData[15] = byte((length >> 8) & 0xFF)
	exifData[16] = byte((length >> 16) & 0xFF)
	exifData[17] = byte((length >> 24) & 0xFF)
	
	return append(exifData, []byte(jsonStr)...)
}

func GenerateSticker(inputData []byte, isVideo bool, pack string, author string) ([]byte, error) {
	tmpDir := os.TempDir()
	inputExt := ".jpg"
	if isVideo {
		inputExt = ".mp4"
	}
	
	timestamp := time.Now().UnixNano()
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

	webpData, err := ioutil.ReadFile(outputPath)
	if err != nil {
		return nil, err
	}

	return webpData, nil
}
