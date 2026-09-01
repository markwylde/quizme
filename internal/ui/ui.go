// Package ui renders a questionnaire as a native desktop form.
package ui

import (
	"errors"

	"github.com/markwylde/interrogate/internal/questionnaire"
)

// Run presents the questionnaire and blocks until the responder submits, saves,
// or dismisses it, returning the status that outcome corresponds to.
func Run(doc *questionnaire.Document) (questionnaire.Status, error) {
	return "", errors.New("the form is not built yet")
}
