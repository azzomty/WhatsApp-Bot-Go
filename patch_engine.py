import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/mdm.go', 'r') as f:
    c = f.read()

old_fetch = '''	session.SubDecks = []SubDeck{}

	// 1. Get Deck Type ID
	deckTypeURL := "https://www.masterduelmeta.com/api/v1/deck-types?name=" + url.QueryEscape(session.TargetDeck.Name)
	req1, _ := http.NewRequest("GET", deckTypeURL, nil)
	req1.Header.Set("User-Agent", "Mozilla/5.0")
	res1, err1 := http.DefaultClient.Do(req1)
	
	var deckID string
	if err1 == nil && res1.StatusCode == 200 {
		defer res1.Body.Close()
		body1, _ := io.ReadAll(res1.Body)
		var deckTypes []struct {
			ID string `json:"_id"`
		}
		json.Unmarshal(body1, &deckTypes)
		if len(deckTypes) > 0 {
			deckID = deckTypes[0].ID
		}
	}

	if deckID != "" {
		// 2. Fetch Top Decks using Deck Type ID
		apiURL := "https://www.masterduelmeta.com/api/v1/top-decks?deckType=" + deckID + "&sort=-created&limit=15"
		req, _ := http.NewRequest("GET", apiURL, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		res, err := http.DefaultClient.Do(req)'''

new_fetch = '''	session.SubDecks = []SubDeck{}

	// Parse URL to determine if it's an engine or deck-type
	u, _ := url.Parse(session.TargetDeck.URL)
	segments := strings.Split(strings.Trim(u.Path, "/"), "/")
	
	isEngine := false
	entityName := session.TargetDeck.Name
	if len(segments) >= 3 {
		if segments[1] == "engines" {
			isEngine = true
		}
		entityName, _ = url.PathUnescape(segments[2])
	}

	// 1. Get ID
	var metaURL string
	if isEngine {
		metaURL = "https://www.masterduelmeta.com/api/v1/engines?name=" + url.QueryEscape(entityName)
	} else {
		metaURL = "https://www.masterduelmeta.com/api/v1/deck-types?name=" + url.QueryEscape(entityName)
	}
	
	req1, _ := http.NewRequest("GET", metaURL, nil)
	req1.Header.Set("User-Agent", "Mozilla/5.0")
	res1, err1 := http.DefaultClient.Do(req1)
	
	var entityID string
	if err1 == nil && res1.StatusCode == 200 {
		defer res1.Body.Close()
		body1, _ := io.ReadAll(res1.Body)
		var items []struct {
			ID string `json:"_id"`
		}
		json.Unmarshal(body1, &items)
		if len(items) > 0 {
			entityID = items[0].ID
		}
	}

	if entityID != "" {
		// 2. Fetch Top Decks using ID
		apiURL := ""
		if isEngine {
			apiURL = "https://www.masterduelmeta.com/api/v1/top-decks?engines=" + entityID + "&sort=-created&limit=15"
		} else {
			apiURL = "https://www.masterduelmeta.com/api/v1/top-decks?deckType=" + entityID + "&sort=-created&limit=15"
		}
		
		req, _ := http.NewRequest("GET", apiURL, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		res, err := http.DefaultClient.Do(req)'''

c = c.replace(old_fetch, new_fetch)

# Also fix the `if deckID != ""` to `if entityID != ""` in the block closing if necessary, but we replaced the whole thing.
# Oh wait, the closing brace we modified earlier was `if deckID != "" {`.
c = c.replace('if deckID != "" {', 'if entityID != "" {')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/mdm.go', 'w') as f:
    f.write(c)
