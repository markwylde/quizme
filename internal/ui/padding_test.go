package ui

import (
	"image"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

// A card's padding has to look even, which is not the same as being even.
//
// Every measurement the layouts work in is a text box, and a text box carries
// the room a capital needs above the letters and the room a descender needs
// below. Pad a card equally and the eye -- which reads the space above the
// capitals against the space below the baseline -- sees the bottom as tight.
// So this test reads the rendered pixels rather than the boxes.
//
// The fixture's last lines carry no descenders, deliberately: a descender is
// ink the eye does not count as the edge of the text.
func TestACardsPaddingLooksEven(t *testing.T) {
	restore := useTheme(t, newTheme())
	defer restore()

	f, doc := build(t, `
title: t
questions:
  - id: folded
    type: select
    prompt: Which of these is NOT one of the three elements a claimant must establish?
    required: true
    options: [A duty of care, An intention to cause harm]
    answer: An intention to cause harm
  - id: open
    type: select
    prompt: A shop displays a jacket in its window. In contract law, what is the ruling?
    required: true
    options: [An offer, An invitation to treat]
`)
	content := f.build()
	win := test.NewWindow(content)
	defer win.Close()

	// Twice over, so the prompts have settled their wrapping before the pixels
	// are read: the first pass is what tells them how wide they are.
	win.Resize(fyne.NewSize(780, 900))
	win.Resize(fyne.NewSize(780, 901))
	win.Resize(fyne.NewSize(780, 900))

	img := win.Canvas().Capture()
	scale := float32(img.Bounds().Dx()) / 780

	for _, q := range doc.Questions {
		c := f.cards[q.ID]
		top, left := positionOn(content, c.card)
		size := c.card.Size()

		x0, x1 := int(left*scale)+1, int((left+size.Width)*scale)-1
		y0, y1 := int(top*scale), int((top+size.Height)*scale)

		first, last, ok := inkRows(img, x0, x1, y0, y1)
		if !ok {
			t.Fatalf("%s: the card drew nothing", q.ID)
		}

		above := float32(first-y0) / scale
		below := float32(y1-last-1) / scale
		if diff := above - below; diff > 2 || diff < -2 {
			t.Errorf("%s: %v of card above the text and %v below it", q.ID, above, below)
		}
	}
}

// positionOn is where an object sits within root.
func positionOn(root, target fyne.CanvasObject) (top, left float32) {
	var walk func(o fyne.CanvasObject, y, x float32) bool
	walk = func(o fyne.CanvasObject, y, x float32) bool {
		if o == nil {
			return false
		}
		y += o.Position().Y
		x += o.Position().X
		if o == target {
			top, left = y, x
			return true
		}
		switch p := o.(type) {
		case *fyne.Container:
			for _, child := range p.Objects {
				if walk(child, y, x) {
					return true
				}
			}
		case fyne.Widget:
			for _, child := range test.WidgetRenderer(p).Objects() {
				if walk(child, y, x) {
					return true
				}
			}
		}
		return false
	}
	walk(root, 0, 0)
	return top, left
}

// inkRows reports the first and last rows in the box holding a pixel unlike the
// card behind it.
func inkRows(img image.Image, x0, x1, y0, y1 int) (first, last int, ok bool) {
	for y := y0; y < y1; y++ {
		// The card's own colour, read from the far right of the row, where a
		// question draws nothing.
		br, bg, bb, _ := img.At(x1-2, y).RGBA()
		for x := x0; x < x1; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if channelDiff(r, br)+channelDiff(g, bg)+channelDiff(b, bb) > 0x6000 {
				if !ok {
					first, ok = y, true
				}
				last = y
				break
			}
		}
	}
	return first, last, ok
}

func channelDiff(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}
