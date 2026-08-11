package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

func main() {
	query := "cats site:pinterest.com"
	escaped := url.QueryEscape(query)
	req, _ := http.NewRequest("GET", "https://images.search.yahoo.com/search/images?p="+escaped, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	
	client := &http.Client{}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	bodyStr := string(body)
	
	fmt.Println("Status:", resp.StatusCode)
	
	// Yahoo images are usually inside <img src='...'> or JSON data 'imgurl'
	re := regexp.MustCompile(`imgurl=(https?://[^&]+)`)
	matches := re.FindAllStringSubmatch(bodyStr, -1)
	
	fmt.Println("Found matches:", len(matches))
	for i, m := range matches {
		if i < 5 {
			// URL decode the extracted URL
			decoded, _ := url.QueryUnescape(m[1])
			fmt.Println("Image:", decoded)
		}
	}
}
