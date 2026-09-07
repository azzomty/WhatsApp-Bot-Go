with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

if "var seenGroupLinks =" not in content:
    content = content.replace("import (", "import (\n\t\"sync\"\n")
    content = content.replace("var AutoJoinGroups bool", "var (\n\tAutoJoinGroups bool\n\tseenGroupLinks = make(map[string]bool)\n\tseenLinksMutex sync.Mutex\n)")

    # Just in case var AutoJoinGroups is inside var ( )
    if "var AutoJoinGroups bool" not in content:
        content = content.replace("	AutoJoinGroups bool\n", "	AutoJoinGroups bool\n\tseenGroupLinks = make(map[string]bool)\n\tseenLinksMutex sync.Mutex\n")

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'w') as f:
    f.write(content)
