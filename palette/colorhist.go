package palette

import (
	"image"
	"image/color"
)

// ColorHistogram is the histogram of colors in a image.
type ColorHistogram Histogram

// NewColorHistogram returns the ColorHistogram for the given image where the histogram keys are encoded RGB values.
func NewColorHistogram(img image.Image) ColorHistogram {
	bounds := img.Bounds()

	area := 0
	m := make(map[int]float64)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
			if rgba.A == 0 {
				// Skip invisible pixels.
				continue
			}
			m[rgbaToInt(rgba)]++
			area++
		}
	}

	// Normalise by the number of pixels
	return ColorHistogram(mapToHistogram(m, float64(area)))
}

// Cdf returns the cumulative distribution function (CDF) for this histogram.
func (h ColorHistogram) Cdf() (Cdf, error) {
	return Histogram(h).Cdf()
}

// Combine combines two histograms together. The resulting histogram is
// not normalised.
func (h ColorHistogram) Combine(hist ColorHistogram) ColorHistogram {
	return ColorHistogram(Histogram(h).Combine(Histogram(hist)))
}

// Normalise ensures the area under the histogram equals one.
func (h ColorHistogram) Normalise() ColorHistogram {
	return ColorHistogram(Histogram(h).Normalise())
}

func (h ColorHistogram) String() string {
	return Histogram(h).String()
}
