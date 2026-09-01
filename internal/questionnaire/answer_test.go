package questionnaire

import (
	"reflect"
	"strings"
	"testing"
)

func f64(v float64) *float64 { return &v }

func TestNormalizeAnswer(t *testing.T) {
	tests := []struct {
		name    string
		q       *Question
		raw     any
		want    any
		wantErr string
	}{
		// select
		{name: "select valid", q: &Question{Type: TypeSelect, Options: []string{"a", "b"}}, raw: "b", want: "b"},
		{name: "select numeric option", q: &Question{Type: TypeSelect, Options: []string{"1", "2"}}, raw: 2, want: "2"},
		{name: "select boolean option", q: &Question{Type: TypeSelect, Options: []string{"true", "false"}}, raw: true, want: "true"},
		{name: "select not an option", q: &Question{Type: TypeSelect, Options: []string{"a", "b"}}, raw: "c", wantErr: "is not one of the options"},
		{name: "select given a list", q: &Question{Type: TypeSelect, Options: []string{"a", "b"}}, raw: []any{"a"}, wantErr: "expected a single value"},

		// multiselect
		{name: "multiselect valid", q: &Question{Type: TypeMultiselect, Options: []string{"a", "b", "c"}}, raw: []any{"a", "c"}, want: []string{"a", "c"}},
		{name: "multiselect single element", q: &Question{Type: TypeMultiselect, Options: []string{"a", "b"}}, raw: []any{"a"}, want: []string{"a"}},
		{name: "multiselect repeated", q: &Question{Type: TypeMultiselect, Options: []string{"a", "b"}}, raw: []any{"a", "a"}, wantErr: "selected more than once"},
		{name: "multiselect unknown", q: &Question{Type: TypeMultiselect, Options: []string{"a", "b"}}, raw: []any{"z"}, wantErr: "is not one of the options"},
		{name: "multiselect not a list", q: &Question{Type: TypeMultiselect, Options: []string{"a", "b"}}, raw: "a", wantErr: "expected a list"},

		// text
		{name: "text", q: &Question{Type: TypeText}, raw: "hello", want: "hello"},
		{name: "textarea multiline", q: &Question{Type: TypeTextarea}, raw: "one\ntwo", want: "one\ntwo"},
		{name: "text from number", q: &Question{Type: TypeText}, raw: 42, want: "42"},

		// number
		{name: "number int", q: &Question{Type: TypeNumber}, raw: 5, want: 5.0},
		{name: "number float", q: &Question{Type: TypeNumber}, raw: 2.5, want: 2.5},
		{name: "number in bounds", q: &Question{Type: TypeNumber, Min: f64(1), Max: f64(30)}, raw: 30, want: 30.0},
		{name: "number below min", q: &Question{Type: TypeNumber, Min: f64(1)}, raw: 0, wantErr: "below the minimum"},
		{name: "number above max", q: &Question{Type: TypeNumber, Max: f64(30)}, raw: 31, wantErr: "above the maximum"},
		{name: "number from text", q: &Question{Type: TypeNumber}, raw: "7", want: 7.0},
		{name: "number nonsense", q: &Question{Type: TypeNumber}, raw: true, wantErr: "expected a number"},

		// boolean
		{name: "boolean true", q: &Question{Type: TypeBoolean}, raw: true, want: true},
		{name: "boolean false", q: &Question{Type: TypeBoolean}, raw: false, want: false},
		{name: "boolean from string", q: &Question{Type: TypeBoolean}, raw: "yes", wantErr: "expected true or false"},

		// scale
		{name: "scale in range", q: &Question{Type: TypeScale}, raw: 3, want: 3},
		{name: "scale at min", q: &Question{Type: TypeScale}, raw: 1, want: 1},
		{name: "scale at max", q: &Question{Type: TypeScale}, raw: 5, want: 5},
		{name: "scale below range", q: &Question{Type: TypeScale}, raw: 0, wantErr: "outside the scale range 1 to 5"},
		{name: "scale above range", q: &Question{Type: TypeScale}, raw: 6, wantErr: "outside the scale range 1 to 5"},
		{name: "scale custom bounds", q: &Question{Type: TypeScale, Min: f64(0), Max: f64(10)}, raw: 9, want: 9},
		{name: "scale fractional", q: &Question{Type: TypeScale}, raw: 2.5, wantErr: "expected a whole number"},

		// rank
		{name: "rank permutation", q: &Question{Type: TypeRank, Options: []string{"a", "b", "c"}}, raw: []any{"c", "a", "b"}, want: []string{"c", "a", "b"}},
		{name: "rank incomplete", q: &Question{Type: TypeRank, Options: []string{"a", "b", "c"}}, raw: []any{"a", "b"}, wantErr: "must list all 3 options"},
		{name: "rank repeated", q: &Question{Type: TypeRank, Options: []string{"a", "b"}}, raw: []any{"a", "a"}, wantErr: "appears more than once"},
		{name: "rank unknown option", q: &Question{Type: TypeRank, Options: []string{"a", "b"}}, raw: []any{"a", "z"}, wantErr: `"z" is not one of the options`},

		// blank
		{name: "nil is blank", q: &Question{Type: TypeSelect, Options: []string{"a"}}, raw: nil, want: nil},
		{name: "empty string is blank", q: &Question{Type: TypeText}, raw: "", want: nil},
		{name: "whitespace is blank", q: &Question{Type: TypeText}, raw: "   ", want: nil},
		{name: "empty list is blank", q: &Question{Type: TypeMultiselect, Options: []string{"a", "b"}}, raw: []any{}, want: nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeAnswer(tc.q, tc.raw)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("got %v, want error containing %q", got, tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("error = %q, want it to contain %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestRecordedAnswersAreValidatedAtLoad(t *testing.T) {
	got := loadErr(t, `
title: t
questions:
  - id: q
    type: select
    prompt: Pick
    options: [a, b]
    answer: z
`)
	if !strings.Contains(got, "recorded `answer` is not valid") {
		t.Errorf("error = %q", got)
	}
}

func TestRecordedAnswersAreNormalized(t *testing.T) {
	doc := mustLoad(t, `
title: t
questions:
  - id: n
    type: number
    prompt: How many?
    answer: 7
  - id: r
    type: rank
    prompt: Order these
    options: [a, b]
    answer: [b, a]
  - id: s
    type: scale
    prompt: How much?
    answer: 4
`)
	if got := doc.Question("n").Answer; got != 7.0 {
		t.Errorf("number answer = %#v, want 7.0", got)
	}
	if got := doc.Question("r").Answer; !reflect.DeepEqual(got, []string{"b", "a"}) {
		t.Errorf("rank answer = %#v", got)
	}
	if got := doc.Question("s").Answer; got != 4 {
		t.Errorf("scale answer = %#v, want int 4", got)
	}
}

func TestHasAnswer(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want bool
	}{
		{"nil", nil, false},
		{"empty string", "", false},
		{"blank string", "  ", false},
		{"empty list", []string{}, false},
		{"text", "x", true},
		{"list", []string{"a"}, true},
		{"zero number", 0.0, true},
		{"false boolean", false, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			q := &Question{Answer: tc.v}
			if got := q.HasAnswer(); got != tc.want {
				t.Errorf("HasAnswer(%#v) = %v, want %v", tc.v, got, tc.want)
			}
		})
	}
}
