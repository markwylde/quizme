package ui

import (
	"encoding/json"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/markwylde/quizme/internal/questionnaire"
)

// --- The split -----------------------------------------------------------

func TestThePromptLivesInTheHeaderAndTheRestInTheBody(t *testing.T) {
	f, doc := build(t, uiDoc)
	for _, q := range doc.Questions {
		c := f.cards[q.ID]
		if c.header == nil || c.body == nil {
			t.Fatalf("question %q has no header/body split", q.ID)
		}
		prompt, found := search[*promptText](c.header, first[*promptText]())
		if !found {
			t.Errorf("the prompt for %q is not in its header", q.ID)
		} else if prompt.Text() != q.Prompt {
			t.Errorf("the header for %q carries the prompt %q", q.ID, prompt.Text())
		}
		if _, ok := search[*promptText](c.body, first[*promptText]()); ok {
			t.Errorf("the prompt for %q is still in its body as well", q.ID)
		}
		for what, obj := range map[string]fyne.CanvasObject{
			"control": c.control,
			"comment": c.comment.root,
			"warning": c.warning,
		} {
			if _, ok := search[fyne.CanvasObject](c.body, func(o fyne.CanvasObject) bool { return o == obj }); !ok {
				t.Errorf("the %s for %q is not in its body", what, q.ID)
			}
		}
	}
}

func TestHelpTextLivesInTheBody(t *testing.T) {
	f, _ := build(t, `
title: t
questions:
  - {id: q, type: text, prompt: "Name?", help: "Anything you like"}
`)
	c := f.cards["q"]
	if _, ok := search[*widget.Label](c.body, func(l *widget.Label) bool { return l.Text == "Anything you like" }); !ok {
		t.Error("the help text should fold away with the rest of the question")
	}
}

// Everything a question shows shares the prompt's left edge. The controls
// answer the prompt, and starting them a chevron's width to its left meant
// nothing on the card lined up with anything else.
func TestTheBodyLinesUpWithThePrompt(t *testing.T) {
	f, _ := build(t, uiDoc)
	c := f.cards["storage"]

	w := test.NewWindow(f.build())
	defer w.Close()
	w.Resize(fyne.NewSize(780, 900))

	inner := c.body.Objects[0]
	if got, want := inner.Position().X, promptInset(); got != want {
		t.Errorf("the body starts at x=%v, want the prompt's column at x=%v", got, want)
	}
	if got, want := c.header.stack.Position().X, promptInset(); got != want {
		t.Errorf("the prompt starts at x=%v, want x=%v", got, want)
	}
	// Which is the point: one left edge, not two.
	if inner.Position().X != c.header.stack.Position().X {
		t.Errorf("the body is at x=%v and the prompt at x=%v",
			inner.Position().X, c.header.stack.Position().X)
	}
	// And the chevron sits in the space that leaves.
	if got := c.header.chevron.Position().X; got < 0 || got >= promptInset() {
		t.Errorf("the chevron is at x=%v, want it inside the %v column", got, promptInset())
	}
}

// The indent must not become a gap of its own when there is nothing to indent,
// or a folded question would leave a band of empty card behind.
func TestThePromptColumnTakesNoRoomForNothing(t *testing.T) {
	shown := widget.NewLabel("something")
	hidden := widget.NewLabel("nothing")
	hidden.Hide()
	column := &promptColumn{}

	if got := column.MinSize([]fyne.CanvasObject{hidden}); got != (fyne.Size{}) {
		t.Errorf("MinSize = %v with nothing showing, want no size at all", got)
	}
	got := column.MinSize([]fyne.CanvasObject{shown})
	if got.Height != shown.MinSize().Height {
		t.Errorf("height = %v, want the %v its content asks for", got.Height, shown.MinSize().Height)
	}
	if want := shown.MinSize().Width + promptInset(); got.Width != want {
		t.Errorf("width = %v, want %v: the content plus the column it is indented past", got.Width, want)
	}
}

// --- Folding -------------------------------------------------------------

