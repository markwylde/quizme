package ui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/markwylde/quizme/internal/config"
	"github.com/markwylde/quizme/internal/questionnaire"
)

// radioLimit is the point past which a list of options stops being worth
// showing in full and becomes a dropdown.
const radioLimit = 7

// Run presents the questionnaire and blocks until the responder submits, saves,
// or dismisses it, returning the status that outcome corresponds to.
//
// icon is the image to show for the window and the running application, as the
// bytes of an SVG. It is passed in rather than embedded here because the
// drawing lives at the repository root, where the packaging tools also look for
// it, and a package can only embed what sits beside it. Empty leaves whatever
// the platform gives a bare binary.
func Run(doc *questionnaire.Document, icon []byte, opts Options) (questionnaire.Status, error) {
	if !config.ValidTextSize(opts.TextSize) {
		opts.TextSize = config.DefaultTextSize
	}
	a := app.NewWithID("com.markwylde.quizme")
	a.Settings().SetTheme(newScaledTheme(opts.TextSize))
	if len(icon) > 0 {
		a.SetIcon(iconResource(icon))
	}

	title := doc.Title
	if title == "" {
		title = "Questionnaire"
	}
	win := a.NewWindow(title)
	if len(icon) > 0 {
		win.SetIcon(iconResource(icon))
	}

	f := newForm(doc, win)
	f.textSize = opts.TextSize
	f.onTextSize = opts.OnTextSize
	win.SetContent(f.build())
	f.registerSizeShortcuts(win.Canvas())
	win.Resize(fyne.NewSize(780, 760))
	win.CenterOnScreen()
	win.SetCloseIntercept(f.requestClose)
	win.ShowAndRun()

	return f.outcome, nil
}

// Options are the responder's own preferences for how the form is shown, which
// come from outside the questionnaire.
type Options struct {
	// TextSize is the size to open at, as a percentage of the default. Anything
	// the form does not offer opens at the default.
	TextSize int
	// OnTextSize, if set, is told every size the responder changes to, so it
	// can be remembered for next time.
	OnTextSize func(percent int)
}

// iconResource wraps the icon's bytes for the toolkit. The name matters: Fyne
// decides how to read a resource from its extension, so an SVG has to say so.
func iconResource(svg []byte) fyne.Resource {
	return &fyne.StaticResource{StaticName: "icon.svg", StaticContent: svg}
}

// snapshot is a question's state when the form opened, used to tell whether the
// responder has actually changed anything.
type snapshot struct {
	answer  string
	comment string
}

// form is one questionnaire on screen.
type form struct {
	doc *questionnaire.Document
	win fyne.Window

	outcome questionnaire.Status

	cards     map[string]*card
	baseline  map[string]snapshot
	list      *fyne.Container
	scroll    *container.Scroll
	status    *widget.Label
	progress  *progressBar
	footerRow *fyne.Container

	// textSize is the size the form is shown at, as a percentage of the
	// default, and onTextSize hears about every change to it.
	textSize   int
	onTextSize func(int)
	smaller    *widget.Button
	larger     *widget.Button

	// building is set while the page is being assembled. Restoring an answer
	// already in the file goes through the same control callbacks a responder's
	// gesture does -- Fyne's SetSelected fires OnChanged -- and a restore is
	// not a gesture. How an answered question opens is decided once, in
	// buildCard, rather than falling out of however many times a control
	// happens to report itself on the way up.
	building bool
}

// foldState is how a question is presented, and why.
//
// The distinction between the two open states is what keeps auto-collapse from
// fighting the responder: a question they opened themselves stays open, whatever
// they answer next.
type foldState int

const (
	foldAuto   foldState = iota // expanded, and the responder has not said otherwise
	foldOpen                    // expanded because the responder opened it
	foldClosed                  // collapsed
)

// card is one question: the header that folds it, its control, its comment box,
// and the warning shown when it is required and unanswered.
type card struct {
	question *questionnaire.Question
	root     *fyne.Container
	header   *questionHeader
	body     *fyne.Container
	control  fyne.CanvasObject
	warning  *widget.Label
	rank     *rankWidget
	comment  *commentField
	done     *widget.Button
	card     *cardBox

	fold foldState
}

