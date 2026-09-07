import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'r') as f:
    content = f.read()

old_dialer = """	dialer := &net.Dialer{}
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				if strings.HasPrefix(addr, "luluvdo.com:") || strings.HasPrefix(addr, "lulustream.com:") {
					return dialer.DialContext(ctx, network, "104.26.6.79:443")
				}
				return dialer.DialContext(ctx, network, addr)
			},
		},
	}"""

new_dialer = """	client := &http.Client{}"""

content = content.replace(old_dialer, new_dialer)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/stardima.go', 'w') as f:
    f.write(content)
