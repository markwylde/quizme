package questionnaire

import (
	"strings"
	"testing"
)

// visibleIDs renders the currently visible questions as a comma-separated list,
// which makes the table cases below readable.
func visibleIDs(d *Document) string {
	ids := make([]string, 0, len(d.Questions))
	for _, q := range d.VisibleQuestions() {
		ids = append(ids, q.ID)
	}
	return strings.Join(ids, ",")
}

func answer(t *testing.T, d *Document, id string, v any) {
	t.Helper()
	q := d.Question(id)
	if q == nil {
		t.Fatalf("no question %q", id)
	}
	normalized, err := NormalizeAnswer(q, v)
	if err != nil {
		t.Fatalf("NormalizeAnswer(%s, %#v): %v", id, v, err)
	}
	q.Answer = normalized
}

const gatedDoc = `
title: t
questions:
  - id: storage
    type: select
    prompt: Where?
    options: [in-place, sidecar, both]
  - id: why
    type: text
    prompt: Why sidecar?
    show_if: {storage: sidecar}
  - id: detail
    type: text
    prompt: More detail
    show_if: {why: given}
`

func TestVisibilityChain(t *testing.T) {
	doc := mustLoad(t, gatedDoc)

	if got := visibleIDs(doc); got != "storage" {
		t.Errorf("with nothing answered, visible = %q, want %q", got, "storage")
	}

	answer(t, doc, "storage", "in-place")
	if got := visibleIDs(doc); got != "storage" {
		t.Errorf("condition unsatisfied, visible = %q", got)
	}

	answer(t, doc, "storage", "sidecar")
	if got := visibleIDs(doc); got != "storage,why" {
		t.Errorf("condition satisfied, visible = %q", got)
	}

	answer(t, doc, "why", "given")
	if got := visibleIDs(doc); got != "storage,why,detail" {
		t.Errorf("chained condition, visible = %q", got)
	}

	// Withdrawing the first answer must collapse the whole chain, not just the
	// question directly gated on it.
	answer(t, doc, "storage", "both")
	if got := visibleIDs(doc); got != "storage" {
		t.Errorf("after withdrawing the controlling answer, visible = %q, want %q", got, "storage")
	}
}

func TestVisibilityMultipleConditions(t *testing.T) {
	doc := mustLoad(t, `
title: t
questions:
  - {id: a, type: boolean, prompt: "A?"}
  - {id: b, type: boolean, prompt: "B?"}
  - id: both
    type: text
    prompt: Both true
    show_if: {a: true, b: true}
`)
	answer(t, doc, "a", true)
	if got := visibleIDs(doc); got != "a,b" {
		t.Errorf("one of two conditions met, visible = %q", got)
	}
	answer(t, doc, "b", true)
	if got := visibleIDs(doc); got != "a,b,both" {
		t.Errorf("both conditions met, visible = %q", got)
	}
	answer(t, doc, "b", false)
	if got := visibleIDs(doc); got != "a,b" {
		t.Errorf("condition withdrawn, visible = %q", got)
	}
}

func TestVisibilityDependencyOrderIndependentOfDocumentOrder(t *testing.T) {
	// The gated question is written before the one it depends on.
	doc := mustLoad(t, `
title: t
questions:
  - id: gated
    type: text
    prompt: Gated
    show_if: {gate: yes}
  - {id: gate, type: select, prompt: "Gate?", options: [yes, no]}
`)
	answer(t, doc, "gate", "yes")
	if got := visibleIDs(doc); got != "gated,gate" {
		t.Errorf("visible = %q, want document order %q", got, "gated,gate")
	}
}

func TestVisibilityMatchShorthands(t *testing.T) {
	tests := []struct {
		name   string
		answer any
		expect string // YAML for the show_if value
		want   bool
	}{
		{"scalar equal", "sidecar", "sidecar", true},
		{"scalar different", "in-place", "sidecar", false},
		{"list contains expected", []string{"yaml", "json"}, "json", true},
		{"list lacks expected", []string{"yaml"}, "json", false},
		{"expected list contains answer", "json", "[json, toml]", true},
		{"expected list lacks answer", "yaml", "[json, toml]", false},
		{"lists overlap", []string{"yaml", "toml"}, "[json, toml]", true},
		{"lists disjoint", []string{"yaml"}, "[json, toml]", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := mustLoad(t, `
title: t
questions:
  - {id: fmt, type: multiselect, prompt: "Formats?", options: [yaml, json, toml]}
  - id: follow
    type: text
    prompt: Follow up
    show_if: {fmt: `+tc.expect+`}
`)
			doc.Question("fmt").Answer = tc.answer
			if got := doc.Evaluate().Visible("follow"); got != tc.want {
				t.Errorf("visible = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestVisibilityNumericMatch(t *testing.T) {
	// Written as an int in show_if, normalized to float64 as an answer.
	doc := mustLoad(t, `
title: t
questions:
  - {id: n, type: number, prompt: "How many?"}
  - id: follow
    type: text
    prompt: Follow up
    show_if: {n: 3}
`)
	answer(t, doc, "n", 3)
	if !doc.Evaluate().Visible("follow") {
		t.Error("numeric show_if did not match an equal numeric answer")
	}
}

func TestUnansweredIgnoresHiddenRequired(t *testing.T) {
	doc := mustLoad(t, `
title: t
questions:
  - {id: gate, type: select, prompt: "Gate?", options: [open, shut]}
  - id: needed
    type: text
    prompt: Required when open
    required: true
    show_if: {gate: open}
`)
	if !doc.IsComplete() {
		t.Errorf("a hidden required question blocked completion: %v", doc.Unanswered())
	}

	answer(t, doc, "gate", "open")
	if doc.IsComplete() {
		t.Error("a visible required question with no answer should block completion")
	}
	if got := doc.Unanswered(); len(got) != 1 || got[0].ID != "needed" {
		t.Errorf("Unanswered = %v", got)
	}

	answer(t, doc, "needed", "here")
	if !doc.IsComplete() {
		t.Errorf("still incomplete: %v", doc.Unanswered())
	}
}

func TestUnansweredBlankCountsAsMissing(t *testing.T) {
	doc := mustLoad(t, `
title: t
questions:
  - {id: q, type: text, prompt: "Name?", required: true, answer: "  "}
`)
	if doc.IsComplete() {
		t.Error("a whitespace-only answer should not satisfy a required question")
	}
}

func TestUnansweredReportsEveryMissingQuestion(t *testing.T) {
	doc := mustLoad(t, `
title: t
questions:
  - {id: a, type: text, prompt: "A?", required: true}
  - {id: b, type: text, prompt: "B?", required: true}
  - {id: c, type: text, prompt: "C?"}
`)
	got := doc.Unanswered()
	if len(got) != 2 || got[0].ID != "a" || got[1].ID != "b" {
		t.Errorf("Unanswered = %v, want a and b in document order", got)
	}
}
