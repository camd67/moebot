package util

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const lineSpace int = 16
const xBorder int = 30
const yBorder int = 14
const maxWidth int = 500

func MakeGif(text string) []byte {
	const defaultText string = "Hover to view"
	var frames []*image.Paletted

	text, imageSize := FormatTextSize(text, defaultText)
	fmt.Println(text)
	fmt.Println(imageSize.Size().X)
	fmt.Println(imageSize.Size().Y)
	frames = AddImageFrame(frames, imageSize, defaultText, color.RGBA{0xff, 0xff, 0xff, 0xff}, color.RGBA{0x00, 0x00, 0x00, 0xff})

	frames = AddImageFrame(frames, imageSize, text, color.RGBA{0x00, 0x00, 0x00, 0xff}, color.RGBA{0xff, 0xff, 0xff, 0xff})

	buf := new(bytes.Buffer)
	gif.EncodeAll(buf, &gif.GIF{
		Image:     frames,
		LoopCount: 1,
		Delay:     []int{300, 100000},
	})
	return buf.Bytes()
}

func AddImageFrame(frames []*image.Paletted, size image.Rectangle, text string, imageColor color.RGBA, textColor color.RGBA) []*image.Paletted {
	lines := strings.Split(text, "\n")
	img, d := UniformColorImage(size,
		imageColor, textColor, fixed.Point26_6{fixed.Int26_6(xBorder / 2 * 64), fixed.Int26_6(lineSpace * 64)})
	for i, s := range lines {
		d.Dot.X = fixed.Int26_6(xBorder / 2 * 64)
		d.Dot.Y = fixed.Int26_6((i + 1) * lineSpace * 64)

		d.DrawString(s)
	}
	return append(frames, img)
}

func UniformColorImage(size image.Rectangle, imageColor color.RGBA, textColor color.RGBA, startPoint fixed.Point26_6) (result *image.Paletted, drawer *font.Drawer) {
	var palette = []color.Color{
		color.RGBA{0x00, 0x00, 0x00, 0xff},
		color.RGBA{0xff, 0xff, 0xff, 0xff},
	}
	img := image.NewPaletted(size, palette)
	SetBackground(img, imageColor)
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(textColor),
		Face: basicfont.Face7x13,
		Dot:  startPoint,
	}
	return img, d
}

func SetBackground(img *image.Paletted, imgColor color.RGBA) {
	bounds := img.Bounds()
	for x := 0; x < bounds.Size().X; x++ {
		for y := 0; y < bounds.Size().Y; y++ {
			img.Set(x, y, imgColor)
		}
	}
}

func FormatTextSize(text string, def string) (string, image.Rectangle) {
	var r image.Rectangle
	if font.MeasureString(basicfont.Face7x13, def).Ceil() >= font.MeasureString(basicfont.Face7x13, text).Ceil() {
		r = image.Rect(0, 0,
			font.MeasureString(basicfont.Face7x13, def).Ceil()+xBorder,
			yBorder+lineSpace,
		)
		return text, r
	}

	lines := []string{text}
	currentLine := text

	for font.MeasureString(basicfont.Face7x13, currentLine).Ceil() > maxWidth-xBorder {
		lineFragments := strings.Split(currentLine, " ")
		lines = append(lines, "")
		for i := len(lineFragments) - 2; i >= 0 && font.MeasureString(basicfont.Face7x13, currentLine).Ceil() > maxWidth-xBorder; i-- {
			currentLine = strings.Join(lineFragments[:i], " ")
			lines[len(lines)-2] = currentLine
			lines[len(lines)-1] = strings.Join(lineFragments[i:], " ")
		}
		currentLine = lines[len(lines)-1]
	}
	size := image.Point{X: 0, Y: 0}
	if len(lines) > 1 {
		size.X = maxWidth
	} else {
		size.X = font.MeasureString(basicfont.Face7x13, currentLine).Ceil()
	}
	size.Y = lineSpace*len(lines) + yBorder
	r = image.Rect(0, 0, size.X, size.Y)
	return strings.Join(lines, "\n"), r
}
