package main
import (
	"fmt"
	"github.com/kkdai/youtube/v2"
)
func main() {
	client := youtube.Client{}
	video, err := client.GetVideo("YQHsXMglC9A")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Title: %s\n", video.Title)
	fmt.Printf("Author: %s\n", video.Author)
	fmt.Printf("PublishDate: %v\n", video.PublishDate)
	if len(video.Thumbnails) > 0 {
		fmt.Printf("Thumbnail: %s\n", video.Thumbnails[0].URL)
	}
}
