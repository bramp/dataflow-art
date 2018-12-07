package palette

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"sort"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func newRGBImage(width, height int, colors ...color.Color) image.Image {
	bounds := image.Rect(0, 0, width, height)
	img := image.NewRGBA(bounds)

	// Fill the image with equal sized bands of colors
	for i, c := range colors {
		rect := image.Rect((i*width)/len(colors), 0, ((i+1)*width)/len(colors), height)
		draw.Draw(img, rect, &image.Uniform{c}, image.ZP, draw.Src)
	}

	return img
}

var (
	black = RGBA{0, 0, 0, 255}
	white = RGBA{255, 255, 255, 255}

	red   = RGBA{255, 0, 0, 255}
	green = RGBA{0, 255, 0, 255}
	blue  = RGBA{0, 0, 255, 255}
)

type fakeRand struct {
	R float64
}

func (r *fakeRand) Float64() float64 {
	// chosen by fair dice roll.
	// guaranteed to be random.
	return r.R
}

// sortColors sorts the bins so we get repeatable results.
func sortBins(values []bin) []bin {
	sort.Slice(values, func(i, j int) bool {
		return values[i].String() < values[j].String()
	})
	return values
}

func TestNewColorHistogram(t *testing.T) {
	tests := []struct {
		desc string
		img  image.Image
		want Histogram
	}{
		{
			desc: "0x0 image",
			img:  newRGBImage(0, 0, black),
			want: []bin{},
		},
		{
			desc: "100x100 black image",
			img:  newRGBImage(150, 150, black),
			want: []bin{{black.Int(), 1.0}},
		},
		{
			desc: "100x100 black/white image",
			img:  newRGBImage(150, 150, black, white),
			want: []bin{{black.Int(), 0.5}, {white.Int(), 0.5}},
		},
		{
			desc: "100x100 red/green/blue image",
			img:  newRGBImage(150, 150, red, green, blue),
			want: []bin{{blue.Int(), 1.0 / 3}, {green.Int(), 1.0 / 3}, {red.Int(), 1.0 / 3}},
		},
	}

	for _, test := range tests {
		got := NewColorHistogram(test.img)
		if diff := pretty.Compare(sortBins(got), sortBins(test.want)); diff != "" {
			t.Errorf("NewColorHistogram(%q): diff: (-got +want)\n%s", test.desc, diff)
		}
	}
}

func TestHistogram_Cdf(t *testing.T) {
	tests := []struct {
		x    float64
		want int
	}{
		{
			x:    0.0,
			want: blue.Int(),
		},
		{
			x:    0.5,
			want: green.Int(),
		},
		{
			x:    math.Nextafter(1.0, 0.0), // Just less than 1.0
			want: red.Int(),
		},
	}

	// Don't allow empty CDFs
	var hist ColorHistogram
	if _, err := hist.Cdf(); err != errEmptyCdf {
		t.Errorf("hist.Cdf(%s) err = %s, want %s", hist, err, errEmptyCdf)
	}

	hist = ColorHistogram([]bin{{blue.Int(), 1.0 / 3}, {green.Int(), 1.0 / 3}, {red.Int(), 1.0 / 3}})
	cdf, err := hist.Cdf()
	if err != nil {
		t.Fatalf("hist.Cdf(%s) err = %q, want nil", hist, err)
	}

	for _, test := range tests {
		got := cdf.Rand(&fakeRand{test.x})
		if got != test.want {
			t.Errorf("cdf.Rand(%f) = 0x%x want 0x%x", test.x, got, test.want)
		}
	}
}

func TestHistogram_Combine(t *testing.T) {
	tests := []struct {
		desc string
		imgs []image.Image
		want Histogram
	}{
		{
			desc: "black and black",
			imgs: []image.Image{newRGBImage(1, 1, black), newRGBImage(1, 1, black)},
			want: []bin{{black.Int(), 1.0}},
		},
		{
			desc: "black, black and black",
			imgs: []image.Image{newRGBImage(1, 1, black), newRGBImage(1, 1, black), newRGBImage(1, 1, black)},
			want: []bin{{black.Int(), 1.0}},
		},
		{
			desc: "red and green",
			imgs: []image.Image{newRGBImage(1, 1, red), newRGBImage(1, 1, green)},
			want: []bin{{green.Int(), 0.5}, {red.Int(), 0.5}},
		},
		{
			desc: "red, green and blue",
			imgs: []image.Image{newRGBImage(1, 1, red), newRGBImage(1, 1, green), newRGBImage(1, 1, blue)},
			want: []bin{{blue.Int(), 1.0 / 3}, {green.Int(), 1.0 / 3}, {red.Int(), 1.0 / 3}},
		},
	}

	for _, test := range tests {
		got := NewColorHistogram(test.imgs[0])
		for _, img := range test.imgs[1:] {
			got = got.Combine(NewColorHistogram(img))
		}
		got = got.Normalise()
		if diff := pretty.Compare(sortBins(got), sortBins(test.want)); diff != "" {
			t.Errorf("NewColorHistogram.Combine(%q): diff: (-got +want)\n%s", test.desc, diff)
		}
	}
}
