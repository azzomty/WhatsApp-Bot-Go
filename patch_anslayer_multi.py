import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

# Add AnslayerAccount definitions and commands
accounts_def = """type AnslayerAccount struct {
	Email    string `json:"email"`
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
}

var (
	ansAccounts      []AnslayerAccount
	ansAccountsMutex sync.RWMutex
	ansAccountsFile  = "anslayer_accounts.json"
)

func loadAnslayerAccounts() {
	data, err := os.ReadFile(ansAccountsFile)
	if err == nil {
		json.Unmarshal(data, &ansAccounts)
	}
}

func saveAnslayerAccounts() {
	ansAccountsMutex.RLock()
	data, _ := json.MarshalIndent(ansAccounts, "", "  ")
	ansAccountsMutex.RUnlock()
	os.WriteFile(ansAccountsFile, data, 0644)
}

func getAnslayerAccount(idx int) (AnslayerAccount, bool) {
	ansAccountsMutex.RLock()
	defer ansAccountsMutex.RUnlock()
	if idx >= 0 && idx < len(ansAccounts) {
		return ansAccounts[idx], true
	}
	return AnslayerAccount{}, false
}
"""

if "type AnslayerAccount struct" not in content:
    content = content.replace("func loadAnslayerUsers() {", accounts_def + "\nfunc loadAnslayerUsers() {")
    content = content.replace("loadAnslayerUsers()", "loadAnslayerUsers()\n\tloadAnslayerAccounts()")

# Update reqHeaders
content = content.replace(
"""func reqHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+anslayerToken)
	req.Header.Set("Client-Id", anslayerClientID)
	req.Header.Set("Client-Secret", anslayerClientSec)
	req.Header.Set("User-Agent", "okhttp/3.12.13")
}""",
"""func reqHeaders(req *http.Request, token string) {
	req.Header.Set("Accept", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else {
		req.Header.Set("Authorization", "Bearer "+anslayerToken)
	}
	req.Header.Set("Client-Id", anslayerClientID)
	req.Header.Set("Client-Secret", anslayerClientSec)
	req.Header.Set("User-Agent", "okhttp/3.12.13")
}""")

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
    f.write(content)
print("Patched basic accounts logic")
