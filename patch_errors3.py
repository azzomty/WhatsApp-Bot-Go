import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/anslayer.go', 'r') as f:
    content = f.read()

# fix checkAndReplyBatch in monitorAnimeComments
old1 = "replied, _ := checkAndReplyBatch(epIDFloat, msg, 0, 30)"
new1 = "replied, _ := checkAndReplyBatch(epIDFloat, msg, 0, 30, AnslayerAccount{})"
content = content.replace(old1, new1)

old2 = "replied, hasComments = checkAndReplyBatch(epIDFloat, msg, oldOffset, 30)"
new2 = "replied, hasComments = checkAndReplyBatch(epIDFloat, msg, oldOffset, 30, AnslayerAccount{})"
content = content.replace(old2, new2)

# fix reqHeaders in test connection
old3 = """	req, _ := http.NewRequest("GET", "https://anslayer.com/anime/public/animes/get-published-animes", nil)
	reqHeaders(req)"""
new3 = """	req, _ := http.NewRequest("GET", "https://anslayer.com/anime/public/animes/get-published-animes", nil)
	reqHeaders(req, "")"""
content = content.replace(old3, new3)

# fix monitorFavComments in test (wait, what is line 635?)
# Ah, I replaced monitorFavComments signature to take AnslayerAccount!
# Did I miss a call to monitorFavComments somewhere?
# Let's check line 635.