// collapsed reports whether the question is folded to its header.
func (c *card) collapsed() bool { return c.fold == foldClosed }

// pinned reports whether the responder has decided this question's presentation
// for themselves, which an answer must not then undo.
func (c *card) pinned() bool { return c.fold != foldAuto }

func newForm(doc *questionnaire.Document, win fyne.Window) *form {
	f := &form{
		doc:      doc,
		win:      win,
		cards:    map[string]*card{},
		baseline: map[string]snapshot{},
		textSize: config.DefaultTextSize,
	}
	for _, q := range doc.Questions {
		f.baseline[q.ID] = snapshot{answer: answerKey(q.Answer), comment: q.Comment}
	}
	return f
}

// build assembles the whole window: a fixed header, one scrolling page of
// questions, and a fixed footer holding the actions.
func (f *form) build() fyne.CanvasObject {
	f.building = true
	defer func() { f.building = false }()

	f.list = container.NewVBox()

	// The title and intro scroll away with the questions. They are context,
	// read once before starting; pinning them would spend a fifth of the window
	// on them for the rest of the session.
	f.list.Add(f.header())
	for _, q := range f.doc.Questions {
		c := f.buildCard(q)
		f.cards[q.ID] = c
		f.list.Add(c.root)
	}

	f.scroll = container.NewVScroll(container.NewPadded(f.list))

	// The footer holds the progress indicator and the actions, which are needed
	// from anywhere in the page, so it stays pinned. It must also exist before
	// the first visibility pass has anything to report to.
	footer := f.footer()
	f.syncVisibility()

	return container.NewBorder(nil, footer, nil, nil, f.scroll)
}

func (f *form) header() fyne.CanvasObject {
	title := widget.NewLabel(f.doc.Title)
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.SizeName = theme.SizeNameHeadingText
	title.Wrapping = fyne.TextWrapWord

	// The size controls sit level with the title, out of the way of the
	// questions, and scroll away with it; the shortcuts work from anywhere.
	items := []fyne.CanvasObject{container.NewBorder(nil, nil, nil, f.sizeButtons(), title)}
	if intro := strings.TrimSpace(f.doc.Intro); intro != "" {
		items = append(items, f.intro(intro))
	}
	items = append(items, widget.NewSeparator())
	return container.NewVBox(items...)
}

// introLeading is the space between the intro's lines. The intro is the one
// paragraph on the page that is read rather than answered, and the leading a
// label sets for itself is tight for that.
const introLeading = 6

// intro lays the questionnaire's introduction out a line at a time.
//
// One label per authored line, rather than one label holding all of them: the
// leading between rows inside a single label is the toolkit's own and cannot be
// set, while the space between two labels is ours to choose.
func (f *form) intro(intro string) fyne.CanvasObject {
	var lines []fyne.CanvasObject
	for _, line := range strings.Split(intro, "\n") {
		body := widget.NewLabel(strings.TrimRight(line, " \t"))
		body.Wrapping = fyne.TextWrapWord
		lines = append(lines, body)
	}
	return container.New(&tightStack{gap: introLeading}, lines...)
}

func (f *form) footer() fyne.CanvasObject {
	// Progress and outstanding-required are different questions -- "how far
	// through am I" and "what is stopping me submitting" -- so a bar answers
	// the first at a glance and a label answers the second in words.
	f.progress = newProgressBar()
	// One label rather than two side by side: two labels only share a baseline
	// while they share a text size and a row, and neither survives the text
	// growing. One sentence wraps like any other.
	f.status = widget.NewLabel("")
	f.status.Wrapping = fyne.TextWrapWord

	submit := widget.NewButton("Submit", f.submit)
	submit.Importance = widget.HighImportance
	dismiss := widget.NewButton("Dismiss", f.requestClose)
	// Starting over sits with the other form-wide actions, where it can be
	// reached from anywhere, but quietly: it is the one that throws work away.
	clear := widget.NewButton("Clear all answers", f.requestClear)
	clear.Importance = widget.LowImportance

	actions := container.NewHBox(clear, dismiss, submit)
	row := &footerLayout{}
	f.footerRow = container.New(row, f.status, actions)
	row.owner = f.footerRow

	// The bar spans the window along the top edge of the footer, where it also
	// does the job the separator was doing.
	// Inset to the cards' edges, so the status text and the last button line up
	// with the page above rather than with the window.
	edge := func() float32 { return 2 * theme.Size(theme.SizeNamePadding) }
	vertical := func() float32 { return theme.Size(theme.SizeNamePadding) }
	return container.NewVBox(f.progress, container.New(&themedPadding{
		top: vertical, bottom: vertical, left: edge, right: edge,
	}, f.footerRow))
}

