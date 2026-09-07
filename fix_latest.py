with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

old = """	if len(res.Response.Episodes.Data) > 0 {
		// Usually the first item is the newest episode if ordered by latest
		return res.Response.Episodes.Data[0].EpisodeID
	}"""

new = """	if len(res.Response.Episodes.Data) > 0 {
		// The last item in the array is the newest episode
		lastIdx := len(res.Response.Episodes.Data) - 1
		return res.Response.Episodes.Data[lastIdx].EpisodeID
	}"""

if old in content:
    content = content.replace(old, new)
    with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
        f.write(content)
    print("Fixed latest episode logic!")
