import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

new_monitor = """func monitorComments(epID string, stopCh chan struct{}) {
	epIDFloat, _ := strconv.ParseFloat(epID, 64)
	oldOffset := 30

	for {
		select {
		case <-stopCh:
			return
		default:
			anslayerMutex.Lock()
			msg := anslayerReplyMsg
			anslayerMutex.Unlock()

			if msg == "" {
				time.Sleep(10 * time.Second)
				continue
			}

			// Check newest first
			if replied := checkAndReplyBatch(epIDFloat, msg, 0, 30); replied {
				time.Sleep(65 * time.Second)
				continue
			}

			// If no new comments to reply to, check older comments
			if replied := checkAndReplyBatch(epIDFloat, msg, oldOffset, 30); replied {
				oldOffset += 1 // Next time we can fetch from the same offset or advance slightly. Since we replied to one, the list shifted. We'll just advance oldOffset when the whole page is exhausted.
				time.Sleep(65 * time.Second)
				continue
			}

			// If we didn't reply to anything in oldOffset, advance it
			oldOffset += 30
			time.Sleep(10 * time.Second)
		}
	}
}

func checkAndReplyBatch(epIDFloat float64, msg string, offset, limit int) bool {
	params := map[string]interface{}{
		"_order_by": "latest_first",
		"hide_irrelevant": "Yes",
		"episode_id": epIDFloat,
		"_limit": limit,
		"myfirst": "Yes",
		"_offset": offset,
	}
	b, _ := json.Marshal(params)
	u := "https://anslayer.com/anime/public/episode-comments/get-episode-comments?json=" + url.QueryEscape(string(b))
	
	req, _ := http.NewRequest("GET", u, nil)
	reqHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	var res struct {
		Response struct {
			Data []struct {
				CommentID string `json:"episode_comment_id"`
				UserID    string `json:"user_id"`
			} `json:"data"`
		} `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return false
	}
	
	if len(res.Response.Data) == 0 {
		return false // No comments in this batch
	}
	
	for _, c := range res.Response.Data {
		if c.UserID == "" || c.CommentID == "" { continue }
		if hasAnslayerUser(c.UserID) { continue }
		
		// Found one to reply to
		payload := url.Values{}
		payload.Set("episode_comment_id", c.CommentID)
		payload.Set("reply_text", msg)
		payload.Set("spoiler", "No")
		payload.Set("recipient_id", "")
		payload.Set("notification_type", "reply")
		
		reqR, _ := http.NewRequest("POST", "https://anslayer.com/anime/public/episode-comments/create-episode-comment-reply", strings.NewReader(payload.Encode()))
		reqHeaders(reqR)
		reqR.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		
		respR, errR := http.DefaultClient.Do(reqR)
		if errR == nil {
			respR.Body.Close()
			if respR.StatusCode == 200 {
				saveAnslayerUser(c.UserID)
				fmt.Println("Replied to user:", c.UserID)
				return true
			}
		}
	}
	return false
}"""

# Replace old monitorComments
content = re.sub(r'func monitorComments\(epID string, stopCh chan struct\{\}\) \{.*', new_monitor, content, flags=re.DOTALL)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
    f.write(content)