// buildCard lays out one question. Every question gets a comment box, whatever
// its type: the note is often the most precise thing the responder has to say.
//
// The prompt lives in the header, which is also the control that folds the
// question; everything else lives in the body, which is what folding hides.
func (f *form) buildCard(q *questionnaire.Question) *card {
	c := &card{question: q}

	c.header = newQuestionHeader(q, func() { f.toggle(c) })

	var parts []fyne.CanvasObject
	if help := strings.TrimSpace(q.Help); help != "" {
		body := widget.NewLabel(help)
		body.Wrapping = fyne.TextWrapWord
		body.SizeName = theme.SizeNameCaptionText
		// A label starts its text an inner padding below its own box. The
		// answer on a folded question has that padding taken back by the stack
		// above it, so the help has to have it taken back too -- otherwise the
		// first line under a prompt jumps down as the question opens.
		parts = append(parts, alignFirstLine(body))
	}

	c.control = f.control(q, c)
	parts = append(parts, c.control)

	c.comment = newCommentField(q.Comment, func(s string) {
		q.Comment = s
		c.header.Sync() // so a folded question still says it carries a note
	}, f.scrollable, f.changeTextSize)
	// An answer that is typed or ticked together has no single settling
	// gesture, so those questions get one to press. It rides on the comment
	// toggle's row rather than taking a row of its own.
	if needsDone(q.Type) {
		c.done = widget.NewButton("Done", func() { f.finished(q) })
		c.done.Importance = widget.LowImportance
		c.comment.SetTrailing(c.done)
	}
	parts = append(parts, c.comment.root)

	c.warning = widget.NewLabel("")
	c.warning.Importance = widget.DangerImportance
	c.warning.Hide()
	parts = append(parts, c.warning)

	// Indented to the prompt's column: the controls answer the prompt, so they
	// line up with it rather than with the card's edge.
	c.body = container.New(&promptColumn{}, container.NewVBox(parts...))

	// The card and the gap around it do the separating a hairline rule was
	// failing to do, so the separator goes: a card edge and a rule together
	// only look fussy.
	c.card = newCardBox(container.New(cardPadding(),
		container.New(&cardStack{gap: headerBodyGap}, c.header, c.body)))
	c.header.SetHoverReporter(c.card.SetHovered)
	c.root = container.NewPadded(c.card)

	// A question that arrives already answered opens folded. It is settled
	// work, and a questionnaire reopened -- or authored with answers in it --
	// should open showing what is left rather than what is done. Everything
	// still unanswered opens in full, so nothing outstanding is ever hidden
	// from a responder seeing the page for the first time.
	if q.HasAnswer() {
		c.fold = foldClosed
		f.applyFold(c)
	}
	return c
}

// How far into itself each kind of control draws, measured from its own left
// edge, in the toolkit as it stands. Fyne's widgets disagree: a radio group
// sets its buttons in a little way, an entry draws its border at the very
// edge, a button fills its whole box, and text inside a label starts at the
// inner padding.
//
// A card lines up when every one of them draws at the same place as the
// prompt's text, so the form makes up the difference. The numbers are the
// toolkit's, not ours -- TestEveryControlLinesUpWithThePromptsText measures
// them, and will say so if a new Fyne version moves one.
const (
	choiceInk = 5   // a radio or check group's button
	entryInk  = 0.5 // an entry's border stroke
	buttonInk = 0   // a button's filled background
)

// alignFirstLine lifts a label so its text starts where a folded question's
// answer does, rather than an inner padding lower.
func alignFirstLine(label fyne.CanvasObject) fyne.CanvasObject {
	return container.New(&themedPadding{top: func() float32 {
		return -theme.Size(theme.SizeNameInnerPadding)
	}}, label)
}

