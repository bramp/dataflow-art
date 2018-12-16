// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// TODO
// * Export the metrics correctly (Reshuffle can kinda do it)
// * Use new GCS package

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"bramp.net/dataflow-art/palette"
	"bramp.net/morebeam"
	"bramp.net/morebeam/csvio"

	"github.com/apache/beam/sdks/go/pkg/beam"
	"github.com/apache/beam/sdks/go/pkg/beam/io/filesystem"
	"github.com/apache/beam/sdks/go/pkg/beam/io/textio"
	"github.com/apache/beam/sdks/go/pkg/beam/log"
	"github.com/apache/beam/sdks/go/pkg/beam/x/beamx"

	"net/http"
	_ "net/http/pprof"
)

const paletteColors = 5

var (
	artPrefix    = flag.String("art", "art", "Path to where the art is kept.")
	index        = flag.String("index", "art/all_data_info_sample.csv", "Index of the art.")
	outputPrefix = flag.String("output", "output", "Output file (required).")
)

func init() {
	beam.RegisterType(reflect.TypeOf((*drawColorPaletteFn)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*extractHistogramFn)(nil)).Elem())

	beam.RegisterType(reflect.TypeOf(Painting{}))
	beam.RegisterType(reflect.TypeOf(HistogramResult{}))

	beam.RegisterType(reflect.TypeOf(palette.ColorHistogram{}))
	beam.RegisterType(reflect.TypeOf(palette.RGBA{}))
}

// Painting represents a single painting in the dataset.
type Painting struct {
	Artist string `csv:"artist"`
	Title  string `csv:"title"`
	Date   string `csv:"date"`
	Genre  string `csv:"genre"`
	Style  string `csv:"style"`

	Filename string `csv:"new_filename"`

	PixelsX   string `csv:"pixelsx"`
	PixelsY   string `csv:"pixelsy"`
	SizeBytes string `csv:"size_bytes"`

	Source      string `csv:"source"`
	ArtistGroup string `csv:"artist_group"`
	InTrain     string `csv:"in_train"`
}

// Decade returns the decade for the given painting. The function returns the decade
// as string (e.g 2010), or the value "".
func (p *Painting) Decade() string {
	year, err := p.Year()
	if err != nil {
		return ""
	}

	decade := int64(math.Floor(float64(year)/10.0)) * 10
	return strconv.FormatInt(decade, 10)
}

// Year returns the year this painting was painted.
// TODO Check that dates expressed as `2000-2010` work.
func (p *Painting) Year() (int, error) {
	// The date field can be expressed in a few different formats, e.g `2000`, `c.2000` or `2000.0`.
	date := strings.TrimSpace(p.Date)
	date = strings.TrimPrefix(date, "c.") // Some dates are formatted `c.2000`
	f, err := strconv.ParseFloat(date, 64)
	return int(f), err
}

// HistogramResult TODO
type HistogramResult struct {
	Key       string                 `json:"key"`
	Files     []string               `json:"files"`
	Histogram palette.ColorHistogram `json:"histogram,omitempty"`
	Palette   []palette.RGBA         `json:"palette"`
}

type extractHistogramFn struct {
	ArtPrefix string `json:"art_prefix"`

	fs filesystem.Interface
}

// ExtractHistogram calculates the color histograms for all the Paintings in the CoGBK.
func ExtractHistogram(s beam.Scope, files beam.PCollection) (beam.PCollection, beam.PCollection) {
	s = s.Scope("ExtractHistogram")
	return beam.ParDo2(s, &extractHistogramFn{
		ArtPrefix: *artPrefix,
	}, files)
}

func (fn *extractHistogramFn) Setup(ctx context.Context) error {
	var err error
	fn.fs, err = filesystem.New(ctx, fn.ArtPrefix)
	if err != nil {
		return fmt.Errorf("filesystem.New(%q) failed: %s", fn.ArtPrefix, err)
	}
	return nil
}

