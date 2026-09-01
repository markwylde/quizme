package questionnaire

import (
	"strings"
	"testing"
)

func TestFlowInsertsAnswerBetweenBraces(t *testing.T) {
	src := `title: t
questions:
  - {id: q, type: text, prompt: "Name?"}
`
	got := render(t, src, StatusSubmitted, func(d *Document) {
		d.Question("q").Answer = "interrogate"
		d.Question("q").Comment = "seems right"
	})
	wantEqual(t, got, `title: t
status: submitted
submitted_at: "2026-09-01T12:00:00Z"
questions:
  - {id: q, type: text, prompt: "Name?", answer: interrogate, comment: seems right}
`)
}

func TestFlowReplacesExistingEntry(t *testing.T) {
	src := `title: t
questions:
  - {id: q, type: text, answer: old, prompt: "Name?"}
`
	got := render(t, src, StatusSubmitted, func(d *Document) {
		d.Question("q").Answer = "new"
	})
	wantEqual(t, got, `title: t
status: submitted
submitted_at: "2026-09-01T12:00:00Z"
questions:
  - {id: q, type: text, answer: new, prompt: "Name?"}
`)
}

func TestFlowRemovesEntries(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{{
		name: "middle entry",
		src:  `  - {id: q, type: text, answer: gone, prompt: "Name?"}`,
		want: `  - {id: q, type: text, prompt: "Name?"}`,
	}, {
		name: "last entry",
		src:  `  - {id: q, type: text, prompt: "Name?", answer: gone}`,
		want: `  - {id: q, type: text, prompt: "Name?"}`,
	}, {
		name: "both owned entries",
		src:  `  - {id: q, type: text, answer: gone, prompt: "Name?", comment: also gone}`,
		want: `  - {id: q, type: text, prompt: "Name?"}`,
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := render(t, "title: t\nquestions:\n"+tc.src+"\n", StatusSaved, func(d *Document) {
				d.Question("q").Answer = nil
				d.Question("q").Comment = ""
			})
			if !strings.Contains(got, tc.want) {
				t.Errorf("got:\n%s\nwant it to contain:\n%s", got, tc.want)
			}
		})
	}
}

func TestFlowMultilineAnswerIsQuotedNotBlocked(t *testing.T) {
	// A block scalar cannot live between braces, so a multi-line answer has to
	// be escaped onto one line instead.
	src := `title: t
questions:
  - {id: q, type: textarea, prompt: "Why?"}
`
	got := render(t, src, StatusSubmitted, func(d *Document) {
		d.Question("q").Answer = "one\ntwo"
	})
	if strings.Contains(got, "|-") {
		t.Errorf("a block scalar was written inside a flow mapping:\n%s", got)
	}
	reloaded, err := LoadBytes("test.yaml", []byte(got))
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if v := reloaded.Question("q").Answer; v != "one\ntwo" {
		t.Errorf("answer round-tripped as %q", v)
	}
}

func TestFlowListAnswer(t *testing.T) {
	src := `title: t
questions:
  - {id: q, type: multiselect, prompt: "Which?", options: [a, b, c]}
`
	got := render(t, src, StatusSubmitted, func(d *Document) {
		d.Question("q").Answer = []string{"a", "c"}
	})
	if !strings.Contains(got, "answer: [a, c]") {
		t.Errorf("list answer not written in flow style:\n%s", got)
	}
}

func TestFlowWithTrickyScalars(t *testing.T) {
	// Braces and commas inside quoted values must not confuse the scanner.
	src := `title: t
questions:
  - {id: q, type: text, prompt: "What about {this}, or that?"}
`
	got := render(t, src, StatusSubmitted, func(d *Document) {
		d.Question("q").Answer = "fine"
	})
	wantEqual(t, got, `title: t
status: submitted
submitted_at: "2026-09-01T12:00:00Z"
questions:
  - {id: q, type: text, prompt: "What about {this}, or that?", answer: fine}
`)
}

func TestFlowSingleQuotedWithEscapes(t *testing.T) {
	src := `title: t
questions:
  - {id: q, type: text, prompt: 'it''s got a } brace'}
`
	got := render(t, src, StatusSubmitted, func(d *Document) {
		d.Question("q").Answer = "ok"
	})
	if !strings.Contains(got, `prompt: 'it''s got a } brace', answer: ok`) {
		t.Errorf("quoting or brace matching went wrong:\n%s", got)
	}
}

func TestFlowSpanningTwoLines(t *testing.T) {
	src := `title: t
questions:
  - {id: q, type: text,
     prompt: "Name?"}
`
	got := render(t, src, StatusSubmitted, func(d *Document) {
		d.Question("q").Answer = "v"
	})
	if !strings.Contains(got, `prompt: "Name?", answer: v}`) {
		t.Errorf("multi-line flow mapping not handled:\n%s", got)
	}
}

func TestFlowNestedCollectionsInValues(t *testing.T) {
	src := `title: t
questions:
  - {id: q, type: select, options: [a, b], prompt: "Pick?"}
`
	got := render(t, src, StatusSubmitted, func(d *Document) {
		d.Question("q").Answer = "b"
	})
	wantEqual(t, got, `title: t
status: submitted
submitted_at: "2026-09-01T12:00:00Z"
questions:
  - {id: q, type: select, options: [a, b], prompt: "Pick?", answer: b}
`)
}

func TestFlowAndBlockQuestionsInOneDocument(t *testing.T) {
	src := `title: t
questions:
  - {id: flow, type: text, prompt: "Flow?"}
  - id: block
    type: text
    prompt: Block?
`
	got := render(t, src, StatusSubmitted, func(d *Document) {
		d.Question("flow").Answer = "a"
		d.Question("block").Answer = "b"
	})
	wantEqual(t, got, `title: t
status: submitted
submitted_at: "2026-09-01T12:00:00Z"
questions:
  - {id: flow, type: text, prompt: "Flow?", answer: a}
  - id: block
    type: text
    prompt: Block?
    answer: b
`)
}
