with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/mdm.go', 'r') as f:
    c = f.read()

import re

old_struct = '''		var data []struct {
			Author struct {
				Username string `json:"username"`
			} `json:"author"`
			DeckType struct {
				Name string `json:"name"`
			} `json:"deckType"`
			RankedType struct {
				ShortName string `json:"shortName"`
			} `json:"rankedType"`
			TournamentNumber string `json:"tournamentNumber"`
			Main             []struct {
				Card struct {
					Name string `json:"name"`
				} `json:"card"`
				Amount int `json:"amount"`
			} `json:"main"`
			Extra []struct {
				Card struct {
					Name string `json:"name"`
				} `json:"card"`
				Amount int `json:"amount"`
			} `json:"extra"`
		}'''

new_struct = '''		var data []struct {
			Author     interface{} `json:"author"`
			DeckType   interface{} `json:"deckType"`
			RankedType interface{} `json:"rankedType"`
			TournamentNumber interface{} `json:"tournamentNumber"`
			Main       []struct {
				Card struct {
					Name string `json:"name"`
				} `json:"card"`
				Amount int `json:"amount"`
			} `json:"main"`
			Extra []struct {
				Card struct {
					Name string `json:"name"`
				} `json:"card"`
				Amount int `json:"amount"`
			} `json:"extra"`
		}'''

old_loop = '''			for _, d := range data {
				if strings.EqualFold(d.DeckType.Name, session.TargetDeck.Name) {
					sd := SubDeck{Author: d.Author.Username}
					if d.TournamentNumber != "" {
						sd.Info = "Tournament " + d.TournamentNumber
					} else if d.RankedType.ShortName != "" {
						sd.Info = d.RankedType.ShortName
					}'''

new_loop = '''			for _, d := range data {
				deckName := ""
				switch v := d.DeckType.(type) {
				case string:
					deckName = v
				case map[string]interface{}:
					if n, ok := v["name"].(string); ok { deckName = n }
				}
				
				// API usually already filters by deck, but we keep check just in case.
				// Since we query ?deck=Name, we can also accept empty deckName (if API format changes)
				if deckName == "" || strings.EqualFold(deckName, session.TargetDeck.Name) {
					authorName := "Unknown"
					switch v := d.Author.(type) {
					case string:
						authorName = v
					case map[string]interface{}:
						if u, ok := v["username"].(string); ok { authorName = u }
					}
					
					sd := SubDeck{Author: authorName}
					
					tNum := ""
					switch v := d.TournamentNumber.(type) {
					case string: tNum = v
					case float64: tNum = fmt.Sprintf("%.0f", v)
					}
					
					rType := ""
					switch v := d.RankedType.(type) {
					case string: rType = v
					case map[string]interface{}:
						if n, ok := v["shortName"].(string); ok { rType = n }
					}
					
					if tNum != "" {
						sd.Info = "Tournament " + tNum
					} else if rType != "" {
						sd.Info = rType
					}'''

if old_struct in c:
    c = c.replace(old_struct, new_struct)
    c = c.replace(old_loop, new_loop)
    with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/mdm.go', 'w') as f:
        f.write(c)
    print("Patched successfully")
else:
    print("Could not find struct to patch")
