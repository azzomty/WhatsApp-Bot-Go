with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

old_logic = r"""	var vidUrl string
	for _, e := range res.Response.Episodes.Data {
		if e.EpisodeID == epID {
			for _, u := range e.EpisodeUrls {
				if u.ServerName == "muilt" {
					vidUrl = u.Url
					break
				}
			}
			break
		}
	}
	
	if vidUrl == "" {
		sendMessage(ctx, "عذرا، لم أجد سيرفر صالح لهذه الحلقة.")
		return
	}
	
	// Fetch muilt links
	req2, _ := http.NewRequest("GET", vidUrl, nil)
	reqHeaders(req2)
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		sendMessage(ctx, "فشل جلب سيرفرات المشاهدة.")
		return
	}
	defer resp2.Body.Close()
	
	var links []string
	json.NewDecoder(resp2.Body).Decode(&links)
	
	if len(links) == 0 {
		sendMessage(ctx, "لا توجد روابط لهذه الحلقة.")
		return
	}"""

new_logic = r"""	var links []string
	
	for _, e := range res.Response.Episodes.Data {
		if e.EpisodeID == epID {
			for _, u := range e.EpisodeUrls {
				if u.ServerName == "muilt" {
					// Fetch muilt links
					req2, _ := http.NewRequest("GET", u.Url, nil)
					reqHeaders(req2)
					resp2, err := http.DefaultClient.Do(req2)
					if err == nil {
						var mLinks []string
						json.NewDecoder(resp2.Body).Decode(&mLinks)
						links = append(links, mLinks...)
						resp2.Body.Close()
					}
				} else {
					// Add backup servers
					links = append(links, u.Url)
				}
			}
			break
		}
	}
	
	if len(links) == 0 {
		sendMessage(ctx, "لا توجد روابط لهذه الحلقة.")
		return
	}"""

if old_logic in content:
    content = content.replace(old_logic, new_logic)
    with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
        f.write(content)
    print("Patched All Servers!")
else:
    print("Could not find old logic")
