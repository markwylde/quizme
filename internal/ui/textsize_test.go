package ui

import (
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/markwylde/quizme/internal/questionnaire"
)

func TestTheThemeScalesEverySize(t *testing.T) {
	base, big := newScaledTheme(100), newScaledTheme(200)
	if got := base.Size(theme.SizeNameText); got != 14 {
		t.Errorf("body text at 100%% is %v, want 14", got)
	}
	if got := big.Size(theme.SizeNameText); got != 28 {
		t.Errorf("body text at 200%% is %v, want 28", got)
	}
	for _, name := range []fyne.ThemeSizeName{
		theme.SizeNamePadding, theme.SizeNameInnerPadding, theme.SizeNameHeadingText,
		theme.SizeNameInlineIcon, // one the form leaves to the default theme
	} {
		if b, g := base.Size(name), big.Size(name); g != 2*b {
			t.Errorf("%s is %v at 100%% and %v at 200%%, want double", name, b, g)
		}
	}
}

func TestScaledFollowsTheThemeInForce(t *testing.T) {
	restore := useTheme(t, newScaledTheme(150))
	defer restore()
	if got := scaled(10); got != 15 {
		t.Errorf("scaled(10) at 150%% = %v, want 15", got)
	}
}

// The form is built once and then rescaled in place, so these build at the
// default size and change it afterwards: a layout that fixed a number while the
// page was being built is exactly what they are here to catch.
func TestControlsStillLineUpAfterScaling(t *testing.T) {
	restore := useTheme(t, newTheme())
	defer restore()

	f, doc := build(t, alignmentDoc)
	win := test.NewWindow(f.build())
	defer win.Close()
	fyne.CurrentApp().Settings().SetTheme(newScaledTheme(150))
	win.Resize(fyne.NewSize(780, 3600))

	for _, q := range doc.Questions {
		c := f.cards[q.ID]
		prompt, ok := inkOf(c.header, c.header.prompt)
		if !ok {
			t.Fatalf("%s: the prompt drew nothing", q.ID)
		}
		control, ok := answerControl(t, f, q)
		if !ok {
			t.Fatalf("%s (%s): no control found", q.ID, q.Type)
		}
		at, ok := inkOf(c.body, control)
		if !ok {
			t.Fatalf("%s (%s): the control drew nothing", q.ID, q.Type)
		}
		if diff := at - prompt; diff > 2 || diff < -2 {
			t.Errorf("%s (%s): at 150%% the %T draws %v from the prompt's text", q.ID, q.Type, control, diff)
		}

		_, promptEnds, ok := inkSpanOf(c.root, c.header.prompt)
		if !ok {
			t.Fatalf("%s: the prompt drew nothing", q.ID)
		}
		controlStarts, _, ok := inkSpanOf(c.root, control)
		if !ok {
			t.Fatalf("%s: the control drew nothing", q.ID)
		}
		want := scaled(promptTail)
		if gap := controlStarts - promptEnds; gap < want-1.5 || gap > want+1.5 {
			t.Errorf("%s (%s): at 150%% %v between the prompt and its control, want %v", q.ID, q.Type, gap, want)
		}
	}

	if got := f.cards["line"].header.prompt.lineHeight(); got < 1.4*fyne.MeasureText("Ag", 14, promptStyle).Height {
		t.Errorf("the prompt's line is %v tall at 150%%: the text did not grow", got)
	}
}

// The demo questionnaire exercises every control; at a large size it still has
// to lay out and draw.
func TestTheDemoRendersAtALargeSize(t *testing.T) {
	restore := useTheme(t, newTheme())
	defer restore()

	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "demo.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	doc, err := questionnaire.LoadBytes("demo.yaml", raw)
	if err != nil {
		t.Fatal(err)
	}
	f := newForm(doc, nil)
	win := test.NewWindow(f.build())
	defer win.Close()
	fyne.CurrentApp().Settings().SetTheme(newScaledTheme(150))
	win.Resize(fyne.NewSize(780, 900))

	if img := win.Canvas().Capture(); img.Bounds().Empty() {
		t.Error("nothing was drawn at 150%")
	}
	if got, min := f.scroll.Content.MinSize().Height, float32(900); got < min {
		t.Errorf("the demo is only %v tall at 150%%", got)
	}
}

const sizeDoc = `
title: Sizes
questions:
  - id: name
    type: text
    prompt: Name?
  - id: pick
    type: select
    prompt: Pick one
    options: [a, b]
`

// sizeForm builds a form in a window with the size shortcuts registered, and
// puts the app's theme back afterwards.
func sizeForm(t *testing.T) (*form, fyne.Window, *[]int) {
	f, win, heard, _ := sizeFormWithKeys(t)
	return f, win, heard
}

// sizeFormWithKeys also returns the handler the window shortcuts went to. The
// test canvas cannot be typed at, so the shortcuts are registered on a handler
// of the same kind a real canvas uses.
func sizeFormWithKeys(t *testing.T) (*form, fyne.Window, *[]int, *fyne.ShortcutHandler) {
	t.Helper()
	restore := useTheme(t, newTheme())
	t.Cleanup(restore)

	f, _ := build(t, sizeDoc)
	var heard []int
	f.onTextSize = func(p int) { heard = append(heard, p) }
	win := test.NewWindow(f.build())
	t.Cleanup(win.Close)
	keys := &fyne.ShortcutHandler{}
	f.registerSizeShortcuts(keys)
	return f, win, &heard, keys
}

