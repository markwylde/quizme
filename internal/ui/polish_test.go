package ui

import (
	"strings"
	"testing"

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
	if c.Entry().Visible() {
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
	if !c.Expanded() || !c.Entry().Visible() {
		t.Error("pressing the toggle should open the field")
	}
	if got := c.toggle.Text; got != "Hide comment" {
		t.Errorf("toggle = %q while open", got)
	}

	c.Toggle()
	if c.Expanded() || c.Entry().Visible() {
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

func TestProgressStartsAtNone(t *testing.T) {
	f, _ := build(t, uiDoc)
	// storage_why is gated, so it is not among the applicable questions.
	if got := f.progress.Text; got != "0 of 7 answered" {
		t.Errorf("progress = %q", got)
	}
}

func TestProgressFollowsAnswers(t *testing.T) {
	f, _ := build(t, uiDoc)
	find[*widget.Entry](t, f, "name").SetText("interrogate")
	if got := f.progress.Text; got != "1 of 7 answered" {
		t.Errorf("progress = %q", got)
	}
	find[*widget.RadioGroup](t, f, "ship").SetSelected("No")
	if got := f.progress.Text; got != "2 of 7 answered" {
		t.Errorf("progress = %q", got)
	}
	// Clearing an answer takes it back off the count.
	find[*widget.Entry](t, f, "name").SetText("")
	if got := f.progress.Text; got != "1 of 7 answered" {
		t.Errorf("progress = %q", got)
	}
}

func TestProgressIgnoresHiddenQuestions(t *testing.T) {
	f, _ := build(t, uiDoc)
	if strings.HasSuffix(f.progress.Text, "of 8 answered") {
		t.Errorf("progress = %q, but one question is hidden", f.progress.Text)
	}
}

func TestProgressTotalGrowsWhenAQuestionAppears(t *testing.T) {
	f, _ := build(t, uiDoc)
	// Answering the gate reveals its follow-up, so the total goes up by two:
	// one for the answer given, one for the question it uncovered.
	find[*widget.RadioGroup](t, f, "storage").SetSelected("sidecar")
	if got := f.progress.Text; got != "1 of 8 answered" {
		t.Errorf("progress = %q, want the revealed question counted", got)
	}
	find[*widget.RadioGroup](t, f, "storage").SetSelected("both")
	if got := f.progress.Text; got != "1 of 7 answered" {
		t.Errorf("progress = %q, want the hidden question dropped", got)
	}
}
