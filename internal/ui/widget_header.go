package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/markwylde/quizme/internal/questionnaire"
)

// questionHeader is the top line of a question's card, and the control that
// folds it.
//
// The whole row is the target rather than a chevron alone: folding is the
// commonest gesture on the page and does not deserve a twenty-pixel button.
// It is a widget rather than a Button because a button cannot hold a wrapping
// bold prompt beside a caption, a drawn mark, and an icon; Fyne's hit test
// walks up to the nearest Tappable, so the labels inside still deliver the tap.
//
// Collapsed, the header is the only thing left of the question, so it carries
// the answer as well as the prompt -- on its own line beneath it, where a
// wrapped prompt cannot strand it: a folded page reads as a review of what the
// responder has said rather than a list of prompts.
type questionHeader struct {
	widget.BaseWidget

	question *questionnaire.Question
	onTap    func()

	chevron *widget.Icon
	prompt  *promptText
	marker  *widget.Label // "Required", or nil when the question is optional
	summary *widget.Label
	note    *widget.Icon
	tick    *tickMark

	stack    *fyne.Container // the prompt, with the answer beneath it
	trailing *fyne.Container // the marks that sit against the prompt's first line
	row      *fyne.Container

	// onHover reports the pointer arriving and leaving. The highlight belongs
	// to the whole card rather than to the header: painted here it was a
	// rounded panel inset by the card's own padding, and read as a second card
	// inside the first.
	onHover func(bool)

	collapsed bool
}

func newQuestionHeader(q *questionnaire.Question, onTap func()) *questionHeader {
	h := &questionHeader{question: q, onTap: onTap}
	h.ExtendBaseWidget(h)

	h.chevron = widget.NewIcon(theme.MenuDropDownIcon())

	h.prompt = newPromptText(q.Prompt)

	// The summary has a line of its own beneath the prompt, so it can take the
	// width that is left and truncate against it. (Beside the prompt it could
	// not: a truncating label reports almost no minimum width, and in a row
	// driven by minimum widths it collapsed to a bare ellipsis.)
	h.summary = captionLabel("")
	h.summary.Truncation = fyne.TextTruncateEllipsis

	// The same icon the comment toggle uses, so a folded question says "there
	// is a note in here" in the vocabulary the open one taught.
	h.note = widget.NewIcon(theme.DocumentCreateIcon())

	h.tick = newTickMark()

	var trailing []fyne.CanvasObject
	if q.Required {
		// The required marker sits on the prompt's own line. On its own row it
		// read as another instruction to take in; up here it is just a label.
		h.marker = captionLabel("Required")
		h.marker.Importance = widget.MediumImportance
		trailing = append(trailing, h.marker)
	}
	trailing = append(trailing, h.note, h.tick)

	// The answer goes under the prompt rather than beside it. A prompt long
	// enough to wrap -- which is most of them, on a real questionnaire -- left
	// the answer stranded out to the right, level with a line of the question
	// it had nothing to do with.
	// The stack gives back exactly the answer's own top padding, so what stands
	// between the prompt and the answer is the prompt's tail and nothing else --
	// the same distance an open question leaves above its controls.
	h.stack = container.New(
		&tightStack{gap: theme.Size(theme.SizeNameInnerPadding)}, h.prompt, h.summary)

	// The chevron and the trailing marks are placed against the prompt's first
	// line of text, which is neither the top of the header nor the middle of
	// it: a label insets its text by the inner padding, and a wrapped prompt is
	// several lines tall.
	h.trailing = container.New(&centredRow{}, trailing...)
	h.row = container.New(&headerLayout{}, h.chevron, h.stack, h.trailing)
	h.sync()
	return h
}

func (h *questionHeader) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.row)
}

// Content exposes the row. A widget's children hang off its renderer, so
// without this the header's parts are invisible to anything walking the tree.
func (h *questionHeader) Content() fyne.CanvasObject { return h.row }

// Objects exposes the row without building a renderer first, matching what the
// other wrapping widgets in this package do.
func (h *questionHeader) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{h.row}
}

// SetCollapsed points the chevron the right way and decides whether the row has
// to speak for the whole question.
func (h *questionHeader) SetCollapsed(collapsed bool) {
	h.collapsed = collapsed
	h.sync()
}

// Collapsed reports which way the header currently reads.
func (h *questionHeader) Collapsed() bool { return h.collapsed }

// Sync brings the row back in step with the question behind it: the answer it
// shows, whether it is ticked, and whether it carries a note.
func (h *questionHeader) Sync() { h.sync() }

func (h *questionHeader) sync() {
	if h.collapsed {
		h.chevron.SetResource(theme.MenuExpandIcon())
	} else {
		h.chevron.SetResource(theme.MenuDropDownIcon())
	}

	// The answer is worth showing only when the control that holds it is out of
	// sight; expanded, it would just be the control said twice.
	text := answerSummary(h.question)
	h.summary.SetText(text)
	show(h.summary, h.collapsed && text != "")

	// The tick follows the answer, not the fold: a question left open because
	// the responder pinned it open is no less settled for being visible.
	show(h.tick, h.question.HasAnswer())
	show(h.note, strings.TrimSpace(h.question.Comment) != "")

	h.Refresh()
}

