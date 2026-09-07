import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

# Replace the command parsing
old_cmd = """		if strings.HasPrefix(afterSlayer, "مفضلة نشر") {
			replyMsg := strings.TrimSpace(strings.TrimPrefix(afterSlayer, "مفضلة نشر"))
			if replyMsg == "" {
				sendMessage(ctx, "يرجى كتابة الرسالة بعد كلمة مفضلة نشر.")
				return
			}
			startFavMarketing(ctx, replyMsg)
			return
		}"""

new_cmd = """		if strings.HasPrefix(afterSlayer, "مفضلة") {
			anslayerMutex.Lock()
			msg := anslayerReplyMsg
			anslayerMutex.Unlock()
			if msg == "" {
				sendMessage(ctx, "يرجى أولاً حفظ قالب الرد باستخدام أمر:\\n.انمي سلاير نشر <رسالتك>")
				return
			}
			startFavMarketing(ctx, msg)
			return
		}"""

if old_cmd in content:
    content = content.replace(old_cmd, new_cmd)

# Replace startFavMarketing to use checkAndReplyBatch for latest episodes
old_monitor = """func monitorFavComments(ctx *BotContext, animeIDs []string, msg string, stopCh chan struct{}) {
	oldOffsets := make(map[string]int)
	for _, id := range animeIDs {
		oldOffsets[id] = 10
	}
	
	for {
		select {
		case <-stopCh:
			return
		default:
			repliedInThisLoop := false
			
			for _, id := range animeIDs {
				idFloat, _ := strconv.ParseFloat(id, 64)
				
				replied, _ := checkAndReplyAnimeBatch(idFloat, msg, 0, 10)
				if replied {
					repliedInThisLoop = true
					time.Sleep(65 * time.Second)
					break
				}
				
				offset := oldOffsets[id]
				replied, hasComments := checkAndReplyAnimeBatch(idFloat, msg, offset, 10)
				if replied {
					oldOffsets[id] += 1
					repliedInThisLoop = true
					time.Sleep(65 * time.Second)
					break
				} else {
					if hasComments {
						oldOffsets[id] += 10
					}
				}
			}
			
			if !repliedInThisLoop {
				time.Sleep(10 * time.Second)
			}
		}
	}
}"""

new_monitor = """func getLatestEpisodeID(animeID string) string {
	u := fmt.Sprintf("https://anslayer.com/anime/public/anime/get-anime-details?anime_id=%s&fetch_episodes=Yes&more_info=No", animeID)
	req, _ := http.NewRequest("GET", u, nil)
	reqHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	
	var res struct {
		Response struct {
			Episodes struct {
				Data []AnslayerEpisode `json:"data"`
			} `json:"episodes"`
		} `json:"response"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	
	if len(res.Response.Episodes.Data) > 0 {
		// Usually the first item is the newest episode if ordered by latest
		return res.Response.Episodes.Data[0].EpisodeID
	}
	return ""
}

func monitorFavComments(ctx *BotContext, animeIDs []string, msg string, stopCh chan struct{}) {
	oldOffsets := make(map[string]int)
	for _, id := range animeIDs {
		oldOffsets[id] = 30 // Start looking back from 30 for episode comments
	}
	
	for {
		select {
		case <-stopCh:
			return
		default:
			repliedInThisLoop := false
			
			for _, animeID := range animeIDs {
				// Fetch the latest episode ID for this anime dynamically
				latestEpID := getLatestEpisodeID(animeID)
				if latestEpID == "" {
					continue
				}
				
				epIDFloat, _ := strconv.ParseFloat(latestEpID, 64)
				
				// 1. Check Newest first (offset 0)
				replied, _ := checkAndReplyBatch(epIDFloat, msg, 0, 30) // Use existing episode batch func
				if replied {
					repliedInThisLoop = true
					time.Sleep(65 * time.Second)
					break
				}
				
				// 2. If no new comment, check older comments
				offset := oldOffsets[animeID]
				replied, hasComments := checkAndReplyBatch(epIDFloat, msg, offset, 30)
				if replied {
					oldOffsets[animeID] += 1
					repliedInThisLoop = true
					time.Sleep(65 * time.Second)
					break
				} else {
					if hasComments {
						oldOffsets[animeID] += 30
					}
				}
			}
			
			if !repliedInThisLoop {
				time.Sleep(10 * time.Second)
			}
		}
	}
}"""

if old_monitor in content:
    content = content.replace(old_monitor, new_monitor)
    with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
        f.write(content)
    print("Patched Fav Monitor!")
else:
    print("Could not find monitorFavComments")