func TestFoldingHidesTheBodyButKeepsIt(t *testing.T) {
	f, _ := build(t, uiDoc)
	c := f.cards["name"]

	c.comment.Toggle() // so the comment field is showing too
	if !c.body.Visible() {
		t.Fatal("the question should start with its body showing")
	}
	expanded := c.root.MinSize().Height
	field := find[*formEntry](t, f, "name")

	f.toggle(c)
	if c.body.Visible() {
		t.Error("folding should hide the body")
	}
	// The height is what says the body is really out of the page rather than
	// merely flagged: a box layout gives no room to a hidden child.
	folded := c.root.MinSize().Height
	if folded >= expanded {
		t.Errorf("a folded question asks for %v, expanded it asked for %v", folded, expanded)
	}
	if folded > c.header.MinSize().Height+expanded/2 {
		t.Errorf("a folded question should be about its header: %v against a header of %v", folded, c.header.MinSize().Height)
	}
	// Hidden, not discarded: the bindings depend on these being the same
	// objects when the question comes back.
	if got := find[*formEntry](t, f, "name"); got != field {
		t.Error("the field was rebuilt rather than hidden")
	}

	f.toggle(c)
	if !c.body.Visible() {
		t.Error("unfolding should bring the body back")
	}
	if got := c.root.MinSize().Height; got != expanded {
		t.Errorf("the unfolded question asks for %v, want the %v it started at", got, expanded)
	}
	if !c.comment.FieldVisible() {
		t.Error("the comment field should come back as the responder left it")
	}
}

func TestUnansweredQuestionsStartExpanded(t *testing.T) {
	f, doc := build(t, uiDoc) // nothing in the fixture carries an answer
	for _, q := range doc.Questions {
		if f.cards[q.ID].collapsed() {
			t.Errorf("question %q started collapsed with no answer to show", q.ID)
		}
	}
}

// A questionnaire arriving with answers in it -- reopened, or authored that way
// -- opens showing what is left rather than what is already settled.
func TestAnsweredQuestionsStartFolded(t *testing.T) {
	f, doc := build(t, `
title: Reopened
questions:
  - id: storage
    type: select
    prompt: Where do answers live?
    options: [in-place, sidecar, both]
    answer: sidecar
    comment: settled last time
  - id: ship
    type: boolean
    prompt: Ship before review?
    answer: false
  - id: urgency
    type: scale
    prompt: How urgent?
    answer: 4
  - id: formats
    type: multiselect
    prompt: Which formats?
    options: [yaml, json]
    answer: [yaml]
  - id: order
    type: rank
    prompt: Rank these.
    options: [correctness, speed]
    answer: [speed, correctness]
`)
	for _, q := range doc.Questions {
		c := f.cards[q.ID]
		if !c.collapsed() {
			t.Errorf("question %q arrived answered but opened in full", q.ID)
		}
		if !c.header.tick.Visible() {
			t.Errorf("question %q carries an answer but shows no tick", q.ID)
		}
		if c.header.summary.Text == "" {
			t.Errorf("question %q folded without showing what was answered", q.ID)
		}
		if !c.header.summary.Visible() {
			t.Errorf("question %q hides the answer on its folded row", q.ID)
		}
	}

	// And they open on request, with the restored answer in the control.
	storage := f.cards["storage"]
	test.Tap(storage.header)
	if storage.collapsed() {
		t.Fatal("a question folded on load should still open on request")
	}
	if got := find[*widget.RadioGroup](t, f, "storage").Selected; got != "sidecar" {
		t.Errorf("the reopened control holds %q, want the answer from the file", got)
	}
}

// An unanswered question is left open even when its neighbours fold, and a
// comment is not an answer.
func TestOnlyAnsweredQuestionsFoldOnLoad(t *testing.T) {
	f, _ := build(t, `
title: Half done
questions:
  - id: done
    type: select
    prompt: Settled?
    options: [yes, no]
    answer: yes
  - id: outstanding
    type: select
    prompt: Not settled?
    required: true
    options: [yes, no]
  - id: noted
    type: text
    prompt: Only a note?
    comment: something to say
`)
	if !f.cards["done"].collapsed() {
		t.Error("the answered question should have folded")
	}
	if f.cards["outstanding"].collapsed() {
		t.Error("the unanswered question should be open")
	}
	if f.cards["noted"].collapsed() {
		t.Error("a comment is not an answer, so that question should be open")
	}
}

