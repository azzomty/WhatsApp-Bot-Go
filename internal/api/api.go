package api

import (
	"encoding/json"
	"net/http"

	"whatsapp-bot/internal/store"
)

type SyncPayload struct {
	AllowedUsers   []string                     `json:"allowedUsers"`
	AllowedCmds    map[string]map[string]bool   `json:"allowedCommands"`
	CommandAliases map[string]map[string]string `json:"commandAliases"`
	CustomOutputs  map[string]map[string]string `json:"customOutputs"`
}

func StartServer() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	http.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload SyncPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		store.SetSyncState(payload.AllowedUsers, payload.AllowedCmds, payload.CommandAliases, payload.CustomOutputs)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"synced"}`))
	})

}
