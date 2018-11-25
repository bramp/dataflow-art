// Package palette provides tools to extract color palettes from images.
package palette

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
)

func (p ColorPalette) String() string {
	var sb strings.Builder
	for _, c := range p {
		r, g, b, _ := c.RGBA()
		fmt.Fprintf(&sb, "r:%d g:%d b:%d", r, g, b)
	}
	return sb.String()
}

type colorSpace interface {
	ColorToObservation(c color.Color) clusters.Observation
	ClustersToColor(c clusters.Cluster) color.Color
}

type lab struct{}
type rgb struct{}

// Lab represents the CIE L*a*b* color space, which provides a more accurate
// representation of what humans see.
var Lab = lab{}

// RGB represents the RGB color space.
var RGB = rgb{}

func (lab) ClustersToColor(c clusters.Cluster) color.Color {
	return colorful.Lab(c.Center[0], c.Center[1], c.Center[2])
}

func (rgb) ClustersToColor(c clusters.Cluster) color.Color {
	return color.RGBA{
		R: uint8(c.Center[0] * 255),
		G: uint8(c.Center[1] * 255),
		B: uint8(c.Center[2] * 255),
		A: 255,
	}
}

func (lab) ColorToObservation(c color.Color) clusters.Observation {
	cc, ok := colorful.MakeColor(c)
	if !ok {
		panic(fmt.Sprintf("Unable to colorful.MakeColor(%s)", c))
	}

	l, a, b := cc.Lab()
	return clusters.Coordinates{
		l, a, b,
	}
}

func (rgb) ColorToObservation(c color.Color) clusters.Observation {
	r, g, b, _ := c.RGBA()
	return clusters.Coordinates{
		// Map the RGB values between 0.0 to 1.0
		float64(r) / 65535.0,
		float64(g) / 65535.0,
		float64(b) / 65535.0,
	}
}

// Extractor extracts a Color Palette from a image.
type Extractor struct {
	ColorSpace colorSpace // Color Space to use, either palette.RGB or palette.Lab
	Samples    int        // Number of pixels to sample. 0 means all pixels.

	random *rand.Rand // Source of random numbers (useful for testing)
}

// FromImage returns the palette of n colors for the given image.
func (p Extractor) FromImage(img image.Image, n int) (ColorPalette, error) {
	var d clusters.Observations
	if p.Samples == 0 {
		d = p.allObservations(img)
	} else {
		d = p.sampledObservations(img)
	}

	return p.fromObservations(d, n)
}

// FromHistogram returns the palette of n colors for the given ColorHistogram.
func (p Extractor) FromHistogram(hist ColorHistogram, n int) (ColorPalette, error) {
	return p.fromObservations(p.histogramObservations(hist), n)
}

func (p Extractor) fromObservations(d clusters.Observations, n int) (ColorPalette, error) {
	km, err := kmeans.NewWithOptions(0.0001, 1000, nil)
	if err != nil {
		return nil, err
	}

	clusters, err := km.Partition(d, n)
	if err != nil {
		return nil, err
	}
	return p.clustersToPalette(clusters), nil
}

func (p Extractor) allObservations(img image.Image) clusters.Observations {
	bounds := img.Bounds()

	// Use every pixel
	d := make(clusters.Observations, 0, bounds.Dx()*bounds.Dy())
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			d = append(d, p.ColorSpace.ColorToObservation(img.At(x, y)))
		}
	}
	return d
}

func (p Extractor) sampledObservations(img image.Image) clusters.Observations {
	if p.random == nil {
		p.random = rand.New(rand.NewSource(0))
	}

	bounds := img.Bounds()

	minX, maxX := bounds.Min.X, bounds.Max.X
	minY, maxY := bounds.Min.Y, bounds.Max.Y

	// Sample the pixels with replacement
	d := make(clusters.Observations, p.Samples)
	for i := 0; i < p.Samples; i++ {
		x := p.random.Intn(maxX-minX) + minX
		y := p.random.Intn(maxY-minY) + minY

		d[i] = p.ColorSpace.ColorToObservation(img.At(x, y))
	}
	return d
}

func (p Extractor) histogramObservations(hist ColorHistogram) clusters.Observations {
	if p.random == nil {
		p.random = rand.New(rand.NewSource(0))
	}

	cdf := hist.Cdf()

	// Pick samples from the histogram
	d := make(clusters.Observations, p.Samples)
	for i := 0; i < p.Samples; i++ {
		d[i] = p.ColorSpace.ColorToObservation(cdf.Rand(p.random))
	}

	return d
}

func (p Extractor) clustersToPalette(clusters clusters.Clusters) ColorPalette {
	var pal ColorPalette
	for _, c := range clusters {
		pal = append(pal, p.ColorSpace.ClustersToColor(c))
	}
	return pal
}
