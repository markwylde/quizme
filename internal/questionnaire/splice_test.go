package questionnaire

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var fixedTime = time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

// render loads src, applies the answers described by set, and returns the
// rewritten document as a string.
func render(t *testing.T, src string, status Status, set func(d *Document)) string {
	t.Helper()
	doc, err := LoadBytes("test.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	if set != nil {
		set(doc)
	}
	out, err := doc.Render(status, fixedTime)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	// The result must still be a questionnaire.
	if _, err := LoadBytes("test.yaml", out); err != nil {
		t.Fatalf("rendered document no longer loads: %v\n---\n%s", err, out)
	}
	return string(out)
}

func wantEqual(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("rendered document differs\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestSpliceInsertsAnswerIntoQuestion(t *testing.T) {
	src := `title: t
questions:
  - id: name
    type: text
    prompt: What should we call it?
`
	got := render(t, src, StatusSubmitted, func(d *Document) {
		d.Question("name").Answer = "interrogate"
	})
	wantEqual(t, got, `title: t
status: submitted
submitted_at: "2026-09-01T12:00:00Z"
questions:
  - id: name
    type: text
    prompt: What should we call it?
    answer: interrogate
`)
}

func TestSpliceReplacesExistingAnswerInPlace(t *testing.T) {
	src := `title: t
status: saved
submitted_at: "2026-01-01T00:00:00Z"
questions:
  - id: name
    type: text
    answer: old value
    prompt: What should we call it?
`
	got := render(t, src, StatusSubmitted, func(d *Document) {
		d.Question("name").Answer = "new value"
	})
	// The answer keeps the position its author gave it, above `prompt`.
	wantEqual(t, got, `title: t
status: submitted
submitted_at: "2026-09-01T12:00:00Z"
questions:
  - id: name
    type: text
    answer: new value
    prompt: What should we call it?
`)
}

func TestSpliceRemovesClearedAnswer(t *testing.T) {
	src := `title: t
questions:
  - id: name
    type: text
    prompt: Name
    answer: something
    comment: a note
`
	got := render(t, src, StatusSaved, func(d *Document) {
		d.Question("name").Answer = ""
		d.Question("name").Comment = ""
	})
	wantEqual(t, got, `title: t
status: saved
submitted_at: "2026-09-01T12:00:00Z"
questions:
  - id: name
    type: text
    prompt: Name
`)
}

func TestSpliceKeepsCommentsAndBlankLines(t *testing.T) {
	src := `# Top of file.
title: t

# The only question that matters.
questions:
  # Answer carefully.
  - id: name        # inline note
    type: text
    prompt: Name

  # A second question.
  - id: other
    type: text
    prompt: Other
`
	got := render(t, src, StatusSubmitted, func(d *Document) {
		d.Question("name").Answer = "a"
		d.Question("other").Answer = "b"
	})
	wantEqual(t, got, `# Top of file.
title: t

# The only question that matters.
status: submitted
submitted_at: "2026-09-01T12:00:00Z"
questions:
  # Answer carefully.
  - id: name        # inline note
    type: text
    prompt: Name
    answer: a

  # A second question.
  - id: other
    type: text
    prompt: Other
    answer: b
`)
}

func TestSpliceMultilineAnswerUsesBlockScalar(t *testing.T) {
	src := `title: t
questions:
  - id: why
    type: textarea
    prompt: Why?
`
	got := render(t, src, StatusSubmitted, func(d *Document) {
		d.Question("why").Answer = "first line\n\nthird line: with a colon\n  and an indent"
	})
	wantEqual(t, got, `title: t
status: submitted
submitted_at: "2026-09-01T12:00:00Z"
questions:
  - id: why
    type: textarea
    prompt: Why?
    answer: |-
      first line

      third line: with a colon
        and an indent
`)
}

func TestSpliceMultilineAnswerRoundTrips(t *testing.T) {
	answer := "line one\n\nline three: colon\n  indented\n"
	src := `title: t
questions:
  - id: why
    type: textarea
    prompt: Why?
`
	out := render(t, src, StatusSubmitted, func(d *Document) {
		d.Question("why").Answer = answer
	})
	reloaded, err := LoadBytes("test.yaml", []byte(out))
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got := reloaded.Question("why").Answer; got != answer {
		t.Errorf("answer round-tripped as %q, want %q", got, answer)
	}
}

func TestSpliceListAnswers(t *testing.T) {
	src := `title: t
questions:
  - id: formats
    type: multiselect
    prompt: Which?
    options: [yaml, json, toml]
  - id: order
    type: rank
    prompt: Rank
    options:
      - a
      - b
`
	got := render(t, src, StatusSubmitted, func(d *Document) {
		d.Question("formats").Answer = []string{"yaml", "toml"}
		d.Question("order").Answer = []string{"b", "a"}
	})
	wantEqual(t, got, `title: t
status: submitted
submitted_at: "2026-09-01T12:00:00Z"
questions:
  - id: formats
    type: multiselect
    prompt: Which?
    options: [yaml, json, toml]
    answer:
      - yaml
      - toml
  - id: order
    type: rank
    prompt: Rank
    options:
      - a
      - b
    answer:
      - b
      - a
`)
}

func TestSpliceScalarTypes(t *testing.T) {
	src := `title: t
questions:
  - {id: n, type: number, prompt: "How many?"}
  - {id: b, type: boolean, prompt: "Ship it?"}
  - {id: s, type: scale, prompt: "Urgency?"}
`
	got := render(t, src, StatusSubmitted, func(d *Document) {
		d.Question("n").Answer = 7.0
		d.Question("b").Answer = false
		d.Question("s").Answer = 4
	})
	// The questions were written as flow mappings, so the answers are spliced
	// between the braces rather than expanding them into block style.
	wantEqual(t, got, `title: t
status: submitted
submitted_at: "2026-09-01T12:00:00Z"
questions:
  - {id: n, type: number, prompt: "How many?", answer: 7}
  - {id: b, type: boolean, prompt: "Ship it?", answer: false}
  - {id: s, type: scale, prompt: "Urgency?", answer: 4}
`)
}

func TestSpliceCommentsOnQuestions(t *testing.T) {
	src := `title: t
questions:
  - id: q
    type: text
    prompt: Name
`
	got := render(t, src, StatusSaved, func(d *Document) {
		d.Question("q").Comment = "not sure about this one"
	})
	wantEqual(t, got, `title: t
status: saved
submitted_at: "2026-09-01T12:00:00Z"
questions:
  - id: q
    type: text
    prompt: Name
    comment: not sure about this one
`)
}

func TestSpliceHiddenQuestionAnswerIsDropped(t *testing.T) {
	src := `title: t
questions:
  - id: gate
    type: select
    prompt: Gate
    options: [open, shut]
    answer: open
  - id: gated
    type: text
    prompt: Only when open
    show_if: {gate: open}
    answer: written earlier
`
	got := render(t, src, StatusSubmitted, func(d *Document) {
		d.Question("gate").Answer = "shut"
	})
	if strings.Contains(got, "written earlier") {
		t.Errorf("answer to a hidden question survived:\n%s", got)
	}
	if !strings.Contains(got, "show_if: {gate: open}") {
		t.Errorf("the question itself should remain:\n%s", got)
	}
}

func TestSpliceDismissedRecordsOnlyStatus(t *testing.T) {
	src := `title: t
questions:
  - id: q
    type: text
    prompt: Name
    answer: from an earlier session
`
	got := render(t, src, StatusDismissed, func(d *Document) {
		d.Question("q").Answer = "typed but discarded"
		d.Question("q").Comment = "also discarded"
	})
	wantEqual(t, got, `title: t
status: dismissed
submitted_at: "2026-09-01T12:00:00Z"
questions:
  - id: q
    type: text
    prompt: Name
    answer: from an earlier session
`)
}

func TestSplicePreservesQuoteStylesAndIndentation(t *testing.T) {
	src := `title: "Quoted title"
intro: 'single quoted'
questions:
    - id: q
      type: text
      prompt: >-
        A folded
        prompt
`
	got := render(t, src, StatusSubmitted, func(d *Document) {
		d.Question("q").Answer = "x"
	})
	for _, want := range []string{`title: "Quoted title"`, `intro: 'single quoted'`, "prompt: >-", "      answer: x"} {
		if !strings.Contains(got, want) {
			t.Errorf("rendered document missing %q:\n%s", want, got)
		}
	}
}

func TestSplicePreservesCRLF(t *testing.T) {
	src := "title: t\r\nquestions:\r\n  - id: q\r\n    type: text\r\n    prompt: Name\r\n"
	doc, err := LoadBytes("test.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	doc.Question("q").Answer = "value"
	out, err := doc.Render(StatusSubmitted, fixedTime)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	got := string(out)
	if strings.Contains(strings.ReplaceAll(got, "\r\n", ""), "\n") {
		t.Errorf("output mixes line endings:\n%q", got)
	}
	if !strings.Contains(got, "    answer: value\r\n") {
		t.Errorf("answer not written with CRLF:\n%q", got)
	}
}

func TestSpliceStatusOnlyDocumentWithoutMetaKeys(t *testing.T) {
	// No status or submitted_at in the source: both are added before questions.
	src := `questions:
  - {id: q, type: text, prompt: "Name?"}
`
	got := render(t, src, StatusSaved, nil)
	wantEqual(t, got, `status: saved
submitted_at: "2026-09-01T12:00:00Z"
questions:
  - {id: q, type: text, prompt: "Name?"}
`)
}

func TestSpliceRepeatedSaves(t *testing.T) {
	// Answering an already-answered document must be idempotent in shape:
	// values change, the file's structure does not grow.
	src := `title: t
questions:
  - id: q
    type: text
    prompt: Name
`
	dir := t.TempDir()
	path := filepath.Join(dir, "q.yaml")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	doc, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	doc.Question("q").Answer = "first"
	if err := doc.Save(StatusSaved, fixedTime); err != nil {
		t.Fatalf("first save: %v", err)
	}

	doc.Question("q").Answer = "second"
	doc.Question("q").Comment = "changed my mind"
	if err := doc.Save(StatusSubmitted, fixedTime.Add(time.Hour)); err != nil {
		t.Fatalf("second save: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	wantEqual(t, string(got), `title: t
status: submitted
submitted_at: "2026-09-01T13:00:00Z"
questions:
  - id: q
    type: text
    prompt: Name
    answer: second
    comment: changed my mind
`)
}

func TestSaveIsAtomicOnFailure(t *testing.T) {
	src := `title: t
questions:
  - {id: q, type: text, prompt: "Name?"}
`
	dir := t.TempDir()
	path := filepath.Join(dir, "q.yaml")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	doc, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	doc.Question("q").Answer = "value"

	original := writeAll
	writeAll = func(f *os.File, data []byte) error {
		// Write half the bytes, then fail, the worst case for a naive writer.
		f.Write(data[:len(data)/2])
		return os.ErrClosed
	}
	defer func() { writeAll = original }()

	if err := doc.Save(StatusSubmitted, fixedTime); err == nil {
		t.Fatal("Save should have failed")
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != src {
		t.Errorf("original file was modified:\n%s", got)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("a temporary file was left behind: %v", entries)
	}
}

func TestSavePreservesFileMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "q.yaml")
	src := "title: t\nquestions:\n  - {id: q, type: text, prompt: \"Name?\"}\n"
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	doc, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	doc.Question("q").Answer = "v"
	if err := doc.Save(StatusSubmitted, fixedTime); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("mode = %o, want 600", got)
	}
}

func TestSpliceNoTrailingNewlinePreserved(t *testing.T) {
	src := "title: t\nquestions:\n  - {id: q, type: text, prompt: \"Name?\"}"
	doc, err := LoadBytes("test.yaml", []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	doc.Question("q").Answer = "v"
	out, err := doc.Render(StatusSubmitted, fixedTime)
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasSuffix(string(out), "\n") {
		t.Errorf("a file with no trailing newline gained one:\n%q", string(out))
	}
}
