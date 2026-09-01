package ui

import (
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/markwylde/interrogate/internal/questionnaire"
)

// TestRenderPreview draws the form headlessly and writes it to testdata, so the
// layout can be reviewed without a desktop. Run with -run RenderPreview to
// refresh the images.
func TestRenderPreview(t *testing.T) {
	if os.Getenv("INTERROGATE_RENDER") == "" {
		t.Skip("set INTERROGATE_RENDER=1 to refresh the preview images")
	}

	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "demo.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	doc, err := questionnaire.LoadBytes("demo.yaml", raw)
	if err != nil {
		t.Fatal(err)
	}

	for _, variant := range []struct {
		name  string
		theme fyne.ThemeVariant
	}{{"light", 1}, {"dark", 0}} {
		t.Run(variant.name, func(t *testing.T) {
			doc, err := questionnaire.LoadBytes("demo.yaml", raw)
			if err != nil {
				t.Fatal(err)
			}
			test.ApplyTheme(t, &fixedVariantTheme{Theme: newTheme(), variant: variant.theme})

			f := newForm(doc, nil)
			win := test.NewWindow(f.build())
			defer win.Close()
			win.Resize(fyne.NewSize(780, 900))

			img := win.Canvas().Capture()
			out, err := os.Create(filepath.Join("testdata", "preview-"+variant.name+".png"))
			if err != nil {
				t.Fatal(err)
			}
			defer out.Close()
			if err := png.Encode(out, img); err != nil {
				t.Fatal(err)
			}
			t.Logf("wrote testdata/preview-%s.png (%v)", variant.name, img.Bounds().Size())
		})
	}
	_ = doc
}

// fixedVariantTheme pins the light or dark variant so both can be rendered in
// one run.
type fixedVariantTheme struct {
	fyne.Theme
	variant fyne.ThemeVariant
}

func (t *fixedVariantTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return t.Theme.Color(name, t.variant)
}

// QuestionCard pins the variant for the cards too, so a preview renders the
// card colour of the variant it is showing rather than the app's current one.
func (t *fixedVariantTheme) QuestionCard(_ fyne.ThemeVariant) color.Color {
	if provider, ok := t.Theme.(cardProvider); ok {
		return provider.QuestionCard(t.variant)
	}
	return color.Transparent
}

// controlsDoc shows one question of every type, for reviewing the controls
// themselves rather than the page.
const controlsDoc = `
title: Every control
intro: One question of each type, for review.
questions:
  - id: sel
    type: select
    prompt: A single choice
    required: true
    options: [in-place, sidecar, both]
  - id: multi
    type: multiselect
    prompt: Any number of choices
    options: [yaml, json, toml]
  - id: text
    type: text
    prompt: A single line
    answer: interrogate
  - id: num
    type: number
    prompt: A number, between bounds
    min: 1
    max: 30
    answer: 12
  - id: sc
    type: scale
    prompt: A point on a scale
    min: 1
    max: 5
    answer: 4
  - id: rk
    type: rank
    prompt: An ordering
    options: [correctness, speed, looks]
  - id: area
    type: textarea
    prompt: Several lines
    answer: |-
      Answers and questions in one file.

      Two files drift apart the moment the questions change.
`

func TestRenderControlsPreview(t *testing.T) {
	if os.Getenv("INTERROGATE_RENDER") == "" {
		t.Skip("set INTERROGATE_RENDER=1 to refresh the preview images")
	}
	doc, err := questionnaire.LoadBytes("controls.yaml", []byte(controlsDoc))
	if err != nil {
		t.Fatal(err)
	}
	test.ApplyTheme(t, &fixedVariantTheme{Theme: newTheme(), variant: 1})

	f := newForm(doc, nil)
	win := test.NewWindow(f.build())
	defer win.Close()
	win.Resize(fyne.NewSize(780, 1500))

	out, err := os.Create(filepath.Join("testdata", "preview-controls.png"))
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	if err := png.Encode(out, win.Canvas().Capture()); err != nil {
		t.Fatal(err)
	}
	t.Log("wrote testdata/preview-controls.png")
}