func shortcut(key fyne.KeyName) *desktop.CustomShortcut {
	return &desktop.CustomShortcut{KeyName: key, Modifier: fyne.KeyModifierShortcutDefault}
}

func TestTheSizeButtonsStepAndStopAtTheLimits(t *testing.T) {
	f, _, heard := sizeForm(t)

	test.Tap(f.larger)
	if f.textSize != 110 || theme.Size(theme.SizeNameText) != 14*1.1 {
		t.Errorf("after A+ the size is %d (text %v), want 110", f.textSize, theme.Size(theme.SizeNameText))
	}
	test.Tap(f.smaller)
	test.Tap(f.smaller)
	if f.textSize != 90 {
		t.Errorf("after A− twice the size is %d, want 90", f.textSize)
	}

	for f.textSize < 200 {
		test.Tap(f.larger)
	}
	if !f.larger.Disabled() || f.smaller.Disabled() {
		t.Errorf("at 200%% want only A+ disabled, got A+ %v A− %v", f.larger.Disabled(), f.smaller.Disabled())
	}
	f.changeTextSize(sizeUp)
	if f.textSize != 200 {
		t.Errorf("increasing past the limit gave %d", f.textSize)
	}

	for f.textSize > 70 {
		test.Tap(f.smaller)
	}
	if !f.smaller.Disabled() || f.larger.Disabled() {
		t.Errorf("at 70%% want only A− disabled, got A− %v A+ %v", f.smaller.Disabled(), f.larger.Disabled())
	}
	f.changeTextSize(sizeDown)
	if f.textSize != 70 {
		t.Errorf("decreasing past the limit gave %d", f.textSize)
	}

	if (*heard)[0] != 110 || (*heard)[len(*heard)-1] != 70 {
		t.Errorf("the callback heard %v", *heard)
	}
	for i := 1; i < len(*heard); i++ {
		if (*heard)[i] == (*heard)[i-1] {
			t.Errorf("the callback heard %d twice in a row: a change that changed nothing was reported", (*heard)[i])
		}
	}
}

func TestTheSizeShortcuts(t *testing.T) {
	f, _, heard, c := sizeFormWithKeys(t)

	c.TypedShortcut(shortcut(fyne.KeyEqual))
	if f.textSize != 110 {
		t.Errorf("after the increase shortcut the size is %d", f.textSize)
	}
	c.TypedShortcut(&desktop.CustomShortcut{
		KeyName: fyne.KeyEqual, Modifier: fyne.KeyModifierShortcutDefault | fyne.KeyModifierShift})
	if f.textSize != 120 {
		t.Errorf("after the shifted (+) shortcut the size is %d", f.textSize)
	}
	c.TypedShortcut(shortcut(fyne.KeyMinus))
	if f.textSize != 110 {
		t.Errorf("after the decrease shortcut the size is %d", f.textSize)
	}
	c.TypedShortcut(shortcut(fyne.Key0))
	if f.textSize != 100 {
		t.Errorf("after the reset shortcut the size is %d", f.textSize)
	}
	if got := *heard; len(got) != 4 || got[3] != 100 {
		t.Errorf("the callback heard %v", got)
	}
}

func TestTheSizeShortcutsWorkWhileTyping(t *testing.T) {
	f, win, _ := sizeForm(t)
	entry := find[*formEntry](t, f, "name")
	win.Canvas().Focus(entry)
	test.Type(entry, "ab")

	// A focused field is handed the shortcut, not the window.
	entry.TypedShortcut(shortcut(fyne.KeyEqual))
	if f.textSize != 110 {
		t.Errorf("the increase shortcut in a field left the size at %d", f.textSize)
	}
	if entry.Text != "ab" {
		t.Errorf("the field reads %q, want the shortcut to type nothing", entry.Text)
	}

	// Its own shortcuts still reach it.
	entry.TypedShortcut(&fyne.ShortcutSelectAll{})
	if entry.SelectedText() != "ab" {
		t.Errorf("select all in the field selected %q", entry.SelectedText())
	}
}

func TestChangingSizeKeepsTheResponderWork(t *testing.T) {
	f, win, _ := sizeForm(t)
	find[*formEntry](t, f, "name").SetText("quizme")
	f.cards["name"].comment.Entry().SetText("a note")
	find[*widget.RadioGroup](t, f, "pick").SetSelected("b")
	f.toggle(f.cards["name"])
	wasDirty := f.dirty()
	fold := f.cards["name"].fold
	name := find[*formEntry](t, f, "name")

	f.setTextSize(150)
	win.Resize(fyne.NewSize(780, 900))

	if got := find[*formEntry](t, f, "name"); got != name || got.Text != "quizme" {
		t.Error("the answer field was replaced or lost its text")
	}
	if f.doc.Question("name").Comment != "a note" || f.doc.Question("pick").Answer != "b" {
		t.Error("an answer or comment changed with the size")
	}
	if f.cards["name"].fold != fold {
		t.Error("a fold changed with the size")
	}
	if f.dirty() != wasDirty {
		t.Error("changing the size changed whether there is unsaved work")
	}
}

func TestOnlyASizeChangeClosesWithoutAsking(t *testing.T) {
	f, _, _ := sizeForm(t)
	f.setTextSize(130)
	if f.dirty() {
		t.Fatal("a size change alone made the form dirty")
	}
	f.requestClose()
	if f.outcome != questionnaire.StatusDismissed {
		t.Errorf("outcome = %q, want dismissed", f.outcome)
	}
}
