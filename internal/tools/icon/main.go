// Command icon draws icon.svg into icon.png.
//
// The drawing is the source; the raster exists because the platform bundlers
// want one. macOS in particular takes its Dock tile from an application
// bundle's icon file, and the packager that builds the bundle cannot read an
// SVG. Rather than keep two drawings in step by hand, this makes the second
// from the first.
//
// It uses the same rasteriser Fyne itself draws SVGs with, so the icon in the
// Dock and the icon in the window come out of the same code.
//
// Usage:
//
//	go run ./internal/tools/icon [-size 1024] [-in icon.svg] [-out icon.png]
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"regexp"
	"strconv"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

func main() {
	in := flag.String("in", "icon.svg", "the drawing to read")
	out := flag.String("out", "icon.png", "the raster to write")
	size := flag.Int("size", 1024, "the width and height to draw, in pixels")
	flag.Parse()

	if err := run(*in, *out, *size); err != nil {
		fmt.Fprintln(os.Stderr, "icon:", err)
		os.Exit(1)
	}
}

// tileRadius is the corner radius of the drawing's outermost rectangle, as a
// fraction of its width, or false if it has no rounded tile to speak of.
func tileRadius(svg []byte) (float64, bool) {
	rect := regexp.MustCompile(`<rect[^>]*>`).Find(svg)
	if rect == nil {
		return 0, false
	}
	width, ok := attr(rect, "width")
	radius, hasRadius := attr(rect, "rx")
	if !ok || !hasRadius || width <= 0 {
		return 0, false
	}
	return radius / width, true
}

func attr(tag []byte, name string) (float64, bool) {
	m := regexp.MustCompile(name + `="([0-9.]+)"`).FindSubmatch(tag)
	if m == nil {
		return 0, false
	}
	v, err := strconv.ParseFloat(string(m[1]), 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// roundCorners clears the pixels outside a rounded rectangle of the given
// radius, one corner at a time.
func roundCorners(img *image.RGBA, radius float64) {
	if radius <= 0 {
		return
	}
	b := img.Bounds()
	centres := [4][2]float64{
		{float64(b.Min.X) + radius, float64(b.Min.Y) + radius},
		{float64(b.Max.X) - radius, float64(b.Min.Y) + radius},
		{float64(b.Min.X) + radius, float64(b.Max.Y) - radius},
		{float64(b.Max.X) - radius, float64(b.Max.Y) - radius},
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			px, py := float64(x)+0.5, float64(y)+0.5
			for i, c := range centres {
				// Only the quadrant beyond each centre is a corner.
				beyondX := (i%2 == 0 && px < c[0]) || (i%2 == 1 && px > c[0])
				beyondY := (i < 2 && py < c[1]) || (i >= 2 && py > c[1])
				if !beyondX || !beyondY {
					continue
				}
				dx, dy := px-c[0], py-c[1]
				if dx*dx+dy*dy > radius*radius {
					img.Set(x, y, color.RGBA{})
				}
			}
		}
	}
}

func run(in, out string, size int) error {
	if size <= 0 {
		return fmt.Errorf("size %d makes no image", size)
	}

	file, err := os.Open(in)
	if err != nil {
		return err
	}
	defer file.Close()

	drawing, err := oksvg.ReadIconStream(file)
	if err != nil {
		return fmt.Errorf("reading %s: %w", in, err)
	}
	drawing.SetTarget(0, 0, float64(size), float64(size))

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	scanner := rasterx.NewScannerGV(size, size, img, img.Bounds())
	drawing.Draw(rasterx.NewDasher(size, size, scanner), 1)

	// The rasteriser draws a rect's corners square, radius or no radius, so the
	// tile's own rounding has to be applied here. macOS does not round an app
	// icon for you -- the icon is expected to arrive the shape it wants to be.
	svg, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	if radius, ok := tileRadius(svg); ok {
		roundCorners(img, radius*float64(size))
	}

	// Written to a temporary file and moved into place, so a failure part way
	// through cannot leave a half-drawn icon behind.
	tmp, err := os.CreateTemp(".", ".icon-*.png")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if err := png.Encode(tmp, img); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), out); err != nil {
		return err
	}
	// CreateTemp is private by default; this is a build artefact meant to be
	// read by whatever picks it up.
	return os.Chmod(out, 0o644)
}
