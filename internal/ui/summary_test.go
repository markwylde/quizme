package ui

import (
	"strings"
	"testing"

	"github.com/markwylde/interrogate/internal/questionnaire"
)

func TestAnswerSummaryRendersEveryType(t *testing.T) {
	tests := []struct {
		name string
		q    questionnaire.Question
		want string
	}{
		{
			name: "select",
			q:    questionnaire.Question{Type: questionnaire.TypeSelect, Answer: "sidecar"},
			want: "sidecar",
		},
		{
			name: "multiselect",
			q:    questionnaire.Question{Type: questionnaire.TypeMultiselect, Answer: []string{"yaml", "json"}},
			want: "yaml, json",
		},
		{
			name: "text",
			q:    questionnaire.Question{Type: questionnaire.TypeText, Answer: "interrogate"},
			want: "interrogate",
		},
		{
			name: "textarea",
			q:    questionnaire.Question{Type: questionnaire.TypeTextarea, Answer: "because it drifts"},
			want: "because it drifts",
		},
		{
			name: "number",
			q:    questionnaire.Question{Type: questionnaire.TypeNumber, Answer: 12.5},
			want: "12.5",
		},
		{
			name: "boolean true",
			q:    questionnaire.Question{Type: questionnaire.TypeBoolean, Answer: true},
			want: "Yes",
		},
		{
			name: "boolean false",
			q:    questionnaire.Question{Type: questionnaire.TypeBoolean, Answer: false},
			want: "No",
		},
		{
			name: "scale",
			q:    questionnaire.Question{Type: questionnaire.TypeScale, Answer: 4},
			want: "4",
		},
		{
			name: "rank",
			q:    questionnaire.Question{Type: questionnaire.TypeRank, Answer: []string{"correctness", "speed"}},
			want: "1. correctness, 2. speed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			q := tc.q
			if got := answerSummary(&q); got != tc.want {
				t.Errorf("answerSummary = %q, want %q", got, tc.want)
			}
		})
	}
}

// Every supported type needs a rendering; a type added later without one would
// summarise to a Go value.
func TestEveryTypeHasASummary(t *testing.T) {
	answers := map[questionnaire.Type]any{
		questionnaire.TypeSelect:      "a",
		questionnaire.TypeMultiselect: []string{"a"},
		questionnaire.TypeText:        "a",
		questionnaire.TypeTextarea:    "a",
		questionnaire.TypeNumber:      1.0,
		questionnaire.TypeBoolean:     false,
		questionnaire.TypeScale:       1,
		questionnaire.TypeRank:        []string{"a"},
	}
	for _, kind := range questionnaire.AllTypes {
		answer, ok := answers[kind]
		if !ok {
			t.Fatalf("no fixture answer for %s", kind)
		}
		q := questionnaire.Question{Type: kind, Answer: answer}
		if got := answerSummary(&q); got == "" {
			t.Errorf("%s summarises to nothing when answered", kind)
		}
	}
}

func TestAnswerSummaryOfAnUnansweredQuestion(t *testing.T) {
	for _, kind := range questionnaire.AllTypes {
		q := questionnaire.Question{Type: kind}
		if got := answerSummary(&q); got != "" {
			t.Errorf("%s with no answer summarises to %q, want nothing", kind, got)
		}
	}
	// A cleared answer is no answer, so it summarises to nothing too.
	for _, q := range []questionnaire.Question{
		{Type: questionnaire.TypeText, Answer: ""},
		{Type: questionnaire.TypeMultiselect, Answer: []string{}},
	} {
		q := q
		if got := answerSummary(&q); got != "" {
			t.Errorf("cleared %s summarises to %q, want nothing", q.Type, got)
		}
	}
}

func TestAnswerSummaryIsOneShortLine(t *testing.T) {
	long := questionnaire.Question{Type: questionnaire.TypeTextarea, Answer: strings.Repeat("a", 200)}
	got := answerSummary(&long)
	if !strings.HasSuffix(got, "…") {
		t.Errorf("summary = %q, want it truncated with an ellipsis", got)
	}
	if len([]rune(got)) > previewLimit+1 {
		t.Errorf("summary is %d runes, want no more than %d", len([]rune(got)), previewLimit+1)
	}

	multi := questionnaire.Question{Type: questionnaire.TypeTextarea, Answer: "first line\nsecond line"}
	got = answerSummary(&multi)
	if strings.Contains(got, "\n") {
		t.Errorf("summary = %q, want a single line", got)
	}
	if got != "first line…" {
		t.Errorf("summary = %q, want %q", got, "first line…")
	}

	many := questionnaire.Question{
		Type:   questionnaire.TypeMultiselect,
		Answer: []string{"one option", "two option", "three option", "four option", "five option"},
	}
	if got := answerSummary(&many); len([]rune(got)) > previewLimit+1 {
		t.Errorf("a long multiselect summary is %d runes: %q", len([]rune(got)), got)
	}
}

func TestAnswerSummaryOfANilQuestion(t *testing.T) {
	if got := answerSummary(nil); got != "" {
		t.Errorf("answerSummary(nil) = %q", got)
	}
}
