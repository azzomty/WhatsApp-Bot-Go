package commands

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

var (
	rouletteMu     sync.Mutex
	// map[groupJID]adminJID
	RouletteAdmins = make(map[string]string)
)

func HandleRoulette(ctx *BotContext) {
	senderNum := ctx.Event.Info.Sender.ToNonAD().String()
	// Restrict to Saudi number (starting with 966)
	if !strings.HasPrefix(senderNum, "966") {
		return
	}

	if !ctx.Event.Info.IsGroup {
		sendMessage(ctx, "هذا الأمر مخصص للقروبات فقط!")
		return
	}

	groupJID := ctx.Event.Info.Chat

	groupInfo, err := ctx.Client.GetGroupInfo(context.Background(), groupJID)
	if err != nil {
		sendMessage(ctx, "حدث خطأ في قراءة معلومات القروب.")
		return
	}

	ownJID := ctx.Client.Store.ID.ToNonAD().String()

	var eligibleParticipants []types.JID
	for _, p := range groupInfo.Participants {
		pStr := p.JID.ToNonAD().String()
		// Exclude bot and the sender
		if pStr != ownJID && pStr != senderNum && !p.IsAdmin && !p.IsSuperAdmin {
			eligibleParticipants = append(eligibleParticipants, p.JID)
		}
	}

	if len(eligibleParticipants) == 0 {
		sendMessage(ctx, "لا يوجد أشخاص لترقيتهم (الكل إما مشرفين أو أنت والبوت فقط)!")
		return
	}

	// Pick a random participant
	rand.Seed(time.Now().UnixNano())
	chosen := eligibleParticipants[rand.Intn(len(eligibleParticipants))]

	// Promote to Admin
	_, err = ctx.Client.UpdateGroupParticipants(context.Background(), groupJID, []types.JID{chosen}, whatsmeow.ParticipantChangePromote)
	if err != nil {
		sendMessage(ctx, "البوت يحتاج صلاحية الإشراف عشان يقدر يعطي إشراف!")
		return
	}

	chosenStr := chosen.ToNonAD().String()

	rouletteMu.Lock()
	RouletteAdmins[groupJID.String()] = chosenStr
	rouletteMu.Unlock()

	msg := fmt.Sprintf("🎰 *عجلة الحظ (روليت الإشراف)* 🎰\n\nوقع الاختيار على: @%s\nمبروك الإشراف المؤقت! 🎉\n\n⚠️ *تحذير:* لو طردت أي شخص من القروب رح ينسحب منك الإشراف فوراً!", strings.Split(chosenStr, "@")[0])
	
	ctx.Client.SendMessage(context.Background(), groupJID, &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(msg),
			ContextInfo: &waProto.ContextInfo{
				MentionedJID: []string{chosenStr},
			},
		},
	})
}

// CheckRouletteDemotion checks if a roulette admin kicked someone, and demotes them.
func CheckRouletteDemotion(client *whatsmeow.Client, groupJID string, sender string) {
	if sender == "" {
		return
	}

	rouletteMu.Lock()
	adminJID, exists := RouletteAdmins[groupJID]
	if exists && adminJID == sender {
		delete(RouletteAdmins, groupJID)
		rouletteMu.Unlock()

		// Demote them!
		targetJID, _ := types.ParseJID(adminJID)
		gJID, _ := types.ParseJID(groupJID)

		client.UpdateGroupParticipants(context.Background(), gJID, []types.JID{targetJID}, whatsmeow.ParticipantChangeDemote)
		
		msg := fmt.Sprintf("🚨 *انتهت اللعبة!* 🚨\n\n@%s قام باستخدام قوته وطرد شخصاً من القروب! تم سحب الإشراف فوراً 💀", strings.Split(adminJID, "@")[0])
		client.SendMessage(context.Background(), gJID, &waProto.Message{
			ExtendedTextMessage: &waProto.ExtendedTextMessage{
				Text: proto.String(msg),
				ContextInfo: &waProto.ContextInfo{
					MentionedJID: []string{adminJID},
				},
			},
		})
		return
	}
	rouletteMu.Unlock()
}
