package youtube

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func init() {
	// تحميل أداة yt-dlp الخارقة في الخلفية إذا لم تكن موجودة
	go func() {
		if _, err := os.Stat("yt-dlp"); os.IsNotExist(err) {
			resp, err := http.Get("https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp")
			if err == nil {
				defer resp.Body.Close()
				out, err := os.Create("yt-dlp")
				if err == nil {
					io.Copy(out, resp.Body)
					out.Close()
					os.Chmod("yt-dlp", 0755)
				}
			}
		}
	}()
}

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

// GetVideoDetails gets all the details needed for the thumbnail message using yt-dlp
func GetVideoDetails(videoID string) (*VideoInfo, error) {
	cmd := exec.Command("./yt-dlp", "--dump-json", videoID)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("فشل استخراج التفاصيل من yt-dlp: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(out, &data); err != nil {
		return nil, err
	}

	durationFloat, _ := data["duration"].(float64)
	uploadDateStr, _ := data["upload_date"].(string)
	
	publishDate, _ := time.Parse("20060102", uploadDateStr)

	title, _ := data["title"].(string)
	uploader, _ := data["uploader"].(string)
	thumbnail, _ := data["thumbnail"].(string)

	info := &VideoInfo{
		ID:          videoID,
		Title:       title,
		Author:      uploader,
		Duration:    time.Duration(durationFloat) * time.Second,
		PublishDate: publishDate,
		Thumbnail:   thumbnail,
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

// DownloadMedia downloads the audio or video using yt-dlp
func DownloadMedia(videoID string, isAudio bool) ([]byte, error) {
	filename := fmt.Sprintf("%s.mp4", videoID)
	if isAudio {
		filename = fmt.Sprintf("%s.mp3", videoID)
	}
	defer os.Remove(filename)

	var cmd *exec.Cmd
	if isAudio {
		cmd = exec.Command("./yt-dlp", "-f", "bestaudio", "--extract-audio", "--audio-format", "mp3", "-o", filename, videoID)
	} else {
		cmd = exec.Command("./yt-dlp", "-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best", "-o", filename, videoID)
	}
	
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("فشل التحميل من yt-dlp: %v", err)
	}

	return os.ReadFile(filename)
}
