package youtube

import (
	"sync"
)

type InteractiveSession struct {
	Query        string
	VideoIDs     []string
	CurrentIndex int
}

var (
	sessions   = make(map[string]*InteractiveSession)
	sessionMux sync.Mutex
)

// SetSession saves the search session for a user
func SetSession(key string, session *InteractiveSession) {
	sessionMux.Lock()
	defer sessionMux.Unlock()
	sessions[key] = session
}

// GetSession retrieves the search session for a user
func GetSession(key string) *InteractiveSession {
	sessionMux.Lock()
	defer sessionMux.Unlock()
	return sessions[key]
}

// DeleteSession deletes the search session
func DeleteSession(key string) {
	sessionMux.Lock()
	defer sessionMux.Unlock()
	delete(sessions, key)
}