func TestSettleDuringBuildDoesNothing(t *testing.T) {
	f, doc := build(t, uiDoc)
	f.building = true
	f.settle(doc.Question("storage"))
	if f.cards["storage"].collapsed() {
		t.Error("a question should not fold while the page is still being built")
	}
}

func TestTappingTheHeaderFoldsAndUnfolds(t *testing.T) {
	f, _ := build(t, uiDoc)
	c := f.cards["storage"]

	test.Tap(c.header)
	if !c.collapsed() {
		t.Fatal("tapping the header should fold the question")
	}
	if !c.header.Collapsed() {
		t.Error("the header should read as collapsed")
	}

	test.Tap(c.header)
	if c.collapsed() {
		t.Error("tapping it again should unfold the question")
	}
}

// Nothing outstanding is hidden from the responder against their will, but they
// are allowed to put it away themselves.
func TestAnUnansweredQuestionCanBeFolded(t *testing.T) {
	f, doc := build(t, uiDoc)
	c := f.cards["storage"]

	test.Tap(c.header)
	if !c.collapsed() {
		t.Fatal("an unanswered question should still fold on request")
	}
	if c.header.tick.Visible() {
		t.Error("a folded unanswered question should carry no tick")
	}
	if len(doc.Unanswered()) == 0 {
		t.Error("folding a required question should not settle it")
	}
}

// --- Settling ------------------------------------------------------------

func TestSelectFoldsWhenAnswered(t *testing.T) {
	f, doc := build(t, uiDoc)
	c := f.cards["storage"]

	group := find[*widget.RadioGroup](t, f, "storage")
	group.SetSelected("sidecar")

	if !c.collapsed() {
		t.Error("a select question should fold once an option is chosen")
	}
	if got := c.header.summary.Text; got != "sidecar" {
		t.Errorf("the folded row shows %q, want the answer", got)
	}
	if !c.header.tick.Visible() {
		t.Error("the folded row should be ticked")
	}
	if got := doc.Question("storage").Answer; got != "sidecar" {
		t.Errorf("answer = %#v", got)
	}
}

func TestADropdownSelectFoldsWhenAnswered(t *testing.T) {
	f, _ := build(t, `
title: t
questions:
  - id: q
    type: select
    prompt: Pick
    options: [a, b, c, d, e, f, g, h]
`)
	sel := find[*widget.Select](t, f, "q")
	sel.SetSelected("d")
	if !f.cards["q"].collapsed() {
		t.Error("a select shown as a dropdown should fold like any other")
	}
}

func TestBooleanFoldsWhenAnswered(t *testing.T) {
	f, _ := build(t, uiDoc)
	group := find[*widget.RadioGroup](t, f, "ship")
	group.SetSelected("No")
	c := f.cards["ship"]
	if !c.collapsed() {
		t.Error("a boolean question should fold once answered")
	}
	if got := c.header.summary.Text; got != "No" {
		t.Errorf("the folded row shows %q, want No", got)
	}
}

func TestScaleFoldsWhenAnswered(t *testing.T) {
	f, _ := build(t, uiDoc)
	scale := find[*scaleWidget](t, f, "urgency")
	scale.choose(4)
	if !f.cards["urgency"].collapsed() {
		t.Error("a scale question should fold once a point is chosen")
	}
}

func TestClearingAnAnswerDoesNotFold(t *testing.T) {
	f, _ := build(t, uiDoc)

	// A scale clears by pressing the chosen point again, which is the one
	// clearing gesture that could plausibly fold the question instead.
	scale := find[*scaleWidget](t, f, "urgency")
	scale.choose(4)
	c := f.cards["urgency"]
	f.toggle(c) // back open, as a responder revising would
	scale.choose(4)
	if c.collapsed() {
		t.Error("clearing an answer folded the question the responder was clearing")
	}

	// And the same for a select cleared through the group.
	group := find[*widget.RadioGroup](t, f, "storage")
	group.SetSelected("both")
	storage := f.cards["storage"]
	f.toggle(storage)
	group.SetSelected("")
	if storage.collapsed() {
		t.Error("clearing a select folded it")
	}
}

