import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    content = f.read()

content = content.replace('Url:           proto.String(respUL.URL)', 'URL:           proto.String(respUL.URL)')
content = content.replace('FileEncSha256: respUL.FileEncSHA256', 'FileEncSHA256: respUL.FileEncSHA256')
content = content.replace('FileSha256:    respUL.FileSHA256', 'FileSHA256:    respUL.FileSHA256')
content = content.replace('Ptt:           proto.Bool(false)', 'PTT:           proto.Bool(false)')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(content)
print("Fixed struct fields")
