package commands

import (
	"bytes"
	"fmt"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/fogleman/gg"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
)

var colorMap = map[string]color.RGBA{
	"white":   {255, 255, 255, 255},
	"black":   {0, 0, 0, 255},
	"red":     {255, 0, 0, 255},
	"green":   {0, 255, 0, 255},
	"blue":    {0, 0, 255, 255},
	"yellow":  {255, 255, 0, 255},
	"purple":  {128, 0, 128, 255},
	"pink":    {255, 192, 203, 255},
	"orange":  {255, 165, 0, 255},
	"gray":    {128, 128, 128, 255},
	"brown":   {165, 42, 42, 255},
	"cyan":    {0, 255, 255, 255},
	"magenta": {255, 0, 255, 255},
	"lime":    {0, 255, 0, 255},
	"teal":    {0, 128, 128, 255},
	"navy":    {0, 0, 128, 255},
	"gold":    {255, 215, 0, 255},
	"silver":  {192, 192, 192, 255},
	"transparent": {0, 0, 0, 0},
}

func parseColor(name string, def color.RGBA) color.RGBA {
	c, ok := colorMap[strings.ToLower(strings.TrimSpace(name))]
	if ok {
		return c
	}
	return def
}

func findFontFile(fontName string) string {
	dir := "/home/lennox/Desktop/font/"
	files, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	searchName := strings.ToLower(strings.ReplaceAll(fontName, " ", ""))
	for _, f := range files {
		if !f.IsDir() {
			ext := strings.ToLower(filepath.Ext(f.Name()))
			if ext == ".ttf" || ext == ".otf" {
				base := strings.ToLower(strings.ReplaceAll(strings.TrimSuffix(f.Name(), ext), " ", ""))
				if base == searchName || strings.Contains(base, searchName) {
					return filepath.Join(dir, f.Name())
				}
			}
		}
	}
	return ""
}

func GenerateTextImage(text, fontName, bgColorName, fgColorName string) ([]byte, error) {
	bgColor := parseColor(bgColorName, color.RGBA{0, 0, 0, 0})
	fgColor := parseColor(fgColorName, color.RGBA{0, 0, 0, 255})

	fontPath := findFontFile(fontName)
	if fontPath == "" {
		return nil, fmt.Errorf("الخط '%s' غير موجود", fontName)
	}

	fontBytes, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, err
	}

	fnt, err := sfnt.Parse(fontBytes)
	if err != nil {
		return nil, err
	}

	face, err := opentype.NewFace(fnt, &opentype.FaceOptions{
		Size:    100, // Big size for clear stickers
		DPI:     72,
		Hinting: 0,
	})
	if err != nil {
		return nil, err
	}

	dc := gg.NewContext(1, 1)
	dc.SetFontFace(face)
	lines := strings.Split(text, "\n")
	maxWidth := 0.0
	totalHeight := 0.0
	lineHeight := 120.0
	
	for _, line := range lines {
		w, _ := dc.MeasureString(line)
		if w > maxWidth {
			maxWidth = w
		}
		totalHeight += lineHeight
	}

	width := int(maxWidth) + 100
	height := int(totalHeight) + 100

	// Stickers look best in 512x512 max
	if width > 512 || height > 512 {
		// Just scale it nicely by making a 512 canvas if it's too big, or let PNG handle it
		if width > 1024 { width = 1024 }
		if height > 1024 { height = 1024 }
	}

	dc = gg.NewContext(width, height)
	
	dc.SetColor(bgColor)
	dc.Clear()

	dc.SetColor(fgColor)
	dc.SetFontFace(face)
	
	y := 50.0 + (lineHeight * 0.8)
	for _, line := range lines {
		w, _ := dc.MeasureString(line)
		x := (float64(width) - w) / 2
		dc.DrawString(line, x, y)
		y += lineHeight
	}

	var buf bytes.Buffer
	// Encode as PNG for transparency!
	err = png.Encode(&buf, dc.Image())
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
