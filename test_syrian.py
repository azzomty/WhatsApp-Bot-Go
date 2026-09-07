import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/commands.go', 'r') as f:
    content = f.read()

# Check where HandleSyrian should be called
print("Count HandleSyrian:", content.count("HandleSyrian"))

# Check if HandleSyrian is in Handle
if "HandleSyrian(ctx)" in content:
    print("Found HandleSyrian(ctx) in commands.go")
else:
    print("NOT FOUND HandleSyrian(ctx)")

# Find the end of Handle function
start_idx = content.find("func Handle(ctx *BotContext) {")
end_idx = content.find("\n}\n", start_idx)
print("Handle function length:", end_idx - start_idx)
