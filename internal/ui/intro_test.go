package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// The gap a tightStack is asked for is the space between the lines of text, not
// between the labels: a label pads itself top and bottom, and two of them in a
// box layout sit a whole label apart.
func TestATightStacksGapIsBetweenTheText(t *testing.T) {
	inner := theme.Size(theme.SizeNameInnerPadding)
	for _, gap := range []float32{0, 4, 8} {
		stack := &tightStack{gap: gap}
		if got, want := stack.overlap(), 2*inner-gap; got != want {
			t.Errorf("gap %v overlaps by %v, want %v", gap, got, want)
		}
	}
	// A gap wider than the padding there is to take back cannot overlap
	// negatively; the labels just sit apart.
	if got := (&tightStack{gap: 1000}).overlap(); got != 0 {
		t.Errorf("overlap = %v for a gap wider than the padding, want 0", got)
	}
}

// Fyne's line spacing is applied between rich-text segments and explicitly not
// between the rows inside one, so the intro is built a line at a time: the
// leading inside a single label cannot be set at all.
func TestTheIntroIsBuiltALineAtATime(t *testing.T) {
	f := &form{}
	obj := f.intro("First line of the intro.\nSecond line of the intro.")

	box, ok := obj.(*fyne.Container)
	if !ok {
		t.Fatalf("the intro is a %T, want a container of lines", obj)
	}
	if got := len(box.Objects); got != 2 {
		t.Fatalf("the intro is %d objects, want one per authored line", got)
	}

	first, ok := box.Objects[0].(*widget.Label)
	if !ok {
		t.Fatalf("the first line is a %T", box.Objects[0])
	}
	second := box.Objects[1].(*widget.Label)
	if first.Text != "First line of the intro." || second.Text != "Second line of the intro." {
		t.Errorf("lines are %q and %q", first.Text, second.Text)
	}
	for _, line := range []*widget.Label{first, second} {
		if line.Wrapping != fyne.TextWrapWord {
			t.Error("an intro line should wrap rather than run off the window")
		}
	}

	box.Resize(box.MinSize())
	inner := theme.Size(theme.SizeNameInnerPadding)
	// From the bottom of the first line's text to the top of the second's.
	textGap := (second.Position().Y + inner) -
		(first.Position().Y + first.MinSize().Height - inner)
	if textGap != introLeading {
		t.Errorf("the intro's lines are %v apart, want the %v asked for", textGap, introLeading)
	}
	// Which has to be more than a label sets for itself, or there was no point.
	if introLeading <= 0 {
		t.Error("the intro leading should add space, not remove it")
	}
}

func TestTheIntroKeepsBlankLinesAsParagraphBreaks(t *testing.T) {
	f := &form{}
	box := f.intro("First paragraph.\n\nSecond paragraph.").(*fyne.Container)
	if got := len(box.Objects); got != 3 {
		t.Errorf("%d objects, want the blank line kept as a break", got)
	}
}
