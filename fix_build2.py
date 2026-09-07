import re
with open("internal/commands/admin_cover.go", "r") as f:
    c = f.read()

target = """	var patches []appstate.PatchInfo
	for _, g := range groups {
		action := &waSyncAction.ClearChatAction{
			MessageRange: &waSyncAction.SyncActionMessageRange{
				LastMessageTimestamp: proto.Int64(time.Now().Unix()),
			},
		}

		patch := appstate.PatchInfo{
			Type: appstate.WAPatchRegularHigh,
			Mutations: []appstate.MutationInfo{{
				Index:   []string{appstate.IndexClearChat, g.JID.String(), "", "1"},
				Version: 7,
				Value: &waSyncAction.SyncActionValue{
					ClearChatAction: action,
				},
			}},
		}
		patches = append(patches, patch)
	}

	if len(patches) > 0 {
		client.SendAppState(context.Background(), patches)
	}"""
new_target = """	var mutations []appstate.MutationInfo
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
	}"""
c = c.replace(target, new_target)

with open("internal/commands/admin_cover.go", "w") as f:
    f.write(c)
