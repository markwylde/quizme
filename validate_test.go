package main

import (
	"os"
	"strings"
	"testing"

	"github.com/markwylde/quizme/internal/questionnaire"
	"github.com/markwylde/quizme/internal/ui"
)

// --validate is for a caller checking a questionnaire it has just written. A
// plain run already validates before it presents anything, so the flag's whole
// value is that it does not present, and does not write.

func TestValidateAcceptsAGoodQuestionnaire(t *testing.T) {
	path := writeSample(t, sample)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	got := invoke(t, []string{"--validate", path}, neverCalled(t))
	if got.code != exitSubmitted {
		t.Errorf("exit code = %d, want %d\nstderr: %s", got.code, exitSubmitted, got.stderr)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Errorf("--validate modified the file:\n%s", after)
	}
}

func TestValidateRejectsABadQuestionnaire(t *testing.T) {
	path := writeSample(t, "title: t\nquestions:\n  - {id: q, type: carousel, prompt: \"What?\"}\n")
	got := invoke(t, []string{"--validate", path}, neverCalled(t))

	if got.code != exitError {
		t.Errorf("exit code = %d, want %d", got.code, exitError)
	}
	if !strings.Contains(got.stderr, "carousel") {
		t.Errorf("stderr should carry the validation errors: %s", got.stderr)
	}
	if got.stdout != "" {
		t.Errorf("stdout = %q, want empty", got.stdout)
	}
}

func TestValidateReportsEveryError(t *testing.T) {
	path := writeSample(t, `title: t
questions:
  - {id: a, type: carousel, prompt: "One"}
  - {id: a, type: text, prompt: "Two"}
  - {id: c, type: select, prompt: "Three"}
`)
	got := invoke(t, []string{"--validate", path}, neverCalled(t))
	for _, want := range []string{"carousel", "repeats the id", "requires `options`"} {
		if !strings.Contains(got.stderr, want) {
			t.Errorf("stderr should mention %q:\n%s", want, got.stderr)
		}
	}
}

func TestValidateNeedsNoDisplay(t *testing.T) {
	// The presenter is where the display check lives, and --validate must never
	// reach it, so the flag works over SSH and in CI.
	path := writeSample(t, sample)
	got := invoke(t, []string{"--validate", path}, func(*questionnaire.Document, ui.Options) (questionnaire.Status, error) {
		t.Error("--validate reached the presenter, so it would fail without a display")
		return questionnaire.StatusDismissed, nil
	})
	if got.code != exitSubmitted {
		t.Errorf("exit code = %d, want %d\nstderr: %s", got.code, exitSubmitted, got.stderr)
	}
}

func TestValidateMissingFile(t *testing.T) {
	got := invoke(t, []string{"--validate", "does-not-exist.yaml"}, neverCalled(t))
	if got.code != exitError {
		t.Errorf("exit code = %d, want %d", got.code, exitError)
	}
	if !strings.Contains(got.stderr, "does-not-exist.yaml") {
		t.Errorf("stderr should name the path: %s", got.stderr)
	}
}

func TestValidateNeedsAPath(t *testing.T) {
	got := invoke(t, []string{"--validate"}, neverCalled(t))
	if got.code != exitError {
		t.Errorf("exit code = %d, want %d", got.code, exitError)
	}
	if !strings.Contains(got.stderr, "no questionnaire given") {
		t.Errorf("stderr = %q", got.stderr)
	}
}

func TestUsageDocumentsValidate(t *testing.T) {
	got := invoke(t, nil, neverCalled(t))
	if !strings.Contains(got.stderr, "--validate") {
		t.Errorf("usage does not mention --validate:\n%s", got.stderr)
	}
	// The usage has to say what it is actually for, since the obvious reading
	// describes something a plain run already does.
	if !strings.Contains(got.stderr, "already validates") {
		t.Errorf("usage does not explain that a plain run validates too:\n%s", got.stderr)
	}
}
