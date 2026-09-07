import re

with open("internal/pinterest/pinterest.go", "r") as f:
    c = f.read()

target = """func SearchPinterest(query string, aspect string, count int) []PinResult {
	query = url.QueryEscape(query)
	pageSize := count + 10
	if pageSize > 100 {
		pageSize = 100
	}"""
new_target = """func SearchPinterest(query string, aspect string, count int) []PinResult {
	query = url.QueryEscape(query)
	pageSize := 100 // Always fetch max to shuffle properly"""
c = c.replace(target, new_target)

with open("internal/pinterest/pinterest.go", "w") as f:
    f.write(c)

