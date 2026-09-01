package ui

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// scaleWidget is a segmented control: one button per point on the scale.
//
// Fyne has a slider, but a slider is a poor fit for a five-point question --
// it invites fiddling for a precision the question does not have, and it never
// shows an unanswered state. Discrete buttons say exactly what the choices are
// and can be left alone.
type scaleWidget struct {
	widget.BaseWidget

	lo, hi   int
	value    *int // nil when unanswered
	buttons  []*widget.Button
	onChange func(*int)
}

func newScaleWidget(lo, hi int, current *int, onChange func(*int)) *scaleWidget {
	s := &scaleWidget{lo: lo, hi: hi, value: current, onChange: onChange}
	s.ExtendBaseWidget(s)
	return s
}

func (s *scaleWidget) CreateRenderer() fyne.WidgetRenderer {
	row := container.NewHBox()
	s.buttons = nil
	for n := s.lo; n <= s.hi; n++ {
		point := n
		b := widget.NewButton(strconv.Itoa(point), func() { s.choose(point) })
		s.buttons = append(s.buttons, b)
		row.Add(b)
	}
	s.paint()
	return widget.NewSimpleRenderer(row)
}

// choose selects a point, or clears the answer when the current point is
// pressed again, so a scale that is not required can be left blank.
func (s *scaleWidget) choose(point int) {
	if s.value != nil && *s.value == point {
		s.value = nil
	} else {
		v := point
		s.value = &v
	}
	s.paint()
	if s.onChange != nil {
		s.onChange(s.value)
	}
}

// paint marks the selected button and leaves the rest plain.
func (s *scaleWidget) paint() {
	for i, b := range s.buttons {
		point := s.lo + i
		if s.value != nil && *s.value == point {
			b.Importance = widget.HighImportance
		} else {
			b.Importance = widget.MediumImportance
		}
		b.Refresh()
	}
}

// Value returns the selected point, or nil when the question is unanswered.
func (s *scaleWidget) Value() *int { return s.value }
