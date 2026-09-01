package stickers

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func GenerateSticker(inputData []byte, isVideo bool, pack string, author string) ([]byte, error) {
	tmpDir := os.TempDir()
	inputExt := ".jpg"
	if isVideo {
		inputExt = ".mp4"
	}
	
	timestamp := time.Now().UnixNano()
	inputPath := filepath.Join(tmpDir, fmt.Sprintf("in_%d%s", timestamp, inputExt))
	outputPath := filepath.Join(tmpDir, fmt.Sprintf("out_%d.webp", timestamp))
	
	err := ioutil.WriteFile(inputPath, inputData, 0644)
	if err != nil {
		return nil, err
	}
	defer os.Remove(inputPath)
	defer os.Remove(outputPath)

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

	webpData, err := ioutil.ReadFile(outputPath)
	if err != nil {
		return nil, err
	}

	return webpData, nil
}