// alignInk sets a control in far enough that it draws where the prompt's text
// does, given how far into itself it already draws.
func alignInk(control fyne.CanvasObject, ink float32) fyne.CanvasObject {
	return container.New(&themedPadding{left: func() float32 {
		return max(theme.Size(theme.SizeNameInnerPadding)-scaled(ink), 0)
	}}, control)
}

// cardOpticalTail is the extra space under a card's last line.
//
// Even padding measured from the text's box does not look even. A line box
// carries the room a capital needs above the letters and the room a descender
// needs below, and the eye reads the space to the capitals at the top against
// the space from the baseline at the bottom -- so a card padded equally reads
// as tight underneath. This is the difference, and it is why the bottom padding
// is deliberately not the same number as the top.
const cardOpticalTail = 2

// cardPadding is the space between a card's edge and the question inside it.
func cardPadding() fyne.Layout {
	pad := func() float32 { return theme.Size(theme.SizeNamePadding) }
	return &themedPadding{
		top:    pad,
		bottom: func() float32 { return pad() + scaled(cardOpticalTail) },
		left:   pad,
		right:  pad,
	}
}

// themedPadding pads its content by amounts read from the theme each time it
// lays out, rather than fixed when the page was built, so the page follows a
// change of text size without being rebuilt. A nil side is no padding.
type themedPadding struct {
	top, bottom, left, right func() float32
}

func (p *themedPadding) sides() (top, bottom, left, right float32) {
	read := func(f func() float32) float32 {
		if f == nil {
			return 0
		}
		return f()
	}
	return read(p.top), read(p.bottom), read(p.left), read(p.right)
}

func (p *themedPadding) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	top, bottom, left, right := p.sides()
	pos := fyne.NewPos(left, top)
	inner := fyne.NewSize(size.Width-left-right, size.Height-top-bottom)
	for _, o := range objects {
		o.Resize(inner)
		o.Move(pos)
	}
}

func (p *themedPadding) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var min fyne.Size
	for _, o := range objects {
		if o.Visible() {
			min = min.Max(o.MinSize())
		}
	}
	top, bottom, left, right := p.sides()
	return fyne.NewSize(min.Width+left+right, min.Height+top+bottom)
}

// headerBodyGap is the space between a question's prompt and the controls
// beneath it. It is measured from the bottom of the prompt's box, which already
// carries the inner padding, so it is smaller than it looks.
const headerBodyGap = 0

// needsDone reports whether a type's answer is composed over several actions,
// and so cannot be folded the moment it changes: folding a field mid-word would
// take it away from the responder still writing in it.
func needsDone(t questionnaire.Type) bool {
	switch t {
	case questionnaire.TypeText, questionnaire.TypeTextarea,
		questionnaire.TypeNumber, questionnaire.TypeMultiselect:
		return true
	}
	return false
}

// control builds the answer control for a question's declared type.
func (f *form) control(q *questionnaire.Question, c *card) fyne.CanvasObject {
	switch q.Type {
	case questionnaire.TypeSelect:
		return f.selectControl(q)
	case questionnaire.TypeMultiselect:
		return f.multiselectControl(q)
	case questionnaire.TypeText:
		return f.textControl(q)
	case questionnaire.TypeTextarea:
		return f.textareaControl(q)
	case questionnaire.TypeNumber:
		return f.numberControl(q)
	case questionnaire.TypeBoolean:
		return f.booleanControl(q)
	case questionnaire.TypeScale:
		return f.scaleControl(q)
	case questionnaire.TypeRank:
		return f.rankControl(q, c)
	}
	// Load rejects unknown types, so this is unreachable in practice.
	return widget.NewLabel(fmt.Sprintf("unsupported question type %q", q.Type))
}

func (f *form) selectControl(q *questionnaire.Question) fyne.CanvasObject {
	current, _ := q.Answer.(string)

	if len(q.Options) >= radioLimit {
		sel := widget.NewSelect(q.Options, func(v string) { f.set(q, v); f.settle(q) })
		sel.PlaceHolder = "Choose one"
		if current != "" {
			sel.SetSelected(current)
		}
		return alignInk(sel, buttonInk)
	}

	group := widget.NewRadioGroup(q.Options, func(v string) { f.set(q, v); f.settle(q) })
	group.Required = false // an optional question must be clearable
	if current != "" {
		group.SetSelected(current)
	}
	return alignInk(group, choiceInk)
}

