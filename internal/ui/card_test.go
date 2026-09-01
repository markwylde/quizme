package ui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
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
