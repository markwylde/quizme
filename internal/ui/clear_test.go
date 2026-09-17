package ui

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/markwylde/quizme/internal/questionnaire"
)

// inWindow builds a form into a test window the form knows about, as Run does,
// so that dialogs and rebuilds have somewhere to go.
func inWindow(t *testing.T, src string) (*form, *questionnaire.Document, fyne.Window) {
	t.Helper()
	f, doc := build(t, src)
	win := test.NewWindow(f.build())
	t.Cleanup(func() {
		if f.outcome == "" { // finishing closes the window itself
			win.Close()
		}
	})
	f.win = win
	win.Resize(fyne.NewSize(780, 900))
	return f, doc, win
}

// fillIn answers and comments on most of uiDoc through its controls.
func fillIn(t *testing.T, f *form) {
	t.Helper()
	find[*widget.RadioGroup](t, f, "storage").SetSelected("sidecar")
	find[*formEntry](t, f, "why").SetText("drift")
	find[*widget.CheckGroup](t, f, "formats").SetSelected([]string{"yaml"})
	find[*formEntry](t, f, "name").SetText("quizme")
	find[*formEntry](t, f, "budget").SetText("12")
	find[*widget.RadioGroup](t, f, "ship").SetSelected("Yes")
	rank := find[*rankWidget](t, f, "order")
	rank.CreateRenderer()
	rank.move(0, 1)
	f.cards["name"].comment.Entry().SetText("a note")
	f.cards["storage"].comment.Entry().SetText("another")
}

// button finds a button by its label anywhere under root.
func button(root fyne.CanvasObject, label string) (*widget.Button, bool) {
	return deepSearch[*widget.Button](root, func(b *widget.Button) bool { return b.Text == label })
}

// deepSearch is search that also looks inside widgets, which a dialog's
// content is buried in.
func deepSearch[T fyne.CanvasObject](obj fyne.CanvasObject, match func(T) bool) (T, bool) {
	var zero T
	if typed, ok := obj.(T); ok && match(typed) {
		return typed, true
	}
	var children []fyne.CanvasObject
	switch o := obj.(type) {
	case *fyne.Container:
		children = o.Objects
	case fyne.Widget:
		children = test.WidgetRenderer(o).Objects()
	}
	for _, child := range children {
		if found, ok := deepSearch[T](child, match); ok {
			return found, true
		}
	}
	return zero, false
}

func TestClearingEmptiesEveryAnswerAndComment(t *testing.T) {
	f, doc, _ := inWindow(t, uiDoc)
	fillIn(t, f)
	f.finished(doc.Question("name"))

	f.clearAll()

	for _, q := range doc.Questions {
		if q.HasAnswer() || q.Answer != nil {
			t.Errorf("%s still answers %#v", q.ID, q.Answer)
		}
		if q.Comment != "" {
			t.Errorf("%s still has the comment %q", q.ID, q.Comment)
		}
	}
	if got := find[*formEntry](t, f, "name").Text; got != "" {
		t.Errorf("the name field shows %q", got)
	}
	if got := f.cards["name"].comment.Entry().Text; got != "" {
		t.Errorf("the comment field shows %q", got)
	}
	if got := find[*widget.RadioGroup](t, f, "storage").Selected; got != "" {
		t.Errorf("the radio group shows %q", got)
	}
	if answered, _ := f.counts(); answered != 0 || f.progress.value != 0 {
		t.Errorf("progress shows %d answered (%v)", answered, f.progress.value)
	}
	if got := f.progressText(); got != "0 of 7 answered" {
		t.Errorf("progress text = %q", got)
	}
}

