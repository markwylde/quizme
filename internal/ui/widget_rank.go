package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// rankWidget orders a fixed set of options.
//
// Rows can be dragged, and each also carries move-up and move-down buttons:
// dragging inside a scrolling page is fiddly, and the buttons make the control
// usable by keyboard and trackpad alike. Whichever is used, the answer is
// always the full list in its current order, so it can never be a partial
// ranking.
type rankWidget struct {
	widget.BaseWidget

	order    []string
	rows     []*rankRow
	box      *fyne.Container
	touched  bool
	onChange func([]string)
}

func newRankWidget(order []string, onChange func([]string)) *rankWidget {
	r := &rankWidget{order: append([]string(nil), order...), onChange: onChange}
	r.ExtendBaseWidget(r)
	return r
}

func (r *rankWidget) CreateRenderer() fyne.WidgetRenderer {
	r.box = container.NewVBox()
	r.rebuild()
	return widget.NewSimpleRenderer(r.box)
}

// rebuild lays the rows out in the current order. The rows are recreated rather
// than reordered because a row's position is the only thing that identifies it.
func (r *rankWidget) rebuild() {
	if r.box == nil {
		return
	}
	r.box.RemoveAll()
	r.rows = nil
	for i, option := range r.order {
		row := newRankRow(r, i, option, len(r.order))
		r.rows = append(r.rows, row)
		r.box.Add(row)
	}
	r.box.Refresh()
}

// move shifts the option at index by delta places, clamped to the list.
func (r *rankWidget) move(index, delta int) bool {
	target := index + delta
	if index < 0 || index >= len(r.order) || target < 0 || target >= len(r.order) {
		return false
	}
	r.order[index], r.order[target] = r.order[target], r.order[index]
	r.touched = true
	r.rebuild()
	r.notify()
	return true
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
}

func newRankRow(list *rankWidget, index int, label string, total int) *rankRow {
	r := &rankRow{list: list, index: index, label: label, total: total}
	r.ExtendBaseWidget(r)
	return r
}

func (r *rankRow) CreateRenderer() fyne.WidgetRenderer {
	up := widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() { r.list.move(r.index, -1) })
	down := widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() { r.list.move(r.index, 1) })
	if r.index == 0 {
		up.Disable()
	}
	if r.index == r.total-1 {
		down.Disable()
	}

	position := widget.NewLabel(ordinal(r.index + 1))
	position.TextStyle = fyne.TextStyle{Monospace: true}

	content := container.NewBorder(nil, nil, position, container.NewHBox(up, down), widget.NewLabel(r.label))
	return widget.NewSimpleRenderer(content)
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
	for r.carried <= -step && r.list.move(r.index, -1) {
		r.carried += step
		r.index--
	}
	for r.carried >= step && r.list.move(r.index, 1) {
		r.carried -= step
		r.index++
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
