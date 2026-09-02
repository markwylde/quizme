package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/markwylde/interrogate/internal/questionnaire"
)

// answerSummary renders a question's answer on one line, for the header of a
// collapsed question.
//
// A folded page is meant to read as a review of what the responder has said, so
// every type needs a rendering a reader recognises as that answer -- not a Go
// value printed. An unanswered question summarises to nothing: the header shows
// the prompt alone, and its lack of a tick says the rest.
//
// Shortening goes through the same preview helper the collapsed comment uses, so
// a long answer and a long comment truncate identically.
func answerSummary(q *questionnaire.Question) string {
	if q == nil || !q.HasAnswer() {
		return ""
	}

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
			return preview(strings.Join(v, ", "))
		}
	case questionnaire.TypeRank:
		if v, ok := q.Answer.([]string); ok {
			ranked := make([]string, 0, len(v))
			for i, option := range v {
				ranked = append(ranked, fmt.Sprintf("%d. %s", i+1, option))
			}
			return preview(strings.Join(ranked, ", "))
		}
	case questionnaire.TypeSelect, questionnaire.TypeText, questionnaire.TypeTextarea:
		if v, ok := q.Answer.(string); ok {
			return preview(v)
		}
	}

	// An answer of an unexpected shape is still worth showing: better a
	// printed value than a question that looks unanswered.
	return preview(fmt.Sprint(q.Answer))
}
