package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
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
	client  *whatsmeow.Client
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

func eventHandler(evt interface{}) {
	switch v := evt.(type) {

	case *events.Message:
		if v.Info.Timestamp.Before(startupTime) {
			return
		}
		if v.Info.IsFromMe {
			// Optionally allow bot's own messages for commands if needed
			if v.Info.Chat.String() == "status@broadcast" {
				return
			}
		}

		commands.AddMessage(v.Info.Chat.String(), v)

		text := ""
		
		unwrap := func(m *waProto.Message) *waProto.Message {
			for m != nil {
				if m.EphemeralMessage != nil && m.EphemeralMessage.Message != nil {
					m = m.EphemeralMessage.Message
					continue
				}
				if m.ViewOnceMessage != nil && m.ViewOnceMessage.Message != nil {
					m = m.ViewOnceMessage.Message
					continue
				}
				if m.ViewOnceMessageV2 != nil && m.ViewOnceMessageV2.Message != nil {
					m = m.ViewOnceMessageV2.Message
					continue
				}
				break
			}
			return m
		}

		uMsg := unwrap(v.Message)
		if uMsg.GetExtendedTextMessage() != nil {
			text = uMsg.GetExtendedTextMessage().GetText()
		} else if uMsg.GetConversation() != "" {
			text = uMsg.GetConversation()
		} else if uMsg.GetImageMessage() != nil {
			text = uMsg.GetImageMessage().GetCaption()
		} else if uMsg.GetVideoMessage() != nil {
			text = uMsg.GetVideoMessage().GetCaption()
		}

		senderLID := getLID(client, v.Info.Sender)
		if strings.Contains(senderLID, "224245258948685") {
			client.SendMessage(context.Background(), v.Info.Chat, client.BuildReaction(v.Info.Chat, v.Info.Sender, v.Info.ID, "👍🏻"))
		}



		// أمر معرفة الـ LID
		if strings.HasPrefix(text, ".lid") {
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

		if v.Message.GetReactionMessage() != nil {
			if v.Info.Sender.ToNonAD().String() != client.Store.ID.ToNonAD().String() {
				client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
					ExtendedTextMessage: &waProto.ExtendedTextMessage{
						Text: proto.String("شفتك تفاعلت! تم تفعيل اقتراحات Pinterest الخاصة بك بناءً على هذا التفاعل."),
					},
				})
			}
		}

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

		ctx := &commands.BotContext{
			Client: client,
			Event:  v,
			ChatID: v.Info.Chat,
			Sender: v.Info.Sender,
			Text:   text,
		}

		senderID := getLID(client, v.Info.Sender)
		if store.IsMuted(senderID) {
			client.SendMessage(context.Background(), v.Info.Chat, client.BuildRevoke(v.Info.Chat, v.Info.Sender, v.Info.ID))
			return
		}

		if strings.Contains(senderID, "224245258948685") {
			client.SendMessage(context.Background(), v.Info.Chat, client.BuildReaction(v.Info.Chat, v.Info.Sender, v.Info.ID, "👍🏻"))
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

				default:
					// Not a valid choice, let commands handle it
					goto CommandHandling
				}

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
					results := pinterest.SearchPinterest(req.Query+suffix, aspect)

					if len(results) > 0 {
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
						if err == nil && len(data) > 5000 {
							resp, err := client.Upload(context.Background(), data, whatsmeow.MediaImage)
							if err == nil {
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
								client.SendMessage(context.Background(), v.Info.Chat, &waProto.Message{
									ImageMessage: imgMsg,
								})
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
		if len(v.Join) > 0 {
			groupStr := v.JID.String()
			if store.GetTargetGroup("welcome") == groupStr || store.GetTargetGroup("primary") == groupStr {
				for _, joiner := range v.Join {
					commands.SendWelcomeMessage(client, v.JID, joiner)
				}
			}
		}
		if len(v.Demote) > 0 || len(v.Leave) > 0 {
			protectedLIDs := []string{
				"966508364121",
				"963992306978",
			}
			affected := append(v.Demote, v.Leave...)
			for _, participant := range affected {
				isProtected := false
				for _, pLID := range protectedLIDs {
					if strings.Contains(participant.String(), pLID) {
						isProtected = true
						break
					}
				}
				if isProtected {
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
								Conversation: proto.String("🚨 تم المساس بأحد الأرقام المحمية! تم سحب إشراف الجميع كإجراء أمني."),
							})
						}
					}
					break
				}
			}
		}
	case *events.LoggedOut:
		fmt.Println("تم تسجيل الخروج من الهاتف! جاري حذف الجلسة...")
		client.Logout(context.Background())
		os.Exit(0)
	}
}

func startRenderServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Go Bot is Alive! 🚀")
	})
	fmt.Printf("[Render Mode] Server listening on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}

func main() {
	api.StartServer()
	go startRenderServer()

	store.LoadAll(".")

	dbLog := waLog.Stdout("Database", "ERROR", true)
	container, err := sqlstore.New(context.Background(), "sqlite3", "file:whatsapp_v2.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}

	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		panic(err)
	}

	clientLog := waLog.Stdout("Client", "INFO", true)
	client = whatsmeow.NewClient(deviceStore, clientLog)
	startupTime = time.Now()

	client.AddEventHandler(eventHandler)

	if client.Store.ID == nil {
		err = client.Connect()
		if err != nil {
			panic(err)
		}

		qrChan, _ := client.GetQRChannel(context.Background())
		for evt := range qrChan {
			if evt.Event == "code" {
				fmt.Println("===========================================")
				fmt.Println("واتساب يرفض كود الربط للأرقام، يرجى مسح كود QR التالي:")
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println("===========================================")
			} else {
				fmt.Println("QR Event:", evt.Event)
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		fmt.Println("تم تسجيل الدخول بنجاح! البوت جاهز.")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
}