func TestRankFoldsOnlyOnceConfirmed(t *testing.T) {
	f, doc := build(t, uiDoc)
	c := f.cards["order"]
	rank := find[*rankWidget](t, f, "order")

	rank.move(0, 1)
	if c.collapsed() {
		t.Fatal("reordering is not finishing: the question should stay open")
	}

	rank.confirm()
	f.settle(doc.Question("order"))
	if !c.collapsed() {
		t.Error("confirming the order should fold the question")
	}
	if !strings.HasPrefix(c.header.summary.Text, "1. ") {
		t.Errorf("the folded row shows %q, want the numbered order", c.header.summary.Text)
	}
}

// The confirm button is what a responder actually presses, so it is worth
// checking that it is wired to the fold and not only to the answer.
func TestTheConfirmButtonFoldsTheRanking(t *testing.T) {
	f, _ := build(t, uiDoc)
	c := f.cards["order"]
	confirm, ok := search[*widget.Button](c.control, func(b *widget.Button) bool { return b.Text == "Use this order" })
	if !ok {
		t.Fatal("no confirm button on the rank control")
	}
	confirm.OnTapped()
	if !c.collapsed() {
		t.Error("the confirm button should fold the question")
	}
}

func TestTypingNeverFolds(t *testing.T) {
	f, _ := build(t, uiDoc)
	for _, id := range []string{"name", "budget", "why"} {
		// "why" is gated behind storage: sidecar, so reveal it first.
		if id == "why" {
			find[*widget.RadioGroup](t, f, "storage").SetSelected("sidecar")
		}
		entry := find[*formEntry](t, f, id)
		entry.SetText("12")
		if f.cards[id].collapsed() {
			t.Errorf("the %s question folded while it was being typed into", id)
		}
	}
}

func TestTickingAMultiselectNeverFolds(t *testing.T) {
	f, _ := build(t, uiDoc)
	group := find[*widget.CheckGroup](t, f, "formats")
	group.SetSelected([]string{"yaml"})
	if f.cards["formats"].collapsed() {
		t.Error("a multiselect folded after one option, before the rest could be ticked")
	}
	group.SetSelected([]string{"yaml", "json"})
	if f.cards["formats"].collapsed() {
		t.Error("a multiselect should stay open until it is finished with")
	}
}

func TestDoneFoldsTheTypedQuestions(t *testing.T) {
	for _, tc := range []struct {
		id   string
		set  func(f *form, t *testing.T)
		want string
	}{
		{id: "name", set: func(f *form, t *testing.T) { find[*formEntry](t, f, "name").SetText("quizme") }, want: "quizme"},
		{id: "budget", set: func(f *form, t *testing.T) { find[*formEntry](t, f, "budget").SetText("12") }, want: "12"},
		{id: "formats", set: func(f *form, t *testing.T) {
			find[*widget.CheckGroup](t, f, "formats").SetSelected([]string{"yaml", "json"})
		}, want: "yaml, json"},
	} {
		t.Run(tc.id, func(t *testing.T) {
			f, _ := build(t, uiDoc)
			c := f.cards[tc.id]
			if c.done == nil {
				t.Fatalf("the %s question has no Done action", tc.id)
			}
			tc.set(f, t)
			c.done.OnTapped()

			if !c.collapsed() {
				t.Errorf("Done did not fold the %s question", tc.id)
			}
			if got := c.header.summary.Text; got != tc.want {
				t.Errorf("the folded row shows %q, want %q", got, tc.want)
			}
			if !c.header.tick.Visible() {
				t.Error("the folded row should be ticked")
			}
		})
	}
}

