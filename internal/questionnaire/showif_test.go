package questionnaire

import (
	"strings"
	"testing"
)

func TestShowIfReferenceErrors(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []string
	}{{
		name: "dangling reference",
		src: `
title: t
questions:
  - {id: a, type: text, prompt: "One"}
  - id: b
    type: text
    prompt: Two
    show_if: {ghost: yes}
`,
		want: []string{`question "b"`, `unknown question "ghost"`},
	}, {
		name: "self reference",
		src: `
title: t
questions:
  - id: a
    type: text
    prompt: One
    show_if: {a: yes}
`,
		want: []string{`question "a"`, "refers to itself"},
	}, {
		name: "two element cycle",
		src: `
title: t
questions:
  - id: a
    type: text
    prompt: One
    show_if: {b: yes}
  - id: b
    type: text
    prompt: Two
    show_if: {a: yes}
`,
		want: []string{"circular `show_if` chain", "a", "b"},
	}, {
		name: "three element cycle",
		src: `
title: t
questions:
  - id: a
    type: text
    prompt: One
    show_if: {c: yes}
  - id: b
    type: text
    prompt: Two
    show_if: {a: yes}
  - id: c
    type: text
    prompt: Three
    show_if: {b: yes}
`,
		want: []string{"circular `show_if` chain", "a -> ", "b", "c"},
	}, {
		name: "show_if not a mapping",
		src: `
title: t
questions:
  - id: a
    type: text
    prompt: One
    show_if: [b]
`,
		want: []string{"must be a mapping"},
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := loadErr(t, tc.src)
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Errorf("error = %q\nwant it to contain %q", got, want)
				}
			}
		})
	}
}

func TestCycleReportedOnce(t *testing.T) {
	src := `
title: t
questions:
  - id: a
    type: text
    prompt: One
    show_if: {b: yes}
  - id: b
    type: text
    prompt: Two
    show_if: {a: yes}
`
	_, err := LoadBytes("test.yaml", []byte(src))
	errs, ok := err.(Errors)
	if !ok {
		t.Fatalf("error is %T, want Errors", err)
	}
	cycles := 0
	for _, e := range errs {
		if strings.Contains(e.Error(), "circular") {
			cycles++
		}
	}
	if cycles != 1 {
		t.Errorf("reported the cycle %d times, want once:\n%v", cycles, err)
	}
}

func TestDeepChainIsNotACycle(t *testing.T) {
	// A long chain and a diamond both revisit nodes without looping.
	mustLoad(t, `
title: t
questions:
  - {id: a, type: boolean, prompt: "One"}
  - id: b
    type: boolean
    prompt: Two
    show_if: {a: true}
  - id: c
    type: boolean
    prompt: Three
    show_if: {a: true}
  - id: d
    type: text
    prompt: Four
    show_if: {b: true, c: true}
`)
}
