package ui

import (
	"image/color"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

const shopPrompt = "A shop displays a jacket in its window marked \"£40\". A customer takes it " +
	"to the till and offers £40. The shop refuses to sell. In contract law, what is the display?"

func TestAPromptWrapsOnWords(t *testing.T) {
	p := newPromptText(shopPrompt)
	p.CreateRenderer()
	p.Resize(fyne.NewSize(400, p.MinSize().Height))

	if len(p.Lines()) < 2 {
		t.Fatalf("the prompt came to %d lines at 400 wide", len(p.Lines()))
	}

	var rejoined []string
	for _, line := range p.Lines() {
		if got := p.measure(line.Text).Width; got > 400-2*p.inset() {
			t.Errorf("the line %q is %v wide, past the %v it had", line.Text, got, 400-2*p.inset())
		}
		if strings.TrimSpace(line.Text) != line.Text {
			t.Errorf("the line %q carries loose whitespace", line.Text)
		}
		rejoined = append(rejoined, line.Text)
	}
	// Nothing lost and nothing invented: the lines are the prompt again.
	if got, want := strings.Join(rejoined, " "), strings.Join(strings.Fields(shopPrompt), " "); got != want {
		t.Errorf("the wrapped lines read %q, want %q", got, want)
	}
}

// The whole reason this is not a widget.Label: the space between the lines.
func TestAPromptsLinesAreLedApart(t *testing.T) {
	p := newPromptText(shopPrompt)
	p.CreateRenderer()
	p.Resize(fyne.NewSize(400, p.MinSize().Height))

	lines := p.Lines()
	if len(lines) < 2 {
		t.Fatal("need at least two lines to measure the leading")
	}
	step := lines[1].Position().Y - lines[0].Position().Y
	if want := p.lineHeight() + promptLeading; step != want {
		t.Errorf("the lines are %v apart, want a line and the %v leading (%v)", step, promptLeading, want)
	}
	if promptLeading <= 0 {
		t.Error("the leading should add space to what the face sets, not remove it")
	}

	// And the box it asks for accounts for every line and every gap.
	n := float32(len(lines))
	want := 2*p.inset() + n*p.lineHeight() + (n-1)*promptLeading
	if got := p.MinSize().Height; got != want {
		t.Errorf("the prompt asks for %v of height, want %v", got, want)
	}
}

// The text is inset the way a label's was, so everything that lines itself up
// against a prompt -- the chevron, the tick, the answer beneath -- keeps the
// measurements it already had.
func TestAPromptSitsInTheSameBoxALabelDid(t *testing.T) {
	p := newPromptText("Call it what?")
	p.CreateRenderer()
	p.Resize(fyne.NewSize(400, p.MinSize().Height))

	if got := len(p.Lines()); got != 1 {
		t.Fatalf("%d lines for a short prompt", got)
	}
	line := p.Lines()[0]
	if got, want := line.Position().X, theme.Size(theme.SizeNameInnerPadding); got != want {
		t.Errorf("the text starts at x=%v, want the inner padding %v", got, want)
	}
	if got, want := line.Position().Y, theme.Size(theme.SizeNameInnerPadding); got != want {
		t.Errorf("the text starts at y=%v, want the inner padding %v", got, want)
	}
	if got, want := p.MinSize().Height, p.lineHeight()+2*theme.Size(theme.SizeNameInnerPadding); got != want {
		t.Errorf("a one-line prompt is %v tall, want %v", got, want)
	}
	// Which is the line the marks are placed against.
	if got, want := promptLineCentre(), theme.Size(theme.SizeNameInnerPadding)+p.lineHeight()/2; got != want {
		t.Errorf("the marks aim at y=%v, the prompt's line is centred at y=%v", got, want)
	}
}

func TestANarrowerPromptTakesMoreLines(t *testing.T) {
	p := newPromptText(shopPrompt)
	p.CreateRenderer()

	p.Resize(fyne.NewSize(600, p.MinSize().Height))
	wide, wideHeight := len(p.Lines()), p.MinSize().Height

	p.Resize(fyne.NewSize(260, p.MinSize().Height))
	narrow, narrowHeight := len(p.Lines()), p.MinSize().Height

	if narrow <= wide {
		t.Errorf("%d lines at 260 wide against %d at 600", narrow, wide)
	}
	if narrowHeight <= wideHeight {
		t.Errorf("the narrower prompt asks for %v, the wider one %v", narrowHeight, wideHeight)
	}
}

func TestAPromptKeepsTheAuthorsOwnLineBreaks(t *testing.T) {
	p := newPromptText("First line.\nSecond line.")
	p.CreateRenderer()
	p.Resize(fyne.NewSize(600, p.MinSize().Height))

	if got := len(p.Lines()); got != 2 {
		t.Fatalf("%d lines, want the two the author wrote", got)
	}
	if p.Lines()[0].Text != "First line." || p.Lines()[1].Text != "Second line." {
		t.Errorf("lines are %q and %q", p.Lines()[0].Text, p.Lines()[1].Text)
	}
}

// Resolved on refresh rather than captured at construction, so a form open
// while the desktop switches between light and dark redraws in the right one.
func TestAPromptRereadsItsColourOnRefresh(t *testing.T) {
	p := newPromptText("Call it what?")
	p.CreateRenderer()
	p.Resize(fyne.NewSize(400, p.MinSize().Height))

	line := p.Lines()[0]
	line.Color = color.NRGBA{R: 1, G: 2, B: 3, A: 4}
	p.Refresh()

	if line.Color == (color.NRGBA{R: 1, G: 2, B: 3, A: 4}) {
		t.Error("Refresh did not re-resolve the prompt's colour")
	}
	if got, want := line.Color, theme.Color(theme.ColorNameForeground); got != want {
		t.Errorf("colour = %v, want the theme's foreground %v", got, want)
	}
}

func TestWrapWordsPutsAnUnbreakableWordOnItsOwnLine(t *testing.T) {
	measure := func(s string) fyne.Size { return fyne.NewSize(float32(len(s)), 10) }
	lines := wrapWords("a bb enormouslylongword c", 10, measure)

	for _, line := range lines {
		if strings.Contains(line, "enormouslylongword") && line != "enormouslylongword" {
			t.Errorf("the long word shares the line %q", line)
		}
	}
	if strings.Contains(strings.Join(lines, " "), "enormouslylong ") {
		t.Error("the long word was cut in half")
	}
}

// The prompt's height depends on the width it is given, which nothing asks it
// for until after it has been measured once. So the case worth pinning is the
// one a responder actually causes: a window that changes width.
func TestNarrowingTheWindowRewrapsThePromptsOnThePage(t *testing.T) {
	f, _ := build(t, `
title: t
questions:
  - id: shop
    type: select
    prompt: A shop displays a jacket in its window marked "40". A customer takes it to the till and offers 40. The shop refuses to sell. In contract law, what is the display?
    options: [An offer, An invitation to treat]
`)
	win := test.NewWindow(f.build())
	defer win.Close()

	win.Resize(fyne.NewSize(900, 600))
	prompt, ok := search[*promptText](f.cards["shop"].header, first[*promptText]())
	if !ok {
		t.Fatal("no prompt in the card")
	}
	wide, wideCard := len(prompt.Lines()), f.cards["shop"].root.MinSize().Height

	win.Resize(fyne.NewSize(420, 600))
	narrow, narrowCard := len(prompt.Lines()), f.cards["shop"].root.MinSize().Height

	if narrow <= wide {
		t.Errorf("the prompt took %d lines at 420 wide and %d at 900", narrow, wide)
	}
	if narrowCard <= wideCard {
		t.Errorf("the card asks for %v at 420 wide and %v at 900: it did not follow the prompt",
			narrowCard, wideCard)
	}
}
