import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'r') as f:
    content = f.read()

content = content.replace('"context"\n\t"encoding/json"', '"context"\n\t"encoding/json"\n\t"os"\n\t"os/exec"')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'w') as f:
    f.write(content)
