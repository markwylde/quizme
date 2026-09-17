package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// tickMark is the "this one is settled" mark on a question's header.
//
// It is drawn rather than set in type or taken from an icon sheet: the embedded
// Inter faces are not guaranteed to carry a check glyph, and a themed SVG
// resource caches its raster under a name that does not change when the theme
// variant does -- which is exactly the case this has to survive. Two strokes
// resolved through the theme on every refresh cost less than either.
type tickMark struct {
	widget.BaseWidget

	short *canvas.Line
	long  *canvas.Line
}

const (
	// tickSize is the mark's extent. It sits beside caption-sized text, so it
	// is scaled to that rather than to a button.
	tickSize = 15
	// tickStroke is thick enough to read as a deliberate mark at tickSize.
	tickStroke = 2
)

func newTickMark() *tickMark {
	t := &tickMark{}
	t.ExtendBaseWidget(t)
	return t
}

func (t *tickMark) CreateRenderer() fyne.WidgetRenderer {
	t.short = canvas.NewLine(t.colour())
	t.long = canvas.NewLine(t.colour())
	t.short.StrokeWidth = scaled(tickStroke)
	t.long.StrokeWidth = scaled(tickStroke)
	return &tickRenderer{mark: t, objects: []fyne.CanvasObject{t.short, t.long}}
}

func (t *tickMark) MinSize() fyne.Size {
	t.ExtendBaseWidget(t)
	return fyne.NewSize(scaled(tickSize), scaled(tickSize))
}

// Refresh re-resolves the stroke colour, so a form open while the desktop
// switches between light and dark redraws the mark in the right green rather
// than keeping the one it was built with.
func (t *tickMark) Refresh() {
	for _, line := range []*canvas.Line{t.short, t.long} {
		if line == nil {
			continue
		}
		line.StrokeColor = t.colour()
		line.StrokeWidth = scaled(tickStroke)
		line.Refresh()
	}
	t.BaseWidget.Refresh()
}

// colour is the theme's success colour, which the form's own theme answers from
// the palette for the current variant.
func (t *tickMark) colour() color.Color { return theme.Color(theme.ColorNameSuccess) }

type tickRenderer struct {
	mark    *tickMark
	objects []fyne.CanvasObject
}

// Layout draws the mark in proportions of whatever square it is given, so it
// stays a tick at any size.
func (r *tickRenderer) Layout(size fyne.Size) {
	side := fyne.Min(size.Width, size.Height)
	offsetX := (size.Width - side) / 2
	offsetY := (size.Height - side) / 2

	at := func(fx, fy float32) fyne.Position {
		return fyne.NewPos(offsetX+side*fx, offsetY+side*fy)
	}

	r.mark.short.Position1 = at(0.18, 0.54)
	r.mark.short.Position2 = at(0.42, 0.78)
	r.mark.long.Position1 = at(0.42, 0.78)
	r.mark.long.Position2 = at(0.84, 0.24)
}

func (r *tickRenderer) MinSize() fyne.Size { return r.mark.MinSize() }

func (r *tickRenderer) Refresh() {
	r.Layout(r.mark.Size())
	for _, o := range r.objects {
		o.Refresh()
	}
}

func (r *tickRenderer) Objects() []fyne.CanvasObject { return r.objects }

func (r *tickRenderer) Destroy() {}
