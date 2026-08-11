sed -i 's/strings.HasPrefix(lowerText, ".بينتريست")/strings.HasPrefix(lowerText, ".بينتريست") || strings.HasPrefix(lowerText, ".بحث")/g' internal/commands/commands.go
sed -i 's/strings.Replace(ctx.Text, ".بينتريست", "", 1)/strings.Replace(strings.Replace(ctx.Text, ".بينتريست", "", 1), ".بحث", "", 1)/g' internal/commands/commands.go
