with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/mdm.go', 'r') as f:
    content = f.read()

content = content.replace('"context"\n', '')
content = content.replace('"go.mau.fi/whatsmeow"\n', '')
content = content.replace('waProto "go.mau.fi/whatsmeow/binary/proto"\n', '')
content = content.replace('"google.golang.org/protobuf/proto"\n', '')

content = content.replace('ctx.Message.Info.Chat.String()', 'ctx.ChatID.String()')
content = content.replace('ctx.Message.Info.Sender.String()', 'ctx.Sender.String()')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/mdm.go', 'w') as f:
    f.write(content)
