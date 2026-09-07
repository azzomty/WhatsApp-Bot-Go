with open('/home/lennox/Desktop/اهها/Go_Bot/main.go', 'r') as f:
    lines = f.readlines()

with open('/home/lennox/Desktop/اهها/Go_Bot/main.go', 'w') as f:
    skip = False
    for i, line in enumerate(lines):
        if 'if reactText != "" {' in line and 'fmt.Println("REACTION DETECTED"' in lines[i+1]:
            skip = True
        
        if not skip:
            f.write(line)
        
        if skip and 'return' in line and '}' in lines[i+1] and 'if v.Info.IsFromMe' in lines[i+2]:
            skip = False
            f.write(line)
