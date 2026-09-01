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
}

// previewLimit is how much of a comment the collapsed toggle shows.
const previewLimit = 44

func newCommentField(current string, onChange func(string)) *commentField {
	c := &commentField{}

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
	c.root = container.NewVBox(c.toggle, c.entry)
	c.apply()

	return c
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

func (c *commentField) apply() {
	if c.expanded {
		c.entry.Show()
	} else {
		c.entry.Hide()
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
func preview(text string) string {
	line := text
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = strings.TrimSpace(line[:i])
	}
	runes := []rune(line)
	if len(runes) <= previewLimit {
		if len(runes) < len([]rune(text)) {
			return line + "…" // there were further lines
		}
		return line
	}
	return strings.TrimRight(string(runes[:previewLimit]), " ") + "…"
}
