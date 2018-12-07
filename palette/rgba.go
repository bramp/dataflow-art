package palette

import (
	"fmt"
	"image/color"
)

const hexColorFormat = "#%02x%02x%02x"

// RGBA is a color.RGBA but implements the encoding.TextMarshaler and encoding.TextUnmarshaler
// interfaces to turn the color into HTML hex form, e.g "#001122".
type RGBA color.RGBA

// RGBA returns the alpha-premultiplied red, green, blue and alpha values
// for the color.
func (c RGBA) RGBA() (uint32, uint32, uint32, uint32) {
	return color.RGBA(c).RGBA()
}

// Int returns the 32 bit value of this color.
func (c RGBA) Int() int {
	return rgbaToInt(c)
}

// MarshalText marshals the color into its HTML hex form, e.g "#001122".
func (c RGBA) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf(hexColorFormat, c.R, c.G, c.B)), nil
}

// UnmarshalText unmarshal from HTML hex form to a color.Color.
func (c *RGBA) UnmarshalText(hex []byte) error {
	var r, g, b uint8
	n, err := fmt.Sscanf(string(hex), hexColorFormat, &r, &g, &b)
	if n != 3 {
		return fmt.Errorf("can't parse hex %q", hex)
	}
	if err != nil {
		return err
	}
	*c = RGBA(color.RGBA{r, g, b, 255})
	return nil
}

//rgbaToInt returns the color encoded as encoded as 0xAARRGGBB
func rgbaToInt(rgba color.Color) int {
	r, g, b, a := rgba.RGBA()
	return int((a&0xFF)<<24 | (r&0xFF)<<16 | (g&0xFF)<<8 | (b & 0xFF))
}

// intToRgba returns the 0xAARRGGBB number as a color.Color.
func intToRgba(x int) color.Color {
	var rgba color.RGBA
	rgba.A = uint8((uint32(x) >> 24) & 0xFF)
	rgba.R = uint8((uint32(x) >> 16) & 0xFF)
	rgba.G = uint8((uint32(x) >> 8) & 0xFF)
	rgba.B = uint8((uint32(x) >> 0) & 0xFF)
	return rgba
}
