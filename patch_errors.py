import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

# Fix undefined acc
content = content.replace("reqHeaders(req, acc.Token)", "reqHeaders(req, \"\")")

# But wait, inside my functions I want it to be acc.Token!
# Where did I replace it?
# In patch_marketing_funcs.py, I did:
# content = content.replace("reqHeaders(req)", "reqHeaders(req, acc.Token)")
# This replaced ALL reqHeaders(req)!

# Let's revert it and do it correctly.
content = content.replace('reqHeaders(req, "")', "reqHeaders(req, \"\")") # just in case
content = content.replace("reqHeaders(req)", 'reqHeaders(req, "")')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
    f.write(content)
print("Reverted all to empty string")
