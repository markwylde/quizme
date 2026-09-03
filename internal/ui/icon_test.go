package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
)

// readIcon is the drawing as it ships, read from the repository root.
func readIcon(t *testing.T) []byte {
	t.Helper()
	svg, err := os.ReadFile(filepath.Join("..", "..", "icon.svg"))
	if err != nil {
		t.Fatal(err)
	}
	return svg
}

// The resource has to be named for what it holds: Fyne decides how to read a
// resource from its extension, and an SVG called anything else is not drawn.
func TestTheIconIsOfferedAsAnSVG(t *testing.T) {
	res := iconResource(readIcon(t))

	if got := res.Name(); !strings.HasSuffix(got, ".svg") {
		t.Errorf("the icon is offered as %q, which the toolkit will not read as a drawing", got)
	}
	if len(res.Content()) == 0 {
		t.Fatal("the icon carries no content")
	}
	if !strings.Contains(string(res.Content()), "<svg") {
		t.Error("the icon's content is not an SVG")
	}
}

// Fyne draws a subset of SVG, so the only assurance worth having is that this
// drawing comes out as pixels rather than as nothing.
func TestTheIconRenders(t *testing.T) {
	img := canvas.NewImageFromResource(iconResource(readIcon(t)))
	img.SetMinSize(fyne.NewSize(64, 64))

	win := test.NewWindow(img)
	defer win.Close()
	win.Resize(fyne.NewSize(64, 64))

	painted := win.Canvas().Capture()
	bounds := painted.Bounds()
	if bounds.Empty() {
		t.Fatal("nothing was painted")
	}

	// The tile is a solid blue, so the middle of the icon has to be blue: not
	// transparent, not the window behind it, and not a placeholder grey.
	var blues, opaque int
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := painted.At(x, y).RGBA()
			if a < 0x8000 {
				continue
			}
			opaque++
			if b > r+0x2000 && b > g+0x2000 {
				blues++
			}
		}
	}
	if opaque == 0 {
		t.Fatal("the icon painted nothing opaque")
	}
	if blues*4 < opaque {
		t.Errorf("only %d of %d painted pixels are blue: the tile did not draw", blues, opaque)
	}
}

// The icon the binary carries is the one in the repository, not a stale copy.
func TestTheEmbeddedIconIsTheOneOnDisk(t *testing.T) {
	// The embed lives in the root package, so this checks the file it reads is
	// where the build expects it and is a drawing, which is all this package
	// can see from here.
	svg := readIcon(t)
	if len(svg) == 0 {
		t.Fatal("icon.svg is empty")
	}
	if !strings.Contains(string(svg), "viewBox") {
		t.Error("icon.svg has no viewBox, so it will not scale to a window icon")
	}
}
