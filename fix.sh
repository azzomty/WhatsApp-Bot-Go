sed -i 's/pinterest.SetPending(ctx.ChatID.String(), query, count)/pinterest.SetPending(ctx.ChatID.String(), query, count, false, "")/g' internal/commands/commands.go
sed -i 's/pinterest.SetLastSearch(v.Info.Chat.String(), req.Query, aspect, overrideCount)/pinterest.SetLastSearch(v.Info.Chat.String(), req.Query, aspect, overrideCount, req.IsVisual, req.Base64Image)/g' main.go
