package palette

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"sort"

	colorful "github.com/lucasb-eyer/go-colorful"
)

const paletteBoxWidth = 50
const paletteBoxHeight = 50

type ColorPalette []color.Color // TODO Perhaps replace with color.Palette

// FromRGBA creates a ColorPalette from the given slice of RGBA colors.
func FromRGBA(pal []color.RGBA) ColorPalette {
	var p ColorPalette
	for _, c := range pal {
		p = append(p, c)
	}
	return p
}

// ToRGBA returns this ColorPalette represented by RGBA colors.
func (p ColorPalette) ToRGBA() []color.RGBA {
	var rgba []color.RGBA
	for _, c := range p {
		rgb := color.RGBAModel.Convert(c).(color.RGBA)
		rgba = append(rgba, rgb)
	}
	return rgba
}

// Draw outputs a png file representing the color palette.
func (p ColorPalette) Draw(out io.Writer) error {
	// Sort the palette so its results are comparable
	sort.Slice(p, func(i, j int) bool {
		ci, _ := colorful.MakeColor(p[i])
		cj, _ := colorful.MakeColor(p[j])

		// Sort by hue
		hi, _, _ := ci.Hsl()
		hj, _, _ := cj.Hsl()
		if hi != hj {
			return hi < hj
		}

		// If hues are the same, sort by RGB value
		if ci.R != cj.R {
			return ci.R < cj.R
		}
		if ci.G != cj.G {
			return ci.G < cj.G
		}
		if ci.B != cj.B {
			return ci.B < cj.B
		}
		return false
	})

	img := image.NewRGBA(image.Rect(0, 0, len(p)*paletteBoxWidth, paletteBoxHeight))

	// Draw a box for each entry in the palete
	for i, c := range p {
		rect := image.Rect(i*paletteBoxWidth, 0, (i+1)*paletteBoxWidth, paletteBoxHeight)
		draw.Draw(img, rect, &image.Uniform{c}, image.ZP, draw.Src)
	}

	return png.Encode(out, img)
}
