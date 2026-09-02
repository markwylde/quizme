package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// commentField is a comment box that rests collapsed.
//
// Every question accepts a comment, but most questions on most questionnaires
// never get one, and an open two-row field under each of them is most of the
// scrolling on a ten-question page. The field is built and hidden rather than
// created on demand, so the answer binding and the form's dirty tracking work
// the same whether it is showing or not.
//
// A collapsed comment is never invisible: once there is text, the toggle shows
// the start of it instead of an invitation to write one.
type commentField struct {
	entry    *widget.Entry
	toggle   *widget.Button
	root     *fyne.Container
	expanded bool

	// trailing is an action sharing the toggle's row, if the question has one.
	trailing fyne.CanvasObject
	// shield keeps the page scrolling while the pointer is over the field.
	shield func(fyne.CanvasObject) fyne.CanvasObject
	// wrapped is what actually goes in the layout: the entry, shielded.
	wrapped fyne.CanvasObject
}

// shielded returns the entry wrapped so it does not swallow page scrolling.
func (c *commentField) shielded() fyne.CanvasObject {
	if c.wrapped != nil {
		return c.wrapped
	}
	c.wrapped = c.entry
	if c.shield != nil {
		c.wrapped = c.shield(c.entry)
	}
	return c.wrapped
}

// previewLimit is how much of a comment the collapsed toggle shows.
const previewLimit = 44

func newCommentField(current string, onChange func(string), shield func(fyne.CanvasObject) fyne.CanvasObject) *commentField {
	c := &commentField{shield: shield}

	c.entry = widget.NewMultiLineEntry()
	c.entry.SetPlaceHolder("Anything worth saying about this answer")
	c.entry.SetMinRowsVisible(2)
	c.entry.Wrapping = fyne.TextWrapWord
	c.entry.SetText(current)
	c.entry.OnChanged = func(s string) {
		onChange(s)
		c.relabel()
	}

	// An icon rather than bare text: at body weight and heading position, a
	// plain label reads as another prompt rather than something to press.
	c.toggle = widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), c.Toggle)
	c.toggle.Importance = widget.LowImportance
	c.toggle.Alignment = widget.ButtonAlignLeading

	// A question that arrives with a comment opens showing it: the responder
	// wrote it last time and should not have to hunt for it.
	c.expanded = strings.TrimSpace(current) != ""
	c.root = container.NewVBox(c.toggle, c.shielded())
	c.apply()

	return c
}

// SetTrailing puts an action on the toggle's own row, at its trailing edge.
//
// A question whose answer has no settling gesture of its own needs something to
// press when it is finished with, and the toggle's row is the one row on the
// card with space going spare.
func (c *commentField) SetTrailing(obj fyne.CanvasObject) {
	c.trailing = obj
	c.root.Objects[0] = container.NewBorder(nil, nil, nil, obj, c.toggle)
	c.root.Refresh()
}

// Toggle opens or closes the field, focusing it on the way open.
func (c *commentField) Toggle() {
	c.expanded = !c.expanded
	c.apply()
	if c.expanded {
		c.focus()
	}
}

// Expanded reports whether the field is currently showing.
func (c *commentField) Expanded() bool { return c.expanded }

// Entry exposes the text field, for tests and for focus handling.
func (c *commentField) Entry() *widget.Entry { return c.entry }

// FieldVisible reports whether the text field is on screen. The entry is
// wrapped in a scroll shield, and it is the wrapper that gets hidden, so asking
// the entry itself would always say yes.
func (c *commentField) FieldVisible() bool { return c.shielded().Visible() }

func (c *commentField) apply() {
	if c.expanded {
		c.shielded().Show()
	} else {
		c.shielded().Hide()
	}
	c.relabel()
	c.root.Refresh()
}

// relabel keeps the toggle honest about what the field holds.
func (c *commentField) relabel() {
	text := strings.TrimSpace(c.entry.Text)
	switch {
	case c.expanded:
		c.toggle.SetText("Hide comment")
	case text == "":
		c.toggle.SetText("Add comment")
	default:
		c.toggle.SetText("Comment: " + preview(text))
	}
}

// focus puts the cursor in the field. There is no canvas in tests, and none
// before the window is shown, so a missing one is not an error.
func (c *commentField) focus() {
	app := fyne.CurrentApp()
	if app == nil || app.Driver() == nil {
		return
	}
	if canvas := app.Driver().CanvasForObject(c.entry); canvas != nil {
		canvas.Focus(c.entry)
	}
}

// preview renders the start of a comment on one line.
func preview(text string) string { return previewTo(text, previewLimit) }

// previewTo renders the start of some text on one line, no longer than limit.
//
// The collapsed comment and a collapsed question's answer both need this, but
// not at the same length: the comment shares its row with a toggle, while the
// answer has a line of its own to spend.
func previewTo(text string, limit int) string {
	line := text
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = strings.TrimSpace(line[:i])
	}
	runes := []rune(line)
	if len(runes) <= limit {
		if len(runes) < len([]rune(text)) {
			return line + "…" // there were further lines
		}
		return line
	}
	return strings.TrimRight(string(runes[:limit]), " ") + "…"
}
