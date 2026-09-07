import re

with open("internal/store/exchange.go", "r") as f:
    c = f.read()

target_vars = """	ExchangeGroup    string
	MyExchangeMsgs   [][]byte
)"""
new_vars = """	ExchangeGroup    string
	MyExchangeMsgs   [][]byte
	StrikeList       = make(map[string]int)
)"""
c = c.replace(target_vars, new_vars)

target_load = """	msgData, err := os.ReadFile("my_exchange.json")
	if err == nil {
		json.Unmarshal(msgData, &MyExchangeMsgs)
	}
}"""
new_load = """	msgData, err := os.ReadFile("my_exchange.json")
	if err == nil {
		json.Unmarshal(msgData, &MyExchangeMsgs)
	}

	strikeData, err := os.ReadFile("strikes.json")
	if err == nil {
		json.Unmarshal(strikeData, &StrikeList)
	}
}"""
c = c.replace(target_load, new_load)

new_funcs = """

func IncrementStrike(id string) int {
	exchangeMu.Lock()
	defer exchangeMu.Unlock()
	StrikeList[id]++
	val := StrikeList[id]
	data, _ := json.Marshal(StrikeList)
	os.WriteFile("strikes.json", data, 0644)
	return val
}

func ResetStrike(id string) {
	exchangeMu.Lock()
	defer exchangeMu.Unlock()
	if StrikeList[id] != 0 {
		StrikeList[id] = 0
		data, _ := json.Marshal(StrikeList)
		os.WriteFile("strikes.json", data, 0644)
	}
}

func GetStrike(id string) int {
	exchangeMu.RLock()
	defer exchangeMu.RUnlock()
	return StrikeList[id]
}
"""

c = c + new_funcs

with open("internal/store/exchange.go", "w") as f:
    f.write(c)

