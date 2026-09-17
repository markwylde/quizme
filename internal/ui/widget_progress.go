package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// progressBar shows a proportion as a filled track.
//
// Fyne's own ProgressBar paints its track in a faded version of the primary
// colour, so an empty bar reads as a partly full one -- misleading at exactly
// the point the reading matters. This one has a neutral track, so empty looks
// empty.
type progressBar struct {
	widget.BaseWidget

	value float64 // 0 to 1

	track *canvas.Rectangle
	fill  *canvas.Rectangle
}

// progressHeight is the bar's thickness. Slim enough to sit in a footer beside
// text without becoming the loudest thing in it.
const progressHeight = 6

func newProgressBar() *progressBar {
	p := &progressBar{}
	p.ExtendBaseWidget(p)
	return p
}

func (p *progressBar) CreateRenderer() fyne.WidgetRenderer {
	p.track = canvas.NewRectangle(theme.Color(theme.ColorNameInputBorder))
	p.fill = canvas.NewRectangle(theme.Color(theme.ColorNamePrimary))
	return &progressRenderer{bar: p}
}

// SetValue sets the proportion filled, clamped to the range it can draw.
func (p *progressBar) SetValue(v float64) {
	switch {
	case v < 0:
		v = 0
	case v > 1:
		v = 1
	}
	if p.value == v {
		return
	}
	p.value = v
	p.Refresh()
}

// Value reports the proportion currently filled.
func (p *progressBar) Value() float64 { return p.value }

func (p *progressBar) MinSize() fyne.Size {
	p.ExtendBaseWidget(p)
	return fyne.NewSize(scaled(110), scaled(progressHeight))
}

type progressRenderer struct {
	bar *progressBar
}

func (r *progressRenderer) Layout(size fyne.Size) {
	radius := scaled(progressHeight) / 2
	r.bar.track.CornerRadius = radius
	r.bar.fill.CornerRadius = radius
	r.bar.track.Resize(size)
	r.bar.track.Move(fyne.NewPos(0, 0))
	r.bar.fill.Resize(fyne.NewSize(size.Width*float32(r.bar.value), size.Height))
	r.bar.fill.Move(fyne.NewPos(0, 0))
}

func (r *progressRenderer) MinSize() fyne.Size { return r.bar.MinSize() }

func (r *progressRenderer) Refresh() {
	r.bar.track.FillColor = theme.Color(theme.ColorNameInputBorder)
	r.bar.fill.FillColor = theme.Color(theme.ColorNamePrimary)
	r.Layout(r.bar.Size())
	r.bar.track.Refresh()
	r.bar.fill.Refresh()
}

func (r *progressRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bar.track, r.bar.fill}
}

func (r *progressRenderer) Destroy() {}
