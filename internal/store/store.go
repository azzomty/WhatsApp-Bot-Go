package store

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

var (
	MutedUsers     = make(map[string]bool)
	AllowedUsers   = make(map[string]bool)
	CommandAliases = make(map[string]map[string]string)
	CustomOutputs  = make(map[string]map[string]string)
	StickerAuthors = make(map[string]map[string]string)
	BaymaxNames    = make(map[string]string)
	Words          = make(map[string]interface{})
	Hebebia        = make([]string, 0)
	CommandBans    = make(map[string]map[string]bool)
	TargetGroups   = make(map[string]string)
	mutedMutex     sync.RWMutex
	allowedMutex   sync.RWMutex
	aliasMutex     sync.RWMutex
	OutputMutex    sync.RWMutex
	stickerMutex   sync.RWMutex
	baymaxMutex    sync.RWMutex
	hebebiaMutex   sync.RWMutex
	zotMutex       sync.RWMutex
	banMutex       sync.RWMutex
	targetMutex    sync.RWMutex
	ZotCounter     int
)

func loadJSON(filename string, v interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()
	return json.NewDecoder(file).Decode(v)
}

func saveJSON(filename string, v interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func LoadAll(baseDir string) {
	loadJSON(baseDir+"/muted_users.json", &MutedUsers)

	var allowed []string
	if err := loadJSON(baseDir+"/allowed_users.json", &allowed); err == nil {
		for _, u := range allowed {
			AllowedUsers[u] = true
		}
	}
	var unlockedLids []string
	if err := loadJSON(baseDir+"/unlocked_lids.json", &unlockedLids); err == nil {
		for _, u := range unlockedLids {
			AllowedUsers[u] = true
		}
	}

	loadJSON(baseDir+"/command_aliases.json", &CommandAliases)
	loadJSON(baseDir+"/custom_outputs.json", &CustomOutputs)
	loadJSON(baseDir+"/sticker_authors.json", &StickerAuthors)
	loadJSON(baseDir+"/baymax_names.json", &BaymaxNames)
	loadJSON(baseDir+"/words.json", &Words)
	loadJSON(baseDir+"/hebebia.json", &Hebebia)
	loadJSON(baseDir+"/command_bans.json", &CommandBans)
	loadJSON(baseDir+"/target_groups.json", &TargetGroups)

	if data, err := os.ReadFile(baseDir + "/zot_counter.txt"); err == nil {
		var zot int
		if _, err := fmt.Sscanf(string(data), "%d", &zot); err == nil {
			ZotCounter = zot
		}
	} else {
		ZotCounter = 184
	}
}

func SaveMuted(baseDir string) {
	mutedMutex.RLock()
	defer mutedMutex.RUnlock()
	saveJSON(baseDir+"/muted_users.json", MutedUsers)
}

func SaveAllowed(baseDir string) {
	allowedMutex.RLock()
	defer allowedMutex.RUnlock()
	var arr []string
	for u := range AllowedUsers {
		arr = append(arr, u)
	}
	saveJSON(baseDir+"/allowed_users.json", arr)
}

func SaveBans(baseDir string) {
	banMutex.RLock()
	defer banMutex.RUnlock()
	saveJSON(baseDir+"/command_bans.json", CommandBans)
}

func IsCommandBanned(userID string, cmd string) bool {
	banMutex.RLock()
	defer banMutex.RUnlock()
	if userBans, ok := CommandBans[userID]; ok {
		return userBans[cmd]
	}
	return false
}

func SetCommandBan(userID string, cmd string, banned bool, baseDir string) {
	banMutex.Lock()
	if CommandBans[userID] == nil {
		CommandBans[userID] = make(map[string]bool)
	}
	if banned {
		CommandBans[userID][cmd] = true
	} else {
		delete(CommandBans[userID], cmd)
	}
	banMutex.Unlock()
	SaveBans(baseDir)
}

func SaveTargetGroups(baseDir string) {
	targetMutex.RLock()
	defer targetMutex.RUnlock()
	saveJSON(baseDir+"/target_groups.json", TargetGroups)
}

func SetTargetGroup(key string, groupID string, baseDir string) {
	targetMutex.Lock()
	TargetGroups[key] = groupID
	targetMutex.Unlock()
	SaveTargetGroups(baseDir)
}

func GetTargetGroup(key string) string {
	targetMutex.RLock()
	defer targetMutex.RUnlock()
	return TargetGroups[key]
}

func SaveAliases(baseDir string) {
	aliasMutex.RLock()
	defer aliasMutex.RUnlock()
	saveJSON(baseDir+"/command_aliases.json", CommandAliases)
}

func SaveOutputs(baseDir string) {
	OutputMutex.RLock()
	defer OutputMutex.RUnlock()
	saveJSON(baseDir+"/custom_outputs.json", CustomOutputs)
}

func SaveStickerAuthors(baseDir string) {
	stickerMutex.RLock()
	defer stickerMutex.RUnlock()
	saveJSON(baseDir+"/sticker_authors.json", StickerAuthors)
}

func SaveBaymaxNames(baseDir string) {
	baymaxMutex.RLock()
	defer baymaxMutex.RUnlock()
	saveJSON(baseDir+"/baymax_names.json", BaymaxNames)
}

func IsMuted(id string) bool {
	mutedMutex.RLock()
	defer mutedMutex.RUnlock()
	return MutedUsers[id]
}

func SetMuted(id string, muted bool, baseDir string) {
	mutedMutex.Lock()
	if muted {
		MutedUsers[id] = true
	} else {
		delete(MutedUsers, id)
	}
	mutedMutex.Unlock()
	SaveMuted(baseDir)
}

func IsAllowed(id string) bool {
	allowedMutex.RLock()
	defer allowedMutex.RUnlock()
	return AllowedUsers[id]
}

func GetAlias(id, cmd string) string {
	aliasMutex.RLock()
	defer aliasMutex.RUnlock()
	if userAliases, ok := CommandAliases[id]; ok {
		for originalCmd, newCmd := range userAliases {
			if cmd == newCmd {
				return originalCmd
			}
		}
	}
	return cmd
}

func SetAlias(id, oldCmd, newCmd string, baseDir string) {
	aliasMutex.Lock()
	if CommandAliases[id] == nil {
		CommandAliases[id] = make(map[string]string)
	}
	CommandAliases[id][oldCmd] = newCmd
	aliasMutex.Unlock()
	SaveAliases(baseDir)
}

func GetCustomOutput(id, cmd, defaultOut string) string {
	OutputMutex.RLock()
	defer OutputMutex.RUnlock()
	if userOutputs, ok := CustomOutputs[id]; ok {
		if out, exists := userOutputs[cmd]; exists {
			return out
		}
	}
	return defaultOut
}

func SetCustomOutput(id, cmd, output string, baseDir string) {
	OutputMutex.Lock()
	if CustomOutputs[id] == nil {
		CustomOutputs[id] = make(map[string]string)
	}
	CustomOutputs[id][cmd] = output
	OutputMutex.Unlock()
	SaveOutputs(baseDir)
}

func GetStickerAuthor(id string) map[string]string {
	stickerMutex.RLock()
	defer stickerMutex.RUnlock()
	if author, ok := StickerAuthors[id]; ok {
		return author
	}
	return map[string]string{"pack": "B O T", "author": "Z E R O"}
}

func SetStickerAuthor(id, pack, author string, baseDir string) {
	stickerMutex.Lock()
	StickerAuthors[id] = map[string]string{"pack": pack, "author": author}
	stickerMutex.Unlock()
	SaveStickerAuthors(baseDir)
}

func GetBaymaxName(id string) string {
	baymaxMutex.RLock()
	defer baymaxMutex.RUnlock()
	if name, ok := BaymaxNames[id]; ok {
		return name
	}
	return ""
}

func SetBaymaxName(id, name, baseDir string) {
	baymaxMutex.Lock()
	BaymaxNames[id] = name
	baymaxMutex.Unlock()
	SaveBaymaxNames(baseDir)
}

func SaveHebebia(baseDir string) {
	hebebiaMutex.RLock()
	defer hebebiaMutex.RUnlock()
	saveJSON(baseDir+"/hebebia.json", Hebebia)
}

func GetHebebia() []string {
	hebebiaMutex.RLock()
	defer hebebiaMutex.RUnlock()
	return Hebebia
}

func AddHebebia(info, baseDir string) {
	hebebiaMutex.Lock()
	Hebebia = append(Hebebia, info)
	hebebiaMutex.Unlock()
	SaveHebebia(baseDir)
}

func DeleteHebebia(index int, baseDir string) string {
	hebebiaMutex.Lock()
	defer hebebiaMutex.Unlock()
	if index >= 0 && index < len(Hebebia) {
		deleted := Hebebia[index]
		Hebebia = append(Hebebia[:index], Hebebia[index+1:]...)
		SaveHebebia(baseDir)
		return deleted
	}
	return ""
}

func GetAndIncrementZotCounter(baseDir string) int {
	zotMutex.Lock()
	defer zotMutex.Unlock()
	val := ZotCounter
	ZotCounter++
	os.WriteFile(baseDir+"/zot_counter.txt", []byte(fmt.Sprintf("%d", ZotCounter)), 0644)
	return val
}

var AllowedCommands = make(map[string]map[string]bool)
var allowedCmdMutex sync.RWMutex

func SetSyncState(allowed []string, cmds map[string]map[string]bool, aliases map[string]map[string]string, outputs map[string]map[string]string) {
	allowedMutex.Lock()
	AllowedUsers = make(map[string]bool)
	for _, u := range allowed {
		AllowedUsers[u] = true
	}
	allowedMutex.Unlock()

	allowedCmdMutex.Lock()
	AllowedCommands = cmds
	allowedCmdMutex.Unlock()

	aliasMutex.Lock()
	CommandAliases = aliases
	aliasMutex.Unlock()

	OutputMutex.Lock()
	CustomOutputs = outputs
	OutputMutex.Unlock()
}

func IsCommandAllowed(userID string, cmd string) bool {
	if IsAllowed(userID) {
		return true
	}
	allowedCmdMutex.RLock()
	defer allowedCmdMutex.RUnlock()
	if userCmds, ok := AllowedCommands[userID]; ok {
		return userCmds[cmd]
	}
	return false
}
