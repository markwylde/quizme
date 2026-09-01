package questionnaire

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The splicer's central promise: everything the tool does not own comes back
// byte for byte. These tests strip the owned entries from the source and the
// rendered output and require the remainder to be identical.

var (
	ownedBlockKey = regexp.MustCompile(`^\s*(answer|comment|status|submitted_at):`)
	ownedFlowKey  = regexp.MustCompile(`,?\s*(answer|comment): [^,}]*`)
)

// stripOwned removes the entries this tool writes, leaving only the parts of
// the document that must never change.
//
// Blank lines are the awkward case: one inside an owned block scalar belongs to
// the answer, while one after it belongs to the document. They are held back
// until the next non-blank line settles which.
func stripOwned(doc string) string {
	var out []string
	var heldBlanks []string
	skipIndent := -1

	for _, line := range strings.Split(doc, "\n") {
		if skipIndent >= 0 && strings.TrimSpace(line) == "" {
			heldBlanks = append(heldBlanks, line)
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " "))
		if skipIndent >= 0 {
			if indent > skipIndent {
				heldBlanks = nil // those blanks were part of the answer
				continue
			}
			out = append(out, heldBlanks...)
			heldBlanks = nil
			skipIndent = -1
		}

		if ownedBlockKey.MatchString(line) {
			skipIndent = indent
			continue
		}
		out = append(out, ownedFlowKey.ReplaceAllString(line, ""))
	}
	out = append(out, heldBlanks...)
	return strings.Join(out, "\n")
}

func requirePreserved(t *testing.T, src, rendered string) {
	t.Helper()
	if got, want := stripOwned(rendered), stripOwned(src); got != want {
		t.Errorf("content outside the tool's own keys changed\n--- before ---\n%s\n--- after ---\n%s", want, got)
	}
}

func TestPreservationAcrossAwkwardDocuments(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{{
		name: "comments everywhere",
		src: `# leading comment
# second line

title: t   # trailing note

# about the questions
questions:

  # first
  - id: a      # inline
    type: text
    prompt: A

  # second
  - id: b
    type: text
    prompt: B
# trailing comment
`,
	}, {
		name: "unusual indentation",
		src: `title: t
questions:
        - id: a
          type: text
          prompt: A
        - id: b
          type: text
          prompt: B
`,
	}, {
		name: "quoting styles",
		src: `title: "double"
intro: 'single'
questions:
  - id: a
    type: text
    prompt: >-
      folded
      prompt
  - id: b
    type: text
    prompt: |
      literal
      prompt
`,
	}, {
		name: "flow mappings",
		src: `title: t
questions:
  - {id: a, type: text, prompt: "A?"}
  - {id: b, type: select, options: [x, y], prompt: "B?"}
`,
	}, {
		name: "extra keys the tool does not know",
		src: `title: t
owner: someone
tags: [alpha, beta]
questions:
  - id: a
    type: text
    prompt: A
    weight: 10
    meta:
      nested: true
      list:
        - one
        - two
`,
	}, {
		name: "answers already present",
		src: `title: t
status: saved
submitted_at: "2020-01-01T00:00:00Z"
questions:
  - id: a
    type: text
    prompt: A
    answer: previous
    comment: an old note
  - id: b
    type: multiselect
    prompt: B
    options: [x, y]
    answer:
      - x
`,
	}, {
		name: "document marker and trailing dots",
		src: `---
title: t
questions:
  - id: a
    type: text
    prompt: A
`,
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := LoadBytes("test.yaml", []byte(tc.src))
			if err != nil {
				t.Fatalf("LoadBytes: %v", err)
			}
			for i, q := range doc.Questions {
				if i%2 == 0 {
					q.Answer = "answered"
					q.Comment = "noted"
				} else {
					q.Answer = nil
					q.Comment = ""
				}
			}
			// Types that cannot take a plain string keep something valid.
			for _, q := range doc.Questions {
				if q.Answer == nil {
					continue
				}
				switch q.Type {
				case TypeSelect:
					q.Answer = q.Options[0]
				case TypeMultiselect, TypeRank:
					q.Answer = q.Options
				case TypeNumber, TypeScale:
					q.Answer = 1
				case TypeBoolean:
					q.Answer = true
				}
			}

			out, err := doc.Render(StatusSubmitted, fixedTime)
			if err != nil {
				t.Fatalf("Render: %v", err)
			}
			if _, err := LoadBytes("test.yaml", out); err != nil {
				t.Fatalf("rendered document no longer loads: %v\n%s", err, out)
			}
			requirePreserved(t, tc.src, string(out))
		})
	}
}

func TestPreservationOnRepeatedSaves(t *testing.T) {
	src := `# A questionnaire that gets answered twice.
title: t

questions:
  # notes survive
  - id: a
    type: text
    prompt: A
  - {id: b, type: text, prompt: "B?"}
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
	for round, values := range [][2]string{{"one", "two"}, {"three", ""}, {"", "four"}} {
		doc.Question("a").Answer = values[0]
		doc.Question("b").Answer = values[1]
		if err := doc.Save(StatusSaved, fixedTime); err != nil {
			t.Fatalf("save %d: %v", round, err)
		}
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		requirePreserved(t, src, string(got))
	}
}

func TestPreservationWithTheFullFixture(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "full.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	doc, err := LoadBytes("full.yaml", raw)
	if err != nil {
		t.Fatal(err)
	}
	answers := map[string]any{
		"storage":     "sidecar",
		"storage_why": "because it drifts\n\nand that is worse",
		"formats":     []string{"yaml", "json"},
		"name":        "interrogate",
		"budget":      12.0,
		"ship_it":     true,
		"urgency":     4,
		"priorities":  []string{"correctness", "looks", "speed"},
	}
	for id, v := range answers {
		q := doc.Question(id)
		normalized, err := NormalizeAnswer(q, v)
		if err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		q.Answer = normalized
		q.Comment = "note for " + id
	}

	out, err := doc.Render(StatusSubmitted, fixedTime)
	if err != nil {
		t.Fatal(err)
	}
	reloaded, err := LoadBytes("full.yaml", out)
	if err != nil {
		t.Fatalf("rendered document no longer loads: %v\n%s", err, out)
	}
	for id, want := range answers {
		normalized, _ := NormalizeAnswer(doc.Question(id), want)
		got := reloaded.Question(id).Answer
		if !equalAnswers(got, normalized) {
			t.Errorf("%s round-tripped as %#v, want %#v", id, got, normalized)
		}
		if reloaded.Question(id).Comment != "note for "+id {
			t.Errorf("%s comment = %q", id, reloaded.Question(id).Comment)
		}
	}
	if reloaded.Status != StatusSubmitted {
		t.Errorf("status = %q", reloaded.Status)
	}
	requirePreserved(t, string(raw), string(out))
}

func equalAnswers(a, b any) bool {
	as, aok := a.([]string)
	bs, bok := b.([]string)
	if aok && bok {
		return strings.Join(as, "\x00") == strings.Join(bs, "\x00")
	}
	return a == b
}