func (h *questionHeader) Tapped(*fyne.PointEvent) {
	if h.onTap != nil {
		h.onTap()
	}
}

// SetHoverReporter says where to report the pointer arriving and leaving.
func (h *questionHeader) SetHoverReporter(report func(bool)) { h.onHover = report }

func (h *questionHeader) MouseIn(*desktop.MouseEvent) { h.hover(true) }

func (h *questionHeader) MouseOut() { h.hover(false) }

func (h *questionHeader) MouseMoved(*desktop.MouseEvent) {}

func (h *questionHeader) hover(in bool) {
	if h.onHover != nil {
		h.onHover(in)
	}
}

// Cursor says the row can be pressed, which is the affordance a header has
// instead of looking like a button.
func (h *questionHeader) Cursor() desktop.Cursor { return desktop.PointerCursor }

// show is Show or Hide by boolean, which reads better than the four-line if
// this file would otherwise repeat.
func show(obj fyne.CanvasObject, visible bool) {
	if visible {
		obj.Show()
		return
	}
	obj.Hide()
}

// tightStack stacks labels as one block of text, gap pixels apart.
//
// The gap is measured between the lines of text, not between the labels: a
// label pads itself top and bottom, so two of them in a box layout sit a whole
// label apart -- around three times the leading of one wrapped paragraph -- and
// read as separate things. The stack takes that padding back out and adds the
// gap it is asked for.
//
// This is the only way to set the leading between lines at all. Fyne applies
// its line spacing between rich-text segments and explicitly not between the
// rows inside one, so the wrapped lines of a label ignore the theme entirely.
type tightStack struct {
	gap float32
}

func (t *tightStack) overlap() float32 {
	// A label's own padding, top and bottom, is what stands between the two
	// lines of text before anything is asked for.
	overlap := 2*theme.Size(theme.SizeNameInnerPadding) - t.gap
	if overlap < 0 {
		return 0
	}
	return overlap
}

func (t *tightStack) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var size fyne.Size
	first := true
	for _, o := range objects {
		if !o.Visible() {
			continue
		}
		min := o.MinSize()
		if min.Width > size.Width {
			size.Width = min.Width
		}
		if !first {
			size.Height -= t.overlap()
		}
		size.Height += min.Height
		first = false
	}
	return size
}

func (t *tightStack) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	// Width first, for every child, before any of them is placed. The prompt
	// wraps its own text, so it only knows how tall it is once it has a width
	// -- and everything below it is placed from that height. Positioning as we
	// went put the answer where a stale height said the prompt ended.
	y := float32(0)
	first := true
	for _, o := range objects {
		if !o.Visible() {
			continue
		}
		if !first {
			y -= t.overlap()
		}
		height := o.MinSize().Height
		o.Resize(fyne.NewSize(size.Width, height))
		o.Move(fyne.NewPos(0, y))
		y += height
		first = false
	}
}

// headerLayout places the fold control and the trailing marks against the first
// line of the prompt, with the prompt and its answer between them.
//
// A border layout cannot do this. Given the whole height of the header it
// stretches its edges to match, so on a prompt long enough to wrap the chevron
// and the tick drift towards the middle of the question; boxed to stop that,
// they sit against the top of the prompt's box instead, which is the inner
// padding above where the text actually starts.
type headerLayout struct{}

func (l *headerLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	chevron, stack, trailing, ok := l.parts(objects)
	if !ok {
		return fyne.Size{}
	}
	c, s, t := chevron.MinSize(), stack.MinSize(), trailing.MinSize()
	return fyne.NewSize(
		promptInset()+s.Width+theme.Size(theme.SizeNamePadding)+t.Width,
		fyne.Max(s.Height, fyne.Max(c.Height, t.Height)),
	)
}

func (l *headerLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	chevron, stack, trailing, ok := l.parts(objects)
	if !ok {
		return
	}
	pad := theme.Size(theme.SizeNamePadding)
	c, t := chevron.MinSize(), trailing.MinSize()

	width := size.Width - promptInset() - t.Width - pad
	if width < 0 {
		width = 0
	}
	// Twice, deliberately. The first gives the prompt its width, which is what
	// tells it how many lines it takes and so how tall the stack really is; the
	// second sizes the stack to that. Skipping the second leaves the card as
	// tall as whatever the stack last thought it was.
	stack.Resize(fyne.NewSize(width, stack.MinSize().Height))
	stack.Move(fyne.NewPos(promptInset(), 0))

	line := promptLineCentre()
	chevron.Resize(c)
	// Centred in the column rather than jammed against its left edge, so a
	// chevron narrower than an inline icon still looks deliberate.
	chevron.Move(fyne.NewPos((promptInset()-pad-c.Width)/2, line-c.Height/2))
	trailing.Resize(t)
	trailing.Move(fyne.NewPos(size.Width-t.Width, line-t.Height/2))
}

func (l *headerLayout) parts(objects []fyne.CanvasObject) (chevron, stack, trailing fyne.CanvasObject, ok bool) {
	if len(objects) != 3 {
		return nil, nil, nil, false
	}
	return objects[0], objects[1], objects[2], true
}

