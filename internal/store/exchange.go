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
	MyExchangeMsg    []byte
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
		MyExchangeMsg = msgData
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

func SetMyExchangeMsg(data []byte) {
	exchangeMu.Lock()
	defer exchangeMu.Unlock()
	MyExchangeMsg = data
	os.WriteFile("my_exchange.json", data, 0644)
}

func GetMyExchangeMsg() []byte {
	exchangeMu.RLock()
	defer exchangeMu.RUnlock()
	return MyExchangeMsg
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
