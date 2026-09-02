package ui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/markwylde/interrogate/internal/questionnaire"
)

// --- Scrolling header ---------------------------------------------------

func TestHeaderScrollsWithTheQuestions(t *testing.T) {
	f, _ := build(t, uiDoc)

	// The header is the first thing in the scrolling list, not pinned above it.
	if len(f.list.Objects) == 0 {
		t.Fatal("the question list is empty")
	}
	head := f.list.Objects[0]
	for id, c := range f.cards {
		if c.root == head {
			t.Fatalf("the first item in the list is the card for %q, not the header", id)
		}
	}
	if _, ok := search[*widget.Label](head, func(l *widget.Label) bool { return l.Text == "Decisions" }); !ok {
		t.Error("the title is not in the scrolling list")
	}
}

func TestActionsStayReachable(t *testing.T) {
	f, _ := build(t, uiDoc)
	// The footer is built separately from the scrolling list, so it cannot
	// scroll away whatever the responder does.
	footer := f.footer()
	if _, ok := search[*widget.Button](footer, func(b *widget.Button) bool { return b.Text == "Submit" }); !ok {
		t.Error("Submit is not in the pinned footer")
	}
	if _, ok := search[*widget.Button](footer, func(b *widget.Button) bool { return b.Text == "Dismiss" }); !ok {
		t.Error("Dismiss is not in the pinned footer")
	}
	for _, obj := range f.list.Objects {
		if _, ok := search[*widget.Button](obj, func(b *widget.Button) bool { return b.Text == "Submit" }); ok {
			t.Error("Submit is inside the scrolling list, so it would scroll away")
		}
	}
}

// --- Collapsible comments -----------------------------------------------

func TestCommentStartsCollapsed(t *testing.T) {
	f, _ := build(t, uiDoc)
	c := f.cards["name"].comment
	if c.Expanded() {
		t.Error("a comment field should rest collapsed")
	}
	if c.FieldVisible() {
		t.Error("the entry should be hidden while collapsed")
	}
	if got := c.toggle.Text; got != "Add comment" {
		t.Errorf("toggle = %q, want %q", got, "Add comment")
	}
}

func TestCommentOpensOnDemand(t *testing.T) {
	f, _ := build(t, uiDoc)
	c := f.cards["name"].comment

	c.Toggle()
	if !c.Expanded() || !c.FieldVisible() {
		t.Error("pressing the toggle should open the field")
	}
	if got := c.toggle.Text; got != "Hide comment" {
		t.Errorf("toggle = %q while open", got)
	}

	c.Toggle()
	if c.Expanded() || c.FieldVisible() {
		t.Error("pressing it again should close the field")
	}
}

func TestExistingCommentStartsOpen(t *testing.T) {
	f, _ := build(t, `
title: t
questions:
  - {id: q, type: text, prompt: "Name?", comment: written last time}
`)
	c := f.cards["q"].comment
	if !c.Expanded() {
		t.Error("a question that arrives with a comment should show it")
	}
	if got := c.Entry().Text; got != "written last time" {
		t.Errorf("entry = %q", got)
	}
}

func TestCollapsedCommentIsNeverInvisible(t *testing.T) {
	f, doc := build(t, uiDoc)
	c := f.cards["name"].comment

	c.Toggle()
	c.Entry().SetText("worth a conversation")
	c.Toggle()

	if c.Expanded() {
		t.Fatal("the field should be closed")
	}
	if !strings.HasPrefix(c.toggle.Text, "Comment: ") {
		t.Errorf("toggle = %q, want it to show the comment", c.toggle.Text)
	}
	if !strings.Contains(c.toggle.Text, "worth a conversation") {
		t.Errorf("toggle = %q, want it to preview the text", c.toggle.Text)
	}
	// And it is still recorded, collapsed or not.
	if doc.Question("name").Comment != "worth a conversation" {
		t.Errorf("comment = %q", doc.Question("name").Comment)
	}
}

func TestCollapsedCommentPreviewIsOneShortLine(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{"short", "fine", "Comment: fine"},
		{"multi-line", "first line\nsecond line", "Comment: first line…"},
		{"long", strings.Repeat("a", 80), "Comment: " + strings.Repeat("a", previewLimit) + "…"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f, _ := build(t, uiDoc)
			c := f.cards["name"].comment
			c.Toggle()
			c.Entry().SetText(tc.text)
			c.Toggle()
			if c.toggle.Text != tc.want {
				t.Errorf("toggle = %q, want %q", c.toggle.Text, tc.want)
			}
		})
	}
}

