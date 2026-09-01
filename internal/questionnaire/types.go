// Package questionnaire loads, validates, and answers questionnaire documents.
//
// A questionnaire is a YAML file authored by an agent and answered by a human.
// Answers live on the questions they belong to, in the same file, so the two
// can never drift apart. Everything in the file that this package does not
// explicitly own is preserved byte for byte; see splice.go.
package questionnaire

import (
	"fmt"
	"strings"
)

// Type is the kind of control a question is answered with, and determines what
// shape a valid answer takes.
type Type string

const (
	TypeSelect      Type = "select"
	TypeMultiselect Type = "multiselect"
	TypeText        Type = "text"
	TypeTextarea    Type = "textarea"
	TypeNumber      Type = "number"
	TypeBoolean     Type = "boolean"
	TypeScale       Type = "scale"
	TypeRank        Type = "rank"
)

// AllTypes lists every supported question type, in the order they are
// documented.
var AllTypes = []Type{
	TypeSelect, TypeMultiselect, TypeText, TypeTextarea,
	TypeNumber, TypeBoolean, TypeScale, TypeRank,
}

// Valid reports whether t is a recognised question type.
func (t Type) Valid() bool {
	for _, known := range AllTypes {
		if t == known {
			return true
		}
	}
	return false
}

// NeedsOptions reports whether a type draws its answer from a fixed option
// list, and therefore requires one.
func (t Type) NeedsOptions() bool {
	switch t {
	case TypeSelect, TypeMultiselect, TypeRank:
		return true
	}
	return false
}

// Status records how far a questionnaire has got. It is the only coordination
// channel between the form and whatever launched it: no lock file, no sidecar.
type Status string

const (
	// StatusPending is a questionnaire authored but not yet answered.
	StatusPending Status = "pending"
	// StatusSubmitted is a questionnaire completed and submitted.
	StatusSubmitted Status = "submitted"
	// StatusSaved is a questionnaire saved with required questions outstanding.
	StatusSaved Status = "saved"
	// StatusDismissed is a questionnaire whose answers the responder discarded.
	StatusDismissed Status = "dismissed"
)

// Valid reports whether s is a recognised status.
func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusSubmitted, StatusSaved, StatusDismissed:
		return true
	}
	return false
}

// Condition is one clause of a question's show_if: the answer to Question must
// match Expected for the clause to hold.
type Condition struct {
	Question string
	Expected any
}

// Question is a single question and, once answered, its answer.
//
// The node fields locate the question inside the source document so answers
// can be spliced into the exact bytes they belong to.
type Question struct {
	ID       string
	Prompt   string
	Help     string
	Type     Type
	Options  []string
	Min      *float64
	Max      *float64
	Required bool
	ShowIf   []Condition

	// Answer is the recorded answer, or nil if the question is unanswered.
	// Its Go type depends on Type; see NormalizeAnswer.
	Answer any
	// Comment is the free-text note attached to the question. Every question
	// accepts one, answered or not.
	Comment string

	pos questionPos
}

// HasAnswer reports whether the question holds an answer that counts as given.
// An empty string or empty list is treated as no answer, so clearing a control
// in the form is indistinguishable from never touching it.
func (q *Question) HasAnswer() bool {
	return !isBlank(q.Answer)
}

// ScaleBounds returns the effective bounds of a scale question, applying the
// documented defaults of 1 and 5.
func (q *Question) ScaleBounds() (int, int) {
	lo, hi := 1, 5
	if q.Min != nil {
		lo = int(*q.Min)
	}
	if q.Max != nil {
		hi = int(*q.Max)
	}
	return lo, hi
}

// Document is a whole questionnaire: its metadata, its questions, and the
// original bytes it was read from.
type Document struct {
	Path  string
	Title string
	Intro string

	Status      Status
	SubmittedAt string

	Questions []*Question

	source  []byte
	lineEnd string
	pos     documentPos
}

// Source returns the exact bytes the document was loaded from.
func (d *Document) Source() []byte { return d.source }

// Question returns the question with the given id, or nil.
func (d *Document) Question(id string) *Question {
	for _, q := range d.Questions {
		if q.ID == id {
			return q
		}
	}
	return nil
}

// isBlank reports whether v should be treated as "no answer given".
func isBlank(v any) bool {
	switch t := v.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(t) == ""
	case []string:
		return len(t) == 0
	case []any:
		return len(t) == 0
	}
	return false
}

// Errors is a collection of validation failures. Loading reports everything
// wrong with a document at once rather than stopping at the first problem, so
// an author fixing a questionnaire sees the whole list.
type Errors []error

func (e Errors) Error() string {
	msgs := make([]string, len(e))
	for i, err := range e {
		msgs[i] = err.Error()
	}
	return strings.Join(msgs, "\n")
}

// Unwrap exposes the individual failures to errors.Is and errors.As.
func (e Errors) Unwrap() []error { return e }

// ErrOrNil returns e as an error, or nil when it is empty.
func (e Errors) ErrOrNil() error {
	if len(e) == 0 {
		return nil
	}
	return e
}

func (e *Errors) addf(format string, args ...any) {
	*e = append(*e, fmt.Errorf(format, args...))
}
