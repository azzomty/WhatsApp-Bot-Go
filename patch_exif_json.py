import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/stickers/stickers.go', 'r') as f:
    content = f.read()

new_logic = """func createExif(pack, author string) []byte {
	packBytes, _ := json.Marshal(pack)
	authorBytes, _ := json.Marshal(author)
	
	jsonStr := fmt.Sprintf(`{"sticker-pack-id":"com.snowcorp.stickerly.android.stickercontentprovider b5e7275f-f1de-4137-961f-57becfad34f2","sticker-pack-name":%s,"sticker-pack-publisher":%s,"emojis":["🤖"]}`, string(packBytes), string(authorBytes))
	
	exifData := []byte{
		0x49, 0x49, 0x2A, 0x00, 0x08, 0x00, 0x00, 0x00, 0x01, 0x00, 0x41, 0x57,
		0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0x16, 0x00, 0x00, 0x00,
	}
	
	length := len(jsonStr)
	exifData[14] = byte(length & 0xFF)
	exifData[15] = byte((length >> 8) & 0xFF)
	exifData[16] = byte((length >> 16) & 0xFF)
	exifData[17] = byte((length >> 24) & 0xFF)
	
	return append(exifData, []byte(jsonStr)...)
}"""

pattern = re.compile(r'func createExif\(pack, author string\) \[\]byte \{.*?\n\}', re.DOTALL)
content = pattern.sub(new_logic, content)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/stickers/stickers.go', 'w') as f:
    f.write(content)
print("Patched createExif")
