package main

import (
	"image"
	"os"
	"path/filepath"
	"testing"
)

// The radius comes out of the drawing rather than being chosen here, so the
// raster is the shape the drawing asks for.
func TestTheTileRadiusComesFromTheDrawing(t *testing.T) {
	svg := []byte(`<svg viewBox="0 0 512 512"><rect width="512" height="512" rx="116" fill="#4A48E6"/></svg>`)

	radius, ok := tileRadius(svg)
	if !ok {
		t.Fatal("no radius found")
	}
	if want := 116.0 / 512.0; radius != want {
		t.Errorf("radius = %v, want %v", radius, want)
	}
}

func TestASquareTileHasNoRadius(t *testing.T) {
	svg := []byte(`<svg viewBox="0 0 512 512"><rect width="512" height="512" fill="#4A48E6"/></svg>`)
	if _, ok := tileRadius(svg); ok {
		t.Error("a rect with no rx should report no radius")
	}
	if _, ok := tileRadius([]byte(`<svg></svg>`)); ok {
		t.Error("a drawing with no rect should report no radius")
	}
}

// The rasteriser draws a rect's corners square whatever its radius, so the
// rounding is applied to the pixels afterwards. This is the difference between
// the tile looking like an app icon and looking like a sticker.
func TestRoundingClearsTheCornersAndNothingElse(t *testing.T) {
	const size = 100
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for i := range img.Pix {
		img.Pix[i] = 0xFF // opaque white everywhere
	}

	roundCorners(img, 20)

	if _, _, _, a := img.At(0, 0).RGBA(); a != 0 {
		t.Error("the top-left corner was not cleared")
	}
	if _, _, _, a := img.At(size-1, size-1).RGBA(); a != 0 {
		t.Error("the bottom-right corner was not cleared")
	}
	if _, _, _, a := img.At(size/2, size/2).RGBA(); a == 0 {
		t.Error("the middle of the tile was cleared")
	}
	// The edges between the corners are part of the tile.
	if _, _, _, a := img.At(size/2, 0).RGBA(); a == 0 {
		t.Error("the top edge was cleared")
	}
	if _, _, _, a := img.At(0, size/2).RGBA(); a == 0 {
		t.Error("the left edge was cleared")
	}
	// And a corner just inside the radius stays.
	if _, _, _, a := img.At(20, 20).RGBA(); a == 0 {
		t.Error("the pixel at the corner's own centre was cleared")
	}
}

// End to end, against the drawing that actually ships.
func TestTheIconDrawsToAPNG(t *testing.T) {
	out := filepath.Join(t.TempDir(), "icon.png")
	if err := run(filepath.Join("..", "..", "..", "icon.svg"), out, 128); err != nil {
		t.Fatal(err)
	}

	file, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		t.Fatal(err)
	}
	if format != "png" {
		t.Errorf("wrote a %s", format)
	}
	if got := img.Bounds().Dx(); got != 128 {
		t.Errorf("drew %d pixels wide, want 128", got)
	}

	// The tile is blue in the middle and clear at the corner.
	r, g, b, a := img.At(64, 100).RGBA()
	if a == 0 || b <= r || b <= g {
		t.Errorf("the middle of the tile is %v, want an opaque blue", img.At(64, 100))
	}
	if _, _, _, corner := img.At(0, 0).RGBA(); corner != 0 {
		t.Error("the corner was left square")
	}
}
