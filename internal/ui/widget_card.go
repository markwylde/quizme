package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// cardBox paints a question's panel behind its content.
//
// It is a widget rather than a bare canvas.Rectangle so that the colour is
// re-read on every refresh. Fyne refreshes widgets when the theme variant
// changes, so a form open while the desktop switches between light and dark
// picks up the right colour instead of keeping the one it started with.
type cardBox struct {
	widget.BaseWidget

	content fyne.CanvasObject
	rect    *canvas.Rectangle

	// settled says to paint the finished-question panel rather than the plain
	// one. It is the card's own state, not the question's, so that the one
	// place that knows whether a question is folded and answered is the one
	// place that decides.
	settled bool
}

func newCardBox(content fyne.CanvasObject) *cardBox {
	c := &cardBox{content: content}
	c.ExtendBaseWidget(c)
	return c
}

func (c *cardBox) CreateRenderer() fyne.WidgetRenderer {
	c.rect = canvas.NewRectangle(c.colour())
	c.rect.CornerRadius = theme.Size(theme.SizeNameCardRadius)
	return widget.NewSimpleRenderer(container.NewStack(c.rect, c.content))
}

// SetSettled chooses which panel the card paints.
func (c *cardBox) SetSettled(settled bool) {
	if c.settled == settled {
		return
	}
	c.settled = settled
	c.Refresh()
}

// Settled reports which panel the card is painting.
func (c *cardBox) Settled() bool { return c.settled }

// Content is what the card paints behind. A widget's children are otherwise
// reachable only through its renderer, which makes the question inside a card
// invisible to anything walking the object tree.
func (c *cardBox) Content() fyne.CanvasObject { return c.content }

func (c *cardBox) Refresh() {
	if c.rect != nil {
		c.rect.FillColor = c.colour()
		c.rect.CornerRadius = theme.Size(theme.SizeNameCardRadius)
		c.rect.Refresh()
	}
	c.BaseWidget.Refresh()
}

// colour resolves the card background through the theme, so it tracks the
// current variant rather than whatever was current when the card was built.
func (c *cardBox) colour() color.Color {
	app := fyne.CurrentApp()
	if app == nil {
		return color.Transparent
	}
	settings := app.Settings()
	provider, ok := settings.Theme().(cardProvider)
	if !ok {
		return color.Transparent // a theme without card colours simply gets none
	}
	if c.settled {
		return provider.SettledCard(settings.ThemeVariant())
	}
	return provider.QuestionCard(settings.ThemeVariant())
}