func (f *form) multiselectControl(q *questionnaire.Question) fyne.CanvasObject {
	group := widget.NewCheckGroup(q.Options, func(v []string) { f.set(q, append([]string(nil), v...)) })
	if current, ok := q.Answer.([]string); ok {
		group.SetSelected(current)
	}
	return alignInk(group, choiceInk)
}

func (f *form) textControl(q *questionnaire.Question) fyne.CanvasObject {
	entry := newFormEntry(false, f.changeTextSize)
	if current, ok := q.Answer.(string); ok {
		entry.SetText(current)
	}
	entry.OnChanged = func(s string) { f.set(q, s) }
	// Enter says the same thing the Done action does, for a field where there is
	// only ever one line to finish.
	entry.OnSubmitted = func(string) { f.finished(q) }
	return f.scrollable(entry)
}

func (f *form) textareaControl(q *questionnaire.Question) fyne.CanvasObject {
	entry := newFormEntry(true, f.changeTextSize)
	entry.SetMinRowsVisible(4)
	entry.Wrapping = fyne.TextWrapWord
	if current, ok := q.Answer.(string); ok {
		entry.SetText(current)
	}
	entry.OnChanged = func(s string) { f.set(q, s) }
	return f.scrollable(entry)
}

// scrollable shields a text field so that scrolling the page over it keeps
// working, and sets it in far enough to line up with the prompt. See
// scrollThrough for the shielding, and alignInk for the alignment.
func (f *form) scrollable(field fyne.CanvasObject) fyne.CanvasObject {
	return alignInk(newScrollThrough(field, f.scrollPage), entryInk)
}

// scrollPage moves the questionnaire under the responder, matching what the
// page scroller would have done with the event itself.
func (f *form) scrollPage(e *fyne.ScrollEvent) {
	if f.scroll == nil {
		return
	}
	offset := f.scroll.Offset
	offset.Y -= e.Scrolled.DY
	offset.X -= e.Scrolled.DX

	limit := f.scroll.Content.Size().Height - f.scroll.Size().Height
	if limit < 0 {
		limit = 0
	}
	if offset.Y < 0 {
		offset.Y = 0
	}
	if offset.Y > limit {
		offset.Y = limit
	}
	if offset.X != 0 {
		offset.X = 0 // the page never scrolls sideways
	}

	f.scroll.Offset = offset
	f.scroll.Refresh()
}

// numberControl accepts free text but only records a value the question's
// bounds allow, so an out-of-range entry is refused with the bound shown rather
// than silently clamped.
func (f *form) numberControl(q *questionnaire.Question) fyne.CanvasObject {
	entry := newFormEntry(false, f.changeTextSize)
	entry.SetPlaceHolder(boundsHint(q))
	if current, ok := q.Answer.(float64); ok {
		entry.SetText(strconv.FormatFloat(current, 'f', -1, 64))
	}
	entry.Validator = func(s string) error {
		if strings.TrimSpace(s) == "" {
			return nil
		}
		_, err := parseBounded(q, s)
		return err
	}
	entry.OnChanged = func(s string) {
		if strings.TrimSpace(s) == "" {
			f.set(q, nil)
			return
		}
		v, err := parseBounded(q, s)
		if err != nil {
			return // the validator is already showing why
		}
		f.set(q, v)
	}
	entry.OnSubmitted = func(string) { f.finished(q) }
	return f.scrollable(entry)
}

func (f *form) booleanControl(q *questionnaire.Question) fyne.CanvasObject {
	// Two explicit choices rather than a single checkbox: an unticked box
	// cannot say the difference between "no" and "not answered".
	group := widget.NewRadioGroup([]string{"Yes", "No"}, func(v string) {
		switch v {
		case "Yes":
			f.set(q, true)
		case "No":
			f.set(q, false)
		default:
			f.set(q, nil)
		}
		f.settle(q)
	})
	group.Horizontal = true
	if current, ok := q.Answer.(bool); ok {
		if current {
			group.SetSelected("Yes")
		} else {
			group.SetSelected("No")
		}
	}
	return alignInk(group, choiceInk)
}

