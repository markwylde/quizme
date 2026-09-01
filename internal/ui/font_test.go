package ui

import (
	"testing"

	"fyne.io/fyne/v2"
)

func TestBundledFontsAreEmbedded(t *testing.T) {
	// The fonts are compiled in, so they are present wherever the binary runs:
	// there is no system font to be missing on Linux or Windows.
	faces := map[string]fyne.Resource{
		"regular":     fontRegular,
		"bold":        fontBold,
		"italic":      fontItalic,
		"bold italic": fontBoldItalic,
	}
	for name, res := range faces {
		if len(res.Content()) == 0 {
			t.Errorf("%s face is empty", name)
		}
		// Every TrueType file starts with one of these signatures.
		sig := string(res.Content()[:4])
		if sig != "\x00\x01\x00\x00" && sig != "true" && sig != "OTTO" {
			t.Errorf("%s face does not look like a font: %q", name, sig)
		}
	}
}

func TestThemeSelectsTheRightFace(t *testing.T) {
	th := newTheme()
	tests := []struct {
		name  string
		style fyne.TextStyle
		want  fyne.Resource
	}{
		{"plain", fyne.TextStyle{}, fontRegular},
		{"bold", fyne.TextStyle{Bold: true}, fontBold},
		{"italic", fyne.TextStyle{Italic: true}, fontItalic},
		{"bold italic", fyne.TextStyle{Bold: true, Italic: true}, fontBoldItalic},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := th.Font(tc.style); got != tc.want {
				t.Errorf("Font(%+v) = %v, want %v", tc.style, got.Name(), tc.want.Name())
			}
		})
	}

	// Monospace has no Inter face, so it must fall through rather than return
	// a proportional font for a control that relies on alignment.
	if got := th.Font(fyne.TextStyle{Monospace: true}); got == fontRegular {
		t.Error("monospace should not fall back to the proportional face")
	}
}
