package ui

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

// Inter is compiled into the binary rather than taken from the system, so a
// questionnaire looks the same on every machine it is answered on. Fyne bundles
// Noto Sans by default, which is a fine reading face but a loose fit for a
// form; Inter was drawn for interfaces and has the tighter metrics and clearer
// digits that a page of prompts and numeric answers wants.
//
// Inter is used under the SIL Open Font License; see fonts/LICENSE-Inter.txt.

//go:embed fonts/Inter-Regular.ttf
var interRegular []byte

//go:embed fonts/Inter-Bold.ttf
var interBold []byte

//go:embed fonts/Inter-Italic.ttf
var interItalic []byte

//go:embed fonts/Inter-BoldItalic.ttf
var interBoldItalic []byte

var (
	fontRegular    = &fyne.StaticResource{StaticName: "Inter-Regular.ttf", StaticContent: interRegular}
	fontBold       = &fyne.StaticResource{StaticName: "Inter-Bold.ttf", StaticContent: interBold}
	fontItalic     = &fyne.StaticResource{StaticName: "Inter-Italic.ttf", StaticContent: interItalic}
	fontBoldItalic = &fyne.StaticResource{StaticName: "Inter-BoldItalic.ttf", StaticContent: interBoldItalic}
)

// Font returns the face for a text style. Monospace and symbol faces fall
// through to the toolkit's own, which Inter does not provide.
func (t formTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Monospace || style.Symbol {
		return t.Theme.Font(style)
	}
	switch {
	case style.Bold && style.Italic:
		return fontBoldItalic
	case style.Bold:
		return fontBold
	case style.Italic:
		return fontItalic
	}
	return fontRegular
}
