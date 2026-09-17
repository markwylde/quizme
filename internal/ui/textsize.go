package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"github.com/markwylde/quizme/internal/config"
)

// sizeChange is what a text size gesture asks for.
type sizeChange int

const (
	sizeNone sizeChange = iota
	sizeUp
	sizeDown
	sizeReset
)

// sizeShortcuts are the key presses that change the text size: the platform's
// shortcut modifier with = or + for larger, - for smaller, and 0 for the
// default. + is reached with Shift on most layouts, so the shifted forms count
// too.
func sizeShortcuts() map[*desktop.CustomShortcut]sizeChange {
	out := map[*desktop.CustomShortcut]sizeChange{}
	for _, mod := range []fyne.KeyModifier{
		fyne.KeyModifierShortcutDefault,
		fyne.KeyModifierShortcutDefault | fyne.KeyModifierShift,
	} {
		out[&desktop.CustomShortcut{KeyName: fyne.KeyEqual, Modifier: mod}] = sizeUp
		out[&desktop.CustomShortcut{KeyName: fyne.KeyPlus, Modifier: mod}] = sizeUp
		out[&desktop.CustomShortcut{KeyName: fyne.KeyMinus, Modifier: mod}] = sizeDown
		out[&desktop.CustomShortcut{KeyName: fyne.Key0, Modifier: mod}] = sizeReset
	}
	return out
}

// sizeChangeFor reads a shortcut as a text size gesture, if it is one.
func sizeChangeFor(s fyne.Shortcut) sizeChange {
	custom, ok := s.(*desktop.CustomShortcut)
	if !ok {
		return sizeNone
	}
	for sc, change := range sizeShortcuts() {
		if sc.KeyName == custom.KeyName && sc.Modifier == custom.Modifier {
			return change
		}
	}
	return sizeNone
}

// registerSizeShortcuts puts the text size shortcuts on the window, for when no
// text field has focus. A focused field is handed shortcuts before the window
// is, which is what formEntry is for.
func (f *form) registerSizeShortcuts(c interface {
	AddShortcut(fyne.Shortcut, func(fyne.Shortcut))
}) {
	for sc, change := range sizeShortcuts() {
		change := change
		c.AddShortcut(sc, func(fyne.Shortcut) { f.changeTextSize(change) })
	}
}

// changeTextSize applies a text size gesture.
func (f *form) changeTextSize(change sizeChange) {
	switch change {
	case sizeUp:
		f.setTextSize(f.textSize + config.TextSizeStep)
	case sizeDown:
		f.setTextSize(f.textSize - config.TextSizeStep)
	case sizeReset:
		f.setTextSize(config.DefaultTextSize)
	}
}

// setTextSize shows the form at a new text size and reports it.
//
// The theme is swapped rather than the page rebuilt: every layout on the page
// reads its sizes from the theme as it lays out, so the answers, the folds and
// the focus the responder had all stay exactly where they were. The size is
// not an answer, so nothing here touches what dirty() compares.
func (f *form) setTextSize(percent int) {
	percent = min(max(percent, config.MinTextSize), config.MaxTextSize)
	if percent == f.textSize {
		return
	}
	f.textSize = percent
	if app := fyne.CurrentApp(); app != nil {
		app.Settings().SetTheme(newScaledTheme(percent))
	}
	f.syncSizeButtons()
	if f.onTextSize != nil {
		f.onTextSize(percent)
	}
}

// sizeButtons are the header's smaller and larger controls.
func (f *form) sizeButtons() fyne.CanvasObject {
	f.smaller = widget.NewButton("A−", func() { f.changeTextSize(sizeDown) })
	f.smaller.Importance = widget.LowImportance
	f.larger = widget.NewButton("A+", func() { f.changeTextSize(sizeUp) })
	f.larger.Importance = widget.LowImportance
	f.syncSizeButtons()
	return container.NewHBox(f.smaller, f.larger)
}

// syncSizeButtons disables whichever direction has run out of sizes.
func (f *form) syncSizeButtons() {
	if f.smaller == nil || f.larger == nil {
		return
	}
	enable(f.smaller, f.textSize > config.MinTextSize)
	enable(f.larger, f.textSize < config.MaxTextSize)
}

func enable(b *widget.Button, on bool) {
	if on {
		b.Enable()
	} else {
		b.Disable()
	}
}

// formEntry is a text field that lets the form's own shortcuts through.
//
// A focused field is given every shortcut before the window is, and Fyne's entry
// keeps the ones it does not recognise to itself -- so without this, the text
// size could not be changed while typing, which is when someone squinting at
// their own words most wants to.
type formEntry struct {
	widget.Entry

	onSize func(sizeChange)
}

func newFormEntry(multiLine bool, onSize func(sizeChange)) *formEntry {
	e := &formEntry{onSize: onSize}
	e.MultiLine = multiLine
	e.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
	e.ExtendBaseWidget(e)
	return e
}

func (e *formEntry) TypedShortcut(s fyne.Shortcut) {
	if change := sizeChangeFor(s); change != sizeNone && e.onSize != nil {
		e.onSize(change)
		return
	}
	e.Entry.TypedShortcut(s)
}
