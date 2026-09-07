import re

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/stickers/stickers.go', 'r') as f:
    content = f.read()

# Replace webpmux in-place calls with temp file calls
old1 = '''		err = exec.Command("./webpmux", "-set", "exif", exifPath, inputPath, "-o", outputPath).Run()'''
new1 = '''		err = exec.Command("./webpmux", "-set", "exif", exifPath, inputPath, "-o", outputPath).Run()'''
# Wait, old1 already writes to outputPath! `inputPath` is input, `outputPath` is output. So that's NOT in-place!
# But for the ffmpeg branch:
old2 = '''		// Inject EXIF using webpmux
		exec.Command("./webpmux", "-set", "exif", exifPath, outputPath, "-o", outputPath).Run()'''
new2 = '''		// Inject EXIF using webpmux
		exifOut := outputPath + ".exif.webp"
		err = exec.Command("./webpmux", "-set", "exif", exifPath, outputPath, "-o", exifOut).Run()
		if err == nil {
			os.Rename(exifOut, outputPath)
		}'''

content = content.replace(old2, new2)

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/stickers/stickers.go', 'w') as f:
    f.write(content)
print("Patched webpmux IO")
