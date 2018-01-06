package util

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	lineSpace = 16
	xBorder   = 30
	yBorder   = 14
	maxWidth  = 500
)

func MakeGif(text string) []byte {
	const defaultText = "Hover to view"
	var frames []*image.Paletted

	text, imageSize := formatTextSize(text, defaultText)

	frames = addImageFrame(frames, imageSize, defaultText, color.RGBA{0xff, 0xff, 0xff, 0xff}, color.RGBA{0x00, 0x00, 0x00, 0xff})

	frames = addImageFrame(frames, imageSize, text, color.RGBA{0x00, 0x00, 0x00, 0xff}, color.RGBA{0xff, 0xff, 0xff, 0xff})

	buf := new(bytes.Buffer)
	gif.EncodeAll(buf, &gif.GIF{
		Image:     frames,
		LoopCount: 1,
		Delay:     []int{300, 100000},
	})
	return buf.Bytes()
}

func addImageFrame(frames []*image.Paletted, size image.Rectangle, text string, imageColor color.RGBA, textColor color.RGBA) []*image.Paletted {
	lines := strings.Split(text, "\n")
	img, d := uniformColorImage(size,
		imageColor, textColor, fixed.Point26_6{fixed.Int26_6(xBorder / 2 * 64), fixed.Int26_6(lineSpace * 64)})
	for i, s := range lines {
		d.Dot.X = fixed.Int26_6(xBorder / 2 * 64)
		d.Dot.Y = fixed.Int26_6((i + 1) * lineSpace * 64)

		d.DrawString(s)
	}
	return append(frames, img)
}

func uniformColorImage(size image.Rectangle, imageColor color.RGBA, textColor color.RGBA, startPoint fixed.Point26_6) (result *image.Paletted, drawer *font.Drawer) {
	var palette = []color.Color{
		color.RGBA{0x00, 0x00, 0x00, 0xff},
		color.RGBA{0xff, 0xff, 0xff, 0xff},
	}
	img := image.NewPaletted(size, palette)
	setBackground(img, imageColor)
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(textColor),
		Face: basicfont.Face7x13,
		Dot:  startPoint,
	}
	return img, d
}

func setBackground(img *image.Paletted, imgColor color.RGBA) {
	bounds := img.Bounds()
	for x := 0; x < bounds.Size().X; x++ {
		for y := 0; y < bounds.Size().Y; y++ {
			img.Set(x, y, imgColor)
		}
	}
}

func formatTextSize(text string, def string) (string, image.Rectangle) {
	var r image.Rectangle
	if font.MeasureString(basicfont.Face7x13, def).Ceil() >= font.MeasureString(basicfont.Face7x13, text).Ceil() {
		r = image.Rect(0, 0,
			font.MeasureString(basicfont.Face7x13, def).Ceil()+xBorder,
			yBorder+lineSpace,
		)
		return text, r
	}

	lines := strings.Split(text, "\n")
	result := []string{}
	for _, line := range lines {
		result = append(result, formatLine(line)...)
	}

	size := image.Point{X: 0, Y: 0}
	if len(result) > 1 {
		size.X = maxWidth
	} else {
		size.X = font.MeasureString(basicfont.Face7x13, text).Ceil()
	}
	size.Y = lineSpace*len(result) + yBorder
	r = image.Rect(0, 0, size.X, size.Y)
	return strings.Join(result, "\n"), r
}

func formatLine(text string) []string {
	result := []string{text}
	currentLine := text

	for font.MeasureString(basicfont.Face7x13, currentLine).Ceil() > maxWidth-xBorder {
		lineFragments := strings.Split(currentLine, " ")
		result = append(result, "")
		for i := len(lineFragments) - 2; i >= 0 && font.MeasureString(basicfont.Face7x13, currentLine).Ceil() > maxWidth-xBorder; i-- {
			currentLine = strings.Join(lineFragments[:i], " ")
			result[len(result)-2] = currentLine
			result[len(result)-1] = strings.Join(lineFragments[i:], " ")
		}
		currentLine = result[len(result)-1]
	}
	return result
}
