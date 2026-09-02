package ui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/markwylde/interrogate/internal/questionnaire"
)

// isInk reports whether an object actually puts something on the screen. A
// transparent rectangle is part of most Fyne widgets -- a hover layer, a
// background waiting to be filled in -- and counting it as the widget's left
// edge measures something no reader can see.
func isInk(obj fyne.CanvasObject) (yes, leaf bool) {
	switch o := obj.(type) {
	case *canvas.Text:
		return o.Text != "" && opaque(o.Color), true
	case *canvas.Rectangle:
		return opaque(o.FillColor) || (o.StrokeWidth > 0 && opaque(o.StrokeColor)), true
	case *canvas.Circle:
		return opaque(o.FillColor) || (o.StrokeWidth > 0 && opaque(o.StrokeColor)), true
	case *canvas.Line:
		return o.StrokeWidth > 0 && opaque(o.StrokeColor), true
	case *canvas.Image:
		return o.Resource != nil || o.File != "" || o.Image != nil, true
	}
	return false, false
}

func opaque(c color.Color) bool {
	if c == nil {
		return false
	}
	_, _, _, a := c.RGBA()
	return a > 0x1000 // a whisper of alpha is not something you can see
}

// leftmostInk is the smallest x of anything drawn under obj, measured in the
// coordinates of obj's own parent.
func leftmostInk(obj fyne.CanvasObject) (float32, bool) {
	found := false
	best := float32(0)

	var walk func(o fyne.CanvasObject, offset float32)
	walk = func(o fyne.CanvasObject, offset float32) {
		if o == nil || !o.Visible() {
			return
		}
		at := offset + o.Position().X

		if ink, leaf := isInk(o); leaf {
			if ink && (!found || at < best) {
				best, found = at, true
			}
			return
		}

		switch p := o.(type) {
		case *fyne.Container:
			for _, child := range p.Objects {
				walk(child, at)
			}
		case fyne.Widget:
			for _, child := range test.WidgetRenderer(p).Objects() {
				walk(child, at)
			}
		}
	}
	walk(obj, 0)
	return best, found
}

// inkOf finds target somewhere under root and reports the leftmost ink it
// draws, in root's parent's coordinates.
func inkOf(root, target fyne.CanvasObject) (float32, bool) {
	var walk func(o fyne.CanvasObject, offset float32) (float32, bool)
	walk = func(o fyne.CanvasObject, offset float32) (float32, bool) {
		if o == nil {
			return 0, false
		}
		at := offset + o.Position().X
		if o == target {
			x, ok := leftmostInk(o)
			return at - o.Position().X + x, ok
		}
		switch p := o.(type) {
		case *fyne.Container:
			for _, child := range p.Objects {
				if x, ok := walk(child, at); ok {
					return x, true
				}
			}
		case fyne.Widget:
			for _, child := range test.WidgetRenderer(p).Objects() {
				if x, ok := walk(child, at); ok {
					return x, true
				}
			}
		}
		return 0, false
	}
	return walk(root, 0)
}

const alignmentDoc = `
title: Alignment
questions:
  - id: sel
    type: select
    prompt: A single choice, with a prompt long enough to be a fair test of the column
    options: [in-place, sidecar, both]
  - id: many
    type: select
    prompt: A choice with enough options to become a dropdown
    options: [a, b, c, d, e, f, g, h]
  - id: multi
    type: multiselect
    prompt: Any number of choices
    options: [yaml, json, toml]
  - id: line
    type: text
    prompt: A single line
  - id: para
    type: textarea
    prompt: Several lines
  - id: num
    type: number
    prompt: A number
    min: 1
    max: 30
  - id: yesno
    type: boolean
    prompt: Yes or no
  - id: scale
    type: scale
    prompt: A point on a scale
  - id: order
    type: rank
    prompt: An ordering
    options: [correctness, speed, looks]
`

// Everything a question draws lines up with its prompt's text. Fyne's widgets
// disagree about how far in they draw -- a radio group sets its buttons in a
// little way, an entry draws its border at the very edge, a button fills its
// whole box -- so the form makes up the difference per control, and this is
// what says the numbers it uses are still the right ones.
func TestEveryControlLinesUpWithThePromptsText(t *testing.T) {
	restore := useTheme(t, newTheme())
	defer restore()

	f, doc := build(t, alignmentDoc)
	win := test.NewWindow(f.build())
	defer win.Close()
	win.Resize(fyne.NewSize(780, 2400))

	for _, q := range doc.Questions {
		c := f.cards[q.ID]

		prompt, ok := inkOf(c.header, c.header.prompt)
		if !ok {
			t.Fatalf("%s: the prompt drew nothing", q.ID)
		}

		// The control a responder answers with, rather than the container it
		// arrives in: a container's leftmost ink can be some hint or label that
		// happens to line up, which would let a misplaced control through.
		control, ok := answerControl(t, f, q)
		if !ok {
			t.Fatalf("%s (%s): no control found", q.ID, q.Type)
		}
		at, ok := inkOf(c.body, control)
		if !ok {
			t.Fatalf("%s (%s): the control drew nothing", q.ID, q.Type)
		}

		if diff := at - prompt; diff > 1 || diff < -1 {
			t.Errorf("%s (%s): the %T draws %v from the prompt's text (%v against %v)",
				q.ID, q.Type, control, diff, at, prompt)
		}
	}
}

// answerControl is the widget a question is actually answered with.
func answerControl(t *testing.T, f *form, q *questionnaire.Question) (fyne.CanvasObject, bool) {
	t.Helper()
	c := f.cards[q.ID]
	switch q.Type {
	case questionnaire.TypeSelect:
		if len(q.Options) >= radioLimit {
			return searchIn[*widget.Select](c.control)
		}
		return searchIn[*widget.RadioGroup](c.control)
	case questionnaire.TypeBoolean:
		return searchIn[*widget.RadioGroup](c.control)
	case questionnaire.TypeMultiselect:
		return searchIn[*widget.CheckGroup](c.control)
	case questionnaire.TypeText, questionnaire.TypeTextarea, questionnaire.TypeNumber:
		return searchIn[*widget.Entry](c.control)
	case questionnaire.TypeScale:
		return searchIn[*scaleWidget](c.control)
	case questionnaire.TypeRank:
		return searchIn[*rankWidget](c.control)
	}
	return nil, false
}

func searchIn[T fyne.CanvasObject](root fyne.CanvasObject) (fyne.CanvasObject, bool) {
	found, ok := search[T](root, first[T]())
	if !ok {
		return nil, false
	}
	return found, true
}
