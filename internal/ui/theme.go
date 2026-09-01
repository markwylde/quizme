package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// The default Fyne look is a generic Material palette. A questionnaire is a
// document someone reads carefully and then commits to, so this theme aims for
// something closer to a considered form than a toolkit demo: a warm off-white
// page, quiet separators, one confident accent for the action that matters, and
// noticeably more breathing room than the defaults give.

type formTheme struct{ fyne.Theme }

func newTheme() fyne.Theme { return formTheme{Theme: theme.DefaultTheme()} }

// palette is one complete set of colours; the two variants below differ only in
// their values, never in which roles they define.
type palette struct {
	background  color.Color
	surface     color.Color
	foreground  color.Color
	muted       color.Color
	primary     color.Color
	onPrimary   color.Color
	separator   color.Color
	inputBorder color.Color
	disabled    color.Color
	disabledBg  color.Color
	hover       color.Color
	pressed     color.Color
	selection   color.Color
	focus       color.Color
	errorCol    color.Color
	success     color.Color
	warning     color.Color
	shadow      color.Color
	overlay     color.Color
	scrollBar   color.Color

	// questionTints cycle behind consecutive questions. A hairline rule was not
	// enough to tell one question from the next while scrolling; a band of
	// colour is. They are deliberately within a few points of the background --
	// enough to read as a change, not enough to compete with the content.
	questionTints []color.Color
}

var lightPalette = palette{
	background:  hex(0xFBFAF8),
	surface:     hex(0xFFFFFF),
	foreground:  hex(0x1C1B19),
	muted:       hex(0x8B877F),
	primary:     hex(0x3B5BDB),
	onPrimary:   hex(0xFFFFFF),
	separator:   hex(0xE6E2DA),
	inputBorder: hex(0xD8D3C9),
	disabled:    hex(0xB5B0A6),
	disabledBg:  hex(0xF1EFEA),
	hover:       rgba(0x1C, 0x1B, 0x19, 0x0E),
	pressed:     rgba(0x1C, 0x1B, 0x19, 0x1A),
	selection:   rgba(0x3B, 0x5B, 0xDB, 0x33),
	focus:       rgba(0x3B, 0x5B, 0xDB, 0x66),
	errorCol:    hex(0xC0392B),
	success:     hex(0x2B8A3E),
	warning:     hex(0xB8860B),
	shadow:      rgba(0x00, 0x00, 0x00, 0x22),
	overlay:     hex(0xFFFFFF),
	scrollBar:   rgba(0x1C, 0x1B, 0x19, 0x40),
	questionTints: []color.Color{
		hex(0xF4F6FC), // blue
		hex(0xF1F8F2), // green
		hex(0xFCF7ED), // amber
		hex(0xFCF3F4), // rose
		hex(0xF7F4FC), // violet
		hex(0xEFF8F8), // teal
	},
}

var darkPalette = palette{
	background:  hex(0x17181A),
	surface:     hex(0x1F2124),
	foreground:  hex(0xE9E7E3),
	muted:       hex(0x8E9299),
	primary:     hex(0x7C93FF),
	onPrimary:   hex(0x0E1020),
	separator:   hex(0x2C2F34),
	inputBorder: hex(0x3A3E44),
	disabled:    hex(0x5C6068),
	disabledBg:  hex(0x232529),
	hover:       rgba(0xFF, 0xFF, 0xFF, 0x12),
	pressed:     rgba(0xFF, 0xFF, 0xFF, 0x1F),
	selection:   rgba(0x7C, 0x93, 0xFF, 0x40),
	focus:       rgba(0x7C, 0x93, 0xFF, 0x77),
	errorCol:    hex(0xFF6B6B),
	success:     hex(0x51CF66),
	warning:     hex(0xFFD43B),
	shadow:      rgba(0x00, 0x00, 0x00, 0x66),
	overlay:     hex(0x1F2124),
	scrollBar:   rgba(0xFF, 0xFF, 0xFF, 0x40),
	questionTints: []color.Color{
		hex(0x191C24), // blue
		hex(0x171D19), // green
		hex(0x201C15), // amber
		hex(0x201819), // rose
		hex(0x1B1823), // violet
		hex(0x141D1E), // teal
	},
}

func (t formTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	p := lightPalette
	if variant == theme.VariantDark {
		p = darkPalette
	}

	switch name {
	case theme.ColorNameBackground:
		return p.background
	case theme.ColorNameForeground:
		return p.foreground
	case theme.ColorNameForegroundOnPrimary:
		return p.onPrimary
	case theme.ColorNamePrimary, theme.ColorNameHyperlink:
		return p.primary
	case theme.ColorNameButton, theme.ColorNameInputBackground, theme.ColorNameMenuBackground:
		return p.surface
	case theme.ColorNameOverlayBackground, theme.ColorNameHeaderBackground:
		return p.overlay
	case theme.ColorNamePlaceHolder:
		return p.muted
	case theme.ColorNameSeparator:
		return p.separator
	case theme.ColorNameInputBorder:
		return p.inputBorder
	case theme.ColorNameDisabled:
		// Disabled text and icons.
		return p.disabled
	case theme.ColorNameDisabledButton:
		// The button's fill, not its label: sharing one value with the above
		// paints a solid block where a faded icon should be.
		return p.disabledBg
	case theme.ColorNameHover:
		return p.hover
	case theme.ColorNamePressed:
		return p.pressed
	case theme.ColorNameSelection:
		return p.selection
	case theme.ColorNameFocus:
		return p.focus
	case theme.ColorNameError:
		return p.errorCol
	case theme.ColorNameSuccess:
		return p.success
	case theme.ColorNameWarning:
		return p.warning
	case theme.ColorNameShadow:
		return p.shadow
	case theme.ColorNameScrollBar:
		return p.scrollBar
	case theme.ColorNameScrollBarBackground:
		return color.Transparent
	}
	return t.Theme.Color(name, variant)
}

func (t formTheme) Size(name fyne.ThemeSizeName) float32 {
	// A form is read before it is filled in, so it gets more space and slightly
	// larger text than a dense application would want.
	switch name {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInnerPadding:
		return 8
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 23
	case theme.SizeNameSubHeadingText:
		return 17
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNameLineSpacing:
		return 4
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameInputRadius:
		return 6
	case theme.SizeNameButtonRadius:
		return 6
	case theme.SizeNameSelectionRadius:
		return 5
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameScrollBarSmall:
		return 5
	}
	return t.Theme.Size(name)
}

func hex(v uint32) color.Color {
	return color.NRGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 0xFF}
}

func rgba(r, g, b, a uint8) color.Color {
	return color.NRGBA{R: r, G: g, B: b, A: a}
}

// tintCount is how many tints a palette cycles through. Both palettes define
// the same number, which the tests assert.
const tintCount = 6

// tintProvider is a theme that can colour the band behind a question.
//
// The tints are looked up through this rather than as theme colour names,
// because a name the running theme has never heard of is a logged error on
// every draw. A theme that does not provide tints simply gets none.
type tintProvider interface {
	QuestionTint(index int, variant fyne.ThemeVariant) color.Color
}

// QuestionTint returns the background for the nth question on screen, cycling
// so that no two adjacent questions share one.
func (t formTheme) QuestionTint(index int, variant fyne.ThemeVariant) color.Color {
	p := lightPalette
	if variant == theme.VariantDark {
		p = darkPalette
	}
	if len(p.questionTints) == 0 {
		return color.Transparent
	}
	if index < 0 {
		index = 0
	}
	return p.questionTints[index%len(p.questionTints)]
}