func TestCollapsedCommentIsStillWritten(t *testing.T) {
	// The field being hidden must not change what reaches the file.
	f, doc := build(t, uiDoc)
	c := f.cards["name"].comment
	c.Toggle()
	c.Entry().SetText("kept")
	c.Toggle()

	find[*widget.RadioGroup](t, f, "storage").SetSelected("both")
	f.submit()
	if f.outcome != questionnaire.StatusSubmitted {
		t.Fatalf("outcome = %q", f.outcome)
	}

	out, err := doc.Render(questionnaire.StatusSubmitted, fixedTestTime())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "comment: kept") {
		t.Errorf("the collapsed comment was not written:\n%s", out)
	}
}

func TestCollapsedCommentCountsAsDirty(t *testing.T) {
	f, _ := build(t, uiDoc)
	c := f.cards["name"].comment
	c.Toggle()
	c.Entry().SetText("a note")
	c.Toggle()
	if !f.dirty() {
		t.Error("a comment written then collapsed is still unsaved work")
	}
}

func TestTogglingAnEmptyCommentIsNotDirty(t *testing.T) {
	f, _ := build(t, uiDoc)
	f.cards["name"].comment.Toggle()
	if f.dirty() {
		t.Error("opening an empty comment field is not a change")
	}
}

// --- Progress -----------------------------------------------------------

func TestProgressStartsEmpty(t *testing.T) {
	f, _ := build(t, uiDoc)
	if got := f.progress.Value(); got != 0 {
		t.Errorf("progress = %v, want 0 on a fresh questionnaire", got)
	}
	// storage_why is gated, so it is not among the applicable questions.
	if got := f.progressText(); got != "0 of 7 answered" {
		t.Errorf("progress label = %q", got)
	}
}

func TestProgressFollowsAnswers(t *testing.T) {
	f, _ := build(t, uiDoc)

	find[*widget.Entry](t, f, "name").SetText("interrogate")
	if got := f.progress.Value(); !closeTo(got, 1.0/7.0) {
		t.Errorf("progress = %v, want one seventh", got)
	}
	find[*widget.RadioGroup](t, f, "ship").SetSelected("No")
	if got := f.progress.Value(); !closeTo(got, 2.0/7.0) {
		t.Errorf("progress = %v, want two sevenths", got)
	}

	// Clearing an answer takes it back off the count.
	find[*widget.Entry](t, f, "name").SetText("")
	if got := f.progress.Value(); !closeTo(got, 1.0/7.0) {
		t.Errorf("progress = %v after clearing", got)
	}
}

func TestProgressReadsFullWhenEverythingApplicableIsAnswered(t *testing.T) {
	f, _ := build(t, `
title: t
questions:
  - {id: a, type: text, prompt: "A?"}
  - {id: b, type: text, prompt: "B?"}
`)
	find[*widget.Entry](t, f, "a").SetText("x")
	find[*widget.Entry](t, f, "b").SetText("y")
	if got := f.progress.Value(); !closeTo(got, 1) {
		t.Errorf("progress = %v, want complete", got)
	}
	if got := f.progressText(); got != "2 of 2 answered" {
		t.Errorf("progress label = %q", got)
	}
}

func TestProgressIgnoresHiddenQuestions(t *testing.T) {
	f, _ := build(t, uiDoc)
	if got := f.progressText(); strings.HasSuffix(got, "of 8 answered") {
		t.Errorf("progress = %q, but one question is hidden", got)
	}
}

func TestProgressTotalGrowsWhenAQuestionAppears(t *testing.T) {
	f, _ := build(t, uiDoc)
	// Answering the gate reveals its follow-up, so the total goes up.
	find[*widget.RadioGroup](t, f, "storage").SetSelected("sidecar")
	if got := f.progressText(); got != "1 of 8 answered" {
		t.Errorf("progress = %q, want the revealed question counted", got)
	}
	find[*widget.RadioGroup](t, f, "storage").SetSelected("both")
	if got := f.progressText(); got != "1 of 7 answered" {
		t.Errorf("progress = %q, want the hidden question dropped", got)
	}
}

func closeTo(got, want float64) bool {
	d := got - want
	if d < 0 {
		d = -d
	}
	return d < 0.0001
}

func TestEmptyProgressBarLooksEmpty(t *testing.T) {
	// The whole reason for a custom bar: Fyne's own paints its track in a faded
	// primary, so nothing-answered reads as partly-done.
	p := newProgressBar()
	p.CreateRenderer()
	p.SetValue(0)
	p.Resize(fyne.NewSize(100, progressHeight))

	if p.fill.Size().Width != 0 {
		t.Errorf("fill width = %v at zero, want 0", p.fill.Size().Width)
	}
	if p.track.FillColor == p.fill.FillColor {
		t.Error("the track and the fill are the same colour, so empty looks full")
	}
}

