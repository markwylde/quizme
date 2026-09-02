package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// The test driver runs an animation straight to its end, so a move in these
// tests lands the rows where they belong without any waiting.

func newTestRank(t *testing.T) (*rankWidget, *[][]string) {
	t.Helper()
	var seen [][]string
	r := newRankWidget([]string{"correctness", "speed", "looks"}, func(order []string) {
		seen = append(seen, order)
	})
	r.CreateRenderer()
	r.Resize(fyne.NewSize(400, r.MinSize().Height))
	return r, &seen
}

func TestRankStacksItsRows(t *testing.T) {
	r, _ := newTestRank(t)

	step := r.rowHeight() + r.gap()
	if step <= 0 {
		t.Fatalf("row step = %v, want a positive height", step)
	}
	for i, row := range r.rows {
		if got, want := row.Position(), fyne.NewPos(0, float32(i)*step); got != want {
			t.Errorf("row %d sits at %v, want %v", i, got, want)
		}
		if got := row.Size().Width; got != 400 {
			t.Errorf("row %d is %v wide, want the full 400", i, got)
		}
	}

	// The container has no layout, so the widget has to add the rows up itself.
	want := 3*r.rowHeight() + 2*r.gap()
	if got := r.MinSize().Height; got != want {
		t.Errorf("MinSize height = %v, want %v for three rows and two gaps", got, want)
	}
}

// The rows outlive a reorder. A row rebuilt in its new place could not travel
// there, and the travel is what tells the responder what moved.
func TestReorderMovesTheSameRowsRatherThanRebuildingThem(t *testing.T) {
	r, _ := newTestRank(t)
	first, second := r.rows[0], r.rows[1]

	r.move(0, 1)

	if r.rows[0] != second || r.rows[1] != first {
		t.Fatal("the rows were rebuilt instead of swapped")
	}
	if got := r.Order(); got[0] != "speed" || got[1] != "correctness" {
		t.Errorf("order = %v", got)
	}

	step := r.rowHeight() + r.gap()
	if got, want := second.Position(), fyne.NewPos(0, 0); got != want {
		t.Errorf("the promoted row ended at %v, want %v", got, want)
	}
	if got, want := first.Position(), fyne.NewPos(0, step); got != want {
		t.Errorf("the demoted row ended at %v, want %v", got, want)
	}
}

func TestReorderIsAnimatedSnappily(t *testing.T) {
	r, _ := newTestRank(t)

	r.move(0, 1)
	if r.anim == nil {
		t.Fatal("a reorder should animate the rows into place")
	}
	if got := r.anim.Duration; got != rankSlide {
		t.Errorf("duration = %v, want the snappy %v", got, rankSlide)
	}
	if r.anim.Curve == nil {
		t.Error("the slide should be eased rather than linear")
	}

	// Rewinding the animation puts the rows back where the travel began, which
	// is the other half of saying they travelled at all.
	step := r.rowHeight() + r.gap()
	promoted := r.rows[0]
	r.anim.Tick(0)
	if got, want := promoted.Position(), fyne.NewPos(0, step); got != want {
		t.Errorf("at the start of the slide the promoted row is at %v, want %v", got, want)
	}
	r.anim.Tick(0.5)
	if got := promoted.Position().Y; got <= 0 || got >= step {
		t.Errorf("halfway through the slide the row is at %v, want it between 0 and %v", got, step)
	}
	r.anim.Tick(1)
	if got, want := promoted.Position(), fyne.NewPos(0, 0); got != want {
		t.Errorf("at the end of the slide the row is at %v, want %v", got, want)
	}
	if r.animating {
		t.Error("the widget still thinks it is animating after the slide finished")
	}
}

// A second press while the rows are still moving carries on from where they
// are, rather than snapping them to the end of the last move first.
func TestASecondMoveCarriesOnFromWhereTheRowsAre(t *testing.T) {
	r, _ := newTestRank(t)

	caught := r.rows[0]
	caught.Move(fyne.NewPos(0, 7)) // as if the pointer caught it mid-flight

	r.move(0, 1)
	if r.anim == nil {
		t.Fatal("no animation for the second move")
	}
	r.anim.Tick(0)
	if got := caught.Position().Y; got != 7 {
		t.Errorf("the slide started from %v, want the %v the row was actually at", got, float32(7))
	}
}

// Reporting the answer lays the page out again, and a layout with no slide in
// flight puts every row straight into its final slot. Started in the wrong
// order, that leaves the slide with nowhere to travel from and the reorder is
// instant -- which is the bug this whole feature exists to fix, so it is worth
// a test through the form itself rather than a hand-built widget.
func TestAReorderInsideTheFormStillSlides(t *testing.T) {
	f, _ := build(t, uiDoc)
	rank := find[*rankWidget](t, f, "order")
	rank.Resize(fyne.NewSize(400, rank.MinSize().Height))

	step := rank.rowHeight() + rank.gap()
	demoted := rank.rows[0]

	demoted.down.OnTapped() // exactly what the responder presses

	if rank.anim == nil {
		t.Fatal("the reorder was not animated: something placed the rows before the slide started")
	}
	rank.anim.Tick(0)
	if got := demoted.Position().Y; got != 0 {
		t.Errorf("the slide starts at %v, want the %v the row was at before the press", got, float32(0))
	}
	rank.anim.Tick(1)
	if got := demoted.Position().Y; got != step {
		t.Errorf("the slide ends at %v, want the next slot at %v", got, step)
	}
}

