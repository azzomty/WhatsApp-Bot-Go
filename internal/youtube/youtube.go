package youtube

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
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

// GetVideoDetails gets all the details using kkdai/youtube/v2
func GetVideoDetails(videoID string) (*VideoInfo, error) {
	client := youtube.Client{}
	video, err := client.GetVideo(videoID)
	if err != nil {
		return nil, err
	}

	var thumb string
	if len(video.Thumbnails) > 0 {
		thumb = fmt.Sprintf("https://i.ytimg.com/vi/%s/maxresdefault.jpg", videoID)
	}

	info := &VideoInfo{
		ID:          videoID,
		Title:       video.Title,
		Author:      video.Author,
		Duration:    video.Duration,
		PublishDate: video.PublishDate,
		Thumbnail:   thumb,
	}

	return info, nil
}

func removeEmojis(s string) string {
	// Simple emoji removal (from original logic)
	// For simplicity, we just return the string as is, or we can use strings.Map to remove non-printable characters.
	// We'll strip basic dots if needed since user wants no ...
	s = strings.ReplaceAll(s, "...", "")
	s = strings.ReplaceAll(s, "..", "")
	return s
}

func FormatCaption(info *VideoInfo) string {
	cleanTitle := removeEmojis(info.Title)
	cleanAuthor := removeEmojis(info.Author)
	return fmt.Sprintf("=== تفاصيل المقطع ===\n\nالعنوان: %s\nالقناة: %s\nالمدة: %v\nتاريخ النشر: %s",
		cleanTitle, cleanAuthor, info.Duration, info.PublishDate.Format("2006-01-02"))
}

// DownloadMedia downloads the audio or video using kkdai/youtube/v2
func DownloadMedia(videoID string, isAudio bool) ([]byte, error) {
	client := youtube.Client{}
	video, err := client.GetVideo(videoID)
	if err != nil {
		return nil, err
	}

	var targetFormat *youtube.Format

	if isAudio {
		formats := video.Formats.Type("audio/mp4")
		if len(formats) > 0 {
			targetFormat = &formats[0]
			for i := range formats {
				if formats[i].ItagNo == 140 { // m4a audio
					targetFormat = &formats[i]
					break
				}
			}
		} else {
			formats := video.Formats.WithAudioChannels()
			if len(formats) > 0 {
				targetFormat = &formats[0]
			}
		}
	} else {
		formats := video.Formats.WithAudioChannels()
		if len(formats) == 0 {
			return nil, fmt.Errorf("no formats found")
		}

		for i := range formats {
			if formats[i].ItagNo == 18 { // 360p mp4
				targetFormat = &formats[i]
				break
			}
		}
		if targetFormat == nil {
			for i := range formats {
				if formats[i].ItagNo == 22 { // 720p mp4
					targetFormat = &formats[i]
					break
				}
			}
		}
		if targetFormat == nil {
			targetFormat = &formats[0]
		}
	}

	if targetFormat == nil {
		return nil, fmt.Errorf("could not find suitable format")
	}

	stream, _, err := client.GetStream(video, targetFormat)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	return io.ReadAll(stream)
}

func SearchVideos(query string, maxResults int, pageToken string) ([]string, string, error) {
	if APIKey == "" {
		return nil, "", fmt.Errorf("API key is missing")
	}

	searchURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?part=snippet&q=%s&type=video&maxResults=%d&key=%s", url.QueryEscape(query), maxResults, APIKey)
	if pageToken != "" {
		searchURL += "&pageToken=" + pageToken
	}

	resp, err := http.Get(searchURL)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	var result struct {
		NextPageToken string `json:"nextPageToken"`
		Items         []struct {
			ID struct {
				VideoId string `json:"videoId"`
			} `json:"id"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, "", err
	}

	var videoIDs []string
	for _, item := range result.Items {
		videoIDs = append(videoIDs, item.ID.VideoId)
	}

	if len(videoIDs) == 0 {
		return nil, "", fmt.Errorf("no videos found")
	}

	return videoIDs, result.NextPageToken, nil
}