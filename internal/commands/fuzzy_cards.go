package commands

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	allCardNames []string
	cardsMu      sync.Mutex
)

func LoadCardNames() {
	// Fire and forget
	go func() {
		client := &http.Client{Timeout: 15 * time.Second}
		req, _ := http.NewRequest("GET", "https://db.ygoprodeck.com/api/v7/cardinfo.php", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		resp, err := client.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		var data struct {
			Data []struct {
				Name string `json:"name"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &data); err == nil {
			var names []string
			for _, c := range data.Data {
				names = append(names, c.Name)
			}
			cardsMu.Lock()
			allCardNames = names
			cardsMu.Unlock()
		}
	}()
}

func levenshtein(a, b string) int {
	a = strings.ToLower(a)
	b = strings.ToLower(b)
	la := len(a)
	lb := len(b)
	d := make([][]int, la+1)
	for i := range d {
		d[i] = make([]int, lb+1)
	}
	for i := 0; i <= la; i++ {
		d[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		d[0][j] = j
	}
	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			min := d[i-1][j] + 1
			if d[i][j-1]+1 < min {
				min = d[i][j-1] + 1
			}
			if d[i-1][j-1]+cost < min {
				min = d[i-1][j-1] + cost
			}
			d[i][j] = min
		}
	}
	return d[la][lb]
}

func GetClosestCardName(query string) string {
	cardsMu.Lock()
	names := allCardNames
	cardsMu.Unlock()

	if len(names) == 0 {
		return query // Fallback if not loaded
	}

	bestMatch := query
	bestDist := 99999

	for _, name := range names {
		// Exact match takes precedence
		if strings.EqualFold(name, query) {
			return name
		}

		dist := levenshtein(query, name)
		if dist < bestDist {
			bestDist = dist
			bestMatch = name
		}

		// If query is a substring of the name, treat it as a very close match
		if strings.Contains(strings.ToLower(name), strings.ToLower(query)) {
			// Substring cost = length difference
			subDist := len(name) - len(query)
			// Favor substring over levenshtein if it's not too long
			if subDist < bestDist {
				bestDist = subDist
				bestMatch = name
			}
		}
	}

	// Only return if it's somewhat close
	if bestDist <= len(query) {
		return bestMatch
	}

	return query
}

func init() {
	LoadCardNames()
}
