with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/mdm.go', 'r') as f:
    c = f.read()

old_code = '''	if err == nil {
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)'''

new_code = '''	if err == nil {
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		if res.StatusCode != 200 {
			sendMessage(ctx, fmt.Sprintf("DEBUG: API Status %d", res.StatusCode))
		}
		
		var dbgErr error'''

c = c.replace(old_code, new_code)

old_unmarshal = '''		if json.Unmarshal(body, &data) == nil {'''
new_unmarshal = '''		dbgErr = json.Unmarshal(body, &data)
		if dbgErr != nil {
			sendMessage(ctx, fmt.Sprintf("DEBUG: JSON Error: %v", dbgErr))
		}
		if dbgErr == nil {'''
		
c = c.replace(old_unmarshal, new_unmarshal)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/mdm.go', 'w') as f:
    f.write(c)
