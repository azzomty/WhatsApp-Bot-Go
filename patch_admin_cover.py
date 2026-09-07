import re
with open("internal/commands/admin_cover.go", "r") as f:
    c = f.read()

c = c.replace('waSyncAction "go.mau.fi/whatsmeow/binary/proto/waSyncAction"', 'waSyncAction "go.mau.fi/whatsmeow/proto/waSyncAction"')
c = c.replace('waProto "go.mau.fi/whatsmeow/binary/proto"', 'waProto "go.mau.fi/whatsmeow/binary/proto"') # wait, let's keep binary/proto if it works. Let me test if I can just build it first with go.mau.fi/whatsmeow/proto/waSyncAction

with open("internal/commands/admin_cover.go", "w") as f:
    f.write(c)