func (f *form) scaleControl(q *questionnaire.Question) fyne.CanvasObject {
	lo, hi := q.ScaleBounds()
	var current *int
	if v, ok := q.Answer.(int); ok {
		current = &v
	}

	// The buttons carry their own numbers, so no separate range labels: they
	// only pushed the two ends of the scale apart with empty space.
	return alignInk(newScaleWidget(lo, hi, current, func(v *int) {
		if v == nil {
			f.set(q, nil)
			return
		}
		f.set(q, *v)
		f.settle(q)
	}), buttonInk)
}

func (f *form) rankControl(q *questionnaire.Question, c *card) fyne.CanvasObject {
	order := q.Options
	if current, ok := q.Answer.([]string); ok && len(current) == len(q.Options) {
		order = current
	}

	c.rank = newRankWidget(order, func(v []string) { f.set(q, v) })
	confirm := widget.NewButton("Use this order", func() {
		c.rank.confirm()
		f.settle(q)
	})

	hint := captionLabel("Drag a row, or use the arrows, to reorder.")
	return container.NewVBox(c.rank, container.NewBorder(nil, nil, hint, confirm))
}

func captionLabel(text string) *widget.Label {
	l := widget.NewLabel(text)
	l.SizeName = theme.SizeNameCaptionText
	return l
}

func boundsHint(q *questionnaire.Question) string {
	switch {
	case q.Min != nil && q.Max != nil:
		return fmt.Sprintf("%s to %s", trimNumber(*q.Min), trimNumber(*q.Max))
	case q.Min != nil:
		return fmt.Sprintf("%s or more", trimNumber(*q.Min))
	case q.Max != nil:
		return fmt.Sprintf("%s or less", trimNumber(*q.Max))
	}
	return "A number"
}

func trimNumber(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

func parseBounded(q *questionnaire.Question, s string) (float64, error) {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, fmt.Errorf("enter a number")
	}
	if q.Min != nil && v < *q.Min {
		return 0, fmt.Errorf("must be %s or more", trimNumber(*q.Min))
	}
	if q.Max != nil && v > *q.Max {
		return 0, fmt.Errorf("must be %s or less", trimNumber(*q.Max))
	}
	return v, nil
}

// set records an answer and re-evaluates which questions apply.
func (f *form) set(q *questionnaire.Question, v any) {
	q.Answer = v
	if c, ok := f.cards[q.ID]; ok {
		c.header.Sync() // the tick and the folded answer follow the answer
		// The warning is about an answer that was missing, so it goes as soon
		// as one arrives rather than waiting for the next refused submit.
		if q.HasAnswer() {
			c.warning.Hide()
		}
	}
	f.syncVisibility()
}

// toggle folds or unfolds a question because the responder asked it to, and
// records that they asked: from here on their choice outranks the automatic one.
func (f *form) toggle(c *card) {
	if c.collapsed() {
		c.fold = foldOpen
	} else {
		c.fold = foldClosed
	}
	f.applyFold(c)
}

// settle folds a question whose answer is finished with.
//
// It is the automatic path, so it defers to a responder who has already decided
// how they want the question presented, and it leaves a question alone that has
// no answer -- clearing one is not finishing with it.
func (f *form) settle(q *questionnaire.Question) {
	if f.building {
		return
	}
	c, ok := f.cards[q.ID]
	if !ok || c.pinned() || !q.HasAnswer() {
		return
	}
	c.fold = foldClosed
	f.applyFold(c)
}

// finished is the responder saying they are done with a question -- the "Done"
// action, or Enter in a single-line field. It reads as the same intent as
// folding the question by hand, so it does the same thing.
func (f *form) finished(q *questionnaire.Question) {
	c, ok := f.cards[q.ID]
	if !ok {
		return
	}
	c.fold = foldClosed
	f.applyFold(c)
}

// expand forces a question open, whatever the responder had it set to, for the
// one case where they have to see it: a submit refused because of it. The fold
// returns to automatic, so answering it folds it away again.
func (f *form) expand(c *card) {
	if !c.collapsed() {
		return
	}
	c.fold = foldAuto
	f.applyFold(c)
}