func TestClearingReachesHiddenAnswers(t *testing.T) {
	f, doc, _ := inWindow(t, uiDoc)
	fillIn(t, f)
	// Hide the answered question behind its condition.
	find[*widget.RadioGroup](t, f, "storage").SetSelected("both")
	if f.cards["why"].root.Visible() {
		t.Fatal("setup: why should be hidden")
	}

	f.clearAll()
	if doc.Question("why").Answer != nil {
		t.Errorf("the hidden answer survived: %#v", doc.Question("why").Answer)
	}

	find[*widget.RadioGroup](t, f, "storage").SetSelected("sidecar")
	if !f.cards["why"].root.Visible() {
		t.Fatal("why should show again")
	}
	if got := find[*formEntry](t, f, "why").Text; got != "" || doc.Question("why").HasAnswer() {
		t.Errorf("the hidden answer came back: %q", got)
	}
}

func TestClearingReturnsARankingToItsAuthoredOrder(t *testing.T) {
	f, doc, _ := inWindow(t, uiDoc)
	fillIn(t, f)
	f.clearAll()

	rank := find[*rankWidget](t, f, "order")
	if got := rank.Order(); !reflect.DeepEqual(got, []string{"correctness", "speed", "looks"}) {
		t.Errorf("order = %v", got)
	}
	if rank.Touched() || doc.Question("order").HasAnswer() {
		t.Error("the cleared ranking still counts as answered")
	}
}

func TestClearingExpandsEverythingAndRehidesConditions(t *testing.T) {
	f, _, _ := inWindow(t, uiDoc)
	fillIn(t, f)
	for _, c := range f.cards {
		c.fold = foldClosed
		f.applyFold(c)
	}

	f.clearAll()
	for id, c := range f.cards {
		if c.collapsed() || c.pinned() {
			t.Errorf("%s is %v after clearing, want expanded and automatic", id, c.fold)
		}
	}
	if f.cards["why"].root.Visible() {
		t.Error("the conditional question is still showing")
	}
}

func TestTheClearButtonIsInThePinnedFooter(t *testing.T) {
	f, _, win := inWindow(t, uiDoc)
	if _, ok := search[*widget.Button](f.scroll.Content, func(b *widget.Button) bool { return b.Text == "Clear all answers" }); ok {
		t.Error("the clear button scrolls with the page")
	}
	b, ok := search[*widget.Button](win.Content(), func(b *widget.Button) bool { return b.Text == "Clear all answers" })
	if !ok {
		t.Fatal("no Clear all answers button")
	}
	f.scroll.ScrollToBottom()
	if !b.Visible() {
		t.Error("the clear button is hidden once scrolled")
	}
}

// clearDialog opens the confirmation and returns what it shows.
func clearDialog(t *testing.T, f *form, win fyne.Window) fyne.CanvasObject {
	t.Helper()
	b, ok := search[*widget.Button](win.Content(), func(b *widget.Button) bool { return b.Text == "Clear all answers" })
	if !ok {
		t.Fatal("no Clear all answers button")
	}
	test.Tap(b)
	top := win.Canvas().Overlays().Top()
	if top == nil {
		t.Fatal("no confirmation was shown")
	}
	return top
}

func TestTheClearConfirmation(t *testing.T) {
	f, doc, win := inWindow(t, uiDoc)
	fillIn(t, f)

	top := clearDialog(t, f, win)
	if _, ok := deepSearch[*widget.Label](top, func(l *widget.Label) bool {
		return l.Text == "This will wipe all answers and comments. Are you sure?"
	}); !ok {
		t.Error("the confirmation does not carry the warning")
	}
	if doc.Question("name").Answer != "quizme" {
		t.Fatal("answers were cleared before confirming")
	}

	cancel, ok := button(top, "Cancel")
	if !ok {
		t.Fatal("no Cancel")
	}
	folds := map[string]foldState{}
	for id, c := range f.cards {
		folds[id] = c.fold
	}
	test.Tap(cancel)
	if doc.Question("name").Answer != "quizme" || doc.Question("name").Comment != "a note" {
		t.Error("cancelling changed an answer or comment")
	}
	for id, c := range f.cards {
		if c.fold != folds[id] {
			t.Errorf("cancelling changed %s's fold", id)
		}
	}

	top = clearDialog(t, f, win)
	confirm, ok := button(top, "Clear all")
	if !ok {
		t.Fatal("no Clear all")
	}
	if confirm.Importance != widget.DangerImportance {
		t.Error("the confirming button should read as dangerous")
	}
	test.Tap(confirm)
	if doc.Question("name").Answer != nil || doc.Question("name").Comment != "" {
		t.Error("confirming did not clear")
	}
}

