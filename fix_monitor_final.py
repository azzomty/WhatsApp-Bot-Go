import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

# Update go monitorComments call
content = content.replace("go monitorComments(epIDStr, ch)", "go monitorComments(ctx, epIDStr, ch)")

# checkAndReplyBatch signature
content = content.replace(
    "func checkAndReplyBatch(epIDFloat float64, msg string, offset, limit int) bool {",
    "func checkAndReplyBatch(epIDFloat float64, msg string, offset, limit int) (bool, bool) {"
)

# return false in checkAndReplyBatch (error handling)
content = content.replace(
"""	if err != nil {
		return false
	}""",
"""	if err != nil {
		return false, false
	}"""
)
content = content.replace(
"""	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return false
	}""",
"""	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return false, false
	}"""
)
content = content.replace(
"""	if len(res.Response.Data) == 0 {
		return false // No comments in this batch
	}""",
"""	if len(res.Response.Data) == 0 {
		return false, false // No comments in this batch
	}"""
)
content = content.replace(
"""				return true
			}
		}
	}
	return false
}""",
"""				return true, true
			}
		}
	}
	return false, true
}"""
)

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

new_monitor = """func monitorComments(ctx *BotContext, epID string, stopCh chan struct{}) {
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
					sendMessage(ctx, "✅ تم الرد على جميع التعليقات القديمة في هذه الحلقة! البوت الآن في وضع الاستعداد للتعليقات الجديدة فقط، يمكنك اختيار حلقة أخرى إذا أردت.")
					finishedNotified = true
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
