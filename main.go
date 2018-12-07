package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"log"
	"os"
	"path/filepath"
	"strings"

	"bramp.net/dataflow-art/palette"
)

const (
	paletteColors = 6
	paletteSuffix = "_pal"
	artPrefix     = "art"
)

func clusterImage(filename string) error {
	reader, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer reader.Close()

	img, _, err := image.Decode(reader)
	if err != nil {
		return fmt.Errorf("failed to decode %q: %s", filename, err)
	}

	log.Printf("Clustering %s %s", filename, img.Bounds().Size())

	hist := palette.NewColorHistogram(img)

	extractor := palette.Extractor{
		ColorSpace: palette.Lab, // or palette.RGB
		Samples:    10000,
	}
	/*
		pal, err := extractor.FromImage(img, paletteColors)
		if err != nil {
			return fmt.Errorf("failed to create paletter for %q: %s", filename, err)
		}
	*/

	pal, err := extractor.FromHistogram(hist, paletteColors)
	if err != nil {
		return fmt.Errorf("failed to extract palette for %q: %s", filename, err)
	}

	f, err := os.Create(filename + paletteSuffix + ".png")
	if err != nil {
		return err
	}
	defer f.Close()

	return palette.DrawColorPalette(f, pal)
}

func main() {
	err := filepath.Walk(artPrefix, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(info.Name(), ".jpg") {
			return clusterImage(path)
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}