func TestProgressBarFillsProportionally(t *testing.T) {
	p := newProgressBar()
	p.CreateRenderer()
	p.Resize(fyne.NewSize(100, progressHeight))

	for _, tc := range []struct{ value, want float64 }{{0, 0}, {0.25, 25}, {0.5, 50}, {1, 100}} {
		p.SetValue(tc.value)
		p.Resize(fyne.NewSize(100, progressHeight))
		if got := float64(p.fill.Size().Width); !closeTo(got, tc.want) {
			t.Errorf("value %v filled %v of 100, want %v", tc.value, got, tc.want)
		}
	}
}

func TestProgressBarClampsOutOfRange(t *testing.T) {
	p := newProgressBar()
	p.SetValue(-1)
	if got := p.Value(); got != 0 {
		t.Errorf("Value = %v after setting -1, want 0", got)
	}
	p.SetValue(5)
	if got := p.Value(); got != 1 {
		t.Errorf("Value = %v after setting 5, want 1", got)
	}
}

// --- Scrolling over a text field ----------------------------------------

// scrollable lists the question types whose control is a text field, and so
// carries an internal scroller of Fyne's that would otherwise eat the gesture.
var scrollableFields = map[string]questionnaire.Type{
	"why":    questionnaire.TypeTextarea,
	"name":   questionnaire.TypeText,
	"budget": questionnaire.TypeNumber,
}

func TestScrollOverATextFieldMovesThePage(t *testing.T) {
	// Fyne's Entry has an internal scroller that consumes scroll events even
	// with nothing to scroll, and events do not bubble. Without the shield, a
	// trackpad scroll stops dead when it crosses a text field -- of any kind,
	// since a single-line entry scrolls its content sideways and so has a live
	// scroller too.
	for id := range scrollableFields {
		t.Run(id, func(t *testing.T) {
			f, _ := build(t, uiDoc)
			f.scroll.Resize(fyne.NewSize(600, 400))
			f.list.Resize(fyne.NewSize(600, 2000))
			f.scroll.Content.Resize(fyne.NewSize(600, 2000))

			shield := findShield(t, f, id)
			before := f.scroll.Offset.Y
			shield.Scrolled(&fyne.ScrollEvent{Scrolled: fyne.NewDelta(0, -40)})

			if f.scroll.Offset.Y == before {
				t.Error("scrolling over a text field did not move the page")
			}
		})
	}
}

func TestScrollOverATextFieldStopsAtTheEnds(t *testing.T) {
	for id := range scrollableFields {
		t.Run(id, func(t *testing.T) {
			f, _ := build(t, uiDoc)
			f.scroll.Resize(fyne.NewSize(600, 400))
			f.scroll.Content.Resize(fyne.NewSize(600, 2000))
			shield := findShield(t, f, id)

			// Scrolling up at the top must not run past the start of the page.
			shield.Scrolled(&fyne.ScrollEvent{Scrolled: fyne.NewDelta(0, 500)})
			if got := f.scroll.Offset.Y; got != 0 {
				t.Errorf("offset = %v after scrolling up at the top, want 0", got)
			}

			// Nor past the end.
			shield.Scrolled(&fyne.ScrollEvent{Scrolled: fyne.NewDelta(0, -99999)})
			if got := f.scroll.Offset.Y; got > 1600 {
				t.Errorf("offset = %v, want it capped at the end of the page", got)
			}
		})
	}
}

func TestScrollOverATextFieldIgnoresSideways(t *testing.T) {
	for id := range scrollableFields {
		t.Run(id, func(t *testing.T) {
			f, _ := build(t, uiDoc)
			f.scroll.Resize(fyne.NewSize(600, 400))
			f.scroll.Content.Resize(fyne.NewSize(600, 2000))
			findShield(t, f, id).Scrolled(&fyne.ScrollEvent{Scrolled: fyne.NewDelta(120, 0)})
			if got := f.scroll.Offset.X; got != 0 {
				t.Errorf("offset.X = %v, want the page never to scroll sideways", got)
			}
		})
	}
}

func TestEveryTextFieldIsShielded(t *testing.T) {
	f, doc := build(t, uiDoc)
	for _, q := range doc.Questions {
		c := f.cards[q.ID]
		// Every comment field is multi-line, whatever the question's type.
		if _, ok := search[*scrollShield](c.comment.shielded(), first[*scrollShield]()); !ok {
			t.Errorf("the comment field for %q is not shielded", q.ID)
		}
		// The control itself, rather than the whole card: the comment box's
		// own shield would otherwise answer for a control that has none.
		if !rendersATextField(q.Type) {
			continue
		}
		if _, ok := search[*scrollShield](c.control, first[*scrollShield]()); !ok {
			t.Errorf("the %s control for %q is not shielded", q.Type, q.ID)
		}
	}
}

