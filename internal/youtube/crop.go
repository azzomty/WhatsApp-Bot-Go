package youtube

import (
	"bytes"
	"image"
	"image/jpeg"
	_ "image/png"
)

// CropTo16x9 crops an image to a 16:9 aspect ratio from the center
func CropTo16x9(imgData []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return imgData, err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	targetWidth := width
	targetHeight := (width * 9) / 16

	if targetHeight > height {
		targetHeight = height
		targetWidth = (height * 16) / 9
	}

	startX := (width - targetWidth) / 2
	startY := (height - targetHeight) / 2

	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}

	if sub, ok := img.(subImager); ok {
		cropped := sub.SubImage(image.Rect(startX, startY, startX+targetWidth, startY+targetHeight))
		var buf bytes.Buffer
		err = jpeg.Encode(&buf, cropped, &jpeg.Options{Quality: 90})
		if err != nil {
			return imgData, err
		}
		return buf.Bytes(), nil
	}

	return imgData, nil
}
