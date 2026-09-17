package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/markwylde/quizme/internal/questionnaire"
)

// summaryLimit is how much of an answer a collapsed question shows. Longer than
// the comment preview because the answer has a line to itself, and generous
// enough that most answers arrive whole; the label truncates against the window
// if one does not.
const summaryLimit = 120

// answerSummary renders a question's answer on one line, for the header of a
// collapsed question.
//
// A folded page is meant to read as a review of what the responder has said, so
// every type needs a rendering a reader recognises as that answer -- not a Go
// value printed. An unanswered question summarises to nothing: the header shows
// the prompt alone, and its lack of a tick says the rest.
//
// Shortening goes through the same helper the collapsed comment uses, at a
// longer limit: the answer sits on a line of its own under the prompt, so there
// is room for most answers in full.
func answerSummary(q *questionnaire.Question) string {
	if q == nil || !q.HasAnswer() {
		return ""
	}

	shorten := func(text string) string { return previewTo(text, summaryLimit) }

	switch q.Type {
	case questionnaire.TypeBoolean:
		if v, ok := q.Answer.(bool); ok {
			if v {
				return "Yes"
			}
			return "No"
		}
	case questionnaire.TypeScale:
		if v, ok := q.Answer.(int); ok {
			return strconv.Itoa(v)
		}
	case questionnaire.TypeNumber:
		if v, ok := q.Answer.(float64); ok {
			return trimNumber(v)
		}
	case questionnaire.TypeMultiselect:
		if v, ok := q.Answer.([]string); ok {
			return shorten(strings.Join(v, ", "))
		}
	case questionnaire.TypeRank:
		if v, ok := q.Answer.([]string); ok {
			ranked := make([]string, 0, len(v))
			for i, option := range v {
				ranked = append(ranked, fmt.Sprintf("%d. %s", i+1, option))
			}
			return shorten(strings.Join(ranked, ", "))
		}
	case questionnaire.TypeSelect, questionnaire.TypeText, questionnaire.TypeTextarea:
		if v, ok := q.Answer.(string); ok {
			return shorten(v)
		}
	}

	// An answer of an unexpected shape is still worth showing: better a
	// printed value than a question that looks unanswered.
	return shorten(fmt.Sprint(q.Answer))
}
