package ui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/markwylde/interrogate/internal/questionnaire"
)

// radioLimit is the point past which a list of options stops being worth
// showing in full and becomes a dropdown.
const radioLimit = 7

// Run presents the questionnaire and blocks until the responder submits, saves,
// or dismisses it, returning the status that outcome corresponds to.
func Run(doc *questionnaire.Document) (questionnaire.Status, error) {
	a := app.NewWithID("com.markwylde.interrogate")
	a.Settings().SetTheme(newTheme())

	title := doc.Title
	if title == "" {
		title = "Questionnaire"
	}
	win := a.NewWindow(title)

	f := newForm(doc, win)
	win.SetContent(f.build())
	win.Resize(fyne.NewSize(780, 760))
	win.CenterOnScreen()
	win.SetCloseIntercept(f.requestClose)
	win.ShowAndRun()

	return f.outcome, nil
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

	cards    map[string]*card
	baseline map[string]snapshot
	list     *fyne.Container
	scroll   *container.Scroll
	summary  *widget.Label
	progress *progressBar
	counted  *widget.Label
}

// card is one question: its prompt, its control, its comment box, and the
// warning shown when it is required and unanswered.
type card struct {
	question *questionnaire.Question
	root     *fyne.Container
	warning  *widget.Label
	rank     *rankWidget
	comment  *commentField
	card     *cardBox
}

func newForm(doc *questionnaire.Document, win fyne.Window) *form {
	f := &form{
		doc:      doc,
		win:      win,
		cards:    map[string]*card{},
		baseline: map[string]snapshot{},
	}
	for _, q := range doc.Questions {
		f.baseline[q.ID] = snapshot{answer: answerKey(q.Answer), comment: q.Comment}
	}
	return f
}

// build assembles the whole window: a fixed header, one scrolling page of
// questions, and a fixed footer holding the actions.
func (f *form) build() fyne.CanvasObject {
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

	items := []fyne.CanvasObject{title}
	if intro := strings.TrimSpace(f.doc.Intro); intro != "" {
		body := widget.NewLabel(intro)
		body.Wrapping = fyne.TextWrapWord
		items = append(items, body)
	}
	items = append(items, widget.NewSeparator())
	return container.NewVBox(items...)
}

func (f *form) footer() fyne.CanvasObject {
	// Progress and outstanding-required are different questions -- "how far
	// through am I" and "what is stopping me submitting" -- so a bar answers
	// the first at a glance and a label answers the second in words.
	f.progress = newProgressBar()
	f.counted = captionLabel("")
	f.summary = widget.NewLabel("")
	f.summary.Importance = widget.MediumImportance
	f.summary.Wrapping = fyne.TextWrapWord

	submit := widget.NewButton("Submit", f.submit)
	submit.Importance = widget.HighImportance
	dismiss := widget.NewButton("Dismiss", f.requestClose)

	status := container.NewHBox(f.counted, f.summary)
	actions := container.NewHBox(dismiss, submit)

	// The bar spans the window along the top edge of the footer, where it also
	// does the job the separator was doing.
	return container.NewVBox(
		f.progress,
		container.NewPadded(container.NewBorder(nil, nil, status, actions)),
	)
}

