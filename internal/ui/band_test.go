package ui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func TestEveryPaletteDefinesTheSameTints(t *testing.T) {
	if got := len(lightPalette.questionTints); got != tintCount {
		t.Errorf("light palette has %d tints, want %d", got, tintCount)
	}
	if got := len(darkPalette.questionTints); got != tintCount {
		t.Errorf("dark palette has %d tints, want %d", got, tintCount)
	}
}

// nearBackground reports how far a tint strays from its palette's background.
// A tint is meant to be felt, not seen; anything further than this is competing
// with the content rather than separating it.
func nearBackground(tint, background color.Color) int {
	tr, tg, tb, _ := tint.RGBA()
	br, bg, bb, _ := background.RGBA()
	worst := 0
	for _, d := range []int{
		int(tr>>8) - int(br>>8),
		int(tg>>8) - int(bg>>8),
		int(tb>>8) - int(bb>>8),
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

func TestTintsStayCloseToTheirBackground(t *testing.T) {
	const limit = 12
	for name, p := range map[string]palette{"light": lightPalette, "dark": darkPalette} {
		for i, tint := range p.questionTints {
			if got := nearBackground(tint, p.background); got > limit {
				t.Errorf("%s tint %d is %d points from the background, limit %d", name, i, got, limit)
			}
		}
	}
}

func TestTintsAreDistinguishableFromEachOther(t *testing.T) {
	// Subtle is the point, but two consecutive tints still have to differ.
	for name, p := range map[string]palette{"light": lightPalette, "dark": darkPalette} {
		for i := range p.questionTints {
			next := p.questionTints[(i+1)%len(p.questionTints)]
			if got := nearBackground(p.questionTints[i], next); got < 2 {
				t.Errorf("%s tints %d and %d differ by only %d points", name, i, i+1, got)
			}
		}
	}
}

func TestQuestionTintCycles(t *testing.T) {
	th := newTheme().(tintProvider)
	first := th.QuestionTint(0, theme.VariantLight)
	if got := th.QuestionTint(tintCount, theme.VariantLight); got != first {
		t.Errorf("tint %d = %v, want it to cycle back to %v", tintCount, got, first)
	}
	if th.QuestionTint(-1, theme.VariantLight) != first {
		t.Error("a negative index should clamp rather than panic")
	}
}

func TestQuestionTintFollowsTheVariant(t *testing.T) {
	th := newTheme().(tintProvider)
	light := th.QuestionTint(0, theme.VariantLight)
	dark := th.QuestionTint(0, theme.VariantDark)
	if light == dark {
		t.Error("light and dark should not share a tint")
	}
	// Dark tints must actually be dark, not lightened light ones.
	r, g, b, _ := dark.RGBA()
	if int(r>>8)+int(g>>8)+int(b>>8) > 200 {
		t.Errorf("the dark tint %v is too light for a dark background", dark)
	}
}

func TestAdjacentQuestionsDifferInTint(t *testing.T) {
	f, doc := build(t, uiDoc)
	var tints []int
	for _, q := range doc.Questions {
		c := f.cards[q.ID]
		if c.band == nil || !c.root.Visible() {
			continue
		}
		tints = append(tints, c.band.Tint())
	}
	if len(tints) < 2 {
		t.Fatal("not enough visible questions to compare")
	}
	for i := 1; i < len(tints); i++ {
		if tints[i] == tints[i-1] {
			t.Errorf("questions %d and %d share tint %d", i-1, i, tints[i])
		}
	}
}

func TestHiddenQuestionDoesNotConsumeATint(t *testing.T) {
	// storage_why sits between storage and formats. While it is hidden, its two
	// neighbours must still differ, which they would not if it took a tint.
	f, _ := build(t, uiDoc)

	before := f.cards["formats"].band.Tint()
	if got := f.cards["storage"].band.Tint(); got == before {
		t.Errorf("storage and formats share tint %d across a hidden question", got)
	}

	// Revealing it pushes everything below along by one.
	find[*widget.RadioGroup](t, f, "storage").SetSelected("sidecar")
	if !f.cards["why"].root.Visible() {
		t.Fatal("the gated question did not appear")
	}
	if got := f.cards["formats"].band.Tint(); got != before+1 {
		t.Errorf("formats tint = %d, want %d once a question appears above it", got, before+1)
	}
	if f.cards["why"].band.Tint() == f.cards["storage"].band.Tint() {
		t.Error("the revealed question shares a tint with the one above it")
	}
}

func TestBandWithoutATintedThemeIsTransparent(t *testing.T) {
	// The test theme knows nothing of question tints. A band under it should
	// paint nothing rather than log an error on every draw.
	b := newBand(0, widget.NewLabel("x"))
	if got := b.colour(); got != color.Transparent {
		t.Errorf("colour = %v, want transparent under a theme without tints", got)
	}
}

func TestBandRereadsItsTintOnRefresh(t *testing.T) {
	// The colour is resolved on refresh, not captured at construction, so a
	// form open across a light/dark switch picks up the right tints.
	b := newBand(0, widget.NewLabel("x"))
	b.CreateRenderer()
	if b.rect == nil {
		t.Fatal("the band did not build its rectangle")
	}
	b.rect.FillColor = color.NRGBA{R: 1, G: 2, B: 3, A: 4}
	b.Refresh()
	if b.rect.FillColor == (color.NRGBA{R: 1, G: 2, B: 3, A: 4}) {
		t.Error("Refresh did not re-resolve the tint")
	}
}

func TestBandExposesItsContent(t *testing.T) {
	label := widget.NewLabel("inside")
	b := newBand(0, label)
	if b.Content() != fyne.CanvasObject(label) {
		t.Error("the band does not expose what it wraps")
	}
}
