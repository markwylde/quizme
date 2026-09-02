package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// rankSlide is how long a row takes to travel to its new place.
//
// Short enough to keep up with a responder pressing an arrow repeatedly, long
// enough to show which two options changed places. An instant reorder is
// accurate and unreadable: the list is simply different afterwards, and the
// reader has to work out what moved.
const rankSlide = 110 * time.Millisecond

// rankWidget orders a fixed set of options.
//
// Rows can be dragged, and each also carries move-up and move-down buttons:
// dragging inside a scrolling page is fiddly, and the buttons make the control
// usable by keyboard and trackpad alike. Whichever is used, the answer is
// always the full list in its current order, so it can never be a partial
// ranking.
//
// The rows are created once and then moved, rather than being rebuilt in a new
// order. A row that is destroyed and recreated cannot travel anywhere, and the
// travel is the whole point: it is what tells the responder what just happened.
type rankWidget struct {
	widget.BaseWidget

	order    []string
	rows     []*rankRow // in the order they are presented
	box      *fyne.Container
	touched  bool
	onChange func([]string)

	// anim carries every row that has moved, in one animation, so a single
	// thing owns the rows' positions for its duration -- and Resize knows to
	// keep its hands off while it does.
	anim      *fyne.Animation
	animating bool
}

func newRankWidget(order []string, onChange func([]string)) *rankWidget {
	r := &rankWidget{order: append([]string(nil), order...), onChange: onChange}
	r.ExtendBaseWidget(r)
	r.build()
	return r
}

func (r *rankWidget) build() {
	objects := make([]fyne.CanvasObject, 0, len(r.order))
	for i, option := range r.order {
		row := newRankRow(r, i, option, len(r.order))
		r.rows = append(r.rows, row)
		objects = append(objects, row)
	}
	// No layout: the rows are placed by position, which is what lets them be
	// animated from one place to another.
	r.box = container.NewWithoutLayout(objects...)
	r.place()
}

func (r *rankWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.box)
}

// Content exposes the rows, since a container without a layout is still the
// only route to them for anything walking the object tree.
func (r *rankWidget) Content() fyne.CanvasObject { return r.box }

// MinSize is the stack of rows. The container has no layout of its own, so it
// cannot work this out: it would report one row's worth of height.
func (r *rankWidget) MinSize() fyne.Size {
	r.ExtendBaseWidget(r)
	var size fyne.Size
	for i, row := range r.rows {
		min := row.MinSize()
		if min.Width > size.Width {
			size.Width = min.Width
		}
		if i > 0 {
			size.Height += r.gap()
		}
		size.Height += min.Height
	}
	return size
}

// Resize widens the rows to the space available. It leaves their positions
// alone while an animation owns them, or a resize mid-flight would snap every
// row to where it was going.
func (r *rankWidget) Resize(size fyne.Size) {
	r.BaseWidget.Resize(size)
	height := r.rowHeight()
	for _, row := range r.rows {
		row.Resize(fyne.NewSize(size.Width, height))
	}
	if !r.animating {
		r.place()
	}
}

func (r *rankWidget) rowHeight() float32 {
	if len(r.rows) == 0 {
		return 0
	}
	return r.rows[0].MinSize().Height
}

// gap is the space between rows, matching what a box layout would have left.
func (r *rankWidget) gap() float32 { return theme.Size(theme.SizeNamePadding) }

// slot is where the row presented in the given place belongs.
func (r *rankWidget) slot(index int) fyne.Position {
	return fyne.NewPos(0, float32(index)*(r.rowHeight()+r.gap()))
}

// place puts every row in its slot at once, with no travel.
func (r *rankWidget) place() {
	for i, row := range r.rows {
		row.Move(r.slot(i))
	}
}

// move shifts the option at index by delta places, clamped to the list.
func (r *rankWidget) move(index, delta int) bool {
	target := index + delta
	if index < 0 || index >= len(r.order) || target < 0 || target >= len(r.order) {
		return false
	}
	// Where the rows are before anything moves. This is captured first, and
	// the slide is started before the answer is reported, because reporting it
	// lays the page out again -- and a layout with no slide in flight places
	// every row in its final slot, which would leave the slide with nowhere to
	// travel from.
	starts := r.positions()

	r.order[index], r.order[target] = r.order[target], r.order[index]
	r.rows[index], r.rows[target] = r.rows[target], r.rows[index]
	r.reindex()
	r.touched = true
	r.slide(starts)

	// The answer is recorded now, in this same gesture: the rows catch up on
	// their own, and nothing about the questionnaire waits on them.
	r.notify()
	return true
}

// positions records where every row currently sits.
func (r *rankWidget) positions() map[*rankRow]fyne.Position {
	at := make(map[*rankRow]fyne.Position, len(r.rows))
	for _, row := range r.rows {
		at[row] = row.Position()
	}
	return at
}

