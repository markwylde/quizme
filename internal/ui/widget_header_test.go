package ui

import (
	"image/color"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/markwylde/interrogate/internal/questionnaire"
)

func TestHeaderShowsEveryPart(t *testing.T) {
	q := &questionnaire.Question{
		ID:       "storage",
		Prompt:   "Where do answers live?",
		Type:     questionnaire.TypeSelect,
		Required: true,
		Answer:   "sidecar",
		Comment:  "worth discussing",
	}
	h := newQuestionHeader(q, func() {})
	h.SetCollapsed(true)

	if _, ok := search[*widget.Icon](h, first[*widget.Icon]()); !ok {
		t.Error("no chevron in the header")
	}
	if h.prompt.Text != q.Prompt {
		t.Errorf("prompt = %q, want %q", h.prompt.Text, q.Prompt)
	}
	if !h.prompt.TextStyle.Bold {
		t.Error("the prompt should be bold")
	}
	if h.prompt.Wrapping != fyne.TextWrapWord {
		t.Error("the prompt should wrap rather than widen the page")
	}
	if h.marker == nil || h.marker.Text != "Required" {
		t.Error("a required question should be marked as one")
	}
	if got := h.summary.Text; got != "sidecar" {
		t.Errorf("summary = %q, want the answer", got)
	}
	if !h.summary.Visible() {
		t.Error("a collapsed question should show its answer")
	}
	if !h.tick.Visible() {
		t.Error("an answered question should be ticked")
	}
	if !h.note.Visible() {
		t.Error("a question carrying a comment should say so")
	}

	// Every part has to be reachable through the tree, or nothing outside the
	// widget can find it.
	if _, ok := search[*tickMark](h, first[*tickMark]()); !ok {
		t.Error("the tick is not reachable in the header's object tree")
	}
	if _, ok := search[*widget.Label](h, func(l *widget.Label) bool { return l.Text == q.Prompt }); !ok {
		t.Error("the prompt is not reachable in the header's object tree")
	}
}

func TestOptionalQuestionHasNoRequiredMarker(t *testing.T) {
	q := &questionnaire.Question{ID: "name", Prompt: "Call it what?", Type: questionnaire.TypeText}
	h := newQuestionHeader(q, func() {})
	if h.marker != nil {
		t.Error("an optional question should carry no Required marker")
	}
}

func TestHeaderTicksOnlyWhenAnswered(t *testing.T) {
	q := &questionnaire.Question{ID: "name", Prompt: "Call it what?", Type: questionnaire.TypeText}
	h := newQuestionHeader(q, func() {})
	if h.tick.Visible() {
		t.Error("an unanswered question should carry no tick")
	}

	q.Answer = "interrogate"
	h.Sync()
	if !h.tick.Visible() {
		t.Error("an answered question should carry a tick")
	}

	// A cleared answer is no answer, so the tick goes with it.
	q.Answer = ""
	h.Sync()
	if h.tick.Visible() {
		t.Error("a cleared answer should take the tick with it")
	}
}

func TestHeaderReportsAComment(t *testing.T) {
	q := &questionnaire.Question{ID: "name", Prompt: "Call it what?", Type: questionnaire.TypeText}
	h := newQuestionHeader(q, func() {})
	if h.note.Visible() {
		t.Error("a question with no comment should carry no note marker")
	}

	q.Comment = "a thought"
	h.Sync()
	if !h.note.Visible() {
		t.Error("a question carrying a comment should show its note marker")
	}

	q.Comment = "   "
	h.Sync()
	if h.note.Visible() {
		t.Error("whitespace is not a comment")
	}
}

func TestHeaderShowsTheAnswerOnlyWhileCollapsed(t *testing.T) {
	q := &questionnaire.Question{
		ID: "ship", Prompt: "Ship before review?", Type: questionnaire.TypeBoolean, Answer: false,
	}
	h := newQuestionHeader(q, func() {})
	if h.summary.Visible() {
		t.Error("an expanded question already shows its answer in the control")
	}

	h.SetCollapsed(true)
	if !h.summary.Visible() || h.summary.Text != "No" {
		t.Errorf("collapsed summary = %q, visible = %v", h.summary.Text, h.summary.Visible())
	}

	h.SetCollapsed(false)
	if h.summary.Visible() {
		t.Error("expanding again should hide the summary")
	}
}