func TestTheTextareaHasADoneAction(t *testing.T) {
	f, _ := build(t, uiDoc)
	find[*widget.RadioGroup](t, f, "storage").SetSelected("sidecar") // reveals "why"
	c := f.cards["why"]
	if c.done == nil {
		t.Fatal("a textarea has no settling gesture of its own, so it needs Done")
	}
	find[*formEntry](t, f, "why").SetText("because it drifts")
	c.done.OnTapped()
	if !c.collapsed() {
		t.Error("Done did not fold the textarea")
	}
}

// The types that settle on their own must not also carry a Done action: it
// would be a button that never has anything left to do.
func TestOnlyTheTypedQuestionsCarryDone(t *testing.T) {
	f, doc := build(t, uiDoc)
	for _, q := range doc.Questions {
		has := f.cards[q.ID].done != nil
		if want := needsDone(q.Type); has != want {
			t.Errorf("%s question %q: Done present = %v, want %v", q.Type, q.ID, has, want)
		}
	}
}

func TestEnterFoldsASingleLineField(t *testing.T) {
	for _, id := range []string{"name", "budget"} {
		t.Run(id, func(t *testing.T) {
			f, doc := build(t, uiDoc)
			w := test.NewWindow(f.build())
			defer w.Close()

			entry := find[*formEntry](t, f, id)
			w.Canvas().Focus(entry)
			test.Type(entry, "12")
			if f.cards[id].collapsed() {
				t.Fatal("typing should not fold the question")
			}

			// Enter through the shield, exactly as the responder's would arrive.
			entry.TypedKey(&fyne.KeyEvent{Name: fyne.KeyReturn})
			if !f.cards[id].collapsed() {
				t.Error("Enter should fold a single-line field")
			}
			if !doc.Question(id).HasAnswer() {
				t.Error("the answer typed before Enter was not recorded")
			}
		})
	}
}

func TestMovingOnWithoutFinishingLeavesTheQuestionOpen(t *testing.T) {
	f, doc := build(t, uiDoc)
	w := test.NewWindow(f.build())
	defer w.Close()

	name := find[*formEntry](t, f, "name")
	w.Canvas().Focus(name)
	test.Type(name, "quizme")

	// Off to another question, without saying they were done with this one.
	budget := find[*formEntry](t, f, "budget")
	w.Canvas().Focus(budget)
	test.Type(budget, "3")

	if f.cards["name"].collapsed() {
		t.Error("the first question folded when focus left it, mid-answer")
	}
	if got := doc.Question("name").Answer; got != "quizme" {
		t.Errorf("answer = %#v, want it recorded even though the question is still open", got)
	}
}

func TestTheRespondersOwnChoiceIsNotOverridden(t *testing.T) {
	f, _ := build(t, uiDoc)
	c := f.cards["storage"]

	// Folded by hand, then answered: it stays folded, as they left it.
	test.Tap(c.header)
	find[*widget.RadioGroup](t, f, "storage").SetSelected("both")
	if !c.collapsed() {
		t.Error("a question the responder folded should stay folded")
	}

	// Opened by hand, then answered again: it stays open.
	test.Tap(c.header)
	if c.collapsed() {
		t.Fatal("tapping should have opened it")
	}
	find[*widget.RadioGroup](t, f, "storage").SetSelected("in-place")
	if c.collapsed() {
		t.Error("a question the responder opened should not fold itself again")
	}
}

// --- The rest of the form ------------------------------------------------

func TestFoldingIsIndependentOfConditionalVisibility(t *testing.T) {
	f, _ := build(t, uiDoc)
	why := f.cards["why"]
	if why.root.Visible() {
		t.Fatal("the gated question should start hidden")
	}

	find[*widget.RadioGroup](t, f, "storage").SetSelected("sidecar")
	if !why.root.Visible() {
		t.Fatal("answering should reveal the gated question")
	}
	if why.collapsed() {
		t.Error("a question revealed by a condition should appear expanded")
	}

	// Fold it, then send it away and bring it back.
	f.toggle(why)
	find[*widget.RadioGroup](t, f, "storage").SetSelected("both")
	if why.root.Visible() {
		t.Error("the gated question should be hidden again")
	}
	find[*widget.RadioGroup](t, f, "storage").SetSelected("sidecar")
	if !why.root.Visible() {
		t.Error("the gated question should be back")
	}
	if !why.collapsed() {
		t.Error("a question hidden and revealed should come back as the responder left it")
	}
}

