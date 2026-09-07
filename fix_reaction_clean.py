with open('/home/lennox/Desktop/اهها/Go_Bot/main.go', 'r') as f:
    content = f.read()

import re

# We will replace the whole visual search logic inside HandleReaction with nothing, OR we can remove the block from main.go.
# Actually, the easiest way is to modify commands.go's HandleReaction to just return immediately!