// reindex tells each row where it now sits. A row draws its ordinal and decides
// which of its arrows it can offer from that, so this is what keeps the numbers
// honest without rebuilding anything.
func (r *rankWidget) reindex() {
	for i, row := range r.rows {
		row.setIndex(i)
	}
}

// slide animates every row that is not where it belongs into its slot, from
// the positions it held when the reorder was made.
func (r *rankWidget) slide(starts map[*rankRow]fyne.Position) {
	if r.anim != nil {
		// A second press mid-flight carries on from wherever the rows have
		// got to, rather than jumping them to the end of the last move first.
		r.anim.Stop()
		r.anim = nil
	}

	type travel struct {
		row      *rankRow
		from, to fyne.Position
	}
	var travels []travel
	for i, row := range r.rows {
		from, ok := starts[row]
		if !ok {
			from = row.Position()
		}
		to := r.slot(i)
		if from == to {
			continue
		}
		travels = append(travels, travel{row: row, from: from, to: to})
	}
	if len(travels) == 0 {
		return
	}

	// Set before anything else can lay the rows out: from here until the last
	// tick, the animation owns their positions.
	r.animating = true
	r.anim = &fyne.Animation{
		Duration: rankSlide,
		Curve:    fyne.AnimationEaseInOut,
		Tick: func(done float32) {
			for _, t := range travels {
				t.row.Move(fyne.NewPos(
					t.from.X+(t.to.X-t.from.X)*done,
					t.from.Y+(t.to.Y-t.from.Y)*done,
				))
			}
			if done >= 1 {
				r.animating = false
			}
			r.box.Refresh()
		},
	}
	r.anim.Start()
}

// confirm marks the current order as the answer without changing it, so a
// responder who agrees with the order as presented has something to press.
func (r *rankWidget) confirm() {
	r.touched = true
	r.notify()
}

func (r *rankWidget) notify() {
	if r.onChange != nil {
		r.onChange(append([]string(nil), r.order...))
	}
}

// Order returns the current ranking.
func (r *rankWidget) Order() []string { return append([]string(nil), r.order...) }

// Touched reports whether the responder has interacted with the control. An
// untouched ranking is not an answer: it is only the order the author happened
// to write the options in.
func (r *rankWidget) Touched() bool { return r.touched }

// rankRow is one draggable option.
type rankRow struct {
	widget.BaseWidget

	list  *rankWidget
	index int
	label string
	total int

	// carried accumulates vertical drag until it is worth a whole row, so a
	// drag moves the option one place at a time rather than jumping.
	carried float32

	position *widget.Label
	up, down *widget.Button
	content  *fyne.Container
}

// newRankRow builds the row's parts once. They are updated in place as the row
// changes position, because the row itself has to survive a reorder to be able
// to travel across one.
func newRankRow(list *rankWidget, index int, label string, total int) *rankRow {
	r := &rankRow{list: list, index: index, label: label, total: total}
	r.ExtendBaseWidget(r)

	r.up = widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() { r.list.move(r.index, -1) })
	r.down = widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() { r.list.move(r.index, 1) })

	r.position = widget.NewLabel(ordinal(index + 1))
	r.position.TextStyle = fyne.TextStyle{Monospace: true}

	r.content = container.NewBorder(nil, nil, r.position,
		container.NewHBox(r.up, r.down), widget.NewLabel(r.label))
	r.applyIndex()
	return r
}

func (r *rankRow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.content)
}

// Content exposes the row's parts, which otherwise hang off its renderer.
func (r *rankRow) Content() fyne.CanvasObject { return r.content }

// setIndex records that the row is now presented in a different place.
func (r *rankRow) setIndex(index int) {
	if r.index == index {
		return
	}
	r.index = index
	r.applyIndex()
}

// applyIndex brings what the row draws from its position back in step: its
// ordinal, and which arrows it has anywhere to go with.
func (r *rankRow) applyIndex() {
	r.position.SetText(ordinal(r.index + 1))
	if r.index == 0 {
		r.up.Disable()
	} else {
		r.up.Enable()
	}
	if r.index == r.total-1 {
		r.down.Disable()
	} else {
		r.down.Enable()
	}
}

func (r *rankRow) Dragged(e *fyne.DragEvent) {
	step := r.Size().Height
	if step <= 0 {
		step = r.MinSize().Height
	}
	if step <= 0 {
		return
	}
	r.carried += e.Dragged.DY
	// move reindexes the rows, so r.index is already the row's new place on the
	// next turn of the loop.
	for r.carried <= -step && r.list.move(r.index, -1) {
		r.carried += step
	}
	for r.carried >= step && r.list.move(r.index, 1) {
		r.carried -= step
	}
}

func (r *rankRow) DragEnd() { r.carried = 0 }

// ordinal renders a 1-based position as "1.", padded so the rows line up.
func ordinal(n int) string {
	if n < 10 {
		return " " + itoa(n) + "."
	}
	return itoa(n) + "."
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
