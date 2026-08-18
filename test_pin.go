package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func main() {
	query := url.QueryEscape("قطط")
    urlStr := fmt.Sprintf("https://www.pinterest.com/resource/BaseSearchResource/get/?source_url=/search/pins/?q=%s&data={\"options\":{\"query\":\"%s\"}}", query, query)
    req, _ := http.NewRequest("GET", urlStr, nil)
	
	req.Header.Set("Cookie", "_b=AZehPVTHje5FSKPWa+hL4qmM/XEDJuxk13yIX8h3VBWeJwNgD6CaB3qWfEhPQT8YcaY=")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        fmt.Println(err)
        return
    }
    b, _ := ioutil.ReadAll(resp.Body)
    fmt.Println(string(b))
}