// applyFold puts a card's presentation in step with its fold state.
//
// The body is hidden rather than discarded, so the control, the comment field
// and their bindings are the same objects folded or not.
func (f *form) applyFold(c *card) {
	show(c.body, !c.collapsed())
	c.header.SetCollapsed(c.collapsed())
	if c.card != nil {
		// Only a folded question with an answer reads as settled. Folded and
		// unanswered is the responder putting something aside, which is not
		// the same thing and must not look like it.
		c.card.SetSettled(c.collapsed() && c.question.HasAnswer())
		c.card.Refresh() // the card is a different height now
	}
	if f.list != nil {
		f.list.Refresh()
	}
}

// syncVisibility shows and hides cards to match the current answers.
//
// The cards are built once and then shown or hidden, rather than rebuilt: a
// question appearing further down the page must not steal focus from the
// control the responder is still typing in.
func (f *form) syncVisibility() {
	vis := f.doc.Evaluate()
	changed := false
	for id, c := range f.cards {
		want := vis.Visible(id)
		if want == c.root.Visible() {
			continue
		}
		changed = true
		if want {
			c.root.Show()
		} else {
			c.root.Hide()
			c.warning.Hide()
		}
	}
	if changed && f.list != nil {
		f.list.Refresh()
	}
	f.updateSummary()
}

func (f *form) updateSummary() {
	if f.status == nil || f.progress == nil {
		return
	}

	answered, total := f.counts()
	if total == 0 {
		f.progress.SetValue(0)
	} else {
		f.progress.SetValue(float64(answered) / float64(total))
	}

	f.status.SetText(f.statusText())
	if f.footerRow != nil {
		f.footerRow.Refresh() // longer text may no longer fit beside the actions
	}
}

// statusText says how far through the questionnaire is and, while anything
// required is outstanding, how much of that is left.
func (f *form) statusText() string {
	text := f.progressText()
	switch missing := len(f.doc.Unanswered()); missing {
	case 0:
	case 1:
		text += " · 1 required question left"
	default:
		text += fmt.Sprintf(" · %d required questions left", missing)
	}
	return text
}

// progressText labels the bar. A bare percentage is hard to act on; the counts
// say how many questions are actually left.
func (f *form) progressText() string {
	answered, total := f.counts()
	return fmt.Sprintf("%d of %d answered", answered, total)
}

// counts reports how many of the applicable questions carry an answer.
//
// It reads over the same visibility pass that drives show_if, so a question the
// responder cannot see is in neither figure. That means the total moves as
// conditions resolve, which is the honest reading of a conditional
// questionnaire: counting questions they will never reach would leave the form
// permanently short of complete with nothing to act on.
func (f *form) counts() (answered, total int) {
	for _, q := range f.doc.VisibleQuestions() {
		total++
		if q.HasAnswer() {
			answered++
		}
	}
	return answered, total
}

// submit finishes the questionnaire, refusing while a visible required question
// is still unanswered.
func (f *form) submit() {
	if missing := f.doc.Unanswered(); len(missing) > 0 {
		f.flagMissing(missing)
		return
	}
	f.finish(questionnaire.StatusSubmitted)
}

// flagMissing marks every outstanding question and moves to the first.
func (f *form) flagMissing(missing []*questionnaire.Question) {
	for _, c := range f.cards {
		c.warning.Hide()
	}
	for _, q := range missing {
		if c, ok := f.cards[q.ID]; ok {
			c.warning.SetText("This question needs an answer before you can submit.")
			c.warning.Show()
			// A warning inside a fold is no warning at all, so the questions
			// standing in the way of a submit are opened whatever the responder
			// had them set to.
			f.expand(c)
		}
	}
	f.list.Refresh()
	// After the refresh, not before: the cards just opened have moved
	// everything below them down the page.
	f.scrollTo(missing[0].ID)

	names := make([]string, 0, len(missing))
	for _, q := range missing {
		names = append(names, "• "+q.Prompt)
	}
	message := widget.NewLabel(strings.Join(names, "\n"))
	message.Wrapping = fyne.TextWrapWord

	if f.win == nil {
		return
	}
	dialog.ShowCustom("Some questions still need answers", "Back to the form",
		container.NewVBox(message), f.win)
}

