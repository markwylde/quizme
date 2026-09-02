package ui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// contrast reports the largest per-channel difference between two colours.
func contrast(a, b color.Color) int {
	ar, ag, ab, _ := a.RGBA()
	br, bg, bb, _ := b.RGBA()
	worst := 0
	for _, d := range []int{
		int(ar>>8) - int(br>>8),
		int(ag>>8) - int(bg>>8),
		int(ab>>8) - int(bb>>8),
	} {
		if d < 0 {
			d = -d
		}
		if d > worst {
			worst = d
		}
	}
	return worst
}

func TestCardContrastsWithThePage(t *testing.T) {
	// Enough to read as a distinct panel, quiet enough not to compete with the
	// question's own content.
	const (
		lower = 4
		upper = 24
	)
	for name, p := range map[string]palette{"light": lightPalette, "dark": darkPalette} {
		got := contrast(p.card, p.background)
		if got < lower {
			t.Errorf("%s card is only %d points from the page; it will not read as a card", name, got)
		}
		if got > upper {
			t.Errorf("%s card is %d points from the page, louder than the limit of %d", name, got, upper)
		}
	}
}

func TestEveryQuestionSharesOneCardColour(t *testing.T) {
	// The cycle of per-question tints this replaced read as a swatch book.
	th := newTheme().(cardProvider)
	light := th.QuestionCard(theme.VariantLight)
	for i := 0; i < 10; i++ {
		if got := th.QuestionCard(theme.VariantLight); got != light {
			t.Fatalf("card colour changed between questions: %v then %v", light, got)
		}
	}
}

func TestCardColourFollowsTheVariant(t *testing.T) {
	th := newTheme().(cardProvider)
	light := th.QuestionCard(theme.VariantLight)
	dark := th.QuestionCard(theme.VariantDark)
	if light == dark {
		t.Error("light and dark should not share a card colour")
	}
	r, g, b, _ := dark.RGBA()
	if int(r>>8)+int(g>>8)+int(b>>8) > 200 {
		t.Errorf("the dark card %v is too light for a dark background", dark)
	}
}

func TestPageIsLighterThanTheCardInLight(t *testing.T) {
	// The page was lightened so the cards have something to sit against.
	pr, pg, pb, _ := lightPalette.background.RGBA()
	cr, cg, cb, _ := lightPalette.card.RGBA()
	if int(pr)+int(pg)+int(pb) <= int(cr)+int(cg)+int(cb) {
		t.Error("the light page should be lighter than the cards on it")
	}
}

func TestEveryQuestionGetsACardBox(t *testing.T) {
	f, doc := build(t, uiDoc)
	for _, q := range doc.Questions {
		if f.cards[q.ID].card == nil {
			t.Errorf("question %q has no card", q.ID)
		}
	}
}

func TestCardWithoutAThemedProviderIsTransparent(t *testing.T) {
	// The test theme knows nothing of card colours. A card under it should
	// paint nothing rather than log an error on every draw.
	c := newCardBox(widget.NewLabel("x"))
	if got := c.colour(); got != color.Transparent {
		t.Errorf("colour = %v, want transparent under a theme without cards", got)
	}
}

func TestCardRereadsItsColourOnRefresh(t *testing.T) {
	// Resolved on refresh, not captured at construction, so a form open across
	// a light/dark switch picks up the right colour.
	c := newCardBox(widget.NewLabel("x"))
	c.CreateRenderer()
	if c.rect == nil {
		t.Fatal("the card did not build its rectangle")
	}
	c.rect.FillColor = color.NRGBA{R: 1, G: 2, B: 3, A: 4}
	c.Refresh()
	if c.rect.FillColor == (color.NRGBA{R: 1, G: 2, B: 3, A: 4}) {
		t.Error("Refresh did not re-resolve the card colour")
	}
}

func TestCardHasRoundedCorners(t *testing.T) {
	c := newCardBox(widget.NewLabel("x"))
	c.CreateRenderer()
	if c.rect.CornerRadius <= 0 {
		t.Errorf("corner radius = %v, want a rounded card", c.rect.CornerRadius)
	}
}

func TestCardExposesItsContent(t *testing.T) {
	label := widget.NewLabel("inside")
	c := newCardBox(label)
	if c.Content() != fyne.CanvasObject(label) {
		t.Error("the card does not expose what it wraps")
	}
}

// --- The settled panel ---------------------------------------------------

