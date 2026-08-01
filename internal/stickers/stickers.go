package stickers

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

// GenerateSticker converts an input file (image or video) to a WebP sticker with EXIF
func GenerateSticker(inputData []byte, isVideo bool, pack string, author string) ([]byte, error) {
	tmpId := uuid.New().String()
	ext := ".jpg"
	if isVideo {
		ext = ".mp4"
	}

	inPath := filepath.Join(os.TempDir(), tmpId+ext)
	outPath := filepath.Join(os.TempDir(), tmpId+".webp")

	err := ioutil.WriteFile(inPath, inputData, 0644)
	if err != nil {
		return nil, err
	}
	defer os.Remove(inPath)
	defer os.Remove(outPath)

	cmd := exec.Command("node", "./add_exif.js", inPath, outPath, pack, author)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("node error: %v %s", err, stderr.String())
	}

	webpData, err := ioutil.ReadFile(outPath)
	if err != nil {
		return nil, err
	}

	return webpData, nil
}