// TestEveryTextFieldIsShielded only sees the types uiDoc happens to use, so
// this keeps the fixture honest: a text-field type it never asks for would go
// unchecked.
func TestTheFixtureCoversEveryTextField(t *testing.T) {
	_, doc := build(t, uiDoc)
	for _, want := range []questionnaire.Type{
		questionnaire.TypeText, questionnaire.TypeTextarea, questionnaire.TypeNumber,
	} {
		found := false
		for _, q := range doc.Questions {
			if q.Type == want {
				found = true
			}
		}
		if !found {
			t.Errorf("uiDoc has no %s question, so the shielding test never checks one", want)
		}
	}
}

func rendersATextField(t questionnaire.Type) bool {
	switch t {
	case questionnaire.TypeText, questionnaire.TypeTextarea, questionnaire.TypeNumber:
		return true
	}
	return false
}

// findShield returns the shield over a question's own control, not whichever
// one the card happens to hold first.
func findShield(t *testing.T, f *form, id string) *scrollShield {
	t.Helper()
	c, ok := f.cards[id]
	if !ok {
		t.Fatalf("no card for %q", id)
	}
	found, ok := search[*scrollShield](c.control, first[*scrollShield]())
	if !ok {
		t.Fatalf("no scroll shield on the control for %q", id)
	}
	return found
}

// --- A shielded field is still a field -----------------------------------

func TestShieldedTextFieldStillAnswersAndCounts(t *testing.T) {
	f, doc := build(t, uiDoc)
	answered, before := f.counts()

	entry := find[*widget.Entry](t, f, "name")
	entry.SetText("interrogate")

	if got := doc.Question("name").Answer; got != "interrogate" {
		t.Errorf("answer = %#v, want %q", got, "interrogate")
	}
	nowAnswered, total := f.counts()
	if nowAnswered != answered+1 {
		t.Errorf("answered = %d, want %d: a shielded field should count toward progress", nowAnswered, answered+1)
	}
	if total != before {
		t.Errorf("total = %d, want %d: shielding should not change what is counted", total, before)
	}
}

func TestShieldedFieldsStillTakeTypingAndSelection(t *testing.T) {
	f, doc := build(t, uiDoc)
	w := test.NewWindow(f.build())
	defer w.Close()

	for _, id := range []string{"name", "budget"} {
		entry := find[*widget.Entry](t, f, id)
		w.Canvas().Focus(entry)
		if w.Canvas().Focused() != entry {
			t.Fatalf("the %s field did not take focus through the shield", id)
		}
		test.Type(entry, "12")
		if entry.Text != "12" {
			t.Errorf("the %s field holds %q after typing, want %q", id, entry.Text, "12")
		}
		entry.TypedShortcut(&fyne.ShortcutSelectAll{})
		if got := entry.SelectedText(); got != "12" {
			t.Errorf("selecting all of the %s field gave %q, want %q", id, got, "12")
		}
	}

	if got := doc.Question("name").Answer; got != "12" {
		t.Errorf("typing into a shielded text field recorded %#v", got)
	}
	if got := doc.Question("budget").Answer; got != 12.0 {
		t.Errorf("typing into a shielded number field recorded %#v", got)
	}
}

// The shield sits over the field, so anything it implements it takes away.
// Scrolling is the whole of its business.
func TestTheShieldAnswersOnlyToScrolling(t *testing.T) {
	var s any = newScrollShield(nil)
	if _, ok := s.(fyne.Scrollable); !ok {
		t.Error("the shield must be scrollable, or it shields nothing")
	}
	for name, taken := range map[string]bool{
		"Tappable":          asserted[fyne.Tappable](s),
		"SecondaryTappable": asserted[fyne.SecondaryTappable](s),
		"DoubleTappable":    asserted[fyne.DoubleTappable](s),
		"Focusable":         asserted[fyne.Focusable](s),
		"Draggable":         asserted[fyne.Draggable](s),
		"Hoverable":         asserted[desktop.Hoverable](s),
		"Mouseable":         asserted[desktop.Mouseable](s),
		"Cursorable":        asserted[desktop.Cursorable](s),
	} {
		if taken {
			t.Errorf("the shield implements %s, so it takes those events from the field", name)
		}
	}
}

func asserted[T any](v any) bool {
	_, ok := v.(T)
	return ok
}

// A single-line field has far less slack than a four-row textarea, so the
// shield must not add so much as a pixel to the height the field asks for.
func TestShieldingDoesNotChangeAFieldsSize(t *testing.T) {
	f, _ := build(t, uiDoc)
	for _, id := range []string{"name", "budget", "why"} {
		entry := find[*widget.Entry](t, f, id)
		if got, want := f.cards[id].control.MinSize(), entry.MinSize(); got != want {
			t.Errorf("the shielded %s control asks for %v, the field itself for %v", id, got, want)
		}
	}
}
