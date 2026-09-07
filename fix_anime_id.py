import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

# 1. Struct
old_struct = """type AnslayerAnime struct {
	AnimeID   int    `json:"anime_id"`
	AnimeName string `json:"anime_name"`
}"""
new_struct = """type AnslayerAnime struct {
	AnimeID   string `json:"anime_id"`
	AnimeName string `json:"anime_name"`
}"""
content = content.replace(old_struct, new_struct)

# 2. HandleAnslayerNumberSelect URL formatter
old_url = """u := fmt.Sprintf("https://anslayer.com/anime/public/anime/get-anime-details?anime_id=%d&fetch_episodes=Yes&more_info=No", selected.AnimeID)"""
new_url = """u := fmt.Sprintf("https://anslayer.com/anime/public/anime/get-anime-details?anime_id=%s&fetch_episodes=Yes&more_info=No", selected.AnimeID)"""
content = content.replace(old_url, new_url)

# 3. downloadAnslayerEpisode URL formatter
old_url2 = """u := fmt.Sprintf("https://anslayer.com/anime/public/anime/get-anime-details?anime_id=%d&fetch_episodes=Yes&more_info=No", anime.AnimeID)"""
new_url2 = """u := fmt.Sprintf("https://anslayer.com/anime/public/anime/get-anime-details?anime_id=%s&fetch_episodes=Yes&more_info=No", anime.AnimeID)"""
content = content.replace(old_url2, new_url2)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'w') as f:
    f.write(content)
