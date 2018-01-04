package util

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"math"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func MakeGif(text string) []byte {
	var images []*image.Paletted
	img, d := UniformColorImage(image.Rect(0, 0, int(math.Max(100, float64(font.MeasureString(basicfont.Face7x13, text).Ceil())))+30, 30),
		color.RGBA{0xff, 0xff, 0xff, 0xff}, color.RGBA{0x00, 0x00, 0x00, 0xff}, fixed.Point26_6{fixed.Int26_6(10 * 64), fixed.Int26_6(16 * 64)})
	d.DrawString("Hover to view")
	images = append(images, img)

	for i := 0; i < 10; i++ {
		img, d := UniformColorImage(image.Rect(0, 0, int(math.Max(100, float64(font.MeasureString(basicfont.Face7x13, text).Ceil())))+30, 30),
			color.RGBA{0x00, 0x00, 0x00, 0xff}, color.RGBA{0xff, 0xff, 0xff, 0xff}, fixed.Point26_6{fixed.Int26_6(10 * 64), fixed.Int26_6(16 * 64)})
		d.DrawString(text)
		images = append(images, img)
	}

	buf := new(bytes.Buffer)
	gif.EncodeAll(buf, &gif.GIF{
		Image:     images,
		LoopCount: 1,
		Delay:     []int{500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 500},
	})
	return buf.Bytes()
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
