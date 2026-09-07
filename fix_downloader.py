with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    content = f.read()

content = content.replace('t"path/filepath"', '"path/filepath"')
content = content.replace('\t"io/ioutil"\n', '"io/ioutil"\n')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(content)
