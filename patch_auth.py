with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    content = f.read()

import re

# Add track_authorization extraction
auth_re_str = r'''				authRe := regexp.MustCompile(`"track_authorization":"([^"]+)"`)
				auth := ""
				if m := authRe.FindStringSubmatch(bodyStr); len(m) > 1 {
					auth = m[1]
				}'''

content = content.replace('progUrl := ""', auth_re_str + '\n				progUrl := ""')

# Append it to the request
reqP_old = 'reqP, _ := http.NewRequest("GET", progUrl + "?client_id=" + clientID, nil)'
reqP_new = 'reqP, _ := http.NewRequest("GET", progUrl + "?client_id=" + clientID + "&track_authorization=" + auth, nil)'
content = content.replace(reqP_old, reqP_new)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(content)
print("Patched auth")