// A folded, answered question is tinted green. On a page of folded rows the
// tint is what carries at a glance; the tick is what confirms it up close.
func TestSettledCardContrastsWithThePage(t *testing.T) {
	const (
		lower = 4
		upper = 24
	)
	for name, p := range map[string]palette{"light": lightPalette, "dark": darkPalette} {
		got := contrast(p.settled, p.background)
		if got < lower {
			t.Errorf("%s settled card is only %d points from the page; it will not read as a card", name, got)
		}
		if got > upper {
			t.Errorf("%s settled card is %d points from the page, louder than the limit of %d", name, got, upper)
		}
	}
}

func TestSettledCardIsGreenButQuiet(t *testing.T) {
	for name, p := range map[string]palette{"light": lightPalette, "dark": darkPalette} {
		r, g, b, _ := p.settled.RGBA()
		red, green, blue := int(r>>8), int(g>>8), int(b>>8)
		if green <= red || green <= blue {
			t.Errorf("%s settled card %v has no green cast", name, p.settled)
		}
		// Subtle: a saturated green would turn a finished questionnaire into a
		// wall of colour.
		if green-red > 24 || green-blue > 24 {
			t.Errorf("%s settled card %v is too saturated for a background", name, p.settled)
		}
		// But different enough from the plain card to be seen beside one.
		if got := contrast(p.settled, p.card); got < 3 {
			t.Errorf("%s settled card is only %d points from the plain card", name, got)
		}
	}
}

func TestSettledCardFollowsTheVariant(t *testing.T) {
	th := newTheme().(cardProvider)
	light := th.SettledCard(theme.VariantLight)
	dark := th.SettledCard(theme.VariantDark)
	if light != lightPalette.settled {
		t.Errorf("light settled = %v, want the palette's %v", light, lightPalette.settled)
	}
	if dark != darkPalette.settled {
		t.Errorf("dark settled = %v, want the palette's %v", dark, darkPalette.settled)
	}
	r, g, b, _ := dark.RGBA()
	if int(r>>8)+int(g>>8)+int(b>>8) > 200 {
		t.Errorf("the dark settled card %v is too light for a dark background", dark)
	}
}

func TestCardBoxPaintsTheSettledPanelWhenSettled(t *testing.T) {
	// The test theme knows nothing of card colours, so this asserts through the
	// form's own theme.
	restore := useTheme(t, newTheme())
	defer restore()

	c := newCardBox(widget.NewLabel("x"))
	c.CreateRenderer()
	plain := c.rect.FillColor

	c.SetSettled(true)
	if !c.Settled() {
		t.Fatal("the card did not take the settled state")
	}
	if c.rect.FillColor == plain {
		t.Error("a settled card should not paint the plain panel")
	}
	if got, want := c.rect.FillColor, lightPalette.settled; got != want {
		t.Errorf("settled fill = %v, want %v", got, want)
	}

	c.SetSettled(false)
	if got := c.rect.FillColor; got != plain {
		t.Errorf("fill = %v after unsettling, want the plain %v", got, plain)
	}
}

func TestOnlyAFoldedAnsweredQuestionReadsAsSettled(t *testing.T) {
	f, doc := build(t, uiDoc)

	// Answered and folded: settled.
	find[*widget.RadioGroup](t, f, "storage").SetSelected("sidecar")
	storage := f.cards["storage"]
	if !storage.collapsed() {
		t.Fatal("answering should have folded it")
	}
	if !storage.card.Settled() {
		t.Error("a folded, answered question should read as settled")
	}

	// Opened again to revise: no longer a finished row, so no tint.
	test.Tap(storage.header)
	if storage.card.Settled() {
		t.Error("an expanded question should paint the plain panel, tick or no tick")
	}

	// Folded by hand with nothing answered: put aside, not finished.
	aside := f.cards["ship"]
	test.Tap(aside.header)
	if !aside.collapsed() {
		t.Fatal("tapping should have folded it")
	}
	if aside.card.Settled() {
		t.Error("a folded question with no answer must not look finished")
	}
	if doc.Question("ship").HasAnswer() {
		t.Error("folding a question is not answering it")
	}
}

func TestQuestionsFoldedOnLoadReadAsSettled(t *testing.T) {
	f, _ := build(t, `
title: Reopened
questions:
  - id: done
    type: select
    prompt: Settled?
    options: [yes, no]
    answer: yes
  - id: outstanding
    type: select
    prompt: Not settled?
    options: [yes, no]
`)
	if !f.cards["done"].card.Settled() {
		t.Error("a question folded on load carries an answer, so it should read as settled")
	}
	if f.cards["outstanding"].card.Settled() {
		t.Error("an unanswered question should paint the plain panel")
	}
}