func (fn *extractHistogramFn) Teardown() error {
	return fn.fs.Close()
}

func (fn *extractHistogramFn) ProcessElement(ctx context.Context, key string, values func(*Painting) bool, errors func(string, string)) HistogramResult {
	log.Infof(ctx, "%q: ExtractHistogram started", key)

	result := HistogramResult{
		Key: key,
	}

	var art Painting
	for values(&art) {
		filename := morebeam.Join(fn.ArtPrefix, art.Filename)
		h, err := fn.extractHistogram(ctx, key, filename)
		if err != nil {
			log.Warnf(ctx, "%q: ExtractHistogram failed: %s", key, err)
			errors(key, err.Error())
			continue
		}

		log.Debugf(ctx, "%q: ExtractHistogram combine %q", key, filename)
		result.Histogram = result.Histogram.Combine(h)
		result.Files = append(result.Files, art.Filename)

		runtime.GC() // Hack to reduce memory usage
	}

	log.Infof(ctx, "%q: ExtractHistogram ended %s", key, result.Histogram)

	return result
}

func (fn *extractHistogramFn) extractHistogram(ctx context.Context, key, filename string) (palette.ColorHistogram, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	log.Debugf(ctx, "%q: ExtractHistogram read %q", key, filename)
	fd, err := fn.fs.OpenRead(ctx, filename)
	if err != nil {
		return nil, fmt.Errorf("fs.OpenRead(%q) failed: %s", filename, err)
	}
	defer fd.Close()

	log.Debugf(ctx, "%q: ExtractHistogram decode %q", key, filename)
	img, _, err := image.Decode(fd)
	if err != nil {
		return nil, fmt.Errorf("image.Decode(%q) failed: %s", filename, err)
	}

	log.Debugf(ctx, "%q: ExtractHistogram histogram %q", key, filename)
	return palette.NewColorHistogram(img), nil
}

func CalculateColorPalette(s beam.Scope, hists beam.PCollection) (beam.PCollection, beam.PCollection) {
	s = s.Scope("CalculateColorPalette")
	return beam.ParDo2(s, calculateColorPaletteFn, hists)
}

func calculateColorPaletteFn(ctx context.Context, result HistogramResult, emit func(HistogramResult), errors func(string, string)) {
	if len(result.Histogram) == 0 {
		// Bail out if the histogram is empty (because no images were successfully processed).
		return
	}

	log.Infof(ctx, "%q: CalculateColorPalette: %s", result.Key, result.Histogram)

	extractor := palette.Extractor{
		ColorSpace: palette.Lab, // or palette.RGB
		Samples:    10000,
	}

	pal, err := extractor.FromHistogram(result.Histogram, paletteColors)
	if err != nil {
		log.Warnf(ctx, "%q: CalculateColorPalette failed: %s", result.Key, err)
		errors(result.Key, err.Error())
		return
	}

	result.Palette = pal.ToRGBA() // Convert to a type that can be serialised.
	emit(result)
}

// PCollection<HistogramResult> -> nothing
func DrawColorPalette(s beam.Scope, outputPrefix string, palettes beam.PCollection) beam.PCollection {
	s = s.Scope("DrawColorPalette")
	fn := &drawColorPaletteFn{
		OutputPrefix: outputPrefix,
	}
	return beam.ParDo(s, fn, palettes)
}

// drawColorPaletteFn is a DoFn which contains some initialisation state. This struct will be copied
// to the remote machines, and thus must be serialisable. Apache Beam choses to use json encoding to
// do that.
type drawColorPaletteFn struct {
	OutputPrefix string `json:"output_prefix"`

	fs filesystem.Interface
}

func (fn *drawColorPaletteFn) Setup(ctx context.Context) error {
	var err error
	fn.fs, err = filesystem.New(ctx, fn.OutputPrefix)
	if err != nil {
		return fmt.Errorf("filesystem.New(%q) failed: %s", fn.OutputPrefix, err)
	}
	return nil
}

