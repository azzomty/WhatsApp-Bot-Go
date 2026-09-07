import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    current_content = f.read()

with open('/home/lennox/Desktop/اهها/Go_Bot/old_yt_downloader.go', 'r') as f:
    old_content = f.read()

# 1. Rename .اغنية to .ساوند in current content
current_content = current_content.replace('if strings.HasPrefix(text, ".اغنية") || strings.HasPrefix(text, ".أغنية") {', 'if strings.HasPrefix(text, ".ساوند") || strings.HasPrefix(text, ".ساوند كلاود") {')

# 2. Extract the .اغنية logic from old_yt_downloader.go
start_index = old_content.find('if strings.HasPrefix(text, ".اغنية")')
end_index = old_content.find('\n\n\tif strings.HasPrefix(text, ".تحميل")', start_index)
if end_index == -1:
    end_index = old_content.find('\n\n\tsession, exists :=', start_index)

old_song_logic = old_content[start_index:end_index]

# Remove emojis
old_song_logic = old_song_logic.replace(" ⏳", "")
old_song_logic = old_song_logic.replace("❌ ", "")

# Insert it right before the new .ساوند logic
insert_point = 'if strings.HasPrefix(text, ".ساوند") || strings.HasPrefix(text, ".ساوند كلاود") {'
current_content = current_content.replace(insert_point, old_song_logic + '\n\n\t' + insert_point)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(current_content)
print("Patched downloader to restore YouTube .اغنية and rename new to .ساوند")
