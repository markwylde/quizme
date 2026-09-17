package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// promptLeading is the space between the wrapped lines of a prompt, on top of
// the line height the face itself sets.
//
// A prompt is usually a sentence or two, so most of them wrap, and the toolkit's
// own leading sets those lines close enough to read as one crowded block.
const promptLeading = 5

// promptTail is the space under a prompt's last line, inside its own box. A
// label pads itself equally on every side; a prompt is followed by the answer
// it asks for, and that reads better closer than a label's own padding sets it.
//
// It is the one number behind both gaps a reader sees: under a folded question
// the answer sits a tail plus the answer's own padding away, less the overlap
// the stack takes back, and under an open one the controls sit a tail below the
// prompt's box. Change it and both move together.
const promptTail = 5

// promptText is a question's prompt: bold, wrapping, and laid out a line at a
// time so the leading between those lines can be set.
//
// A widget.Label would be less code, and was what this replaced, but the space
// between the rows of a wrapped label is unreachable: Fyne applies its line
// spacing between rich-text segments and explicitly not between the rows inside
// one. The only way to choose it is to do the wrapping here.
//
// The box it occupies is the same as the label's was -- the text inset by the
// inner padding on every side -- so everything that lines itself up against a
// prompt keeps the measurements it already had.
type promptText struct {
	widget.BaseWidget

	text string

	lines []*canvas.Text
	box   *fyne.Container

	// wrapped is the width the lines were last measured against, and height
	// what they came to. MinSize has no width to work from, so it answers from
	// the last layout: the same two-pass arrangement Fyne's own wrapping text
	// relies on.
	wrapped float32
	height  float32
}

func newPromptText(text string) *promptText {
	p := &promptText{text: text}
	p.ExtendBaseWidget(p)
	p.box = container.NewWithoutLayout()
	p.height = p.inset() + p.lineHeight() + p.tail()
	return p
}

// Text is what the prompt says.
func (p *promptText) Text() string { return p.text }

// Lines are the wrapped lines as they were last laid out.
func (p *promptText) Lines() []*canvas.Text { return p.lines }

func (p *promptText) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(p.box)
}

// Content exposes the lines, which otherwise hang off the renderer.
func (p *promptText) Content() fyne.CanvasObject { return p.box }

func (p *promptText) MinSize() fyne.Size {
	p.ExtendBaseWidget(p)
	// Wide enough for the longest single word, so the prompt can be narrowed
	// to any width that can still hold its text, and no narrower.
	width := float32(0)
	for _, word := range strings.Fields(p.text) {
		if w := p.measure(word).Width; w > width {
			width = w
		}
	}
	return fyne.NewSize(width+2*p.inset(), p.height)
}

func (p *promptText) Resize(size fyne.Size) {
	p.BaseWidget.Resize(size)
	p.reflow(size.Width)
}

// Refresh re-resolves the colour of every line, so a form open across a
// light/dark switch redraws in the right one, and re-wraps in case the text
// size changed with the theme.
func (p *promptText) Refresh() {
	p.wrapped = 0 // measured against the old theme
	p.reflow(p.Size().Width)
	for _, line := range p.lines {
		line.Color = theme.Color(theme.ColorNameForeground)
		line.TextSize = theme.Size(theme.SizeNameText)
		line.Refresh()
	}
	p.BaseWidget.Refresh()
}

// reflow wraps the text to the given width and stacks the lines.
func (p *promptText) reflow(width float32) {
	if width <= 0 || width == p.wrapped {
		return
	}
	p.wrapped = width

	wrapped := wrapWords(p.text, width-2*p.inset(), p.measure)
	if len(wrapped) != len(p.lines) {
		p.lines = nil
		p.box.RemoveAll()
		for range wrapped {
			line := canvas.NewText("", theme.Color(theme.ColorNameForeground))
			line.TextStyle = promptStyle
			line.TextSize = theme.Size(theme.SizeNameText)
			p.lines = append(p.lines, line)
			p.box.Add(line)
		}
	}

	step := p.lineHeight() + scaled(promptLeading)
	for i, line := range p.lines {
		line.Text = wrapped[i]
		line.Resize(fyne.NewSize(width-2*p.inset(), p.lineHeight()))
		line.Move(fyne.NewPos(p.inset(), p.inset()+float32(i)*step))
		line.Refresh()
	}

	height := p.inset() + float32(len(wrapped))*p.lineHeight() +
		float32(len(wrapped)-1)*scaled(promptLeading) + p.tail()
	if height != p.height {
		p.height = height
		// The prompt is taller or shorter than whatever asked for it thought,
		// so the card around it has to be measured again.
		p.box.Refresh()
	}
}

// promptStyle is the face a prompt is set in.
var promptStyle = fyne.TextStyle{Bold: true}

func (p *promptText) measure(text string) fyne.Size {
	return fyne.MeasureText(text, theme.Size(theme.SizeNameText), promptStyle)
}

// lineHeight is the height of one line of the prompt's own face.
func (p *promptText) lineHeight() float32 { return p.measure("Ag").Height }

// inset matches what a label puts between its text and its edge, above it and
// to the left. Everything that lines itself up against a prompt measures from
// there, so that much of the label's box is kept exactly.
func (p *promptText) inset() float32 { return theme.Size(theme.SizeNameInnerPadding) }

// tail is the space under the last line. See promptTail.
func (p *promptText) tail() float32 { return scaled(promptTail) }

// wrapWords breaks text into lines that fit the given width, keeping any line
// breaks the author wrote themselves.
//
// Greedy, word by word: a paragraph of prose has no need of anything cleverer,
// and a word too long for the width goes on a line of its own rather than being
// cut in half.
func wrapWords(text string, width float32, measure func(string) fyne.Size) []string {
	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		line := words[0]
		for _, word := range words[1:] {
			candidate := line + " " + word
			if measure(candidate).Width <= width {
				line = candidate
				continue
			}
			lines = append(lines, line)
			line = word
		}
		lines = append(lines, line)
	}
	return lines
}