// promptInset is where a question's prompt starts: past the column the fold
// chevron sits in. Everything else the question shows lines up with it, so this
// is the one place that decides how far in that is.
func promptInset() float32 {
	return theme.Size(theme.SizeNameInlineIcon) + theme.Size(theme.SizeNamePadding)
}

// promptLineCentre is the middle of a prompt's first line of text, measured from
// the top of the header.
func promptLineCentre() float32 {
	// "Ag" for an ascender and a descender, so the measurement is a full line
	// whatever the prompt happens to say.
	line := fyne.MeasureText("Ag", theme.Size(theme.SizeNameText), promptStyle).Height
	return theme.Size(theme.SizeNameInnerPadding) + line/2
}

// promptColumn indents whatever it holds to the prompt's own column.
//
// A question's controls used to start at the card's edge, a chevron's width to
// the left of the prompt they answer, so nothing on the card shared a left
// edge. The space this leaves under the chevron is the price of that alignment,
// and a cheap one.
type promptColumn struct{}

func (c *promptColumn) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var size fyne.Size
	for _, o := range objects {
		if !o.Visible() {
			continue
		}
		min := o.MinSize()
		if min.Width > size.Width {
			size.Width = min.Width
		}
		size.Height += min.Height
	}
	if size.Height == 0 {
		return size // nothing showing: take up no room at all
	}
	return fyne.NewSize(size.Width+promptInset(), size.Height)
}

func (c *promptColumn) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	width := size.Width - promptInset()
	if width < 0 {
		width = 0
	}
	y := float32(0)
	for _, o := range objects {
		if !o.Visible() {
			continue
		}
		height := o.MinSize().Height
		o.Resize(fyne.NewSize(width, height))
		o.Move(fyne.NewPos(promptInset(), y))
		y += height
	}
}

// cardStack stacks a question's header and its body, gap apart.
//
// A box layout cannot: it positions each child as it goes, from a minimum size
// the child does not know yet. The header wraps its prompt, so its height is
// only settled once it has a width, and a box layout that had already placed
// the body left the card as tall as whatever the header last thought it was.
type cardStack struct {
	gap float32
}

func (c *cardStack) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var size fyne.Size
	first := true
	for _, o := range objects {
		if !o.Visible() {
			continue
		}
		min := o.MinSize()
		if min.Width > size.Width {
			size.Width = min.Width
		}
		if !first {
			size.Height += c.gap
		}
		size.Height += min.Height
		first = false
	}
	return size
}

func (c *cardStack) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	// Every child gets its width before any of them is placed, and then again
	// as it is placed. Both passes matter: the header wraps its prompt, so it
	// only settles its height once it has a width. The first pass gets it
	// there; the second places everything from the height that settling
	// produced. Without it a question whose prompt wrapped was laid out from
	// the height its prompt had before it knew how wide it was -- which left
	// the answer adrift and the card half as tall again as it needed.
	for _, o := range objects {
		if o.Visible() {
			o.Resize(fyne.NewSize(size.Width, o.MinSize().Height))
		}
	}

	y := float32(0)
	first := true
	for _, o := range objects {
		if !o.Visible() {
			continue
		}
		if !first {
			y += c.gap
		}
		height := o.MinSize().Height
		o.Resize(fyne.NewSize(size.Width, height))
		o.Move(fyne.NewPos(0, y))
		y += height
		first = false
	}
}

// centredRow lays a handful of marks out in a line, each centred on the row's
// own middle. A box layout would hang them all from the top, which puts a
// 15-pixel tick and a label with padding of its own on different centrelines.
type centredRow struct{}

func (r *centredRow) MinSize(objects []fyne.CanvasObject) fyne.Size {
	pad := theme.Size(theme.SizeNamePadding)
	var size fyne.Size
	first := true
	for _, o := range objects {
		if !o.Visible() {
			continue
		}
		min := o.MinSize()
		if !first {
			size.Width += pad
		}
		size.Width += min.Width
		if min.Height > size.Height {
			size.Height = min.Height
		}
		first = false
	}
	return size
}

func (r *centredRow) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	pad := theme.Size(theme.SizeNamePadding)
	x := float32(0)
	first := true
	for _, o := range objects {
		if !o.Visible() {
			continue
		}
		min := o.MinSize()
		if !first {
			x += pad
		}
		o.Resize(min)
		o.Move(fyne.NewPos(x, (size.Height-min.Height)/2))
		x += min.Width
		first = false
	}
}

// The header answers to tapping, to the pointer arriving, and to what the
// cursor should look like; everything else belongs to what is inside it.
var (
	_ fyne.Tappable      = (*questionHeader)(nil)
	_ desktop.Hoverable  = (*questionHeader)(nil)
	_ desktop.Cursorable = (*questionHeader)(nil)
	_ fyne.Layout        = (*tightStack)(nil)
	_ fyne.Layout        = (*headerLayout)(nil)
	_ fyne.Layout        = (*centredRow)(nil)
	_ fyne.Layout        = (*promptColumn)(nil)
	_ fyne.Layout        = (*cardStack)(nil)
)