func (fn *drawColorPaletteFn) Teardown() error {
	return fn.fs.Close()
}

// ProcessElement saves the palette to a png file with the given filename. It emits the name of the file.
func (fn *drawColorPaletteFn) ProcessElement(ctx context.Context, result HistogramResult, errors func(string, string)) {
	filename := morebeam.Join(fn.OutputPrefix, result.Key+".png")

	log.Infof(ctx, "%q: DrawColorPalette for %q", result.Key, filename)
	if err := fn.drawColorPalette(ctx, filename, result.Palette); err != nil {
		log.Warnf(ctx, "%q: DrawColorPalette failed: %s", result.Key, err)
		errors(result.Key, err.Error())
	}
}

func (fn *drawColorPaletteFn) drawColorPalette(ctx context.Context, filename string, rgba []palette.RGBA) error {
	fd, err := fn.fs.OpenWrite(ctx, filename)
	if err != nil {
		return fmt.Errorf("fs.OpenWrite(%q) failed: %s", filename, err)
	}
	defer fd.Close()

	if err := palette.FromRGBA(rgba).Draw(fd); err != nil {
		return fmt.Errorf("palette.Draw(%q) failed: %s", filename, err)
	}

	return nil
}

func WriteIndex(s beam.Scope, indexFilename string, results beam.PCollection) {
	s = s.Scope("WriteIndex")

	objs := beam.ParDo(s.Scope("MarshalToJson"), func(result HistogramResult) (string, error) {
		result.Histogram = nil
		value, err := json.Marshal(result)
		if err != nil {
			return "", err
		}

		return string(value), nil
	}, results)

	textio.Write(s, indexFilename, objs)
}

// GeneratePalette TODO
func GeneratePalette(s beam.Scope, group string, keyFn func(Painting) string, paintings beam.PCollection) beam.PCollection {
	s = s.Scope(fmt.Sprintf("GeneratePalette %q", group))
	outputPrefix := morebeam.Join(*outputPrefix, group)

	// Extract the key we want to group on.
	// PCollection<Painting> -> PCollection<KV<string, Painting>>
	paintingsWithKey := morebeam.AddKey(s, keyFn, paintings)

	// Filter out empty keys
	paintingsByGroup := beam.ParDo(s.Scope("FilterEmptyKeys"), func(key string, art Painting, emit func(key string, art Painting)) {
		if key != "" {
			emit(key, art)
		}
	}, paintingsWithKey)

	// PCollection<string, Painting> -> PCollection<CoGBK<string, Painting>>
	paintingsByGroup = beam.GroupByKey(s.Scope("GroupBy"), paintingsByGroup)

	// PCollection<CoGBK<string, Painting>> -> PCollection<KV<string, ColorHistogram>>
	histograms, errors1 := ExtractHistogram(s, paintingsByGroup)

	// Calculate the color palette for the combined histograms.
	// PCollection<KV<string, ColorHistogram>> -> PCollection<KV<string, []color.RGBA>>
	palettes, errors2 := CalculateColorPalette(s, histograms)

	// PCollection<string, []color.RGBA> -> PCollection<string>
	//errors3 := DrawColorPalette(s, outputPrefix, palettes)

	// PCollection<KV<string, []color.RGBA>> -> nothing
	WriteIndex(s, morebeam.Join(outputPrefix, "index.json"), palettes)

	return beam.Flatten(s.Scope("Flatten Errors"), errors1, errors2 /*, errors3*/)
}

// WriteErrorLog takes multiple PCollection<KV<string,string>>s combines them and
// writes them to the given filename.
func WriteErrorLog(s beam.Scope, filename string, errors ...beam.PCollection) {
	s = s.Scope(fmt.Sprintf("Write %q", filename))

	c := beam.Flatten(s, errors...)
	c = beam.ParDo(s, func(key, value string) string {
		return fmt.Sprintf("%s,%s", key, value)
	}, c)

	textio.Write(s, morebeam.Join(*outputPrefix, filename), c)
}

