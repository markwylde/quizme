package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

// Whatever comes first under a prompt starts in the same place, folded or open.
//
// It did not when the question had help text. A folded question's answer has
// its own top padding taken back by the stack above it; a label in the body
// keeps its, so the first line under the prompt dropped by that padding the
// moment the question opened -- visible as a jump while folding.
func TestTheFirstLineUnderAPromptDoesNotMoveWhenTheQuestionOpens(t *testing.T) {
	restore := useTheme(t, newTheme())
	defer restore()

	f, _ := build(t, `
title: t
questions:
  - id: display
    type: select
    prompt: A shop displays a jacket in its window marked "40". A customer takes it to the till and offers 40. In contract law, what is the display?
    help: Think about who makes the offer and who accepts in a shop.
    required: true
    options: [An offer, An invitation to treat, A collateral warranty]
    answer: A collateral warranty
`)
	content := f.build()
	win := test.NewWindow(content)
	defer win.Close()
	settle := func() {
		win.Resize(fyne.NewSize(780, 900))
		win.Resize(fyne.NewSize(780, 901))
		win.Resize(fyne.NewSize(780, 900))
	}
	settle()

	c := f.cards["display"]
	if !c.collapsed() {
		t.Fatal("an answered question should open folded")
	}

	// Folded: the answer is the first thing under the prompt.
	_, promptEnds, ok := inkSpanOf(c.root, c.header.prompt)
	if !ok {
		t.Fatal("the prompt drew nothing")
	}
	answerStarts, _, ok := inkSpanOf(c.root, c.header.summary)
	if !ok {
		t.Fatal("the answer drew nothing")
	}
	folded := answerStarts - promptEnds

	// Open: the help is.
	f.toggle(c)
	settle()

	_, promptEnds, ok = inkSpanOf(c.root, c.header.prompt)
	if !ok {
		t.Fatal("the prompt drew nothing once open")
	}
	helpStarts, _, ok := inkSpanOf(c.root, c.body)
	if !ok {
		t.Fatal("the body drew nothing")
	}
	open := helpStarts - promptEnds

	if diff := open - folded; diff > 1 || diff < -1 {
		t.Errorf("the first line sits %v under the prompt folded and %v open", folded, open)
	}
	if folded < promptTail-1 || folded > promptTail+1 {
		t.Errorf("the answer sits %v under the prompt, want the %v tail", folded, float32(promptTail))
	}
}
