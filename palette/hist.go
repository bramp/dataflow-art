package palette

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"sort"
)

// ColorHistogram is the histogram of colors in a image.
type ColorHistogram []colorValue

// ColorCdf is a cumulative distribution function (CDF) of colors in a image.
// The value field is the less than or equal to the probability of this color.
type ColorCdf []colorValue

type colorValue struct {
	Color color.Color // The color.
	Value float64     // The value.
}

func (x colorValue) String() string {
	r, g, b, _ := x.Color.RGBA()
	return fmt.Sprintf("(r:%d g:%d b:%d: %f)", r, g, b, x.Value)
}

// NewColorHistogram returns the ColorHistogram for the given image.
func NewColorHistogram(img image.Image) ColorHistogram {
	bounds := img.Bounds()

	m := make(map[color.Color]float64)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			m[img.At(x, y)]++
		}
	}

	// Normalise by the number of pixels
	area := float64(img.Bounds().Dx() * img.Bounds().Dy())
	return mapToHistogram(m, area)
}

// Combine combines two histograms together returning the result.
func (h ColorHistogram) Combine(hist ColorHistogram) ColorHistogram {
	m := make(map[color.Color]float64)
	for _, x := range h {
		m[x.Color] += x.Value
	}
	for _, x := range hist {
		m[x.Color] += x.Value
	}

	// Assumes both histograms are already normalised, and weigh them equally.
	return mapToHistogram(m, 2)
}

// Cdf returns the cumulative distribution function (CDF) for this histogram.
func (h ColorHistogram) Cdf() ColorCdf {
	// Make a copy.
	cdf := make([]colorValue, len(h))
	copy(cdf, h)

	// Sort by the probability of each color.
	sort.Slice(cdf, func(i, j int) bool {
		return cdf[i].Value < cdf[j].Value
	})

	// Convert the value to a running total.
	total := float64(0)
	for i := range cdf {
		cdf[i].Value += total
		total = cdf[i].Value
	}

	return cdf
}

func mapToHistogram(m map[color.Color]float64, factor float64) ColorHistogram {
	hist := make([]colorValue, 0, len(m))
	for color, count := range m {
		hist = append(hist, colorValue{Color: color, Value: count / factor})
	}
	return ColorHistogram(hist)
}

// Rand returns a random value from the cdf, chosen based on the weight of the values.
func (cdf ColorCdf) Rand(random *rand.Rand) color.Color {
	p := random.Float64()
	i := sort.Search(len(cdf), func(i int) bool {
		return cdf[i].Value >= p
	})
	if i >= len(cdf) {
		// TODO REMOVE
		panic("This should not happen")
	}
	return cdf[i].Color
}
