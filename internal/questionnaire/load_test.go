package questionnaire

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func loadFixture(t *testing.T, name string) *Document {
	t.Helper()
	path := filepath.Join("testdata", name)
	doc, err := Load(path)
	if err != nil {
		t.Fatalf("Load(%s): %v", name, err)
	}
	return doc
}

func TestLoadFullFixture(t *testing.T) {
	doc := loadFixture(t, "full.yaml")

	if doc.Title != "Storage decisions" {
		t.Errorf("Title = %q", doc.Title)
	}
	if !strings.HasPrefix(doc.Intro, "A few things to settle") {
		t.Errorf("Intro = %q", doc.Intro)
	}
	if doc.Status != StatusPending {
		t.Errorf("Status = %q, want pending", doc.Status)
	}
	if doc.SubmittedAt != "" {
		t.Errorf("SubmittedAt = %q, want empty for an explicit null", doc.SubmittedAt)
	}
	if len(doc.Questions) != 8 {
		t.Fatalf("got %d questions, want 8", len(doc.Questions))
	}

	// Every declared type is represented, which is what makes this fixture
	// worth keeping in step with AllTypes.
	seen := map[Type]bool{}
	for _, q := range doc.Questions {
		seen[q.Type] = true
	}
	for _, want := range AllTypes {
		if !seen[want] {
			t.Errorf("fixture does not cover type %q", want)
		}
	}

	storage := doc.Question("storage")
	if storage == nil {
		t.Fatal("question storage not found")
	}
	if !storage.Required {
		t.Error("storage should be required")
	}
	if got, want := strings.Join(storage.Options, ","), "in-place,sidecar,both"; got != want {
		t.Errorf("options = %q, want %q", got, want)
	}

	why := doc.Question("storage_why")
	if len(why.ShowIf) != 1 {
		t.Fatalf("storage_why show_if = %+v", why.ShowIf)
	}
	if why.ShowIf[0].Question != "storage" || why.ShowIf[0].Expected != "sidecar" {
		t.Errorf("show_if = %+v", why.ShowIf[0])
	}

	budget := doc.Question("budget")
	if budget.Min == nil || *budget.Min != 1 || budget.Max == nil || *budget.Max != 30 {
		t.Errorf("budget bounds = %v..%v", budget.Min, budget.Max)
	}

	// Inline flow sequences must read the same as block ones.
	if got := strings.Join(doc.Question("formats").Options, ","); got != "yaml,json,toml" {
		t.Errorf("inline options = %q", got)
	}

	if lo, hi := doc.Question("urgency").ScaleBounds(); lo != 1 || hi != 5 {
		t.Errorf("scale bounds = %d..%d", lo, hi)
	}
}

func TestScaleBoundsDefault(t *testing.T) {
	doc := mustLoad(t, `
title: t
questions:
  - id: q
    type: scale
    prompt: How much?
`)
	if lo, hi := doc.Question("q").ScaleBounds(); lo != 1 || hi != 5 {
		t.Errorf("default scale bounds = %d..%d, want 1..5", lo, hi)
	}
}

func mustLoad(t *testing.T, src string) *Document {
	t.Helper()
	doc, err := LoadBytes("test.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	return doc
}

func loadErr(t *testing.T, src string) string {
	t.Helper()
	_, err := LoadBytes("test.yaml", []byte(src))
	if err == nil {
		t.Fatal("expected the document to be rejected, but it loaded")
	}
	return err.Error()
}

func TestLoadRejections(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{{
		name: "no questions key",
		src:  "title: t\n",
		want: "missing required key `questions`",
	}, {
		name: "empty questions",
		src:  "title: t\nquestions: []\n",
		want: "`questions` is empty",
	}, {
		name: "questions not a sequence",
		src:  "title: t\nquestions: nope\n",
		want: "must be a sequence",
	}, {
		name: "not a mapping",
		src:  "- just\n- a list\n",
		want: "expected a YAML mapping",
	}, {
		name: "duplicate ids",
		src: `
title: t
questions:
  - {id: dupe, type: text, prompt: "First"}
  - {id: dupe, type: text, prompt: "Second"}
`,
		want: `question "dupe" at questions[1] repeats the id already used at questions[0]`,
	}, {
		name: "missing id",
		src: `
title: t
questions:
  - {type: text, prompt: "Nameless"}
`,
		want: "missing required key `id`",
	}, {
		name: "missing prompt",
		src: `
title: t
questions:
  - {id: silent, type: text}
`,
		want: `question "silent"`,
	}, {
		name: "missing type",
		src: `
title: t
questions:
  - {id: q, prompt: "What?"}
`,
		want: "missing required key `type`",
	}, {
		name: "unknown type",
		src: `
title: t
questions:
  - {id: q, type: carousel, prompt: "What?"}
`,
		want: `unknown type "carousel"`,
	}, {
		name: "select without options",
		src: `
title: t
questions:
  - {id: q, type: select, prompt: "What?"}
`,
		want: "type select requires `options`",
	}, {
		name: "multiselect with one option",
		src: `
title: t
questions:
  - {id: q, type: multiselect, prompt: "What?", options: [only]}
`,
		want: "requires at least two options, found 1",
	}, {
		name: "rank with one option",
		src: `
title: t
questions:
  - {id: q, type: rank, prompt: "What?", options: [only]}
`,
		want: "requires at least two options, found 1",
	}, {
		name: "duplicate options",
		src: `
title: t
questions:
  - {id: q, type: select, prompt: "What?", options: [a, b, a]}
`,
		want: `option "a" is listed more than once`,
	}, {
		name: "min above max",
		src: `
title: t
questions:
  - {id: q, type: number, prompt: "How many?", min: 10, max: 2}
`,
		want: "is greater than",
	}, {
		name: "unknown status",
		src: `
title: t
status: halfway
questions:
  - {id: q, type: text, prompt: "What?"}
`,
		want: `unknown status "halfway"`,
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := loadErr(t, tc.src); !strings.Contains(got, tc.want) {
				t.Errorf("error = %q\nwant it to contain %q", got, tc.want)
			}
		})
	}
}

func TestLoadReportsEveryProblemAtOnce(t *testing.T) {
	// An author fixing a questionnaire should see the whole list, not the
	// first failure over and over.
	src := `
title: t
questions:
  - {id: a, type: carousel, prompt: "One"}
  - {id: a, type: text, prompt: "Two"}
  - {id: c, type: select, prompt: "Three"}
`
	_, err := LoadBytes("test.yaml", []byte(src))
	errs, ok := err.(Errors)
	if !ok {
		t.Fatalf("error is %T, want Errors", err)
	}
	if len(errs) < 3 {
		t.Errorf("got %d errors, want at least 3:\n%v", len(errs), err)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "nope.yaml"))
	if !os.IsNotExist(err) {
		t.Errorf("err = %v, want a not-exist error", err)
	}
}
