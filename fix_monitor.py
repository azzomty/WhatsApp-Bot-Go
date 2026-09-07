import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

old_func1 = """func checkAndReplyBatch(epIDFloat float64, msg string, offset, limit int) bool {"""
new_func1 = """func checkAndReplyBatch(epIDFloat float64, msg string, offset, limit int) (bool, bool) {"""
content = content.replace(old_func1, new_func1)

old_return1 = """	if err != nil {
		return false
	}"""
new_return1 = """	if err != nil {
		return false, false
	}"""
content = content.replace(old_return1, new_return1)

old_return2 = """	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return false
	}"""
new_return2 = """	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return false, false
	}"""
content = content.replace(old_return2, new_return2)

old_return3 = """	if len(res.Response.Data) == 0 {
		return false // No comments in this batch
	}"""
new_return3 = """	if len(res.Response.Data) == 0 {
		return false, false // No comments in this batch
	}"""
content = content.replace(old_return3, new_return3)

old_return4 = """		if err == nil && resp2.StatusCode == 200 {
			saveAnslayerUser(c.UserID)
			return true
		}
	}
	
	return false
}"""
new_return4 = """		if err == nil && resp2.StatusCode == 200 {
			saveAnslayerUser(c.UserID)
			return true, true
		}
	}
	
	return false, true
}"""
content = content.replace(old_return4, new_return4)

old_monitor = """func monitorComments(epID string, stopCh chan struct{}) {
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
}"""

new_monitor = """func monitorComments(epID string, stopCh chan struct{}) {
	epIDFloat, _ := strconv.ParseFloat(epID, 64)
	oldOffset := 30
	finishedNotified := false

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
			replied, hasComments := checkAndReplyBatch(epIDFloat, msg, 0, 30)
			if replied {
				time.Sleep(65 * time.Second)
				continue
			}

			// If no new comments to reply to, check older comments
			replied, hasComments = checkAndReplyBatch(epIDFloat, msg, oldOffset, 30)
			if replied {
				oldOffset += 1 
				time.Sleep(65 * time.Second)
				continue
			}

			if !hasComments {
				if !finishedNotified {
					// Need a ctx to send message! But monitorComments doesn't have ctx!
					// Let's pass chatID and use a global sender?
					// Actually, monitorComments is passed stopCh but NO ctx.
					// I can just print it to logs, but the user wants a notification!
				}
				// reset to 0 to only monitor new ones
				oldOffset = 0
			} else {
				// advance to older
				oldOffset += 30
			}
			time.Sleep(10 * time.Second)
		}
	}
}"""
content = content.replace(old_monitor, new_monitor)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
    f.write(content)