func TestFoldingChangesNothingThatIsCounted(t *testing.T) {
	folded, _ := build(t, uiDoc)
	open, _ := build(t, uiDoc)

	answer := func(f *form, t *testing.T) {
		t.Helper()
		find[*widget.RadioGroup](t, f, "storage").SetSelected("both")
		find[*formEntry](t, f, "name").SetText("quizme")
		find[*formEntry](t, f, "budget").SetText("7")
	}
	answer(folded, t)
	answer(open, t)

	// One of them is put away entirely; the other is left as it was.
	for _, c := range folded.cards {
		if !c.collapsed() {
			folded.toggle(c)
		}
	}
	for _, c := range open.cards {
		if c.collapsed() {
			open.toggle(c)
		}
	}

	wasAnswered, wasTotal := open.counts()
	gotAnswered, gotTotal := folded.counts()
	if gotAnswered != wasAnswered || gotTotal != wasTotal {
		t.Errorf("progress folded = %d of %d, expanded = %d of %d", gotAnswered, gotTotal, wasAnswered, wasTotal)
	}
	if folded.progress.Value() != open.progress.Value() {
		t.Errorf("progress bar folded = %v, expanded = %v", folded.progress.Value(), open.progress.Value())
	}
	if folded.progressText() != open.progressText() {
		t.Errorf("progress label folded = %q, expanded = %q", folded.progressText(), open.progressText())
	}
	if got, want := len(folded.doc.Unanswered()), len(open.doc.Unanswered()); got != want {
		t.Errorf("outstanding folded = %d, expanded = %d", got, want)
	}
	if folded.dirty() != open.dirty() {
		t.Errorf("dirty folded = %v, expanded = %v", folded.dirty(), open.dirty())
	}
}

func TestARefusedSubmitOpensWhatItFlags(t *testing.T) {
	f, doc := build(t, uiDoc)
	c := f.cards["storage"] // the only required question in the fixture

	// Put it away unanswered, then try to submit.
	test.Tap(c.header)
	if !c.collapsed() {
		t.Fatal("the question should be folded before the submit")
	}
	f.scroll.Resize(fyne.NewSize(600, 400))
	f.scroll.Content.Resize(fyne.NewSize(600, 2000))
	f.scroll.Offset = fyne.NewPos(0, 400)

	f.submit()

	if f.outcome == questionnaire.StatusSubmitted {
		t.Fatal("the submit should have been refused")
	}
	if c.collapsed() {
		t.Error("a question standing in the way of a submit must not stay folded")
	}
	if !c.warning.Visible() {
		t.Error("the warning should be showing on the question that was opened")
	}
	if got := f.scroll.Offset.Y; got != c.root.Position().Y {
		t.Errorf("the page is at %v, want the flagged question at %v", got, c.root.Position().Y)
	}
	if len(doc.Unanswered()) != 1 {
		t.Errorf("outstanding = %d, want 1", len(doc.Unanswered()))
	}

	// And answering it now settles it as any other answer would.
	find[*widget.RadioGroup](t, f, "storage").SetSelected("both")
	if !c.collapsed() {
		t.Error("answering a question opened by a refused submit should fold it")
	}
	if c.warning.Visible() {
		t.Error("the warning should have cleared")
	}
}

