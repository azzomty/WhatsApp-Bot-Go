package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func main() {
    query := "قطط"
    dataJson := fmt.Sprintf(`{"options":{"query":"%s","scope":"pins","bookmarks":[""],"add_refine":[]},"context":{}}`, query)
    dataEncoded := url.QueryEscape(dataJson)
    
    urlStr := fmt.Sprintf("https://www.pinterest.com/resource/BaseSearchResource/get/?source_url=/search/pins/?q=%s&data=%s", url.QueryEscape(query), dataEncoded)
    req, _ := http.NewRequest("GET", urlStr, nil)
	
	req.Header.Set("Cookie", "_b=AZehPVTHje5FSKPWa+hL4qmM/XEDJuxk13yIX8h3VBWeJwNgD6CaB3qWfEhPQT8YcaY=")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")
    req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
    req.Header.Set("X-Requested-With", "XMLHttpRequest")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        fmt.Println(err)
        return
    }
    b, _ := ioutil.ReadAll(resp.Body)
    fmt.Println(string(b))
}
