package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// band paints a question's background tint behind its content.
//
// It is a widget rather than a bare canvas.Rectangle so that the colour is
// re-read on every refresh. Fyne refreshes widgets when the theme variant
// changes, so a form open while the desktop switches between light and dark
// picks up the right tints instead of keeping the ones it started with.
type band struct {
	widget.BaseWidget

	tint    int
	content fyne.CanvasObject
	rect    *canvas.Rectangle
}

func newBand(tint int, content fyne.CanvasObject) *band {
	b := &band{tint: tint, content: content}
	b.ExtendBaseWidget(b)
	return b
}

func (b *band) CreateRenderer() fyne.WidgetRenderer {
	b.rect = canvas.NewRectangle(b.colour())
	return widget.NewSimpleRenderer(container.NewStack(b.rect, b.content))
}

// SetTint changes which colour of the cycle this band paints.
func (b *band) SetTint(tint int) {
	if b.tint == tint {
		return
	}
	b.tint = tint
	b.Refresh()
}

// Tint reports the band's position in the colour cycle.
func (b *band) Tint() int { return b.tint }

// Content is what the band paints behind. A widget's children are otherwise
// reachable only through its renderer, which makes the question inside a band
// invisible to anything walking the object tree.
func (b *band) Content() fyne.CanvasObject { return b.content }

func (b *band) Refresh() {
	if b.rect != nil {
		b.rect.FillColor = b.colour()
		b.rect.Refresh()
	}
	b.BaseWidget.Refresh()
}

// colour resolves this band's tint through the theme, so it tracks the current
// variant rather than whatever was current when the band was built.
func (b *band) colour() color.Color {
	app := fyne.CurrentApp()
	if app == nil {
		return color.Transparent
	}
	settings := app.Settings()
	provider, ok := settings.Theme().(tintProvider)
	if !ok {
		return color.Transparent // a theme without tints simply gets none
	}
	return provider.QuestionTint(b.tint, settings.ThemeVariant())
}
