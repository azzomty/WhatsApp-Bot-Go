import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'r') as f:
    content = f.read()

old_episode = """func HandleStardimaEpisode(ctx *BotContext, number int) {
	stardimaMutex.Lock()
	defer stardimaMutex.Unlock()

	sender := ctx.Msg.Info.Sender.String()
	pending, ok := stardimaPending[sender]
	if !ok || len(pending.Episodes) == 0 {
		sendMessage(ctx, "يرجى اختيار المسلسل والموسم أولاً.")
		return
	}

	if number < 1 || number > len(pending.Episodes) {
		sendMessage(ctx, fmt.Sprintf("رقم الحلقة غير صحيح. يرجى اختيار رقم بين 1 و %d", len(pending.Episodes)))
		return
	}

	selected := pending.Episodes[number-1]
	pending.EpNum = fmt.Sprintf("%d", selected.EpisodeNumber)
	pending.WatchURL = selected.WatchURL
	pending.IsMovie = false
	stardimaPending[sender] = pending

	sendQualityMenu(ctx)
}"""

new_episode = """func HandleStardimaEpisode(ctx *BotContext, number int) {
	stardimaMutex.Lock()
	defer stardimaMutex.Unlock()

	sender := ctx.Msg.Info.Sender.String()
	pending, ok := stardimaPending[sender]
	if !ok || len(pending.Episodes) == 0 {
		sendMessage(ctx, "يرجى اختيار المسلسل والموسم أولاً.")
		return
	}

	if number < 1 || number > len(pending.Episodes) {
		sendMessage(ctx, fmt.Sprintf("رقم الحلقة غير صحيح. يرجى اختيار رقم بين 1 و %d", len(pending.Episodes)))
		return
	}

	selected := pending.Episodes[number-1]
	pending.EpNum = fmt.Sprintf("%d", selected.EpisodeNumber)
	pending.WatchURL = selected.WatchURL
	pending.IsMovie = false
	stardimaPending[sender] = pending

	sendMessage(ctx, "جاري جلب الحلقة وتحميلها...")
	go downloadStardimaEpisodeWithQuality(ctx, pending, "bestvideo[height<=480]+bestaudio/best[height<=480]", false)
}"""

content = content.replace(old_episode, new_episode)

old_movie = """			}
			pending.IsMovie = true
			stardimaPending[sender] = pending
			sendQualityMenu(ctx)
		}
	} else if action == "stardima_season" {"""

new_movie = """			}
			pending.IsMovie = true
			stardimaPending[sender] = pending
			sendMessage(ctx, "جاري جلب الفيلم وتحميله...")
			go downloadStardimaMovieWithQuality(ctx, selected, "bestvideo[height<=480]+bestaudio/best[height<=480]", false)
		}
	} else if action == "stardima_season" {"""

content = content.replace(old_movie, new_movie)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'w') as f:
    f.write(content)
