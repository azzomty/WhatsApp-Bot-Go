import re
with open("internal/commands/admin_cover.go", "r") as f:
    c = f.read()

c = c.replace('"sort"\n', '')
c = c.replace('client.GetJoinedGroups()', 'client.GetJoinedGroups(context.Background())')
c = c.replace('ctx.Client.GetJoinedGroups()', 'ctx.Client.GetJoinedGroups(context.Background())')
c = c.replace('client.SendAppState(context.Background(), patches...)', 'client.SendAppState(context.Background(), patches)')

with open("internal/commands/admin_cover.go", "w") as f:
    f.write(c)