// scrollTo brings a question into view.
func (f *form) scrollTo(id string) {
	c, ok := f.cards[id]
	if !ok || f.scroll == nil {
		return
	}
	f.scroll.Offset = fyne.NewPos(0, c.root.Position().Y)
	f.scroll.Refresh()
}

// requestClose handles both the Dismiss button and the window's close box. Work
// in progress is never thrown away without asking.
func (f *form) requestClose() {
	if !f.dirty() {
		f.finish(questionnaire.StatusDismissed)
		return
	}
	if f.win == nil {
		f.finish(questionnaire.StatusDismissed)
		return
	}

	message := widget.NewLabel("You have answers that have not been saved.")
	message.Wrapping = fyne.TextWrapWord

	var d dialog.Dialog
	save := widget.NewButton("Save", func() {
		d.Hide()
		f.finish(f.savedStatus())
	})
	save.Importance = widget.HighImportance
	discard := widget.NewButton("Discard changes", func() {
		d.Hide()
		f.finish(questionnaire.StatusDismissed)
	})
	discard.Importance = widget.DangerImportance
	keep := widget.NewButton("Keep editing", func() { d.Hide() })

	content := container.NewVBox(message, container.NewHBox(keep, discard, save))
	d = dialog.NewCustomWithoutButtons("Unsaved answers", content, f.win)
	d.Show()
}

// clearWarning is what the responder is asked before their answers are wiped.
const clearWarning = "This will wipe all answers and comments. Are you sure?"

// requestClear asks before clearing every answer and comment.
func (f *form) requestClear() {
	if f.win == nil {
		return
	}
	message := widget.NewLabel(clearWarning)
	message.Wrapping = fyne.TextWrapWord

	var d dialog.Dialog
	cancel := widget.NewButton("Cancel", func() { d.Hide() })
	confirm := widget.NewButton("Clear all", func() {
		d.Hide()
		f.clearAll()
	})
	confirm.Importance = widget.DangerImportance

	content := container.NewVBox(message, container.NewHBox(layout.NewSpacer(), cancel, confirm))
	d = dialog.NewCustomWithoutButtons("Clear all answers", content, f.win)
	d.Show()
}

// clearAll returns the questionnaire to unanswered: no answers, no comments,
// and every question presented as it would be on a fresh copy.
//
// The model is cleared and the page rebuilt from it, rather than each control
// being emptied where it stands. Every kind of control has its own idea of
// empty and its own way of reporting a change, and a page built from an
// unanswered model is already exactly right -- folds, visibility, progress, a
// ranking back in authored order and no longer touched. The baseline is left
// alone, so closing afterwards still asks whenever the file would change.
//
// Hidden questions are cleared too: an answer left behind a show_if would come
// back the moment its condition did.
func (f *form) clearAll() {
	for _, q := range f.doc.Questions {
		q.Answer = nil
		q.Comment = ""
	}
	f.cards = map[string]*card{}
	content := f.build()
	if f.win != nil {
		f.win.SetContent(content)
	}
}

// savedStatus reports the outcome for a save: a questionnaire with everything
// answered is submitted, otherwise it is a partial save.
func (f *form) savedStatus() questionnaire.Status {
	if f.doc.IsComplete() {
		return questionnaire.StatusSubmitted
	}
	return questionnaire.StatusSaved
}

// dirty reports whether the responder has changed anything since the form
// opened, which is what decides whether closing needs to ask.
func (f *form) dirty() bool {
	for _, q := range f.doc.Questions {
		was := f.baseline[q.ID]
		if answerKey(q.Answer) != was.answer || q.Comment != was.comment {
			return true
		}
		if c, ok := f.cards[q.ID]; ok && c.rank != nil && c.rank.Touched() {
			return true
		}
	}
	return false
}

func (f *form) finish(status questionnaire.Status) {
	f.outcome = status
	if f.win != nil {
		f.win.Close()
	}
}

// answerKey renders an answer as a comparable string, so two answers can be
// checked for equality without caring which concrete type they arrived as.
func answerKey(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case []string:
		return "[" + strings.Join(t, "\x00") + "]"
	case string:
		return "s" + t
	default:
		return fmt.Sprintf("%T:%v", v, v)
	}
}
