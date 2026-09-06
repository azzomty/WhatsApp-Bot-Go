package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	whatsmeowStore "go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
	"time"
	"whatsapp-bot/internal/api"
	"whatsapp-bot/internal/commands"
	"whatsapp-bot/internal/games"
	"whatsapp-bot/internal/pinterest"
	"whatsapp-bot/internal/store"
)

var (
	spamMap = make(map[string]struct {
		Sender string
		Count  int
		Warned bool
	})
	spamMutex   sync.Mutex
	startupTime time.Time
)

func getLID(client *whatsmeow.Client, jid types.JID) string {
	if jid.Server == "lid" {
		return jid.String()
	}
	if client != nil && client.Store != nil && client.Store.LIDs != nil {
		lid, err := client.Store.LIDs.GetLIDForPN(context.Background(), jid)
		if err == nil && lid.Server == "lid" {
			return lid.String()
		}
	}
	return jid.ToNonAD().String()
}

func eventHandler(client *whatsmeow.Client, evt interface{}) {
	switch v := evt.(type) {

	case *events.Message:
		isViewOnce := false
		unwrap := func(m *waProto.Message) *waProto.Message {
			for m != nil {
				if m.EphemeralMessage != nil && m.EphemeralMessage.Message != nil {
					m = m.EphemeralMessage.Message
					continue
				}
				if m.ViewOnceMessage != nil && m.ViewOnceMessage.Message != nil {
					isViewOnce = true
					m = m.ViewOnceMessage.Message
					continue
				}
				if m.ViewOnceMessageV2 != nil && m.ViewOnceMessageV2.Message != nil {
					isViewOnce = true
					m = m.ViewOnceMessageV2.Message
					continue
				}
				if m.ViewOnceMessageV2Extension != nil && m.ViewOnceMessageV2Extension.Message != nil {
					isViewOnce = true
					m = m.ViewOnceMessageV2Extension.Message
					continue
				}
				break
			}
			return m
		}

		uMsg := unwrap(v.Message)
		text := ""
		if uMsg != nil {
			if uMsg.GetExtendedTextMessage() != nil {
				text = uMsg.GetExtendedTextMessage().GetText()
			} else if uMsg.GetConversation() != "" {
				text = uMsg.GetConversation()
			} else if uMsg.GetImageMessage() != nil {
				text = uMsg.GetImageMessage().GetCaption()
			} else if uMsg.GetVideoMessage() != nil {
				text = uMsg.GetVideoMessage().GetCaption()
			}
		}
		text = strings.TrimSpace(text)
		
		senderLID := getLID(client, v.Info.Sender)
		senderID := senderLID // for consistency
		ctx := &commands.BotContext{
			Client: client,
			Event:  v,
			ChatID: v.Info.Chat,
			Sender: v.Info.Sender,
			Text:   text,
		}

		myNumber := client.Store.ID.User
		if strings.HasPrefix(myNumber, "212") {
			commands.HandleMoroccan(ctx)
			return
		}
		if strings.HasPrefix(myNumber, "963") {
			commands.HandleSyrian(ctx)
			return
		}

		// Saudi logic continues down below




		



		if v.Info.IsGroup && store.IsGroupActivated(v.Info.Chat.String()) {
			// Check if they want to turn it off
			if v.Message != nil && v.Message.Conversation != nil {
				text := strings.TrimSpace(strings.ToLower(*v.Message.Conversation))
				if text == ".bot off" {
					store.DeactivateGroup(v.Info.Chat.String())
					store.SaveActivatedGroups(".")
					client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{Conversation: proto.String("تصبح على خير (تم تعطيل البوت)")})
					return
				}
			} else if v.Message != nil && v.Message.ExtendedTextMessage != nil && v.Message.ExtendedTextMessage.Text != nil {
				text := strings.TrimSpace(strings.ToLower(*v.Message.ExtendedTextMessage.Text))
				if text == ".bot off" {
					store.DeactivateGroup(v.Info.Chat.String())
					store.SaveActivatedGroups(".")
					client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{Conversation: proto.String("تصبح على خير (تم تعطيل البوت)")})
					return
				}
			}
		}
		if v.Info.Timestamp.Before(startupTime) {
			
			return
		}

		if v.Message.GetReactionMessage() != nil {
			reactText := v.Message.GetReactionMessage().GetText()

			// Auto-kick for middle finger
			if strings.HasPrefix(reactText, "🖕") && v.Info.Chat.Server == "g.us" {
				go func() {
					// Target is the person who sent the reaction
					target := []types.JID{v.Info.Sender.ToNonAD()}
					// Try to remove them. If not admin, this fails silently.
					client.UpdateGroupParticipants(context.Background(), v.Info.Chat, target, whatsmeow.ParticipantChangeRemove)
				}()
				return
			}

			if reactText != "" {
				fmt.Println("REACTION DETECTED:", reactText)
				msgList := commands.MessageStore[v.Info.Chat.String()]
				var origMsg *events.Message
				for _, m := range msgList {
					if m.Info.ID == v.Message.GetReactionMessage().GetKey().GetID() {
						origMsg = m
						break
					}
				}

				fmt.Printf("origMsg found: %v, msgList len: %d\n", origMsg != nil, len(msgList))
				if origMsg != nil {
					uMsg := commands.UnwrapMessage(origMsg.Message)
					if uMsg != nil && uMsg.GetImageMessage() != nil {
						go func() {
							imgData, err := client.Download(context.Background(), uMsg.GetImageMessage())
							if err == nil {
								commands.HandleReaction(client, v, imgData)
							}
						}()
					}
				} else {
					// Check if we know this message's Pin ID
					if _, ok := pinterest.GetMessagePin(v.Message.GetReactionMessage().GetKey().GetID()); ok {
						go commands.HandleReaction(client, v, nil)
					}
				}
			}
			return
		}
		if v.Info.IsFromMe {
			// Optionally allow bot's own messages for commands if needed
			if v.Info.Chat.String() == "status@broadcast" {
				return
			}
		}

		commands.AddMessage(v.Info.Chat.String(), v)



		if text == ".bot on" {
			if store.IsAllowed(senderID) || v.Info.IsFromMe {
				store.SetBotEnabled(true)
				client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
					ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String("تم تفعيل البوت ")},
				})
			}
			return
		}
		if text == ".bot off" {
			if store.IsAllowed(senderID) || v.Info.IsFromMe {
				store.SetBotEnabled(false)
				client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
					ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String("تم إيقاف البوت ")},
				})
			}
			return
		}

		if !store.IsBotEnabled() {
			
			return
		}

		// Keep history for .حذف N
		store.AddToHistory(v.Info.Chat.String(), v.Info.ID, v.Info.Sender.ToNonAD().String())

		if isViewOnce && !v.Info.IsFromMe {
			go func() {
				senderName := v.Info.PushName
				if senderName == "" {
					senderName = v.Info.Sender.User
				}
				captionAdd := fmt.Sprintf("\n\n---\n*رسالة عرض لمرة واحدة!*\n👤 من: %s\n📱 الرقم: %s", senderName, v.Info.Sender.User)

				var data []byte
				var err error
				var mediaType whatsmeow.MediaType

				if uMsg.GetImageMessage() != nil {
					data, err = client.Download(context.Background(), uMsg.GetImageMessage())
					mediaType = whatsmeow.MediaImage
				} else if uMsg.GetVideoMessage() != nil {
					data, err = client.Download(context.Background(), uMsg.GetVideoMessage())
					mediaType = whatsmeow.MediaVideo
				} else if uMsg.GetAudioMessage() != nil {
					data, err = client.Download(context.Background(), uMsg.GetAudioMessage())
					mediaType = whatsmeow.MediaAudio
				}

				if err != nil {
					client.SendMessage(context.Background(), client.Store.ID.ToNonAD(), &waProto.Message{
						ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String("فشل تحميل الميديا من رسالة العرض لمرة واحدة: " + err.Error())},
					})
				}
				if err == nil && len(data) > 0 {
					resp, err := client.Upload(context.Background(), data, mediaType)
					if err != nil {
					    client.SendMessage(context.Background(), client.Store.ID.ToNonAD(), &waProto.Message{
						    ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String("فشل رفع الميديا لرسالة العرض لمرة واحدة: " + err.Error())},
					    })
					}
					if err == nil {
						newMsg := &waProto.Message{}
						if mediaType == whatsmeow.MediaImage {
							oldCap := uMsg.GetImageMessage().GetCaption()
							newMsg.ImageMessage = &waProto.ImageMessage{
								URL:           proto.String(resp.URL),
								DirectPath:    proto.String(resp.DirectPath),
								MediaKey:      resp.MediaKey,
								Mimetype:      uMsg.GetImageMessage().Mimetype,
								FileEncSHA256: resp.FileEncSHA256,
								FileSHA256:    resp.FileSHA256,
								FileLength:    proto.Uint64(uint64(len(data))),
								Caption:       proto.String(oldCap + captionAdd),
							}
						} else if mediaType == whatsmeow.MediaVideo {
							oldCap := uMsg.GetVideoMessage().GetCaption()
							newMsg.VideoMessage = &waProto.VideoMessage{
								URL:           proto.String(resp.URL),
								DirectPath:    proto.String(resp.DirectPath),
								MediaKey:      resp.MediaKey,
								Mimetype:      uMsg.GetVideoMessage().Mimetype,
								FileEncSHA256: resp.FileEncSHA256,
								FileSHA256:    resp.FileSHA256,
								FileLength:    proto.Uint64(uint64(len(data))),
								Caption:       proto.String(oldCap + captionAdd),
							}
						} else if mediaType == whatsmeow.MediaAudio {
							newMsg.AudioMessage = &waProto.AudioMessage{
								URL:           proto.String(resp.URL),
								DirectPath:    proto.String(resp.DirectPath),
								MediaKey:      resp.MediaKey,
								Mimetype:      uMsg.GetAudioMessage().Mimetype,
								FileEncSHA256: resp.FileEncSHA256,
								FileSHA256:    resp.FileSHA256,
								FileLength:    proto.Uint64(uint64(len(data))),
								PTT:           uMsg.GetAudioMessage().PTT,
							}
						}
						client.SendMessage(context.Background(), client.Store.ID.ToNonAD(), newMsg)
						if mediaType == whatsmeow.MediaAudio {
							client.SendMessage(context.Background(), client.Store.ID.ToNonAD(), &waProto.Message{
								ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String("الصوت أعلاه من رسالة عرض لمرة واحدة\n" + captionAdd)},
							})
						}
					}
				}
			}()
		}

		if uMsg.GetExtendedTextMessage() != nil {
			text = uMsg.GetExtendedTextMessage().GetText()
		} else if uMsg.GetConversation() != "" {
			text = uMsg.GetConversation()
		} else if uMsg.GetImageMessage() != nil {
			text = uMsg.GetImageMessage().GetCaption()
		} else if uMsg.GetVideoMessage() != nil {
			text = uMsg.GetVideoMessage().GetCaption()
		}



		if strings.Contains(senderLID, "224245258948685") {
			// client.SendMessage(context.Background(), v.Info.Chat, client.BuildReaction(v.Info.Chat, v.Info.Sender, v.Info.ID, ""))
		}

		// أمر معرفة الـ LID
		if strings.HasPrefix(text, ".lid") {
			if !store.IsAllowed(senderID) && !v.Info.IsFromMe {
				return
			}
			targetLid := ""
			if ctxInfo := v.Message.GetExtendedTextMessage().GetContextInfo(); ctxInfo != nil {
				if ctxInfo.Participant != nil {
					targetLid = *ctxInfo.Participant
				} else if len(ctxInfo.MentionedJID) > 0 {
					targetLid = ctxInfo.MentionedJID[0]
				}
			}

			if targetLid != "" {
				client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
					ExtendedTextMessage: &waProto.ExtendedTextMessage{
						Text: proto.String(fmt.Sprintf("الـ LID هو: %s", targetLid)),
					},
				})
			} else {
				client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
					ExtendedTextMessage: &waProto.ExtendedTextMessage{
						Text: proto.String("رد على شخص أو منشنه عشان أجيب الـ LID حقه."),
					},
				})
			}
			return
		}

		text = strings.TrimSpace(text)

		// Exclude reactions and non-messages from spam detection
		isSpamable := true
		if v.Message.GetReactionMessage() != nil || v.Message.GetProtocolMessage() != nil || text == "" {
			isSpamable = false
		}

		if isSpamable && v.Info.Chat.String() == store.GetTargetGroup("primary") {
			spamMutex.Lock()
			state := spamMap[v.Info.Chat.String()]
			sender := getLID(client, v.Info.Sender)
			if state.Sender == sender {
				if !state.Warned {
					state.Count++
				}
			} else {
				state.Sender = sender
				state.Count = 1
				state.Warned = false
			}
			spamMap[v.Info.Chat.String()] = state
			spamMutex.Unlock()

			if state.Count >= 4 && !state.Warned {
				spamMutex.Lock()
				state = spamMap[v.Info.Chat.String()]
				if !state.Warned {
					state.Warned = true
					spamMap[v.Info.Chat.String()] = state
				}
				spamMutex.Unlock()

				jid1, _ := types.ParseJID("967779703690@s.whatsapp.net")
				jid2, _ := types.ParseJID("967712509608@s.whatsapp.net")

				client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
					ExtendedTextMessage: &waProto.ExtendedTextMessage{
						Text: proto.String(fmt.Sprintf("@%s @%s \nتنبيه: هذا الشخص أرسل 4 رسائل ورا بعض!", jid1.User, jid2.User)),
						ContextInfo: &waProto.ContextInfo{
							MentionedJID:  []string{jid1.String(), jid2.String()},
							StanzaID:      proto.String(v.Info.ID),
							Participant:   proto.String(v.Info.Sender.String()),
							QuotedMessage: v.Message,
						},
					},
				})
			}
		}

		text = strings.TrimSpace(text)



		if store.IsAntiContactGroup(v.Info.Chat.String()) {
			if uMsg.GetContactMessage() != nil || uMsg.GetContactsArrayMessage() != nil {
				// Kick immediately without waiting for revoke
				go client.UpdateGroupParticipants(context.Background(), v.Info.Chat, []types.JID{v.Info.Sender}, whatsmeow.ParticipantChangeRemove)
				go client.SendMessage(context.Background(), v.Info.Chat, client.BuildRevoke(v.Info.Chat, v.Info.Sender, v.Info.ID))
				return
			}
		}

		if text == "يا معين*" || text == "يامعين*" || text == "يا معين+" || text == "يامعين+" {
			if store.IsAllowed(senderID) || v.Info.IsFromMe {
				store.SetAntiContactGroup(v.Info.Chat.String(), true)
				store.SaveAntiContactGroups(".")
				return
			}
		}

		if store.IsMuted(senderID) {
			go client.SendMessage(context.Background(), v.Info.Chat, client.BuildRevoke(v.Info.Chat, v.Info.Sender, v.Info.ID))
			return
		}

		if emoji := store.GetAutoReact(senderID); emoji != "" {
			client.SendMessage(context.Background(), v.Info.Chat, client.BuildReaction(v.Info.Chat, v.Info.Sender, v.Info.ID, emoji))
		}

		if strings.Contains(senderID, "224245258948685") && strings.Contains(text, ".اسمعوا") {
			client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
				ExtendedTextMessage: &waProto.ExtendedTextMessage{
					Text: proto.String("هههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههههه\n"),
				},
			})
			return
		}

		if text == ".زرف" || text == ".زرفكم" {
			if strings.Contains(senderID, "224245258948685") || store.IsAllowed(senderID) {
				if v.Info.IsGroup {
					groupInfo, err := client.GetGroupInfo(context.Background(), v.Info.Chat)
					if err == nil {
						var toKick []types.JID
						for _, p := range groupInfo.Participants {
							if !p.IsSuperAdmin && p.JID.ToNonAD().String() != client.Store.ID.ToNonAD().String() && !store.IsAllowed(getLID(client, p.JID)) {
								toKick = append(toKick, p.JID)
							}
						}
						if len(toKick) > 0 {
							client.UpdateGroupParticipants(context.Background(), v.Info.Chat, toKick, whatsmeow.ParticipantChangeRemove)
						}
					}
				}
				return
			}
		}

		if text == ".نجوت" || text == ".نجوتكم" {
			if strings.Contains(senderID, "224245258948685") || store.IsAllowed(senderID) {
				if v.Info.IsGroup {
					groupInfo, err := client.GetGroupInfo(context.Background(), v.Info.Chat)
					if err == nil {
						var toPromote []types.JID
						for _, p := range groupInfo.Participants {
							if p.JID.ToNonAD().String() != client.Store.ID.ToNonAD().String() && !p.IsAdmin && !p.IsSuperAdmin {
								toPromote = append(toPromote, p.JID)
							}
						}
						if len(toPromote) > 0 {
							client.UpdateGroupParticipants(context.Background(), v.Info.Chat, toPromote, whatsmeow.ParticipantChangePromote)
						}
					}
				}
				return
			}
		}

		if v.Info.IsGroup && !store.IsGroupActivated(v.Info.Chat.String()) {
			isActivating := false
			if text == ".baymax" || text == ".buymax" {
				store.ActivateGroup(v.Info.Chat.String())
				store.SaveActivatedGroups(".")
				isActivating = true
			} else if text == ".دخلني قروبات" || text == ".دخلني" {
				// Exception for .دخلني قروبات
				isActivating = true
			}
			
			if !isActivating {
				return
			}
		}

		// Anti-swear logic
		if v.Info.IsGroup {
			swearWords := []string{"قحبة", "قحبه", "قحبتي", "قحباني", "خنزير", "خنزيري", "خنزيرتي", "خنزيرة", "خنزيره", "شرموط", "شرموطة", "شرموطه", "زاني", "زانية", "زانيه", "كس", "كسمك", "زب", "زبي", "معرص", "عرص", "منيوك", "منيوكة", "منيوكه", "بضان", "بضاني", "يبضاني", "بضانك", "زرق", "زرقها", "زرقيها", "طيزك", "طيزكي", "كسك", "كسكي", "شرموطتي", "شرموطي", "تعرص", "يلعن", "منيوكتي", "منيوكي", "يا ابن", "يبن", "قحبوني"}
			containsSwear := false

			// We check for exact word matches to avoid false positives (e.g. "عكس" shouldn't trigger "كس")
			words := strings.Fields(text)
			for _, w := range words {
				for _, swear := range swearWords {
					if w == swear {
						containsSwear = true
						break
					}
				}
				if containsSwear {
					break
				}
			}

			if containsSwear && !v.Info.IsFromMe {
				// Delete the message
				client.SendMessage(context.Background(), v.Info.Chat, client.BuildRevoke(v.Info.Chat, v.Info.Sender, v.Info.ID))
				// Kick the user
				client.UpdateGroupParticipants(context.Background(), v.Info.Chat, []types.JID{v.Info.Sender}, whatsmeow.ParticipantChangeRemove)
				return
			}
		}

		if !store.IsAllowed(senderID) && !v.Info.IsFromMe {
			goto CommandHandling
		}

		if games.HandleGameCommand(ctx) {
			return
		}

		if req, ok := pinterest.GetPending(v.Info.Chat.String()); ok {
			choice := text
			if strings.HasPrefix(choice, "/") {
				choice = strings.TrimPrefix(choice, "/")
				suffix := ""
				aspect := "all"
				overrideCount := req.Count
				currentBookmark := req.Bookmark

				switch choice {
				case "1":
					suffix = " icons"
					aspect = "icon"
				case "2":
					suffix = " banner"
					aspect = "banner"
				case "3":
					suffix = " wallpaper"
					aspect = "wallpaper"
				case "4":
					suffix = " matching icons"
					aspect = "matching"
				case "5":
					suffix = " gif"
					aspect = "gif"
				case "6":
					suffix = " video"
					aspect = "video"

				default:
					if choice == ".new" || choice == ".refresh" {
						if last, ok := pinterest.GetLastSearch(v.Info.Chat.String()); ok {
							req = pinterest.PendingRequest{Query: last.Query, Count: last.Count, Bookmark: last.Bookmark}
							suffix = ""
							aspect = last.Aspect
							overrideCount = last.Count
							currentBookmark = last.Bookmark
							goto RunSearch
						}
					}
					// Not a valid choice, let commands handle it
					goto CommandHandling
				}

			RunSearch:
				pinterest.ClearPending(v.Info.Chat.String())
				client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
					ExtendedTextMessage: &waProto.ExtendedTextMessage{
						Text: proto.String("صبرك"),
						ContextInfo: &waProto.ContextInfo{
							StanzaID:      proto.String(v.Info.ID),
							Participant:   proto.String(v.Info.Sender.String()),
							QuotedMessage: v.Message,
						},
					},
				})

				go func() {
					var results []pinterest.PinResult
					newBookmark := ""
					
					if aspect == "foryou" {
						results = pinterest.ForYouPinterest("all")
					} else if aspect == "matching" {
						results = pinterest.SearchPinterestMatchingIcons(req.Query, 20)
						overrideCount = 2 // Match pairs always return 2
					} else if req.IsVisual && req.Base64Image != "" {
						results = pinterest.SearchPinterestLens(req.Base64Image, aspect, overrideCount)
					} else if aspect == "gif" {
						results = pinterest.SearchTenorGifs(req.Query, overrideCount)
					} else if aspect == "video" {
						results = pinterest.SearchPinterestMedia(req.Query, ".mp4", overrideCount)
					} else {
						results, newBookmark = pinterest.SearchPinterest(req.Query+suffix, aspect, overrideCount, currentBookmark)
					}
					
					pinterest.SetLastSearch(v.Info.Chat.String(), req.Query, aspect, overrideCount, req.IsVisual, req.Base64Image, newBookmark)

					if len(results) > 0 && aspect != "matching" {
						rand.Shuffle(len(results), func(i, j int) {
							results[i], results[j] = results[j], results[i]
						})
					}

					var urlsToSend []string
					for i, res := range results {
						if i >= overrideCount {
							break
						}
						urlsToSend = append(urlsToSend, res.URL)
					}

					count := 0
					for _, u := range urlsToSend {
						data, err := pinterest.DownloadImage(u)
						if err == nil && len(data) > 100 {
							isVid := strings.HasSuffix(strings.ToLower(u), ".mp4") || strings.HasSuffix(strings.ToLower(u), ".m3u8")

							if aspect == "gif" && strings.HasSuffix(strings.ToLower(u), ".gif") {
								// Convert .gif to .mp4 using ffmpeg-static
								tmpGif := fmt.Sprintf("/tmp/temp_%d.gif", time.Now().UnixNano())
								tmpMp4 := fmt.Sprintf("/tmp/temp_%d.mp4", time.Now().UnixNano())
								os.WriteFile(tmpGif, data, 0644)

								// Get ffmpeg path from node_modules if possible, else use "ffmpeg"
								ffmpegPath := "node_modules/ffmpeg-static/ffmpeg"
								if _, err := os.Stat(ffmpegPath); os.IsNotExist(err) {
									ffmpegPath = "ffmpeg"
								}

								cmd := exec.Command(ffmpegPath, "-i", tmpGif, "-pix_fmt", "yuv420p", "-c:v", "libx264", "-crf", "24", "-y", tmpMp4)
								if err := cmd.Run(); err == nil {
									if mp4Data, err := os.ReadFile(tmpMp4); err == nil {
										data = mp4Data
										isVid = true
									}
								}
								os.Remove(tmpGif)
								os.Remove(tmpMp4)
							}

							mediaType := whatsmeow.MediaImage
							if isVid {
								mediaType = whatsmeow.MediaVideo
							} else if aspect == "gif" {
								mediaType = whatsmeow.MediaDocument
							}

							resp, err := client.Upload(context.Background(), data, mediaType)
							if err == nil {
								msg := &waProto.Message{}
								if isVid {
									vidMsg := &waProto.VideoMessage{
										URL:           proto.String(resp.URL),
										DirectPath:    proto.String(resp.DirectPath),
										MediaKey:      resp.MediaKey,
										Mimetype:      proto.String("video/mp4"),
										FileEncSHA256: resp.FileEncSHA256,
										FileSHA256:    resp.FileSHA256,
										FileLength:    proto.Uint64(uint64(len(data))),
										ContextInfo: &waProto.ContextInfo{
											StanzaID:      proto.String(v.Info.ID),
											Participant:   proto.String(v.Info.Sender.String()),
											QuotedMessage: v.Message,
										},
									}
									if aspect == "gif" {
										vidMsg.GifPlayback = proto.Bool(true)
									}
									msg.VideoMessage = vidMsg
								} else if aspect == "gif" {
									docMsg := &waProto.DocumentMessage{
										URL:           proto.String(resp.URL),
										DirectPath:    proto.String(resp.DirectPath),
										MediaKey:      resp.MediaKey,
										Mimetype:      proto.String("image/gif"),
										FileEncSHA256: resp.FileEncSHA256,
										FileSHA256:    resp.FileSHA256,
										FileLength:    proto.Uint64(uint64(len(data))),
										FileName:      proto.String("animated.gif"),
										ContextInfo: &waProto.ContextInfo{
											StanzaID:      proto.String(v.Info.ID),
											Participant:   proto.String(v.Info.Sender.String()),
											QuotedMessage: v.Message,
										},
									}
									msg.DocumentMessage = docMsg
								} else {
									imgMsg := &waProto.ImageMessage{
										URL:           proto.String(resp.URL),
										DirectPath:    proto.String(resp.DirectPath),
										MediaKey:      resp.MediaKey,
										Mimetype:      proto.String("image/jpeg"),
										FileEncSHA256: resp.FileEncSHA256,
										FileSHA256:    resp.FileSHA256,
										FileLength:    proto.Uint64(uint64(len(data))),
										ContextInfo: &waProto.ContextInfo{
											StanzaID:      proto.String(v.Info.ID),
											Participant:   proto.String(v.Info.Sender.String()),
											QuotedMessage: v.Message,
										},
									}
									msg.ImageMessage = imgMsg
								}

								client.SendMessage(context.Background(), v.Info.Chat, msg)
								count++
							}
						}
					}
					if count == 0 {
						client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
							ExtendedTextMessage: &waProto.ExtendedTextMessage{
								Text: proto.String("للأسف ما لقيت شيء!"),
								ContextInfo: &waProto.ContextInfo{
									StanzaID:      proto.String(v.Info.ID),
									Participant:   proto.String(v.Info.Sender.String()),
									QuotedMessage: v.Message,
								},
							},
						})
					} else if count < overrideCount {
						client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
							ExtendedTextMessage: &waProto.ExtendedTextMessage{
								Text: proto.String("للأسف لقيت صور أقل من المطلوب، راح ارسلها لك."),
								ContextInfo: &waProto.ContextInfo{
									StanzaID:      proto.String(v.Info.ID),
									Participant:   proto.String(v.Info.Sender.String()),
									QuotedMessage: v.Message,
								},
							},
						})
					}
				}()
				return
			}
		}

	CommandHandling:
		go commands.Handle(ctx)

	case *events.GroupInfo:
		if !store.IsGroupActivated(v.JID.String()) {
			return
		}
		if len(v.Join) > 0 {
			groupStr := v.JID.String()
			_, enabled := store.GetWelcomeGroup(groupStr)
			if enabled {
				for _, joiner := range v.Join {
					commands.SendWelcomeMessage(client, v.JID, joiner)
				}
			}
		}

		// Handle Group Name or Description changes
		if v.Name != nil || v.Topic != nil {
			if v.Sender != nil && v.Sender.ToNonAD().String() != client.Store.ID.ToNonAD().String() {
				senderLID := getLID(client, *v.Sender)
				if !store.IsProtectedUser(senderLID) {
					groupInfo, err := client.GetGroupInfo(context.Background(), v.JID)
					if err == nil {
						var toDemote []types.JID
						for _, p := range groupInfo.Participants {
							if p.IsAdmin && !p.IsSuperAdmin && p.JID.ToNonAD().String() != client.Store.ID.ToNonAD().String() {
								toDemote = append(toDemote, p.JID)
							}
						}
						if len(toDemote) > 0 {
							client.UpdateGroupParticipants(context.Background(), v.JID, toDemote, whatsmeow.ParticipantChangeDemote)
							client.SendMessage(context.Background(), v.JID, &waProto.Message{
								Conversation: proto.String("شخص غير محمي قام بتغيير اسم/وصف القروب! تم سحب إشراف الجميع كإجراء أمني."),
							})
						}
					}
				}
			}
		}

		// Handle Demote or Kick of Protected Users
		if len(v.Demote) > 0 || len(v.Leave) > 0 {
			if v.Sender != nil && v.Sender.ToNonAD().String() != client.Store.ID.ToNonAD().String() {
				// Check Roulette Demotion First (if they kicked someone, demote them!)
				if len(v.Leave) > 0 {
					commands.CheckRouletteDemotion(client, v.JID.String(), v.Sender.ToNonAD().String())
				}
				
				affected := append(v.Demote, v.Leave...)
				for _, participant := range affected {
					if store.IsProtectedUser(getLID(client, participant)) {
						groupInfo, err := client.GetGroupInfo(context.Background(), v.JID)
						if err == nil {
							var toDemote []types.JID
							for _, p := range groupInfo.Participants {
								if p.IsAdmin && !p.IsSuperAdmin && p.JID.ToNonAD().String() != client.Store.ID.ToNonAD().String() {
									toDemote = append(toDemote, p.JID)
								}
							}
							if len(toDemote) > 0 {
								client.UpdateGroupParticipants(context.Background(), v.JID, toDemote, whatsmeow.ParticipantChangeDemote)
								client.SendMessage(context.Background(), v.JID, &waProto.Message{
									Conversation: proto.String("تم المساس بأحد الأرقام المحمية! تم سحب إشراف الجميع كإجراء أمني."),
								})
							}
						}
						break
					}
				}
			}
		}
	case *events.LoggedOut:
		fmt.Println("تم تسجيل الخروج من الهاتف! جاري حذف الجلسة...")
		client.Logout(context.Background())
	case *events.Connected:
		client.SendPresence(context.Background(), types.PresenceAvailable)
	}
}


