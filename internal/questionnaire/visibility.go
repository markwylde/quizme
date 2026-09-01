package questionnaire

import "reflect"

// Visibility records which questions a responder can currently see, given the
// answers so far.
type Visibility map[string]bool

// Visible reports whether the question with the given id is currently shown.
// An unknown id is not visible.
func (v Visibility) Visible(id string) bool { return v[id] }

// Evaluate works out which questions are visible for the current answers.
//
// Questions are resolved in dependency order, so a question gated on another
// gated question sees a settled result rather than a half-computed one. Load
// rejects circular show_if chains, so an order always exists.
//
// A hidden question counts as unanswered for anything that depends on it: its
// answer is not recorded, so treating it as live would gate questions on a
// value the responder cannot see.
func (d *Document) Evaluate() Visibility {
	vis := make(Visibility, len(d.Questions))
	for _, id := range d.dependencyOrder() {
		q := d.Question(id)
		if q == nil {
			continue
		}
		vis[id] = d.conditionsHold(q, vis)
	}
	return vis
}

func (d *Document) conditionsHold(q *Question, vis Visibility) bool {
	for _, c := range q.ShowIf {
		other := d.Question(c.Question)
		if other == nil || !vis[other.ID] || !other.HasAnswer() {
			return false
		}
		if !matches(other.Answer, c.Expected) {
			return false
		}
	}
	return true
}

// dependencyOrder returns question ids such that every question follows the
// questions its show_if refers to. Document order is preserved wherever the
// dependencies allow, which keeps the result stable and easy to reason about.
func (d *Document) dependencyOrder() []string {
	const (
		white = 0
		grey  = 1
		black = 2
	)
	colour := make(map[string]int, len(d.Questions))
	order := make([]string, 0, len(d.Questions))

	var visit func(id string)
	visit = func(id string) {
		if colour[id] != white {
			return // already placed, or a cycle that load already rejected
		}
		colour[id] = grey
		if q := d.Question(id); q != nil {
			for _, c := range q.ShowIf {
				visit(c.Question)
			}
		}
		colour[id] = black
		order = append(order, id)
	}

	for _, q := range d.Questions {
		visit(q.ID)
	}
	return order
}

// VisibleQuestions returns the questions to present, in document order.
func (d *Document) VisibleQuestions() []*Question {
	vis := d.Evaluate()
	out := make([]*Question, 0, len(d.Questions))
	for _, q := range d.Questions {
		if vis[q.ID] {
			out = append(out, q)
		}
	}
	return out
}

// Unanswered returns the visible, required questions that still have no answer.
// A hidden required question never blocks submission: it is not applicable, so
// there is nothing to answer.
func (d *Document) Unanswered() []*Question {
	vis := d.Evaluate()
	var out []*Question
	for _, q := range d.Questions {
		if vis[q.ID] && q.Required && !q.HasAnswer() {
			out = append(out, q)
		}
	}
	return out
}

// IsComplete reports whether every visible required question has an answer.
func (d *Document) IsComplete() bool { return len(d.Unanswered()) == 0 }

// matches decides whether an answer satisfies a show_if clause.
//
// Two shorthands make the common cases readable:
//
//   - a list-valued answer matches a single expected value it contains, so a
//     multiselect can gate a follow-up without naming every combination;
//   - a list-valued expectation means "any of these".
//
// Together: two lists match when they overlap.
func matches(answer, expected any) bool {
	answers, answerIsList := asList(answer)
	wanted, expectIsList := asList(expected)

	switch {
	case !answerIsList && !expectIsList:
		return scalarEqual(answer, expected)
	case answerIsList && !expectIsList:
		return containsScalar(answers, expected)
	case !answerIsList && expectIsList:
		return containsScalar(wanted, answer)
	default:
		for _, a := range answers {
			if containsScalar(wanted, a) {
				return true
			}
		}
		return false
	}
}

func asList(v any) ([]any, bool) {
	switch t := v.(type) {
	case []string:
		out := make([]any, len(t))
		for i, s := range t {
			out[i] = s
		}
		return out, true
	case []any:
		return t, true
	}
	return nil, false
}

func containsScalar(list []any, want any) bool {
	for _, v := range list {
		if scalarEqual(v, want) {
			return true
		}
	}
	return false
}

// scalarEqual compares two YAML scalars. Numbers decode as int or float64
// depending on how they were written, and options are matched by their textual
// form, so both are normalized before comparing.
func scalarEqual(a, b any) bool {
	if reflect.DeepEqual(a, b) {
		return true
	}
	if af, aok := numeric(a); aok {
		if bf, bok := numeric(b); bok {
			return af == bf
		}
	}
	as, aok := scalarString(a)
	bs, bok := scalarString(b)
	return aok && bok && as == bs
}
