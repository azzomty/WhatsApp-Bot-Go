with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    c = f.read()

c = c.replace('\t"net/url"\n', '')
c = c.replace('\t"regexp"\n', '')
c = c.replace('\t"github.com/kkdai/youtube/v2"\n', '')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(c)
