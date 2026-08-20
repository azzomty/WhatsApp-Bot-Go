package store

import (
	"encoding/json"
	"os"
	"sync"
)

var (
	exchangeMu       sync.RWMutex
	FavoriteList     = make(map[string]bool)
	ExchangeGroup    string
	MyExchangeMsgs   [][]byte
	StrikeList       = make(map[string]int)
)

func LoadExchange() {
	exchangeMu.Lock()
	defer exchangeMu.Unlock()

	favData, err := os.ReadFile("favorites.json")
	if err == nil {
		json.Unmarshal(favData, &FavoriteList)
	}

	egData, err := os.ReadFile("exchange_group.json")
	if err == nil {
		ExchangeGroup = string(egData)
	}

	msgData, err := os.ReadFile("my_exchange.json")
	if err == nil {
		json.Unmarshal(msgData, &MyExchangeMsgs)
	}

	strikeData, err := os.ReadFile("strikes.json")
	if err == nil {
		json.Unmarshal(strikeData, &StrikeList)
	}
}

func ToggleFavorite(id string) bool {
	exchangeMu.Lock()
	defer exchangeMu.Unlock()
	
	status := !FavoriteList[id]
	if status {
		FavoriteList[id] = true
	} else {
		delete(FavoriteList, id)
	}
	
	data, _ := json.Marshal(FavoriteList)
	os.WriteFile("favorites.json", data, 0644)
	return status
}

func SetExchangeGroup(id string) {
	exchangeMu.Lock()
	defer exchangeMu.Unlock()
	ExchangeGroup = id
	os.WriteFile("exchange_group.json", []byte(id), 0644)
}

func GetExchangeGroup() string {
	exchangeMu.RLock()
	defer exchangeMu.RUnlock()
	return ExchangeGroup
}

func AddMyExchangeMsg(data []byte) {
	exchangeMu.Lock()
	defer exchangeMu.Unlock()
	MyExchangeMsgs = append(MyExchangeMsgs, data)
	b, _ := json.Marshal(MyExchangeMsgs)
	os.WriteFile("my_exchange.json", b, 0644)
}

func ClearMyExchangeMsgs() {
	exchangeMu.Lock()
	defer exchangeMu.Unlock()
	MyExchangeMsgs = make([][]byte, 0)
	os.WriteFile("my_exchange.json", []byte("[]"), 0644)
}

func GetMyExchangeMsgs() [][]byte {
	exchangeMu.RLock()
	defer exchangeMu.RUnlock()
	return MyExchangeMsgs
}

func GetFavorites() []string {
	exchangeMu.RLock()
	defer exchangeMu.RUnlock()
	var list []string
	for k := range FavoriteList {
		list = append(list, k)
	}
	return list
}


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
