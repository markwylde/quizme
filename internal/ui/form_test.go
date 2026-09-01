package ui

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/markwylde/interrogate/internal/questionnaire"
)

func TestMain(m *testing.M) {
	// A test app gives real widgets without opening a window.
	test.NewApp()
	m.Run()
}

func build(t *testing.T, src string) (*form, *questionnaire.Document) {
	t.Helper()
	doc, err := questionnaire.LoadBytes("test.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	f := newForm(doc, nil)
	f.build()
	return f, doc
}

// find walks a card's widget tree for the first object of the given type.
func find[T fyne.CanvasObject](t *testing.T, f *form, id string) T {
	t.Helper()
	c, ok := f.cards[id]
	if !ok {
		t.Fatalf("no card for question %q", id)
	}
	var zero T
	found, ok := search[T](c.root, first[T]())
	if !ok {
		t.Fatalf("no %T in the card for %q", zero, id)
	}
	return found
}

// first matches any object of the wanted type.
func first[T fyne.CanvasObject]() func(T) bool { return func(T) bool { return true } }

func search[T fyne.CanvasObject](obj fyne.CanvasObject, match func(T) bool) (T, bool) {
	var zero T
	if typed, ok := obj.(T); ok && match(typed) {
		return typed, true
	}
	if parent, ok := obj.(interface{ Objects() []fyne.CanvasObject }); ok {
		for _, child := range parent.Objects() {
			if found, ok := search[T](child, match); ok {
				return found, true
			}
		}
	}
	switch p := obj.(type) {
	case *fyne.Container:
		for _, child := range p.Objects {
			if found, ok := search[T](child, match); ok {
				return found, true
			}
		}
	}
	return zero, false
}

const uiDoc = `
title: Decisions
intro: A few things to settle.
questions:
  - id: storage
    type: select
    prompt: Where do answers live?
    required: true
    options: [in-place, sidecar, both]
  - id: why
    type: textarea
    prompt: Why sidecar?
    show_if: {storage: sidecar}
  - id: formats
    type: multiselect
    prompt: Which formats?
    options: [yaml, json, toml]
  - id: name
    type: text
    prompt: Call it what?
  - id: budget
    type: number
    prompt: How many days?
    min: 1
    max: 30
  - id: ship
    type: boolean
    prompt: Ship before review?
  - id: urgency
    type: scale
    prompt: How urgent?
  - id: order
    type: rank
    prompt: Rank these.
    options: [correctness, speed, looks]
`

func TestEveryQuestionGetsACard(t *testing.T) {
	f, doc := build(t, uiDoc)
	if len(f.cards) != len(doc.Questions) {
		t.Errorf("built %d cards for %d questions", len(f.cards), len(doc.Questions))
	}
}

func TestSelectRecordsChoice(t *testing.T) {
	f, doc := build(t, uiDoc)
	group := find[*widget.RadioGroup](t, f, "storage")
	group.SetSelected("sidecar")
	if got := doc.Question("storage").Answer; got != "sidecar" {
		t.Errorf("answer = %#v, want %q", got, "sidecar")
	}
}

func TestSelectFallsBackToADropdownWhenOptionsAreMany(t *testing.T) {
	src := `title: t
questions:
  - id: q
    type: select
    prompt: Pick
    options: [a, b, c, d, e, f, g, h]
`
	f, doc := build(t, src)
	sel := find[*widget.Select](t, f, "q")
	sel.SetSelected("g")
	if got := doc.Question("q").Answer; got != "g" {
		t.Errorf("answer = %#v", got)
	}
}

func TestMultiselectRecordsEveryChoice(t *testing.T) {
	f, doc := build(t, uiDoc)
	group := find[*widget.CheckGroup](t, f, "formats")
	group.SetSelected([]string{"yaml", "toml"})
	got, ok := doc.Question("formats").Answer.([]string)
	if !ok || !reflect.DeepEqual(got, []string{"yaml", "toml"}) {
		t.Errorf("answer = %#v", doc.Question("formats").Answer)
	}

	// None selected is a valid state, not an answer.
	group.SetSelected(nil)
	if doc.Question("formats").HasAnswer() {
		t.Errorf("an empty selection should not count as an answer: %#v", doc.Question("formats").Answer)
	}
}

func TestTextRecordsTyping(t *testing.T) {
	f, doc := build(t, uiDoc)
	entry := find[*widget.Entry](t, f, "name")
	entry.SetText("interrogate")
	if got := doc.Question("name").Answer; got != "interrogate" {
		t.Errorf("answer = %#v", got)
	}
}

func TestNumberRespectsBounds(t *testing.T) {
	f, doc := build(t, uiDoc)
	entry := find[*widget.Entry](t, f, "budget")

	entry.SetText("12")
	if got := doc.Question("budget").Answer; got != 12.0 {
		t.Errorf("in-bounds answer = %#v, want 12", got)
	}

	entry.SetText("99")
	if got := doc.Question("budget").Answer; got != 12.0 {
		t.Errorf("an out-of-bounds entry changed the answer to %#v", got)
	}
	if err := entry.Validate(); err == nil {
		t.Error("an out-of-bounds entry should report a validation error")
	} else if !strings.Contains(err.Error(), "30") {
		t.Errorf("the error should name the bound: %v", err)
	}

	entry.SetText("0")
	if err := entry.Validate(); err == nil || !strings.Contains(err.Error(), "1") {
		t.Errorf("below the minimum should name it: %v", err)
	}

	entry.SetText("not a number")
	if err := entry.Validate(); err == nil {
		t.Error("non-numeric text should be refused")
	}

	entry.SetText("")
	if doc.Question("budget").HasAnswer() {
		t.Errorf("clearing the field should clear the answer, got %#v", doc.Question("budget").Answer)
	}
}

func TestBooleanDistinguishesNoFromUnanswered(t *testing.T) {
	f, doc := build(t, uiDoc)
	group := find[*widget.RadioGroup](t, f, "ship")

	if doc.Question("ship").HasAnswer() {
		t.Error("an untouched boolean should be unanswered")
	}
	group.SetSelected("No")
	if got := doc.Question("ship").Answer; got != false {
		t.Errorf("answer = %#v, want false", got)
	}
	group.SetSelected("Yes")
	if got := doc.Question("ship").Answer; got != true {
		t.Errorf("answer = %#v, want true", got)
	}
}

func TestScaleSelectsAndClears(t *testing.T) {
	f, doc := build(t, uiDoc)
	scale := find[*scaleWidget](t, f, "urgency")
	scale.CreateRenderer() // build the buttons

	scale.choose(4)
	if got := doc.Question("urgency").Answer; got != 4 {
		t.Errorf("answer = %#v, want 4", got)
	}
	scale.choose(4) // pressing the same point clears it
	if doc.Question("urgency").HasAnswer() {
		t.Errorf("pressing the selected point should clear it, got %#v", doc.Question("urgency").Answer)
	}
}

func TestScaleStaysWithinItsRange(t *testing.T) {
	f, _ := build(t, uiDoc)
	scale := find[*scaleWidget](t, f, "urgency")
	scale.CreateRenderer()
	if scale.lo != 1 || scale.hi != 5 {
		t.Errorf("scale bounds = %d..%d, want 1..5", scale.lo, scale.hi)
	}
	if len(scale.buttons) != 5 {
		t.Errorf("built %d buttons, want 5", len(scale.buttons))
	}
}

func TestRankAlwaysProducesAFullPermutation(t *testing.T) {
	f, doc := build(t, uiDoc)
	rank := find[*rankWidget](t, f, "order")
	rank.CreateRenderer()

	if doc.Question("order").HasAnswer() {
		t.Error("an untouched ranking is the author's order, not an answer")
	}

	rank.move(0, 1)
	got, ok := doc.Question("order").Answer.([]string)
	if !ok {
		t.Fatalf("answer = %#v", doc.Question("order").Answer)
	}
	if !reflect.DeepEqual(got, []string{"speed", "correctness", "looks"}) {
		t.Errorf("order = %v", got)
	}

	// Whatever sequence of moves is made, the answer stays a permutation.
	for _, move := range [][2]int{{2, -1}, {0, 1}, {1, 1}, {0, -1}, {2, 1}} {
		rank.move(move[0], move[1])
		current := doc.Question("order").Answer.([]string)
		if err := permutationOf(current, []string{"correctness", "speed", "looks"}); err != nil {
			t.Fatalf("after %v: %v", move, err)
		}
	}
}

func TestRankConfirmAnswersWithoutReordering(t *testing.T) {
	f, doc := build(t, uiDoc)
	rank := find[*rankWidget](t, f, "order")
	rank.CreateRenderer()

	rank.confirm()
	got, ok := doc.Question("order").Answer.([]string)
	if !ok || !reflect.DeepEqual(got, []string{"correctness", "speed", "looks"}) {
		t.Errorf("answer = %#v, want the options in their original order", doc.Question("order").Answer)
	}
}

func TestRankMoveAtTheEndsIsRefused(t *testing.T) {
	f, _ := build(t, uiDoc)
	rank := find[*rankWidget](t, f, "order")
	rank.CreateRenderer()
	if rank.move(0, -1) {
		t.Error("moving the first row up should be refused")
	}
	if rank.move(2, 1) {
		t.Error("moving the last row down should be refused")
	}
}

func permutationOf(got, want []string) error {
	if len(got) != len(want) {
		return fmt.Errorf("length %d, want %d", len(got), len(want))
	}
	seen := map[string]int{}
	for _, v := range got {
		seen[v]++
	}
	for _, v := range want {
		if seen[v] != 1 {
			return fmt.Errorf("%q appears %d times in %v", v, seen[v], got)
		}
	}
	return nil
}

func TestConditionalQuestionsAppearAndDisappear(t *testing.T) {
	f, _ := build(t, uiDoc)
	why := f.cards["why"]

	if why.root.Visible() {
		t.Error("a gated question should start hidden")
	}

	find[*widget.RadioGroup](t, f, "storage").SetSelected("sidecar")
	if !why.root.Visible() {
		t.Error("the gated question should appear once its condition holds")
	}

	find[*widget.RadioGroup](t, f, "storage").SetSelected("both")
	if why.root.Visible() {
		t.Error("the gated question should disappear once its condition fails")
	}
}

func TestConditionalQuestionKeepsItsCardInDocumentOrder(t *testing.T) {
	f, _ := build(t, uiDoc)
	// Cards are built once and shown or hidden, never rebuilt, so a question
	// appearing cannot steal focus from the control being typed in.
	want := []string{"storage", "why", "formats", "name", "budget", "ship", "urgency", "order"}
	var got []string
	for _, obj := range f.list.Objects {
		for id, c := range f.cards {
			if c.root == obj {
				got = append(got, id)
			}
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("card order = %v, want %v", got, want)
	}
}

func TestSubmitRefusedWhileARequiredQuestionIsBlank(t *testing.T) {
	f, _ := build(t, uiDoc)

	f.submit()
	if f.outcome != "" {
		t.Errorf("outcome = %q, want the form to stay open", f.outcome)
	}
	if !f.cards["storage"].warning.Visible() {
		t.Error("the outstanding question should be marked")
	}
	if !strings.Contains(f.cards["storage"].warning.Text, "needs an answer") {
		t.Errorf("warning = %q", f.cards["storage"].warning.Text)
	}

	find[*widget.RadioGroup](t, f, "storage").SetSelected("in-place")
	f.submit()
	if f.outcome != questionnaire.StatusSubmitted {
		t.Errorf("outcome = %q, want submitted", f.outcome)
	}
}

func TestHiddenRequiredQuestionDoesNotBlockSubmit(t *testing.T) {
	f, _ := build(t, `
title: t
questions:
  - {id: gate, type: select, prompt: "Gate?", options: [open, shut]}
  - id: needed
    type: text
    prompt: Needed when open
    required: true
    show_if: {gate: open}
`)
	find[*widget.RadioGroup](t, f, "gate").SetSelected("shut")
	f.submit()
	if f.outcome != questionnaire.StatusSubmitted {
		t.Errorf("outcome = %q, want submitted with the required question hidden", f.outcome)
	}
}

func TestSummaryCountsOutstandingQuestions(t *testing.T) {
	f, _ := build(t, `
title: t
questions:
  - {id: a, type: text, prompt: "A?", required: true}
  - {id: b, type: text, prompt: "B?", required: true}
`)
	if got := f.summary.Text; got != "· 2 required questions left" {
		t.Errorf("summary = %q", got)
	}
	find[*widget.Entry](t, f, "a").SetText("x")
	if got := f.summary.Text; got != "· 1 required question left" {
		t.Errorf("summary = %q", got)
	}
	find[*widget.Entry](t, f, "b").SetText("y")
	if got := f.summary.Text; got != "" {
		t.Errorf("summary = %q, want empty once nothing is outstanding", got)
	}
}

func TestClosingWithNothingEnteredDismissesWithoutAsking(t *testing.T) {
	f, _ := build(t, uiDoc)
	if f.dirty() {
		t.Fatal("an untouched form should not be dirty")
	}
	f.requestClose()
	if f.outcome != questionnaire.StatusDismissed {
		t.Errorf("outcome = %q, want dismissed", f.outcome)
	}
}

func TestDirtyTracksEveryKindOfChange(t *testing.T) {
	tests := []struct {
		name   string
		change func(f *form, doc *questionnaire.Document)
	}{
		{"an answer", func(f *form, _ *questionnaire.Document) {
			find[*widget.Entry](t, f, "name").SetText("x")
		}},
		{"a comment", func(f *form, doc *questionnaire.Document) {
			doc.Question("name").Comment = "a note"
		}},
		{"a ranking", func(f *form, _ *questionnaire.Document) {
			rank := find[*rankWidget](t, f, "order")
			rank.CreateRenderer()
			rank.confirm()
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f, doc := build(t, uiDoc)
			tc.change(f, doc)
			if !f.dirty() {
				t.Errorf("changing %s should mark the form dirty", tc.name)
			}
		})
	}
}

func TestAnAlreadyAnsweredQuestionnaireStartsClean(t *testing.T) {
	f, _ := build(t, `
title: t
questions:
  - id: q
    type: text
    prompt: Name
    answer: from last time
    comment: an old note
`)
	if f.dirty() {
		t.Error("reopening an answered questionnaire should not look like unsaved work")
	}
	if got := find[*widget.Entry](t, f, "q").Text; got != "from last time" {
		t.Errorf("the existing answer should be shown, got %q", got)
	}
}

func TestSavedStatusReflectsCompleteness(t *testing.T) {
	f, _ := build(t, uiDoc)
	if got := f.savedStatus(); got != questionnaire.StatusSaved {
		t.Errorf("savedStatus = %q, want saved while a required question is blank", got)
	}
	find[*widget.RadioGroup](t, f, "storage").SetSelected("both")
	if got := f.savedStatus(); got != questionnaire.StatusSubmitted {
		t.Errorf("savedStatus = %q, want submitted once nothing is outstanding", got)
	}
}

func TestExistingAnswersPopulateEveryControl(t *testing.T) {
	f, _ := build(t, `
title: t
questions:
  - {id: sel, type: select, prompt: "Pick?", options: [a, b], answer: b}
  - {id: multi, type: multiselect, prompt: "Which?", options: [x, y], answer: [y]}
  - {id: num, type: number, prompt: "How many?", answer: 12}
  - {id: bool, type: boolean, prompt: "Ship?", answer: false}
  - {id: sc, type: scale, prompt: "Urgency?", answer: 3}
  - {id: rk, type: rank, prompt: "Order?", options: [p, q], answer: [q, p]}
`)
	if got := find[*widget.RadioGroup](t, f, "sel").Selected; got != "b" {
		t.Errorf("select = %q", got)
	}
	if got := find[*widget.CheckGroup](t, f, "multi").Selected; !reflect.DeepEqual(got, []string{"y"}) {
		t.Errorf("multiselect = %v", got)
	}
	if got := find[*widget.Entry](t, f, "num").Text; got != "12" {
		t.Errorf("number = %q", got)
	}
	if got := find[*widget.RadioGroup](t, f, "bool").Selected; got != "No" {
		t.Errorf("boolean = %q", got)
	}
	sc := find[*scaleWidget](t, f, "sc")
	if sc.Value() == nil || *sc.Value() != 3 {
		t.Errorf("scale = %v", sc.Value())
	}
	if got := find[*rankWidget](t, f, "rk").Order(); !reflect.DeepEqual(got, []string{"q", "p"}) {
		t.Errorf("rank = %v", got)
	}
}

func TestCommentBoxOnEveryQuestionType(t *testing.T) {
	f, doc := build(t, uiDoc)
	for _, q := range doc.Questions {
		c := f.cards[q.ID]
		if c.comment == nil {
			t.Errorf("question %q (%s) has no comment field", q.ID, q.Type)
			continue
		}
		// Collapsed, but one action away for every type.
		if c.comment.Expanded() {
			t.Errorf("question %q starts with its comment field open", q.ID)
		}
		c.comment.Toggle()
		if !c.comment.Expanded() {
			t.Errorf("question %q did not open its comment field", q.ID)
		}
		c.comment.Entry().SetText("a note on " + q.ID)
		if q.Comment != "a note on "+q.ID {
			t.Errorf("comment for %q was not recorded, got %q", q.ID, q.Comment)
		}
	}
}

func TestCommentWithoutAnAnswerIsKept(t *testing.T) {
	f, doc := build(t, uiDoc)
	c := f.cards["name"]
	c.comment.Toggle()
	c.comment.Entry().SetText("worth discussing")

	if doc.Question("name").HasAnswer() {
		t.Error("the question should still be unanswered")
	}
	if doc.Question("name").Comment != "worth discussing" {
		t.Errorf("comment = %q", doc.Question("name").Comment)
	}
}
