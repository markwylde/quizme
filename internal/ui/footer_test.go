package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// footerAt lays the form out at a text size and window width, and returns the
// status label and the actions in window coordinates.
func footerAt(t *testing.T, percent int, width float32) (f *form, status, actions fyne.Position, statusSize, actionsSize fyne.Size) {
	t.Helper()
	restore := useTheme(t, newScaledTheme(percent))
	t.Cleanup(restore)

	f, _ = build(t, uiDoc)
	win := test.NewWindow(f.build())
	t.Cleanup(win.Close)
	win.Resize(fyne.NewSize(width, 700))
	// A change of rows refreshes the footer, which lays it out again.
	win.Resize(fyne.NewSize(width, 701))

	row := f.footerRow
	d := fyne.CurrentApp().Driver()
	acts := row.Objects[1]
	return f, d.AbsolutePositionForObject(f.status), d.AbsolutePositionForObject(acts), f.status.Size(), acts.Size()
}

func TestTheFooterNeverOverlaps(t *testing.T) {
	for _, percent := range []int{70, 100, 130, 150, 200} {
		_, s, a, ss, as := footerAt(t, percent, 780)
		sameRow := s.Y < a.Y+as.Height && a.Y < s.Y+ss.Height
		if sameRow && s.X+ss.Width > a.X+1 {
			t.Errorf("at %d%% the status (to %v) runs under the actions (from %v)", percent, s.X+ss.Width, a.X)
		}
	}
}

func TestTheFooterIsOneRowWhileItFits(t *testing.T) {
	f, s, a, ss, as := footerAt(t, 100, 780)
	if f.status.MinSize().Height > ss.Height+1 {
		t.Error("the status is squeezed")
	}
	// Centred on each other, so the text sits level with the buttons' labels.
	if mid, amid := s.Y+ss.Height/2, a.Y+as.Height/2; mid-amid > 1 || amid-mid > 1 {
		t.Errorf("the status is centred at %v and the actions at %v", mid, amid)
	}
	if got := len(f.status.Text); got == 0 {
		t.Error("no status text")
	}
}

func TestTheActionsMoveBelowWhenTheStatusWouldNotFit(t *testing.T) {
	f, s, a, ss, as := footerAt(t, 130, 780)
	if a.Y < s.Y+ss.Height {
		t.Fatalf("at 130%% the actions (y %v) should sit below the status (to %v)", a.Y, s.Y+ss.Height)
	}
	// The status is one line: it had the whole width.
	one := fyne.MeasureText("Ag", 14*1.3, fyne.TextStyle{}).Height
	if ss.Height > 2*one+2*scaled(8) {
		t.Errorf("the status is %v tall: it wrapped although it had the row to itself", ss.Height)
	}
	// The actions keep to the trailing edge.
	rowEnd := fyne.CurrentApp().Driver().AbsolutePositionForObject(f.footerRow).X + f.footerRow.Size().Width
	if end := a.X + as.Width; end < rowEnd-1 || end > rowEnd+1 {
		t.Errorf("the actions end at %v, the row at %v", end, rowEnd)
	}
	if _, ok := search[*widget.Button](f.footerRow, func(b *widget.Button) bool { return b.Text == "Submit" }); !ok {
		t.Error("Submit went missing")
	}
}
