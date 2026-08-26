package youtube

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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

// GetVideoDetails gets all the details using official Google API
func GetVideoDetails(videoID string) (*VideoInfo, error) {
	if APIKey == "" {
		return nil, fmt.Errorf("API key is missing")
	}

	searchURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=snippet,contentDetails,statistics&id=%s&key=%s", url.QueryEscape(videoID), APIKey)
	resp, err := http.Get(searchURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				Title        string `json:"title"`
				ChannelTitle string `json:"channelTitle"`
				PublishedAt  string `json:"publishedAt"`
				Thumbnails   struct {
					Maxres struct {
						Url string `json:"url"`
					} `json:"maxres"`
					High struct {
						Url string `json:"url"`
					} `json:"high"`
				} `json:"thumbnails"`
			} `json:"snippet"`
			ContentDetails struct {
				Duration string `json:"duration"`
			} `json:"contentDetails"`
			Statistics struct {
				ViewCount string `json:"viewCount"`
				LikeCount string `json:"likeCount"`
			} `json:"statistics"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("no results found")
	}

	item := result.Items[0]
	pubDate, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)

	thumb := item.Snippet.Thumbnails.Maxres.Url
	if thumb == "" {
		thumb = item.Snippet.Thumbnails.High.Url
	}

	dStr := item.ContentDetails.Duration
	dStr = strings.TrimPrefix(dStr, "PT")
	dStr = strings.ToLower(dStr)
	duration, _ := time.ParseDuration(dStr)

	info := &VideoInfo{
		ID:          videoID,
		Title:       item.Snippet.Title,
		Author:      item.Snippet.ChannelTitle,
		Duration:    duration,
		PublishDate: pubDate,
		Thumbnail:   thumb,
	}

	return info, nil
}

func removeEmojis(s string) string {
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

func ParseCookies(filename string) http.CookieJar {
	jar, _ := cookiejar.New(nil)
	file, err := os.Open(filename)
	if err != nil {
		return jar
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var cookies []*http.Cookie
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 7 {
			cookie := &http.Cookie{
				Name:   parts[5],
				Value:  parts[6],
				Domain: parts[0],
				Path:   parts[2],
			}
			cookies = append(cookies, cookie)
		}
	}
	u, _ := url.Parse("https://youtube.com")
	jar.SetCookies(u, cookies)
	u2, _ := url.Parse("https://www.youtube.com")
	jar.SetCookies(u2, cookies)
	return jar
}

func DownloadMedia(videoID string, isAudio bool) ([]byte, error) {
	link := "https://www.youtube.com/watch?v=" + videoID

	ext := "mp4"
	format := "bestvideo[vcodec^=avc1][height<=720]+bestaudio[acodec^=mp4a]/best[ext=mp4][height<=720]/best"
	if isAudio {
		ext = "m4a"
		format = "bestaudio[ext=m4a]/bestaudio"
	}

	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("yt_%d.%s", time.Now().UnixNano(), ext))
	defer os.Remove(tmpFile)

	cmd := exec.Command("./yt-dlp", "--ffmpeg-location", "node_modules/ffmpeg-static/ffmpeg", "--merge-output-format", ext, "-f", format, "-o", tmpFile, link)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outStr := string(output)
		if len(outStr) > 200 {
			outStr = outStr[len(outStr)-200:]
		}
		return nil, fmt.Errorf("yt-dlp failed: %v, output: %s", err, outStr)
	}

	return os.ReadFile(tmpFile)
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

func DownloadDirectURL(url string) (string, error) {
	outPath := fmt.Sprintf("/tmp/direct_video_%d.mp4", time.Now().UnixNano())
	// Use yt-dlp to download directly from the URL
	cmd := exec.Command("./yt-dlp", "-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best", "-o", outPath, url)
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to download video: %v", err)
	}
	return outPath, nil
}
