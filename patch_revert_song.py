import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    current_content = f.read()

with open('/home/lennox/Desktop/اهها/Go_Bot/old_downloader.go', 'r') as f:
    old_content = f.read()

# 1. Rename .اغنية to .ساوند in current content
current_content = current_content.replace('if twoWordCmd == ".اغنية" {', 'if twoWordCmd == ".ساوند" || twoWordCmd == ".ساوند كلاود" {')

# 2. Extract the .اغنية logic from old_content
old_song_logic_match = re.search(r'(if twoWordCmd == "\.اغنية" \{.*?\n\t\})', old_content, re.DOTALL)
if old_song_logic_match:
    old_song_logic = old_song_logic_match.group(1)
    
    # Remove emojis from old_song_logic
    old_song_logic = old_song_logic.replace(" ⏳", "")
    old_song_logic = old_song_logic.replace("❌ ", "")
    
    # Insert it right before the new .ساوند logic
    current_content = current_content.replace('if twoWordCmd == ".ساوند" || twoWordCmd == ".ساوند كلاود" {', old_song_logic + '\n\n\tif twoWordCmd == ".ساوند" || twoWordCmd == ".ساوند كلاود" {')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(current_content)
print("Patched downloader to restore old .اغنية and rename new to .ساوند")
