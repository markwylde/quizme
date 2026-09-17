package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// footerLayout puts the status beside the actions while there is room, and the
// actions on a row of their own beneath it when there is not.
//
// A border layout gives its edges whatever they ask for and the middle what is
// left, even when that is less than nothing -- so as the text grew, the status
// ran underneath the buttons. Here the status only shares the row when the
// whole of it fits; otherwise it takes the full width and wraps, and the
// actions keep their place at the trailing edge below.
//
// Objects are the status and then the actions.
type footerLayout struct {
	// owner is the container this lays out, refreshed when the choice of rows
	// changes, so the footer's height follows.
	owner *fyne.Container
	// width is the last width laid out at. How tall the footer has to be depends
	// on it, and a layout is asked its minimum size without being told.
	width float32
	// stacked is whether that width put the actions on their own row.
	stacked bool
}

func (l *footerLayout) gap() float32 { return theme.Size(theme.SizeNamePadding) }

// fits reports whether the status's text and the actions fit on one row of
// the given width.
func (l *footerLayout) fits(status, actions fyne.CanvasObject, width float32) bool {
	inset := theme.Size(theme.SizeNameInnerPadding)
	return naturalWidth(status)-inset+l.gap()+actions.MinSize().Width <= width
}

// naturalWidth is how wide an object would like to be. A wrapping label reports
// a minimum of next to nothing, so its text is measured instead.
func naturalWidth(o fyne.CanvasObject) float32 {
	label, ok := o.(*widget.Label)
	if !ok {
		return o.MinSize().Width
	}
	size := label.SizeName
	if size == "" {
		size = theme.SizeNameText
	}
	text := fyne.MeasureText(label.Text, theme.Size(size), label.TextStyle)
	return text.Width + 2*theme.Size(theme.SizeNameInnerPadding)
}

func (l *footerLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) != 2 {
		return
	}
	status, actions := objects[0], objects[1]
	a := actions.MinSize()

	stacked := !l.fits(status, actions, size.Width)
	changed := stacked != l.stacked || size.Width != l.width
	l.width, l.stacked = size.Width, stacked

	// A label sets its text in by the inner padding; taking that back puts the
	// text on the edge, where the buttons' own backgrounds start.
	inset := theme.Size(theme.SizeNameInnerPadding)

	if !stacked {
		width := size.Width - l.gap() - a.Width + inset
		status.Resize(fyne.NewSize(width, status.MinSize().Height))
		h := status.MinSize().Height // measured again, now it has wrapped
		status.Resize(fyne.NewSize(width, h))
		status.Move(fyne.NewPos(-inset, (size.Height-h)/2))
		actions.Resize(a)
		actions.Move(fyne.NewPos(size.Width-a.Width, (size.Height-a.Height)/2))
	} else {
		status.Resize(fyne.NewSize(size.Width+inset, status.MinSize().Height))
		h := status.MinSize().Height
		status.Resize(fyne.NewSize(size.Width+inset, h))
		status.Move(fyne.NewPos(-inset, 0))
		actions.Resize(a)
		actions.Move(fyne.NewPos(size.Width-a.Width, h+l.gap()))
	}

	if changed && l.owner != nil {
		l.owner.Refresh()
	}
}

func (l *footerLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) != 2 {
		return fyne.Size{}
	}
	status, actions := objects[0], objects[1]
	a := actions.MinSize()
	s := status.MinSize()
	width := fyne.Max(a.Width, s.Width)
	if l.width == 0 || l.fits(status, actions, l.width) {
		return fyne.NewSize(width, fyne.Max(a.Height, s.Height))
	}
	return fyne.NewSize(width, s.Height+l.gap()+a.Height)
}