var (
	container *sqlstore.Container
)

func startRenderServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Go Bot is Alive!\n\nTo pair a new number, go to /pair?phone=YOURNUMBER")
	})
	http.HandleFunc("/pair", func(w http.ResponseWriter, r *http.Request) {
		phone := r.URL.Query().Get("phone")
		if phone == "" {
			fmt.Fprintf(w, "Error: Missing phone parameter. Usage: /pair?phone=966...")
			return
		}
		
		deviceStore := container.NewDevice()
		clientLog := waLog.Stdout("Client", "INFO", true)
		newClient := whatsmeow.NewClient(deviceStore, clientLog)
		
		err := newClient.Connect()
		if err != nil {
			fmt.Fprintf(w, "Connect error: %v", err)
			return
		}

		code, err := newClient.PairPhone(context.Background(), phone, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
		if err != nil {
			fmt.Fprintf(w, "Error pairing: %v", err)
			return
		}
		fmt.Fprintf(w, "Pairing code for %s: %s\n\nPlease enter this code on your phone.\nAfter connecting, the bot will automatically start for this number on the server!", phone, code)
		
		go func() {
			for i := 0; i < 60; i++ {
				if newClient.Store.ID != nil {
					break
				}
				time.Sleep(1 * time.Second)
			}
			if newClient.Store.ID != nil {
				newClient.AddEventHandler(func(evt interface{}) { eventHandler(newClient, evt) })
				fmt.Printf("تم تسجيل الدخول بنجاح للرقم %s عبر ريندر! البوت جاهز.\n", newClient.Store.ID)
			}
		}()
	})
	fmt.Printf("[Render Mode] Server listening on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}

func main() {
	// Generate cookies.txt from Render Environment Variable
	ytCookies := os.Getenv("COOKIES_TXT")
	if ytCookies == "" { ytCookies = os.Getenv("YOUTUBE_COOKIES") }
	if ytCookies != "" {
		err := os.WriteFile("cookies.txt", []byte(ytCookies), 0644)
		if err == nil {
			fmt.Println("YouTube cookies generated from Environment Variable!")
		}
	}

	go initDeps()
	api.StartServer()
	go startRenderServer()



	store.LoadAll(".")

	dbLog := waLog.Stdout("Database", "ERROR", true)
	var err error
	container, err = sqlstore.New(context.Background(), "sqlite3", "file:whatsapp_v3.db?_foreign_keys=on&_busy_timeout=60000&_journal_mode=WAL", dbLog)
	if err != nil {
		panic(err)
	}

	devices, err := container.GetAllDevices(context.Background())
	if err != nil {
		panic(err)
	}


	for _, deviceStore := range devices {
		go startClient(deviceStore)
		time.Sleep(5 * time.Second) // Stagger startup to prevent SQLite locking
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

func startClient(deviceStore *whatsmeowStore.Device) {
	clientLog := waLog.Stdout("Client", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	startupTime = time.Now().Add(-2 * time.Minute)

	client.AddEventHandler(func(evt interface{}) { eventHandler(client, evt) })

	err := client.Connect()
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	fmt.Printf("تم تسجيل الدخول بنجاح للرقم %s! البوت جاهز.\n", client.Store.ID)
}