// The same thing, from the other side: a layout arriving while the slide is in
// flight must not finish it early.
func TestAFormRefreshMidSlideDoesNotFinishIt(t *testing.T) {
	f, _ := build(t, uiDoc)
	rank := find[*rankWidget](t, f, "order")
	rank.Resize(fyne.NewSize(400, rank.MinSize().Height))

	demoted := rank.rows[0]
	demoted.down.OnTapped()

	// The test driver runs an animation to its end the moment it starts, so the
	// in-flight state has to be put back by hand to have anything to check.
	rank.animating = true
	rank.anim.Tick(0.5)
	midway := demoted.Position().Y

	f.syncVisibility() // as any later answer on the page would cause
	f.list.Refresh()   //
	rank.Resize(rank.Size())

	if got := demoted.Position().Y; got != midway {
		t.Errorf("the row jumped to %v mid-slide, want it left at %v", got, midway)
	}
}

// The answer is recorded when the move happens. A questionnaire must never wait
// on an animation.
func TestTheAnswerIsRecordedBeforeTheSlideFinishes(t *testing.T) {
	var order []string
	r := newRankWidget([]string{"a", "b"}, func(o []string) { order = o })
	r.CreateRenderer()
	r.Resize(fyne.NewSize(200, r.MinSize().Height))

	// Stop the animation from running at all, then move.
	r.animating = true
	r.move(0, 1)

	if len(order) != 2 || order[0] != "b" {
		t.Errorf("order reported as %v, want the new order regardless of the slide", order)
	}
	if !r.Touched() {
		t.Error("the move should count as an interaction immediately")
	}
}

func TestReorderKeepsTheOrdinalsAndArrowsHonest(t *testing.T) {
	r, _ := newTestRank(t)

	r.move(0, 1)

	for i, row := range r.rows {
		if got, want := row.position.Text, ordinal(i+1); got != want {
			t.Errorf("row %d is numbered %q, want %q", i, got, want)
		}
		if got, want := row.index, i; got != want {
			t.Errorf("row %d thinks it is at %d", i, got)
		}
	}
	if !r.rows[0].up.Disabled() {
		t.Error("the top row should have nowhere up to go")
	}
	if r.rows[1].up.Disabled() {
		t.Error("a middle row should be able to move up")
	}
	if !r.rows[2].down.Disabled() {
		t.Error("the bottom row should have nowhere down to go")
	}
}

func TestRefusedMoveDoesNotAnimate(t *testing.T) {
	r, _ := newTestRank(t)

	if r.move(0, -1) {
		t.Error("the top row cannot move up")
	}
	if r.anim != nil {
		t.Error("a refused move should not animate anything")
	}
	if r.Touched() {
		t.Error("a refused move is not an interaction")
	}
}

// Resizing mid-slide must not snap the rows to where they were going.
func TestResizeLeavesASlideAlone(t *testing.T) {
	r, _ := newTestRank(t)

	caught := r.rows[0]
	r.animating = true
	caught.Move(fyne.NewPos(0, 9))

	r.Resize(fyne.NewSize(500, r.MinSize().Height))
	if got := caught.Position().Y; got != 9 {
		t.Errorf("the row was moved to %v by a resize mid-slide", got)
	}
	if got := caught.Size().Width; got != 500 {
		t.Errorf("the row is %v wide, want the resize to have applied to its width", got)
	}

	r.animating = false
	r.Resize(fyne.NewSize(500, r.MinSize().Height))
	if got := caught.Position().Y; got != 0 {
		t.Errorf("with no slide in flight the resize should have placed the row, got %v", got)
	}
}

// A drag walks the row through the list one place at a time, and each step
// animates like a press does.
func TestDragMovesOnePlacePerRowHeight(t *testing.T) {
	r, _ := newTestRank(t)
	row := r.rows[0]
	step := row.MinSize().Height

	row.Dragged(&fyne.DragEvent{Dragged: fyne.NewDelta(0, step/3)})
	if got := r.Order()[0]; got != "correctness" {
		t.Errorf("a third of a row of drag already moved it: order = %v", r.Order())
	}

	row.Dragged(&fyne.DragEvent{Dragged: fyne.NewDelta(0, step)})
	if got := r.Order(); got[0] != "speed" || got[1] != "correctness" {
		t.Errorf("order = %v, want the dragged row one place down", got)
	}
	if row.index != 1 {
		t.Errorf("the dragged row thinks it is at %d, want 1", row.index)
	}
	if r.anim == nil {
		t.Error("a drag step should animate like a press does")
	}

	row.DragEnd()
	if row.carried != 0 {
		t.Errorf("carried = %v after the drag ended", row.carried)
	}
}

// The rows have to stay reachable through the object tree, or nothing outside
// the widget can find the controls inside them.
func TestRankRowsAreReachable(t *testing.T) {
	r, _ := newTestRank(t)
	if _, ok := search[*rankRow](r, first[*rankRow]()); !ok {
		t.Error("no rank row in the widget's object tree")
	}
	if _, ok := search[*widget.Label](r, func(l *widget.Label) bool { return l.Text == "correctness" }); !ok {
		t.Error("a row's own label is not reachable")
	}
}