func buildPipeline2(p *beam.Pipeline, s beam.Scope) {

	paintings := csvio.Read(s, *index, reflect.TypeOf(Painting{}))
	//paintings = morebeam.Reshuffle(s, paintings)

	/*
		errors1 := GeneratePalette(s, "filename", func(art Painting) string {
			return strings.TrimSpace(art.Filename)
		}, paintings)
	*/

	errors1 := GeneratePalette(s, "year", func(art Painting) string {
		year, err := art.Year()
		if err != nil {
			return ""
		}
		return strconv.Itoa(year)
	}, paintings)

	errors2 := GeneratePalette(s, "artist", func(art Painting) string {
		return strings.TrimSpace(art.Artist)
	}, paintings)

	errors3 := GeneratePalette(s, "decade", func(art Painting) string {
		return art.Decade()
	}, paintings)

	errors4 := GeneratePalette(s, "style", func(art Painting) string {
		return strings.TrimSpace(art.Style)
	}, paintings)

	WriteErrorLog(s, "errors.log", errors1, errors2, errors3, errors4)
}

// GroupByDecade takes a PCollection<Painting> and returns a PCollection<CoGBK<string, Painting>>
// of the paintings group by decade.
func GroupByDecade(s beam.Scope, paintings beam.PCollection) beam.PCollection {
	s = s.Scope("GroupBy Decade")

	// PCollection<Painting> -> PCollection<KV<string, Painting>>
	paintingsWithKey := morebeam.AddKey(s, func(art Painting) string {
		return art.Decade()
	}, paintings)

	// PCollection<string, Painting> -> PCollection<CoGBK<string, Painting>>
	return beam.GroupByKey(s, paintingsWithKey)
}

// Used in the simple example
func buildPipeline(s beam.Scope) {
	// nothing -> PCollection<Painting>
	paintings := csvio.Read(s, *index, reflect.TypeOf(Painting{}))

	// PCollection<Painting> -> PCollection<CoGBK<string, Painting>>
	paintingsByGroup := GroupByDecade(s, paintings)

	// PCollection<CoGBK<string, Painting>> -> (PCollection<KV<string, ColorHistogram>>, PCollection<KV<string, string>>)
	histograms, errors1 := ExtractHistogram(s, paintingsByGroup)

	// Calculate the color palette for the combined histograms.
	// PCollection<KV<string, ColorHistogram>> -> (PCollection<KV<string, []color.RGBA>>, PCollection<KV<string, string>>)
	palettes, errors2 := CalculateColorPalette(s, histograms)

	// PCollection<KV<string, []color.RGBA>> -> PCollection<KV<string, string>>
	errors3 := DrawColorPalette(s, *outputPrefix, palettes)

	// PCollection<KV<string, []color.RGBA>> -> nothing
	WriteIndex(s, morebeam.Join(*outputPrefix, "index.json"), palettes)

	// PCollection<KV<string, string>> -> nothing
	WriteErrorLog(s, "errors.log", errors1, errors2, errors3)
}

func main() {
	// If beamx or Go flags are used, flags must be parsed first.
	flag.Parse()

	// beam.Init() is an initialization hook that must called on startup. On
	// distributed runners, it is used to intercept control.
	beam.Init()

	ctx := context.Background()

	// Input validation is done as usual. Note that it must be after Init().
	if *outputPrefix == "" {
		log.Fatal(ctx, "--output not provided")
	}

	go func() {
		// HTTP Server for pprof (and other debugging)
		log.Info(ctx, http.ListenAndServe("localhost:8080", nil))
	}()

	p := beam.NewPipeline()
	s := p.Root()

	buildPipeline1(s)
	// buildPipeline2(p, s)

	if err := beamx.Run(ctx, p); err != nil {
		log.Fatalf(ctx, "Failed to execute job: %v", err)
	}
}
