package commands

import (
	"context"
	"fmt"
		"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/appstate"
	waSyncAction "go.mau.fi/whatsmeow/proto/waSyncAction"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

// ClearAllGroups clears messages for all joined groups
func ClearAllGroups(client *whatsmeow.Client) {
	groups, err := client.GetJoinedGroups(context.Background())
	if err != nil {
		return
	}

	var mutations []appstate.MutationInfo
	for _, g := range groups {
		action := &waSyncAction.ClearChatAction{
			MessageRange: &waSyncAction.SyncActionMessageRange{
				LastMessageTimestamp: proto.Int64(time.Now().Unix()),
			},
		}

		mutations = append(mutations, appstate.MutationInfo{
			Index:   []string{appstate.IndexClearChat, g.JID.String(), "", "1"},
			Version: 7,
			Value: &waSyncAction.SyncActionValue{
				ClearChatAction: action,
			},
		})
	}

	if len(mutations) > 0 {
		client.SendAppState(context.Background(), appstate.PatchInfo{
			Type: appstate.WAPatchRegularHigh,
			Mutations: mutations,
		})
	}
}

// StartGroupClearer starts a background task to clear groups every 2 hours
func StartGroupClearer(client *whatsmeow.Client) {
	go func() {
		for {
			ClearAllGroups(client)
			time.Sleep(2 * time.Hour)
		}
	}()
}

// HandleAdminCover extracts the optimal set of admins to cover all groups
func HandleAdminCover(ctx *BotContext) {
	sendMessage(ctx, "جاري فحص جميع القروبات واستخراج أفضل المشرفين للتواصل معهم...")

	groups, err := ctx.Client.GetJoinedGroups(context.Background())
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء جلب القروبات.")
		return
	}

	// map[adminJID][]groupJID
	adminToGroups := make(map[string][]string)
	// map[groupJID]bool
	allGroups := make(map[string]bool)

	ownJID := ctx.Client.Store.ID.ToNonAD().String()

	for _, g := range groups {
		// Only consider groups where we are actually a participant
		groupStr := g.JID.String()
		hasOtherAdmins := false

		for _, p := range g.Participants {
			pStr := p.JID.ToNonAD().String()
			if pStr == ownJID {
				continue
			}
			if p.IsAdmin || p.IsSuperAdmin {
				adminToGroups[pStr] = append(adminToGroups[pStr], groupStr)
				hasOtherAdmins = true
			}
		}

		if hasOtherAdmins {
			allGroups[groupStr] = true
		}
	}

	if len(allGroups) == 0 {
		sendMessage(ctx, "لم أجد أي قروبات تحتوي على مشرفين غيرك!")
		return
	}

	// Greedy Set Cover
	uncoveredGroups := make(map[string]bool)
	for g := range allGroups {
		uncoveredGroups[g] = true
	}

	type selectedAdmin struct {
		JID   string
		Count int
	}
	var result []selectedAdmin

	for len(uncoveredGroups) > 0 {
		bestAdmin := ""
		bestCoverage := 0
		var bestCoveredGroups []string

		for admin, grps := range adminToGroups {
			count := 0
			var covered []string
			for _, g := range grps {
				if uncoveredGroups[g] {
					count++
					covered = append(covered, g)
				}
			}
			if count > bestCoverage {
				bestCoverage = count
				bestAdmin = admin
				bestCoveredGroups = covered
			}
		}

		if bestCoverage == 0 {
			break // Should not happen, but safe break
		}

		result = append(result, selectedAdmin{JID: bestAdmin, Count: bestCoverage})
		
		for _, g := range bestCoveredGroups {
			delete(uncoveredGroups, g)
		}
	}

	// Format output
	msg := fmt.Sprintf("✅ *تم استخراج المشرفين بنجاح!*\n\n")
	msg += fmt.Sprintf("إجمالي القروبات المشتركة: %d\n", len(allGroups))
	msg += fmt.Sprintf("عدد المشرفين المطلوب مراسلتهم لتغطية كل القروبات: %d\n\n", len(result))

	for i, sa := range result {
		phone := sa.JID
		if len(phone) > 15 {
			phone = phone[:len(phone)-15] // roughly remove @s.whatsapp.net
		}
		msg += fmt.Sprintf("%d. @%s (يغطي %d قروبات)\n", i+1, phone, sa.Count)
	}

	msg += "\n💡 *تلميح:* هؤلاء هم أقل عدد من المشرفين تحتاج تكلمهم عشان تغطي كل قروباتك بدون تكرار!"

	ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{ // Wait, it's waProto.Message! I need to import waProto!
		ExtendedTextMessage: &waProto.ExtendedTextMessage{ // Let's just use ctx.Reply or something similar, or import waProto.
			Text: proto.String(msg),
		},
	})
}
