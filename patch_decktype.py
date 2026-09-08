import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/mdm.go', 'r') as f:
    c = f.read()

old_fetch = '''	// Fetch from Top Decks API (last 30 days)
	apiURL := "https://www.masterduelmeta.com/api/v1/top-decks?deck=" + url.QueryEscape(session.TargetDeck.Name)
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	res, err := http.DefaultClient.Do(req)

	session.SubDecks = []SubDeck{}

	if err == nil {
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		if res.StatusCode != 200 {
			sendMessage(ctx, fmt.Sprintf("DEBUG: API Status %d", res.StatusCode))
		}
		
		var dbgErr error'''

new_fetch = '''	session.SubDecks = []SubDeck{}

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
		res, err := http.DefaultClient.Do(req)

		if err == nil {
			defer res.Body.Close()
			body, _ := io.ReadAll(res.Body)
			
			var dbgErr error'''

c = c.replace(old_fetch, new_fetch)

# Now we need to close the if block correctly.
# Currently:
#		dbgErr = json.Unmarshal(body, &data)
#		if dbgErr != nil { ... }
#		if dbgErr == nil { ... for loop ... }
#	} // ends if err == nil
# We need to add one more `}` for `if deckID != ""`

old_fallback = '''	}

	// Fallback to web scraping if API returns nothing or fails'''
new_fallback = '''		}
	}

	// Fallback to web scraping if API returns nothing or fails'''

c = c.replace(old_fallback, new_fallback)

# ALSO, we need to REMOVE the debug API Status 403 because it's no longer there in new_fetch. We didn't copy it over, which is fine since the user already confirmed it works.
# Let's save and see if it builds!
with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/mdm.go', 'w') as f:
    f.write(c)
