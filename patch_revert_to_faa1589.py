with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    current_content = f.read()

with open('/home/lennox/Desktop/اهها/Go_Bot/faa1589_downloader.go', 'r') as f:
    faa_content = f.read()

# Extract .اغنية block from faa_content
start_idx = faa_content.find('if strings.HasPrefix(text, ".اغنية") || strings.HasPrefix(text, ".أغنية") {')
end_idx = faa_content.find('\n\n\tif strings.HasPrefix(text, ".تحميل")', start_idx)
if end_idx == -1:
    end_idx = faa_content.find('\n\n\tsession, exists :=', start_idx)

faa_song_logic = faa_content[start_idx:end_idx]

# Extract .اغنية block from current_content (which is currently the old_yt_downloader version)
cur_start_idx = current_content.find('if strings.HasPrefix(text, ".اغنية") || strings.HasPrefix(text, ".أغنية") {')
cur_end_idx = current_content.find('\n\n\tif strings.HasPrefix(text, ".ساوند") || strings.HasPrefix(text, ".ساوند كلاود") {', cur_start_idx)

# Replace the block in current_content
new_content = current_content[:cur_start_idx] + faa_song_logic + current_content[cur_end_idx:]

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(new_content)
print("Patched downloader to restore faa1589 .اغنية")