// buildCard lays out one question. Every question gets a comment box, whatever
// its type: the note is often the most precise thing the responder has to say.
func (f *form) buildCard(q *questionnaire.Question) *card {
	c := &card{question: q}

	prompt := widget.NewLabel(q.Prompt)
	prompt.TextStyle = fyne.TextStyle{Bold: true}
	prompt.Wrapping = fyne.TextWrapWord

	// The required marker sits on the prompt's own line. On its own row it read
	// as another instruction to take in; beside the prompt it is just a label.
	var heading fyne.CanvasObject = prompt
	if q.Required {
		marker := captionLabel("Required")
		marker.Importance = widget.MediumImportance
		heading = container.NewBorder(nil, nil, nil, marker, prompt)
	}

	parts := []fyne.CanvasObject{heading}
	if help := strings.TrimSpace(q.Help); help != "" {
		body := widget.NewLabel(help)
		body.Wrapping = fyne.TextWrapWord
		body.SizeName = theme.SizeNameCaptionText
		parts = append(parts, body)
	}

	parts = append(parts, f.control(q, c))

	c.comment = newCommentField(q.Comment, func(s string) { q.Comment = s }, f.scrollable)
	parts = append(parts, c.comment.root)

	c.warning = widget.NewLabel("")
	c.warning.Importance = widget.DangerImportance
	c.warning.Hide()
	parts = append(parts, c.warning)

	// The card and the gap around it do the separating a hairline rule was
	// failing to do, so the separator goes: a card edge and a rule together
	// only look fussy.
	c.card = newCardBox(container.NewPadded(container.NewVBox(parts...)))
	c.root = container.NewPadded(c.card)
	return c
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
		sel := widget.NewSelect(q.Options, func(v string) { f.set(q, v) })
		sel.PlaceHolder = "Choose one"
		if current != "" {
			sel.SetSelected(current)
		}
		return sel
	}

	group := widget.NewRadioGroup(q.Options, func(v string) { f.set(q, v) })
	group.Required = false // an optional question must be clearable
	if current != "" {
		group.SetSelected(current)
	}
	return group
}

func (f *form) multiselectControl(q *questionnaire.Question) fyne.CanvasObject {
	group := widget.NewCheckGroup(q.Options, func(v []string) { f.set(q, append([]string(nil), v...)) })
	if current, ok := q.Answer.([]string); ok {
		group.SetSelected(current)
	}
	return group
}

func (f *form) textControl(q *questionnaire.Question) fyne.CanvasObject {
	entry := widget.NewEntry()
	if current, ok := q.Answer.(string); ok {
		entry.SetText(current)
	}
	entry.OnChanged = func(s string) { f.set(q, s) }
	return entry
}

func (f *form) textareaControl(q *questionnaire.Question) fyne.CanvasObject {
	entry := widget.NewMultiLineEntry()
	entry.SetMinRowsVisible(4)
	entry.Wrapping = fyne.TextWrapWord
	if current, ok := q.Answer.(string); ok {
		entry.SetText(current)
	}
	entry.OnChanged = func(s string) { f.set(q, s) }
	return f.scrollable(entry)
}

// scrollable shields a multi-line field so that scrolling the page over it
// keeps working. See scrollThrough for why this is needed.
func (f *form) scrollable(field fyne.CanvasObject) fyne.CanvasObject {
	return newScrollThrough(field, f.scrollPage)
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
	entry := widget.NewEntry()
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
	return entry
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
	})
	group.Horizontal = true
	if current, ok := q.Answer.(bool); ok {
		if current {
			group.SetSelected("Yes")
		} else {
			group.SetSelected("No")
		}
	}
	return group
}

func (f *form) scaleControl(q *questionnaire.Question) fyne.CanvasObject {
	lo, hi := q.ScaleBounds()
	var current *int
	if v, ok := q.Answer.(int); ok {
		current = &v
	}

	// The buttons carry their own numbers, so no separate range labels: they
	// only pushed the two ends of the scale apart with empty space.
	return newScaleWidget(lo, hi, current, func(v *int) {
		if v == nil {
			f.set(q, nil)
			return
		}
		f.set(q, *v)
	})
}

func (f *form) rankControl(q *questionnaire.Question, c *card) fyne.CanvasObject {
	order := q.Options
	if current, ok := q.Answer.([]string); ok && len(current) == len(q.Options) {
		order = current
	}

	c.rank = newRankWidget(order, func(v []string) { f.set(q, v) })
	confirm := widget.NewButton("Use this order", func() { c.rank.confirm() })

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
	f.syncVisibility()
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
	if f.summary == nil || f.progress == nil || f.counted == nil {
		return
	}
	f.counted.SetText(f.progressText())

	answered, total := f.counts()
	if total == 0 {
		f.progress.SetValue(0)
	} else {
		f.progress.SetValue(float64(answered) / float64(total))
	}

	missing := len(f.doc.Unanswered())
	switch missing {
	case 0:
		f.summary.SetText("")
	case 1:
		f.summary.SetText("· 1 required question left")
	default:
		f.summary.SetText(fmt.Sprintf("· %d required questions left", missing))
	}
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
		}
	}
	f.list.Refresh()
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