func TestHeaderChevronPointsTheRightWay(t *testing.T) {
	q := &questionnaire.Question{ID: "q", Prompt: "?", Type: questionnaire.TypeText}
	h := newQuestionHeader(q, func() {})
	open := h.chevron.Resource

	h.SetCollapsed(true)
	if h.chevron.Resource == open {
		t.Error("the chevron should change when the question collapses")
	}

	h.SetCollapsed(false)
	if h.chevron.Resource != open {
		t.Error("the chevron should return to its open form")
	}
}

// The answer is capped before it reaches the header, and truncates against the
// width it is given, so no answer can push the header wider than its window.
func TestHeaderSummaryNeverWidensThePage(t *testing.T) {
	const windowWidth = 780 // what the form opens at

	long := &questionnaire.Question{
		ID:     "q",
		Prompt: "Call it what?",
		Type:   questionnaire.TypeTextarea,
		Answer: strings.Repeat("interrogate ", 40),
	}
	h := newQuestionHeader(long, func() {})
	h.SetCollapsed(true)

	if got := len([]rune(h.summary.Text)); got > summaryLimit+1 {
		t.Errorf("the header carries %d runes of answer, want no more than %d", got, summaryLimit+1)
	}
	if h.summary.Truncation != fyne.TextTruncateEllipsis {
		t.Error("the summary should truncate against the width it is given")
	}
	if got := h.MinSize().Width; got > windowWidth {
		t.Errorf("a long answer asks for %v of width, more than the %v window", got, windowWidth)
	}
}

// A prompt long enough to wrap -- most of them, on a real questionnaire -- used
// to strand the answer out to the right, level with a line of the question it
// had nothing to do with. It belongs under the prompt.
func TestTheAnswerSitsUnderThePrompt(t *testing.T) {
	q := &questionnaire.Question{
		ID:       "shop",
		Type:     questionnaire.TypeSelect,
		Required: true,
		Prompt: "A shop displays a jacket in its window marked \"£40\". A customer takes it " +
			"to the till and offers £40. The shop refuses to sell. In contract law, what is the display?",
		Answer: "An invitation to treat",
	}
	h := newQuestionHeader(q, func() {})
	h.SetCollapsed(true)

	w := test.NewWindow(h)
	defer w.Close()
	w.Resize(fyne.NewSize(700, 200))

	if got, want := h.summary.Position().X, h.prompt.Position().X; got != want {
		t.Errorf("the answer starts at x=%v, want the prompt's own column at x=%v", got, want)
	}
	if h.summary.Position().Y <= h.prompt.Position().Y {
		t.Errorf("the answer is at y=%v, want it below the prompt at y=%v",
			h.summary.Position().Y, h.prompt.Position().Y)
	}
	if h.summary.Position().Y < h.prompt.Position().Y+h.prompt.Size().Height {
		t.Error("the answer overlaps the prompt rather than sitting under it")
	}
}

// The chevron and the tick belong against the prompt's first line. Left to the
// border they stretch to the whole height of the card and float in the middle
// of a wrapped prompt.
func TestTheMarkersSitAgainstTheFirstLine(t *testing.T) {
	q := &questionnaire.Question{
		ID:       "shop",
		Type:     questionnaire.TypeSelect,
		Required: true,
		Prompt:   strings.Repeat("a long prompt that has to wrap several times ", 6),
		Answer:   "An invitation to treat",
	}
	h := newQuestionHeader(q, func() {})
	h.SetCollapsed(true)

	w := test.NewWindow(h)
	defer w.Close()
	w.Resize(fyne.NewSize(500, 300))

	promptHeight := h.prompt.Size().Height
	if promptHeight <= 0 {
		t.Fatal("the prompt was not laid out")
	}
	for name, obj := range map[string]fyne.CanvasObject{
		"chevron": h.chevron,
		"tick":    h.tick,
		"marker":  h.marker,
	} {
		// Its own line-ish: within the first line of the prompt, not halfway
		// down a prompt that wrapped to six.
		if got := obj.Position().Y; got > promptHeight/2 {
			t.Errorf("the %s is at y=%v, adrift in a prompt %v tall", name, got, promptHeight)
		}
	}
}

func TestTappingTheHeaderToggles(t *testing.T) {
	taps := 0
	q := &questionnaire.Question{ID: "q", Prompt: "Where do answers live?", Type: questionnaire.TypeText}
	h := newQuestionHeader(q, func() { taps++ })

	test.Tap(h)
	if taps != 1 {
		t.Fatalf("taps = %d, want 1", taps)
	}
}

