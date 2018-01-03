package util

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"math"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	CaseInsensitive = iota
	CaseSensitive
)

func IntContains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func StrContains(s []string, e string, caseInsensitive int) bool {
	for _, a := range s {
		if caseInsensitive == CaseInsensitive {
			if strings.EqualFold(e, a) {
				return true
			}
		} else {
			if a == e {
				return true
			}
		}
	}
	return false
}

func StrContainsPrefix(s []string, e string, caseInsensitive int) bool {
	for _, a := range s {
		if caseInsensitive == CaseInsensitive {
			if strings.HasPrefix(strings.ToUpper(a), strings.ToUpper(e)) {
				return true
			}
		} else {
			if strings.HasPrefix(a, e) {
				return true
			}
		}
	}
	return false
}

func MakeAlphaOnly(s string) string {
	reg := regexp.MustCompile("[^A-Za-z ]+")
	return reg.ReplaceAllString(s, "")
}

func NormalizeNewlines(s string) string {
	reg := regexp.MustCompile("(\r\n|\r|\n)")
	return reg.ReplaceAllString(s, "\n")
}

/*
Converts a user's ID into a mention.
This is useful when you don't have a User object, but want to mention them
*/
func UserIdToMention(userId string) string {
	return fmt.Sprintf("<@%s>", userId)
}

func FindRoleByName(roles []*discordgo.Role, toFind string) *discordgo.Role {
	toFind = strings.ToUpper(toFind)
	for _, r := range roles {
		if strings.ToUpper(r.Name) == toFind {
			return r
		}
	}
	return nil
}

func FindRoleById(roles []*discordgo.Role, toFind string) *discordgo.Role {
	// for some reason roleIds have spaces in them...
	toFind = strings.TrimSpace(toFind)
	for _, r := range roles {
		if r.ID == toFind {
			return r
		}
	}
	return nil
}

func GetSpoilerContents(messageParams []string) (title string, text string) {
	if messageParams == nil {
		return "", ""
	}
	reg := regexp.MustCompile("^(\\[.+?\\])")
	return strings.Replace(strings.Replace(reg.FindString(strings.Join(messageParams, " ")), "]", "", 1), "[", "", 1), reg.ReplaceAllString(strings.Join(messageParams, " "), "")
}

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
