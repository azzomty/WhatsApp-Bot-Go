with open('/home/lennox/Desktop/اهها/Go_Bot/internal/stickers/stickers.go', 'r') as f:
    c = f.read()

c = c.replace(
    '''err = exec.Command("./webpmux", "-set", "exif", exifPath, outputPath, "-o", exifOut).Run()''',
    '''out, errWebp := exec.Command("./webpmux", "-set", "exif", exifPath, outputPath, "-o", exifOut).CombinedOutput()
		if errWebp != nil {
			fmt.Printf("webpmux failed: %v, output: %s\\n", errWebp, string(out))
		}
		err = errWebp'''
)

c = c.replace(
    '''err = exec.Command("./webpmux", "-set", "exif", exifPath, inputPath, "-o", outputPath).Run()''',
    '''out, errWebp := exec.Command("./webpmux", "-set", "exif", exifPath, inputPath, "-o", outputPath).CombinedOutput()
		if errWebp != nil {
			fmt.Printf("webpmux failed: %v, output: %s\\n", errWebp, string(out))
		}
		err = errWebp'''
)

if '"fmt"' not in c:
    c = c.replace('import (', 'import (\n\t"fmt"')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/stickers/stickers.go', 'w') as f:
    f.write(c)
