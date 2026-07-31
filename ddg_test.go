package main
import (
	"fmt"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"regexp"
)
func main() {
	req, _ := http.NewRequest("GET", "https://duckduckgo.com/?q=matching+icons+anime+site:pinterest.com", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, _ := http.DefaultClient.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	
	re := regexp.MustCompile(`vqd=([a-zA-Z0-9_-]+)`)
	match := re.FindStringSubmatch(string(body))
	if len(match) < 2 { return }
	vqd := match[1]
	
	req2, _ := http.NewRequest("GET", "https://duckduckgo.com/i.js?l=us-en&o=json&q=matching+icons+anime+site:pinterest.com&vqd="+vqd+"&f=,,,&p=1", nil)
	req2.Header.Set("User-Agent", "Mozilla/5.0")
	resp2, _ := http.DefaultClient.Do(req2)
	body2, _ := ioutil.ReadAll(resp2.Body)
	resp2.Body.Close()
	
	var data map[string]interface{}
	json.Unmarshal(body2, &data)
	results := data["results"].([]interface{})
	for i := 0; i < 10 && i < len(results); i++ {
		res := results[i].(map[string]interface{})
		fmt.Println(res["image"])
	}
}
