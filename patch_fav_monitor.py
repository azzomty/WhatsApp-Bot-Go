import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

fav_monitor = """
var favStopChan chan struct{}

func startFavMarketing(ctx *BotContext, msg string) {
	if favStopChan != nil {
		close(favStopChan) // Stop previous
	}
	favStopChan = make(chan struct{})
	
	// Fetch favorites
	u := "https://anslayer.com/anime/public/animes/get-published-animes?json=%7B%22_offset%22%3A0%2C%22_limit%22%3A100%2C%22_order_by%22%3A%22latest_first%22%2C%22list_type%22%3A%22favorites%22%2C%22just_info%22%3A%22Yes%22%2C%22user_id%22%3A9174886%7D"
	req, _ := http.NewRequest("GET", u, nil)
	reqHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		sendMessage(ctx, "فشل جلب المفضلة.")
		return
	}
	defer resp.Body.Close()
	
	var res struct {
		Response struct {
			Data []AnslayerAnime `json:"data"`
		} `json:"response"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	
	if len(res.Response.Data) == 0 {
		sendMessage(ctx, "قائمة المفضلة فارغة.")
		return
	}
	
	animeIDs := make([]string, 0)
	for _, a := range res.Response.Data {
		animeIDs = append(animeIDs, a.AnimeID)
	}
	
	sendMessage(ctx, fmt.Sprintf("✅ تم جلب %d أنمي من المفضلة، سيبدأ البوت الآن بنشر التعليقات عليها جميعاً بشكل دوري!", len(animeIDs)))
	
	go monitorFavComments(ctx, animeIDs, msg, favStopChan)
}

func checkAndReplyAnimeBatch(animeIDFloat float64, msg string, offset, limit int) (bool, bool) {
	params := map[string]interface{}{
		"_order_by": "latest_first",
		"hide_irrelevant": "Yes",
		"anime_id": animeIDFloat,
		"_limit": limit,
		"myfirst": "Yes",
		"_offset": offset,
	}
	b, _ := json.Marshal(params)
	u := "https://anslayer.com/anime/public/anime-comments/get-anime-comments?json=" + url.QueryEscape(string(b))
	
	req, _ := http.NewRequest("GET", u, nil)
	reqHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, false
	}
	defer resp.Body.Close()
	
	var res struct {
		Response struct {
			Data []struct {
				CommentID string `json:"anime_comment_id"`
				UserID    string `json:"user_id"`
			} `json:"data"`
		} `json:"response"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	
	if len(res.Response.Data) == 0 {
		return false, false
	}
	
	hasComments := true
	for _, c := range res.Response.Data {
		// Ensure we don't reply to ourselves (bot user id 9174886)
		if c.UserID == "9174886" {
			continue
		}
		
		// Check if we already replied to this comment globally
		anslayerMutex.Lock()
		_, repliedBefore := repliedComments[c.CommentID]
		anslayerMutex.Unlock()
		
		if !repliedBefore {
			cIDFloat, _ := strconv.ParseFloat(c.CommentID, 64)
			success := postAnslayerReply("anime", cIDFloat, msg)
			if success {
				anslayerMutex.Lock()
				repliedComments[c.CommentID] = true
				anslayerMutex.Unlock()
				fmt.Println("Replied to anime comment:", c.CommentID)
				return true, hasComments
			}
		}
	}
	return false, hasComments
}

func monitorFavComments(ctx *BotContext, animeIDs []string, msg string, stopCh chan struct{}) {
	oldOffsets := make(map[string]int)
	for _, id := range animeIDs {
		oldOffsets[id] = 10 // Start looking back if no new ones
	}
	
	finishedCount := 0

	for {
		select {
		case <-stopCh:
			return
		default:
			repliedInThisLoop := false
			
			for _, id := range animeIDs {
				idFloat, _ := strconv.ParseFloat(id, 64)
				
				// 1. Check Newest first (offset 0)
				replied, _ := checkAndReplyAnimeBatch(idFloat, msg, 0, 10)
				if replied {
					repliedInThisLoop = true
					time.Sleep(65 * time.Second) // wait 1 minute between comments globally to avoid spam
					break // break the anime loop, start from beginning next tick
				}
				
				// 2. If no new comment, check older comments
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
					} else {
						// Exhausted this anime's comments
					}
				}
			}
			
			if !repliedInThisLoop {
				time.Sleep(10 * time.Second) // Check again after 10s if absolutely nothing to reply to
			}
		}
	}
}
"""

if "func startFavMarketing" not in content:
    content += "\n" + fav_monitor

# Now we need to modify HandleAnslayerCommand to parse .انمي سلاير مفضلة نشر
old_marketing = """		if strings.HasPrefix(afterSlayer, "نشر") {
			replyMsg := strings.TrimSpace(strings.TrimPrefix(afterSlayer, "نشر"))
			if replyMsg == "" {
				sendMessage(ctx, "يرجى كتابة الرسالة بعد كلمة نشر.")
				return
			}
			anslayerMutex.Lock()
			anslayerReplyMsg = replyMsg
			anslayerMutex.Unlock()
			sendMessage(ctx, "✅ تم حفظ قالب الرد بنجاح!\\nالرسالة:\\n"+anslayerReplyMsg)
			return
		}"""

new_marketing = """		if strings.HasPrefix(afterSlayer, "مفضلة نشر") {
			replyMsg := strings.TrimSpace(strings.TrimPrefix(afterSlayer, "مفضلة نشر"))
			if replyMsg == "" {
				sendMessage(ctx, "يرجى كتابة الرسالة بعد كلمة مفضلة نشر.")
				return
			}
			startFavMarketing(ctx, replyMsg)
			return
		}

		if strings.HasPrefix(afterSlayer, "نشر") {
			replyMsg := strings.TrimSpace(strings.TrimPrefix(afterSlayer, "نشر"))
			if replyMsg == "" {
				sendMessage(ctx, "يرجى كتابة الرسالة بعد كلمة نشر.")
				return
			}
			anslayerMutex.Lock()
			anslayerReplyMsg = replyMsg
			anslayerMutex.Unlock()
			sendMessage(ctx, "✅ تم حفظ قالب الرد بنجاح!\\nالرسالة:\\n"+anslayerReplyMsg)
			return
		}"""

if "مفضلة نشر" not in content:
    content = content.replace(old_marketing, new_marketing)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
    f.write(content)
