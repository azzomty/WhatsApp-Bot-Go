import re
with open("main.go", "r") as f:
    c = f.read()

target = """		// Handle Demote or Kick of Protected Users
		if len(v.Demote) > 0 || len(v.Leave) > 0 {
			if v.Sender != nil && v.Sender.ToNonAD().String() != client.Store.ID.ToNonAD().String() {"""

new_target = """		// Handle Demote or Kick of Protected Users
		if len(v.Demote) > 0 || len(v.Leave) > 0 {
			if v.Sender != nil && v.Sender.ToNonAD().String() != client.Store.ID.ToNonAD().String() {
				// Check Roulette Demotion First (if they kicked someone, demote them!)
				if len(v.Leave) > 0 {
					commands.CheckRouletteDemotion(client, v.JID.String(), v.Sender.ToNonAD().String())
				}
				"""
c = c.replace(target, new_target)

with open("main.go", "w") as f:
    f.write(c)

