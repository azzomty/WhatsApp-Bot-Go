package stickers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// GenerateSticker converts an input file (image or video) to a WebP sticker with EXIF
func GenerateSticker(inputData []byte, isVideo bool, pack string, author string) ([]byte, error) {
	req, err := http.NewRequest("POST", "http://127.0.0.1:4321/sticker", bytes.NewBuffer(inputData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-pack", url.QueryEscape(pack))
	req.Header.Set("x-author", url.QueryEscape(author))
	if isVideo {
		req.Header.Set("x-is-video", "true")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("sticker server error: %s", string(body))
	}

	return ioutil.ReadAll(resp.Body)
}
