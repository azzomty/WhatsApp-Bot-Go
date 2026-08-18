package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	// patch internal/commands/commands.go
	content, _ := ioutil.ReadFile("internal/commands/commands.go")
	str := string(content)
	str = strings.Replace(str, "results = pinterest.SearchPinterest(last.Query, last.Aspect)", "results = pinterest.SearchPinterest(last.Query, last.Aspect, last.Count)", 1)
	ioutil.WriteFile("internal/commands/commands.go", []byte(str), 0644)

	// patch main.go
	content, _ = ioutil.ReadFile("main.go")
	str = string(content)
	str = strings.Replace(str, "results = pinterest.SearchPinterest(req.Query+\" gif\", \"gif\")", "results = pinterest.SearchPinterest(req.Query+\" gif\", \"gif\", overrideCount)", 1)
	str = strings.Replace(str, "results = pinterest.SearchPinterest(req.Query+suffix, aspect)", "results = pinterest.SearchPinterest(req.Query+suffix, aspect, overrideCount)", 1)
	str = strings.Replace(str, "results = pinterest.SearchPinterestMedia(req.Query, \".mp4\")", "results = pinterest.SearchPinterestMedia(req.Query, \".mp4\", overrideCount)", 1)
	ioutil.WriteFile("main.go", []byte(str), 0644)

	// patch internal/pinterest/media_search.go
	content, _ = ioutil.ReadFile("internal/pinterest/media_search.go")
	str = string(content)
	str = strings.Replace(str, "func SearchPinterestMedia(query string, ext string) []PinResult {", "func SearchPinterestMedia(query string, ext string, count int) []PinResult {", 1)
	str = strings.Replace(str, "pins := SearchPinterest(searchQuery, \"all\")", "pins := SearchPinterest(searchQuery, \"all\", count)", 1)
	ioutil.WriteFile("internal/pinterest/media_search.go", []byte(str), 0644)

	// patch internal/pinterest/pinterest.go
	content, _ = ioutil.ReadFile("internal/pinterest/pinterest.go")
	str = string(content)
	str = strings.Replace(str, "pins := SearchPinterest(\"matching icons \"+query, \"all\")", "pins := SearchPinterest(\"matching icons \"+query, \"all\", 10)", 1)
	ioutil.WriteFile("internal/pinterest/pinterest.go", []byte(str), 0644)

	fmt.Println("Done")
}
