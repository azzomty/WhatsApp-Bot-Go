package youtube

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"

	"github.com/kkdai/youtube/v2"
)

// SearchAndDownloadAudio searches YouTube and returns the best audio stream as bytes
func SearchAndDownloadAudio(query string) ([]byte, string, error) {
	// 1. Search YouTube
	searchURL := "https://www.youtube.com/results?search_query=" + url.QueryEscape(query)
	resp, err := http.Get(searchURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to search: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	// 2. Extract first video ID using regex
	re := regexp.MustCompile(`"videoId":"([a-zA-Z0-9_-]{11})"`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return nil, "", fmt.Errorf("no results found")
	}
	videoID := matches[1]

	// 3. Download Audio using kkdai/youtube/v2
	client := youtube.Client{}
	video, err := client.GetVideo(videoID)
	if err != nil {
		return nil, "", err
	}

	// Try to find the best audio format
	formats := video.Formats.WithAudioChannels()
	if len(formats) == 0 {
		return nil, "", fmt.Errorf("no audio format found")
	}
	
	// Sort by bitrate (descending) to get the best quality
	formats.Sort()
	bestFormat := formats[0]

	stream, _, err := client.GetStream(video, &bestFormat)
	if err != nil {
		return nil, "", err
	}
	defer stream.Close()

	// Read stream into buffer
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(stream)
	if err != nil {
		return nil, "", err
	}

	return buf.Bytes(), video.Title, nil
}
