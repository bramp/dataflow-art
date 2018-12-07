package palette

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"strings"
)

const paletteBoxWidth = 50
const paletteBoxHeight = 50

// ColorPalette is a slice of color.Color.
type ColorPalette []color.Color

// FromRGBA creates a ColorPalette from the given slice of RGBA colors.
func FromRGBA(pal []RGBA) ColorPalette {
	p := make(ColorPalette, len(pal))
	for i, c := range pal {
		p[i] = c
	}
	return p
}

// ToRGBA returns this ColorPalette represented by RGBA colors.
func (p ColorPalette) ToRGBA() []RGBA {
	rgba := make([]RGBA, len(p))
	for i, c := range p {
		rgba[i] = RGBA(color.RGBAModel.Convert(c).(color.RGBA))
	}
	return rgba
}

// Draw outputs a png file representing the color palette.
func (p ColorPalette) Draw(out io.Writer) error {
	img := image.NewRGBA(image.Rect(0, 0, len(p)*paletteBoxWidth, paletteBoxHeight))

	// Draw a box for each entry in the palete
	for i, c := range p {
		rect := image.Rect(i*paletteBoxWidth, 0, (i+1)*paletteBoxWidth, paletteBoxHeight)
		draw.Draw(img, rect, &image.Uniform{c}, image.ZP, draw.Src)
	}

	return png.Encode(out, img)
}

func (p ColorPalette) String() string {
	var sb strings.Builder
	for _, c := range p {
		r, g, b, _ := c.RGBA()
		fmt.Fprintf(&sb, "r:%d g:%d b:%d", r, g, b)
	}
	return sb.String()
}
