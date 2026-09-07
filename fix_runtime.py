import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'r') as f:
    content = f.read()

content = content.replace('NumberOfEpisodes int    `json:"number_of_episodes"`', 'NumberOfEpisodes int    `json:"number_of_episodes"`\n\t\tRuntime          int    `json:"runtime"`')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/media.go', 'w') as f:
    f.write(content)
