package youtube

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/kkdai/youtube/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var APIKey = "AIzaSyD1Rnl8ANpNH2DBN1GibmpcU_VYXVGbiu4"

type VideoInfo struct {
	ID          string
	Title       string
	Author      string
	Duration    time.Duration
	Views       int
	Likes       int
	PublishDate time.Time
	Thumbnail   string
}

// SearchVideo searches YouTube via the official API and returns the video ID
func SearchVideo(query string) (string, error) {
	if APIKey == "" {
		return "", fmt.Errorf("API key is missing")
	}

	searchURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?part=snippet&q=%s&type=video&maxResults=1&key=%s", url.QueryEscape(query), APIKey)
	resp, err := http.Get(searchURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Items []struct {
			ID struct {
				VideoId string `json:"videoId"`
			} `json:"id"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Items) == 0 {
		return "", fmt.Errorf("no results found")
	}

	return result.Items[0].ID.VideoId, nil
}

// GetVideoDetails gets all the details needed for the thumbnail message
func GetVideoDetails(videoID string) (*VideoInfo, error) {
	client := youtube.Client{}
	vid, err := client.GetVideo(videoID)
	if err != nil {
		return nil, err
	}

	// We can use the YouTube Data API to get likes if needed, but youtube/v2 doesn't provide likes directly.
	// We'll leave likes as 0 or we can fetch it via API.
	
	info := &VideoInfo{
		ID:          vid.ID,
		Title:       vid.Title,
		Author:      vid.Author,
		Duration:    vid.Duration,
		PublishDate: vid.PublishDate,
	}

	if len(vid.Thumbnails) > 0 {
		// Get highest quality thumbnail
		info.Thumbnail = vid.Thumbnails[len(vid.Thumbnails)-1].URL
	}

	return info, nil
}

func removeEmojis(s string) string {
	// A simple regex to remove common emojis
	re := regexp.MustCompile(`[\x{1F600}-\x{1F64F}\x{1F300}-\x{1F5FF}\x{1F680}-\x{1F6FF}\x{1F1E0}-\x{1F1FF}\x{2600}-\x{26FF}\x{2700}-\x{27BF}]`)
	return re.ReplaceAllString(s, "")
}

func FormatCaption(info *VideoInfo) string {
	p := message.NewPrinter(language.Arabic)
	cleanTitle := removeEmojis(info.Title)
	cleanAuthor := removeEmojis(info.Author)
	return p.Sprintf("*العنوان:* %s\n*القناة:* %s\n*المدة:* %v\n*تاريخ النشر:* %s\n\nجاري التحميل...",
		cleanTitle, cleanAuthor, info.Duration, info.PublishDate.Format("2006-01-02"))
}

// DownloadMedia downloads the audio or video
func DownloadMedia(videoID string, isAudio bool) ([]byte, error) {
	client := youtube.Client{}
	vid, err := client.GetVideo(videoID)
	if err != nil {
		return nil, err
	}

	formats := vid.Formats.WithAudioChannels()
	if !isAudio {
		formats = vid.Formats
	}
	
	if len(formats) == 0 {
		return nil, fmt.Errorf("no format found")
	}
	formats.Sort()

	var bestFormat youtube.Format
	if isAudio {
		bestFormat = formats[0]
	} else {
		// Try to find a good quality video with audio (like 720p mp4)
		found := false
		for _, f := range formats {
			if f.AudioChannels > 0 && (f.QualityLabel == "720p" || f.QualityLabel == "480p" || f.QualityLabel == "360p") {
				bestFormat = f
				found = true
				break
			}
		}
		if !found {
			bestFormat = formats[0] // fallback
		}
	}

	stream, _, err := client.GetStream(vid, &bestFormat)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, stream)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
