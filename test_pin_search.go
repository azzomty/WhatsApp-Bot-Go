package main
import (
	"fmt"
	"whatsapp-bot/internal/pinterest"
)
func main() {
	res := pinterest.SearchPinterest("anime", "all")
	fmt.Println("Initial Pins:", len(res))
}
