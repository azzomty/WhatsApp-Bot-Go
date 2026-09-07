with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    content = f.read()

content = content.replace('t"net/url"\\n\\t"regexp"', '"net/url"\\n\\t"regexp"')
content = content.replace('t"net/url"\n\t"regexp"', '"net/url"\n\t"regexp"')
content = content.replace('t"net/url"', '"net/url"')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(content)
