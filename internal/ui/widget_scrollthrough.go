package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// scrollThrough makes a text field stop swallowing the page's scrolling.
//
// Fyne gives every Entry an internal scroller, not only the multi-line ones: a
// single-line field truncates and scrolls its content sideways, so its scroller
// is live too. Scroll.Scrolled consumes the event whether or not it had
// anywhere to scroll. Events do not bubble, and the driver dispatches to the
// deepest object under the cursor implementing the interface -- so a trackpad
// scroll that crosses any text field simply stops until the pointer leaves it.
//
// The fix is to be deeper still: a transparent layer over the field that
// implements Scrollable and forwards to the page. Every other kind of event is
// matched by its own interface, which this layer does not implement, so taps,
// hovering, text selection and focus all reach the field as before.
type scrollThrough struct {
	widget.BaseWidget

	field  fyne.CanvasObject
	onto   func(*fyne.ScrollEvent)
	shield *scrollShield
}

func newScrollThrough(field fyne.CanvasObject, onto func(*fyne.ScrollEvent)) *scrollThrough {
	s := &scrollThrough{field: field, onto: onto, shield: newScrollShield(onto)}
	s.ExtendBaseWidget(s)
	return s
}

func (s *scrollThrough) CreateRenderer() fyne.WidgetRenderer {
	// The shield is stacked last, so it is the deepest match under the cursor.
	return widget.NewSimpleRenderer(container.NewStack(s.field, s.shield))
}

// Objects exposes both layers, so the object tree stays walkable without
// building a renderer first.
func (s *scrollThrough) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{s.field, s.shield}
}

// Content is the field being shielded.
func (s *scrollThrough) Content() fyne.CanvasObject { return s.field }

// scrollShield is the transparent layer itself. It draws nothing and answers
// only to scrolling.
type scrollShield struct {
	widget.BaseWidget

	onto func(*fyne.ScrollEvent)
}

func newScrollShield(onto func(*fyne.ScrollEvent)) *scrollShield {
	s := &scrollShield{onto: onto}
	s.ExtendBaseWidget(s)
	return s
}

func (s *scrollShield) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewWithoutLayout())
}

// Scrolled forwards the event to whatever the shield was pointed at.
func (s *scrollShield) Scrolled(e *fyne.ScrollEvent) {
	if s.onto != nil {
		s.onto(e)
	}
}
