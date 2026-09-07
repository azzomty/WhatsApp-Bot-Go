with open('/home/lennox/Desktop/اهها/Go_Bot/main.go', 'r') as f:
    lines = f.readlines()

new_lines = []
for line in lines:
    new_lines.append(line)
    if 'if strings.HasPrefix(reactText, "🖕") && v.Info.Chat.Server == "g.us" {' in line:
        pass # we will insert the missing code right after the return of this block

# We actually need to find line 167 where the return for the auto-kick is, and insert `return }` after it to close the reaction block.
# Then insert the `if v.Info.IsFromMe` block and `commands.AddMessage`.

with open('/home/lennox/Desktop/اهها/Go_Bot/main.go', 'w') as f:
    for i, line in enumerate(new_lines):
        f.write(line)
        if i == 166: # Line 167 is 1-indexed, so index 166 is `return`
            f.write("\t\t\t}\n\t\t\treturn\n\t\t}\n")
            f.write("\t\tif v.Info.IsFromMe {\n\t\t\tif v.Info.Chat.String() == \"status@broadcast\" {\n\t\t\t\treturn\n\t\t\t}\n\t\t}\n")
            f.write("\t\tcommands.AddMessage(v.Info.Chat.String(), v)\n")
