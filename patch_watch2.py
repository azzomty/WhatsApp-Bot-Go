import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

old_ep = """	if epID == "" {
		sendMessage(ctx, "لم يتم العثور على الحلقة.")
		return true
	}
	
	anslayerMutex.Lock()"""

new_ep = """	if epID == "" {
		sendMessage(ctx, "لم يتم العثور على الحلقة.")
		return true
	}
	
	if session.Mode == "watch" {
		sendMessage(ctx, "جاري جلب الحلقة وتحميلها...")
		go downloadAnslayerEpisode(ctx, session.SelectedAnime, epNumStr, epID)
		delete(ansSessions, ctx.Sender.User)
		return true
	}
	
	anslayerMutex.Lock()"""

content = content.replace(old_ep, new_ep)

# Add download function
add_func = """
func downloadAnslayerEpisode(ctx *BotContext, anime AnslayerAnime, epNum string, epID string) {
	// First get the episode details to find episode_urls
	u := fmt.Sprintf("https://anslayer.com/anime/public/anime/get-anime-details?anime_id=%d&fetch_episodes=Yes&more_info=No", anime.AnimeID)
	req, _ := http.NewRequest("GET", u, nil)
	reqHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		sendMessage(ctx, "خطأ في الاتصال.")
		return
	}
	
	var res struct {
		Response struct {
			Episodes struct {
				Data []struct {
					EpisodeID string `json:"episode_id"`
					EpisodeUrls []struct {
						ServerName string `json:"episode_server_name"`
						Url        string `json:"episode_url"`
					} `json:"episode_urls"`
				} `json:"data"`
			} `json:"episodes"`
		} `json:"response"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	resp.Body.Close()
	
	var vidUrl string
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
	}
	
	// Choose first fembed or ok.ru link, or just first link
	targetLink := links[0]
	
	// Download using yt-dlp via existing function DownloadM3U8WithQuality (works for any url yt-dlp supports)
	data, err := DownloadM3U8WithQuality(targetLink, "bestvideo[height<=720]+bestaudio/best[height<=720]")
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء التحميل: " + err.Error())
		return
	}
	sendVideoDataWithSplit(ctx, data, anime.AnimeName, epNum, false)
}
"""

content += add_func

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
    f.write(content)
