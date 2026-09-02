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

	// hovered lifts the whole card while the pointer is on its header. The
	// highlight covers the card rather than sitting inside it: painted within
	// the header it was a rounded panel inset by the card's own padding, and
	// read as a second card inside the first.
	hovered bool
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

// SetHovered lifts or drops the whole card.
func (c *cardBox) SetHovered(hovered bool) {
	if c.hovered == hovered {
		return
	}
	c.hovered = hovered
	c.Refresh()
}

// Hovered reports whether the card is currently lifted.
func (c *cardBox) Hovered() bool { return c.hovered }

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

// over composites a translucent colour onto an opaque one. The theme's hover
// colour is an overlay, so painting it as a fill would show the page through the
// card rather than lifting it.
func over(base, top color.Color) color.Color {
	br, bg, bb, _ := base.RGBA()
	tr, tg, tb, ta := top.RGBA()
	// RGBA reports alpha-premultiplied values, so the overlay's own colour is
	// already scaled by its alpha and only the base needs attenuating.
	blend := func(b, t uint32) uint8 {
		return uint8((t + b*(0xFFFF-ta)/0xFFFF) >> 8)
	}
	return color.NRGBA{R: blend(br, tr), G: blend(bg, tg), B: blend(bb, tb), A: 0xFF}
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
	base := provider.QuestionCard(settings.ThemeVariant())
	if c.settled {
		base = provider.SettledCard(settings.ThemeVariant())
	}
	if c.hovered {
		return over(base, theme.Color(theme.ColorNameHover))
	}
	return base
}
