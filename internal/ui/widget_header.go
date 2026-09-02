package ui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/markwylde/interrogate/internal/questionnaire"
)

// questionHeader is the top line of a question's card, and the control that
// folds it.
//
// The whole row is the target rather than a chevron alone: folding is the
// commonest gesture on the page and does not deserve a twenty-pixel button.
// It is a widget rather than a Button because a button cannot hold a wrapping
// bold prompt beside a caption, a drawn mark, and an icon; Fyne's hit test
// walks up to the nearest Tappable, so the labels inside still deliver the tap.
//
// Collapsed, the row is the only thing left of the question, so it carries the
// answer as well as the prompt: a folded page reads as a review of what the
// responder has said rather than a list of prompts.
type questionHeader struct {
	widget.BaseWidget

	question *questionnaire.Question
	onTap    func()

	chevron *widget.Icon
	prompt  *widget.Label
	marker  *widget.Label // "Required", or nil when the question is optional
	summary *widget.Label
	note    *widget.Icon
	tick    *tickMark

	row        *fyne.Container
	background *canvas.Rectangle

	collapsed bool
	hovered   bool
}

func newQuestionHeader(q *questionnaire.Question, onTap func()) *questionHeader {
	h := &questionHeader{question: q, onTap: onTap}
	h.ExtendBaseWidget(h)

	h.chevron = widget.NewIcon(theme.MenuDropDownIcon())

	h.prompt = widget.NewLabel(q.Prompt)
	h.prompt.TextStyle = fyne.TextStyle{Bold: true}
	h.prompt.Wrapping = fyne.TextWrapWord

	// The summary asks for the width its text needs, which answerSummary has
	// already capped. Fyne's own truncation is no use here: a truncating label
	// reports almost no minimum width, so beside the prompt it collapses to a
	// bare ellipsis and shows none of the answer.
	h.summary = captionLabel("")

	// The same icon the comment toggle uses, so a folded question says "there
	// is a note in here" in the vocabulary the open one taught.
	h.note = widget.NewIcon(theme.DocumentCreateIcon())

	h.tick = newTickMark()

	trailing := []fyne.CanvasObject{h.summary}
	if q.Required {
		// The required marker sits on the prompt's own line. On its own row it
		// read as another instruction to take in; up here it is just a label.
		h.marker = captionLabel("Required")
		h.marker.Importance = widget.MediumImportance
		trailing = append(trailing, h.marker)
	}
	trailing = append(trailing, h.note, h.tick)

	h.row = container.NewBorder(nil, nil, h.chevron, container.NewHBox(trailing...), h.prompt)
	h.sync()
	return h
}

func (h *questionHeader) CreateRenderer() fyne.WidgetRenderer {
	h.background = canvas.NewRectangle(h.tint())
	h.background.CornerRadius = theme.Size(theme.SizeNameSelectionRadius)
	return widget.NewSimpleRenderer(container.NewStack(h.background, h.row))
}

// Content exposes the row. A widget's children hang off its renderer, so
// without this the header's parts are invisible to anything walking the tree.
func (h *questionHeader) Content() fyne.CanvasObject { return h.row }

// Objects exposes the row without building a renderer first, matching what the
// other wrapping widgets in this package do.
func (h *questionHeader) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{h.row}
}

// SetCollapsed points the chevron the right way and decides whether the row has
// to speak for the whole question.
func (h *questionHeader) SetCollapsed(collapsed bool) {
	h.collapsed = collapsed
	h.sync()
}

// Collapsed reports which way the header currently reads.
func (h *questionHeader) Collapsed() bool { return h.collapsed }

// Sync brings the row back in step with the question behind it: the answer it
// shows, whether it is ticked, and whether it carries a note.
func (h *questionHeader) Sync() { h.sync() }

func (h *questionHeader) sync() {
	if h.collapsed {
		h.chevron.SetResource(theme.MenuExpandIcon())
	} else {
		h.chevron.SetResource(theme.MenuDropDownIcon())
	}

	// The answer is worth showing only when the control that holds it is out of
	// sight; expanded, it would just be the control said twice.
	text := answerSummary(h.question)
	h.summary.SetText(text)
	show(h.summary, h.collapsed && text != "")

	// The tick follows the answer, not the fold: a question left open because
	// the responder pinned it open is no less settled for being visible.
	show(h.tick, h.question.HasAnswer())
	show(h.note, strings.TrimSpace(h.question.Comment) != "")

	h.Refresh()
}

// Refresh re-resolves the hover tint for the same reason the card does: a form
// open across a light/dark switch must not keep the colour it started with.
func (h *questionHeader) Refresh() {
	if h.background != nil {
		h.background.FillColor = h.tint()
		h.background.CornerRadius = theme.Size(theme.SizeNameSelectionRadius)
		h.background.Refresh()
	}
	h.BaseWidget.Refresh()
}

// tint is the hover highlight, and nothing at all when the pointer is elsewhere:
// the card behind it already provides the panel.
func (h *questionHeader) tint() color.Color {
	if !h.hovered {
		return color.Transparent
	}
	return theme.Color(theme.ColorNameHover)
}

func (h *questionHeader) Tapped(*fyne.PointEvent) {
	if h.onTap != nil {
		h.onTap()
	}
}

func (h *questionHeader) MouseIn(*desktop.MouseEvent) {
	h.hovered = true
	h.Refresh()
}

func (h *questionHeader) MouseOut() {
	h.hovered = false
	h.Refresh()
}

func (h *questionHeader) MouseMoved(*desktop.MouseEvent) {}

// show is Show or Hide by boolean, which reads better than the four-line if
// this file would otherwise repeat.
func show(obj fyne.CanvasObject, visible bool) {
	if visible {
		obj.Show()
		return
	}
	obj.Hide()
}

// The header answers to tapping and to the pointer arriving; everything else
// belongs to what is inside it.
var (
	_ fyne.Tappable     = (*questionHeader)(nil)
	_ desktop.Hoverable = (*questionHeader)(nil)
)
