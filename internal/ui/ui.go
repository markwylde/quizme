// Package ui renders a questionnaire as a native desktop form.
package ui

import "fyne.io/fyne/v2/app"

// probe keeps the toolkit dependency wired up while the form is built out.
func probe() { _ = app.New }