func TestARefusedSubmitOpensEveryFlaggedQuestionInOrder(t *testing.T) {
	f, _ := build(t, `
title: t
questions:
  - {id: first, type: text, prompt: "First?", required: true}
  - {id: second, type: text, prompt: "Second?", required: true}
  - {id: third, type: text, prompt: "Third?", required: true}
`)
	for _, id := range []string{"first", "third"} {
		test.Tap(f.cards[id].header)
	}
	f.scroll.Resize(fyne.NewSize(600, 400))
	f.scroll.Content.Resize(fyne.NewSize(600, 2000))
	f.scroll.Offset = fyne.NewPos(0, 300)

	f.submit()

	for _, id := range []string{"first", "second", "third"} {
		c := f.cards[id]
		if c.collapsed() {
			t.Errorf("%q was left folded despite being flagged", id)
		}
		if !c.warning.Visible() {
			t.Errorf("%q was not flagged", id)
		}
	}
	if got, want := f.scroll.Offset.Y, f.cards["first"].root.Position().Y; got != want {
		t.Errorf("the page moved to %v, want the first flagged question at %v", got, want)
	}
}

func TestFoldingReachesNeitherTheFileNorTheOutput(t *testing.T) {
	answer := func(f *form, t *testing.T) {
		t.Helper()
		find[*widget.RadioGroup](t, f, "storage").SetSelected("both")
		find[*formEntry](t, f, "name").SetText("quizme")
		f.cards["name"].comment.Entry().SetText("a note")
	}

	folded, foldedDoc := build(t, uiDoc)
	answer(folded, t)
	for _, c := range folded.cards {
		if !c.collapsed() {
			folded.toggle(c)
		}
	}
	folded.submit()
	if folded.outcome != questionnaire.StatusSubmitted {
		t.Fatalf("outcome = %q", folded.outcome)
	}

	open, openDoc := build(t, uiDoc)
	answer(open, t)
	for _, c := range open.cards {
		if c.collapsed() {
			open.toggle(c)
		}
	}
	open.submit()

	foldedBytes, err := foldedDoc.Render(questionnaire.StatusSubmitted, fixedTestTime())
	if err != nil {
		t.Fatal(err)
	}
	openBytes, err := openDoc.Render(questionnaire.StatusSubmitted, fixedTestTime())
	if err != nil {
		t.Fatal(err)
	}
	if string(foldedBytes) != string(openBytes) {
		t.Errorf("folding changed the file written:\n--- folded ---\n%s\n--- expanded ---\n%s", foldedBytes, openBytes)
	}
	for _, word := range []string{"fold", "collaps", "expand"} {
		if strings.Contains(strings.ToLower(string(foldedBytes)), word) {
			t.Errorf("the written file mentions %q; fold state belongs to the session, not the document", word)
		}
	}

	foldedJSON, err := json.Marshal(foldedDoc.Result(questionnaire.StatusSubmitted))
	if err != nil {
		t.Fatal(err)
	}
	openJSON, err := json.Marshal(openDoc.Result(questionnaire.StatusSubmitted))
	if err != nil {
		t.Fatal(err)
	}
	if string(foldedJSON) != string(openJSON) {
		t.Errorf("folding changed what is printed:\n%s\nvs\n%s", foldedJSON, openJSON)
	}
}

// Nothing about the folds is written down; what a reopened questionnaire shows
// folded is decided from the answers in it, and nothing else.
func TestReopeningFollowsTheAnswersNotTheFolds(t *testing.T) {
	f, doc := build(t, uiDoc)
	find[*widget.RadioGroup](t, f, "storage").SetSelected("both")
	for _, c := range f.cards {
		if !c.collapsed() {
			f.toggle(c)
		}
	}
	f.submit()

	out, err := doc.Render(questionnaire.StatusSubmitted, fixedTestTime())
	if err != nil {
		t.Fatal(err)
	}

	again, reopened := build(t, string(out))
	for _, q := range reopened.Questions {
		if got, want := again.cards[q.ID].collapsed(), q.HasAnswer(); got != want {
			t.Errorf("question %q came back collapsed = %v, want %v for HasAnswer = %v",
				q.ID, got, want, q.HasAnswer())
		}
	}

	// The responder folded every question, answered or not, before submitting.
	// None of that was kept: the unanswered ones are open again.
	open := 0
	for _, q := range reopened.Questions {
		if !again.cards[q.ID].collapsed() {
			open++
		}
	}
	if open == 0 {
		t.Error("everything came back folded, so some fold state survived the file")
	}
}
