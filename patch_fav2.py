import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

# Replace the broken checkAndReplyAnimeBatch and monitorFavComments
broken = """func checkAndReplyAnimeBatch(animeIDFloat float64, msg string, offset, limit int) (bool, bool) {
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
}"""

fixed = """func checkAndReplyAnimeBatch(animeIDFloat float64, msg string, offset, limit int) (bool, bool) {
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
	
	for _, c := range res.Response.Data {
		if c.UserID == "" || c.CommentID == "" { continue }
		if hasAnslayerUser(c.UserID) { continue }
		
		payload := url.Values{}
		payload.Set("anime_comment_id", c.CommentID)
		payload.Set("reply_text", msg)
		payload.Set("spoiler", "No")
		payload.Set("recipient_id", "")
		payload.Set("notification_type", "reply")
		
		reqR, _ := http.NewRequest("POST", "https://anslayer.com/anime/public/anime-comments/create-anime-comment-reply", strings.NewReader(payload.Encode()))
		reqHeaders(reqR)
		reqR.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		
		respR, errR := http.DefaultClient.Do(reqR)
		if errR == nil {
			respR.Body.Close()
			if respR.StatusCode == 200 {
				ansUsersMutex.Lock()
				anslayerUsers[c.UserID] = true
				ansUsersMutex.Unlock()
				
				f, _ := os.OpenFile("anslayer_users.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				f.WriteString(c.UserID + "\\n")
				f.Close()
				
				fmt.Println("Replied to user:", c.UserID, "on anime comment:", c.CommentID)
				return true, true
			}
		}
	}
	return false, true
}

func monitorFavComments(ctx *BotContext, animeIDs []string, msg string, stopCh chan struct{}) {
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

if broken in content:
    content = content.replace(broken, fixed)
    with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
        f.write(content)
    print("Replaced!")
else:
    print("Could not find broken text")