const answeredDoc = `title: t
questions:
  - id: name
    type: text
    prompt: Name?
    required: true
    answer: quizme
    comment: from last time
  - id: pick
    type: select
    prompt: Pick
    options: [a, b]
`

func writeDoc(t *testing.T, body string) (*questionnaire.Document, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "q.yaml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := questionnaire.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return doc, path
}

func TestClearingAnAnsweredQuestionnaireAsksOnClose(t *testing.T) {
	doc, path := writeDoc(t, answeredDoc)
	f := newForm(doc, nil)
	f.build()
	f.clearAll()

	if !f.dirty() {
		t.Fatal("clearing answers that are in the file should count as a change")
	}

	// Saving writes the cleared questionnaire, which still loads.
	if err := doc.Save(f.savedStatus(), time.Now()); err != nil {
		t.Fatal(err)
	}
	saved, err := questionnaire.Load(path)
	if err != nil {
		t.Fatalf("the saved file does not validate: %v", err)
	}
	if q := saved.Question("name"); q.HasAnswer() || q.Comment != "" {
		t.Errorf("the saved file still carries %#v / %q", q.Answer, q.Comment)
	}
}

func TestDiscardingAfterClearingLeavesTheFile(t *testing.T) {
	doc, path := writeDoc(t, answeredDoc)
	f := newForm(doc, nil)
	f.build()
	f.clearAll()
	if raw, _ := os.ReadFile(path); string(raw) != answeredDoc {
		t.Error("clearing wrote the file")
	}

	// Discarding is a dismissal, which records only the status: the answers
	// that were in the file stay in it.
	if err := doc.Save(questionnaire.StatusDismissed, time.Now()); err != nil {
		t.Fatal(err)
	}
	saved, err := questionnaire.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if q := saved.Question("name"); q.Answer != "quizme" || q.Comment != "from last time" {
		t.Errorf("discarding lost what was in the file: %#v / %q", q.Answer, q.Comment)
	}
}

func TestClearingThisSessionsAnswersClosesWithoutAsking(t *testing.T) {
	f, _, _ := inWindow(t, uiDoc)
	fillIn(t, f)
	if !f.dirty() {
		t.Fatal("setup: the form should be dirty")
	}
	f.clearAll()
	if f.dirty() {
		t.Fatal("a form cleared back to how it opened should not be dirty")
	}
	f.requestClose()
	if f.outcome != questionnaire.StatusDismissed {
		t.Errorf("outcome = %q, want dismissed", f.outcome)
	}
}

func TestSubmittingAfterClearingIsRefused(t *testing.T) {
	f, _, _ := inWindow(t, uiDoc)
	fillIn(t, f)
	f.clearAll()
	f.submit()
	if f.outcome != "" {
		t.Errorf("outcome = %q, want the form to stay open", f.outcome)
	}
	if !f.cards["storage"].warning.Visible() {
		t.Error("the outstanding required question is not flagged")
	}
}

func TestClearingKeepsTheTextSize(t *testing.T) {
	restore := useTheme(t, newTheme())
	defer restore()
	f, _, _ := inWindow(t, uiDoc)
	f.setTextSize(150)
	f.clearAll()
	if f.textSize != 150 || scaled(10) != 15 {
		t.Errorf("after clearing the size is %d (scale %v)", f.textSize, scaled(10)/10)
	}
	if f.larger == nil || f.larger.Disabled() {
		t.Error("the rebuilt size buttons are missing or wrong")
	}
}
