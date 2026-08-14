package main
import (
	"fmt"
	"whatsapp-bot/internal/pinterest"
)
func main() {
	res := pinterest.SearchPinterestMedia("anime", ".gif")
	fmt.Println("GIF Results:", len(res))
}