// The prompt is a plain label, so a tap on it reaches the header only because
// the header is the nearest Tappable above it. Tapping the canvas exercises
// exactly that dispatch.
func TestTappingThePromptReachesTheHeader(t *testing.T) {
	taps := 0
	q := &questionnaire.Question{ID: "q", Prompt: "Where do answers live?", Type: questionnaire.TypeText}
	h := newQuestionHeader(q, func() { taps++ })

	w := test.NewWindow(h)
	defer w.Close()
	w.Resize(fyne.NewSize(500, 80))

	if asserted[fyne.Tappable](h.prompt) {
		t.Fatal("the prompt is tappable itself, so this proves nothing about the header")
	}
	test.TapCanvas(w.Canvas(), fyne.NewPos(250, 30))
	if taps != 1 {
		t.Errorf("taps = %d after tapping over the prompt, want 1", taps)
	}
}

func TestHeaderTintsOnHover(t *testing.T) {
	q := &questionnaire.Question{ID: "q", Prompt: "?", Type: questionnaire.TypeText}
	h := newQuestionHeader(q, func() {})
	h.CreateRenderer()

	if h.background.FillColor != color.Transparent {
		t.Error("an un-hovered header should add nothing to the card behind it")
	}

	h.MouseIn(&desktop.MouseEvent{})
	if h.background.FillColor == color.Transparent {
		t.Error("hovering should tint the header, or nothing says it can be pressed")
	}

	h.MouseOut()
	if h.background.FillColor != color.Transparent {
		t.Error("the tint should go when the pointer leaves")
	}
}

// --- The tick -----------------------------------------------------------

func TestTickIsDrawnInTheSuccessColour(t *testing.T) {
	th := &tunableTheme{Theme: newTheme(), success: color.NRGBA{R: 1, G: 2, B: 3, A: 255}}
	restore := useTheme(t, th)
	defer restore()

	mark := newTickMark()
	mark.CreateRenderer()
	if got := mark.short.StrokeColor; got != th.success {
		t.Errorf("stroke = %v, want the theme's success colour %v", got, th.success)
	}
}

// The colour is resolved on refresh rather than captured at construction, so a
// form open while the desktop switches between light and dark redraws the tick
// in the right green instead of keeping the one it started with.
func TestTickRereadsItsColourOnRefresh(t *testing.T) {
	th := &tunableTheme{Theme: newTheme(), success: lightPalette.success}
	restore := useTheme(t, th)
	defer restore()

	mark := newTickMark()
	mark.CreateRenderer()
	was := mark.long.StrokeColor

	th.success = darkPalette.success // as a variant switch would
	mark.Refresh()

	if mark.long.StrokeColor == was {
		t.Error("Refresh did not re-resolve the tick colour")
	}
	if got := mark.long.StrokeColor; got != darkPalette.success {
		t.Errorf("stroke = %v, want %v", got, darkPalette.success)
	}
}

// And the two presentations really do carry different greens, so the refresh
// above has something to pick up.
func TestSuccessColourFollowsTheVariant(t *testing.T) {
	th := newTheme()
	light := th.Color(theme.ColorNameSuccess, theme.VariantLight)
	dark := th.Color(theme.ColorNameSuccess, theme.VariantDark)
	if light != lightPalette.success {
		t.Errorf("light success = %v, want the palette's %v", light, lightPalette.success)
	}
	if dark != darkPalette.success {
		t.Errorf("dark success = %v, want the palette's %v", dark, darkPalette.success)
	}
	if light == dark {
		t.Error("light and dark should not share a success colour")
	}
}

func TestTickDrawsTwoStrokes(t *testing.T) {
	mark := newTickMark()
	r := mark.CreateRenderer()
	if got := len(r.Objects()); got != 2 {
		t.Fatalf("the tick is drawn from %d objects, want 2", got)
	}
	r.Layout(fyne.NewSize(tickSize, tickSize))
	// The short stroke comes down to the corner the long one rises from.
	if mark.short.Position2 != mark.long.Position1 {
		t.Errorf("the strokes do not meet: %v and %v", mark.short.Position2, mark.long.Position1)
	}
	if mark.long.Position2.Y >= mark.long.Position1.Y {
		t.Error("the long stroke should rise to the right")
	}
}

// tunableTheme is the form's theme with one colour under the test's control, so
// a variant switch can be simulated without a desktop.
type tunableTheme struct {
	fyne.Theme
	success color.Color
}

func (t *tunableTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameSuccess {
		return t.success
	}
	return t.Theme.Color(name, variant)
}

// useTheme installs a theme for the duration of one test.
func useTheme(t *testing.T, th fyne.Theme) func() {
	t.Helper()
	settings := fyne.CurrentApp().Settings()
	was := settings.Theme()
	settings.SetTheme(th)
	return func() { settings.SetTheme(was) }
}
