with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

bad = "func loadAnslayerUsers()\n\tloadAnslayerAccounts() {"
good = "func loadAnslayerUsers() {\n\tloadAnslayerAccounts()"
if bad in content:
    content = content.replace(bad, good)
    with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
        f.write(content)
    print("Fixed syntax")
