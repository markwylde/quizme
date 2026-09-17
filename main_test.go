package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/markwylde/quizme/internal/questionnaire"
	"github.com/markwylde/quizme/internal/ui"
)

// answered returns a presenter that fills in the given answers and leaves the
// form with the given status, standing in for a person at the keyboard.
func answered(status questionnaire.Status, answers map[string]any) presenter {
	return func(doc *questionnaire.Document, _ ui.Options) (questionnaire.Status, error) {
		for id, v := range answers {
			q := doc.Question(id)
			if q == nil {
				continue
			}
			normalized, err := questionnaire.NormalizeAnswer(q, v)
			if err != nil {
				return "", err
			}
			q.Answer = normalized
		}
		return status, nil
	}
}

const sample = `title: Sample
questions:
  - id: name
    type: text
    prompt: What should we call it?
    required: true
  - id: notes
    type: textarea
    prompt: Anything else?
`

func writeSample(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "q.yaml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

type result struct {
	code   int
	stdout string
	stderr string
}

func invoke(t *testing.T, args []string, show presenter) result {
	t.Helper()
	var out, errBuf bytes.Buffer
	code := run(args, &out, &errBuf, show)
	return result{code: code, stdout: out.String(), stderr: errBuf.String()}
}

func TestExitCodeSubmitted(t *testing.T) {
	path := writeSample(t, sample)
	got := invoke(t, []string{path}, answered(questionnaire.StatusSubmitted, map[string]any{"name": "quizme"}))
	if got.code != exitSubmitted {
		t.Errorf("exit code = %d, want %d\nstderr: %s", got.code, exitSubmitted, got.stderr)
	}
	if !strings.Contains(got.stdout, `"status": "submitted"`) {
		t.Errorf("stdout = %s", got.stdout)
	}

	saved, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(saved), "answer: quizme") {
		t.Errorf("answer not written to the file:\n%s", saved)
	}
}

func TestExitCodeSaved(t *testing.T) {
	path := writeSample(t, sample)
	got := invoke(t, []string{path}, answered(questionnaire.StatusSaved, map[string]any{"notes": "partial"}))
	if got.code != exitSaved {
		t.Errorf("exit code = %d, want %d\nstderr: %s", got.code, exitSaved, got.stderr)
	}
	var out questionnaire.Output
	if err := json.Unmarshal([]byte(got.stdout), &out); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, got.stdout)
	}
	if out.Complete {
		t.Error("a questionnaire with a required question outstanding should not report complete")
	}
	if len(out.Unanswered) != 1 || out.Unanswered[0] != "name" {
		t.Errorf("unanswered = %v, want [name]", out.Unanswered)
	}
}

func TestExitCodeDismissed(t *testing.T) {
	path := writeSample(t, sample)
	got := invoke(t, []string{path}, answered(questionnaire.StatusDismissed, map[string]any{"name": "typed then discarded"}))
	if got.code != exitDismissed {
		t.Errorf("exit code = %d, want %d\nstderr: %s", got.code, exitDismissed, got.stderr)
	}

	var out questionnaire.Output
	if err := json.Unmarshal([]byte(got.stdout), &out); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, got.stdout)
	}
	if len(out.Answers) != 0 {
		t.Errorf("a dismissed questionnaire reported answers: %v", out.Answers)
	}
	if out.Status != questionnaire.StatusDismissed {
		t.Errorf("status = %q", out.Status)
	}

	saved, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(saved), "typed then discarded") {
		t.Errorf("a dismissed answer was written to the file:\n%s", saved)
	}
	if !strings.Contains(string(saved), "status: dismissed") {
		t.Errorf("the dismissal was not recorded:\n%s", saved)
	}
}

func TestExitCodeErrorOnMissingFile(t *testing.T) {
	got := invoke(t, []string{filepath.Join(t.TempDir(), "nope.yaml")}, neverCalled(t))
	if got.code != exitError {
		t.Errorf("exit code = %d, want %d", got.code, exitError)
	}
	if !strings.Contains(got.stderr, "nope.yaml") {
		t.Errorf("stderr should name the path: %s", got.stderr)
	}
	if got.stdout != "" {
		t.Errorf("stdout should stay empty on failure: %q", got.stdout)
	}
}

func TestInvalidQuestionnaireOpensNothing(t *testing.T) {
	path := writeSample(t, "title: t\nquestions:\n  - {id: q, type: carousel, prompt: \"What?\"}\n")
	got := invoke(t, []string{path}, neverCalled(t))
	if got.code != exitError {
		t.Errorf("exit code = %d, want %d", got.code, exitError)
	}
	if !strings.Contains(got.stderr, "carousel") {
		t.Errorf("stderr should explain the problem: %s", got.stderr)
	}
	if got.stdout != "" {
		t.Errorf("stdout = %q, want empty", got.stdout)
	}
}

// neverCalled fails the test if the form is opened.
func neverCalled(t *testing.T) presenter {
	t.Helper()
	return func(*questionnaire.Document, ui.Options) (questionnaire.Status, error) {
		t.Error("the form was opened for a questionnaire that should have been rejected first")
		return questionnaire.StatusDismissed, nil
	}
}

func TestUsageErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"no argument", nil, "no questionnaire given"},
		{"two arguments", []string{"a.yaml", "b.yaml"}, "expected one questionnaire, got 2"},
		{"unknown flag", []string{"--wat"}, "flag provided but not defined"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := invoke(t, tc.args, neverCalled(t))
			if got.code != exitError {
				t.Errorf("exit code = %d, want %d", got.code, exitError)
			}
			if !strings.Contains(got.stderr, tc.want) {
				t.Errorf("stderr = %q, want it to contain %q", got.stderr, tc.want)
			}
			if got.stdout != "" {
				t.Errorf("stdout = %q, want empty", got.stdout)
			}
		})
	}
}

func TestPresenterFailureIsReported(t *testing.T) {
	path := writeSample(t, sample)
	got := invoke(t, []string{path}, func(*questionnaire.Document, ui.Options) (questionnaire.Status, error) {
		return "", os.ErrPermission
	})
	if got.code != exitError {
		t.Errorf("exit code = %d, want %d", got.code, exitError)
	}
	if !strings.Contains(got.stderr, "permission denied") {
		t.Errorf("stderr = %q", got.stderr)
	}
}

func TestUnusableOutcomeIsRejected(t *testing.T) {
	path := writeSample(t, sample)
	got := invoke(t, []string{path}, func(*questionnaire.Document, ui.Options) (questionnaire.Status, error) {
		return questionnaire.StatusPending, nil
	})
	if got.code != exitError {
		t.Errorf("exit code = %d, want %d", got.code, exitError)
	}
}

func TestStdoutIsParseableJSONOnly(t *testing.T) {
	path := writeSample(t, sample)
	got := invoke(t, []string{path}, answered(questionnaire.StatusSubmitted, map[string]any{
		"name":  "quizme",
		"notes": "a note\nover two lines",
	}))

	var out questionnaire.Output
	if err := json.Unmarshal([]byte(got.stdout), &out); err != nil {
		t.Fatalf("stdout is not clean JSON: %v\n%s", err, got.stdout)
	}
	if out.Path != path || out.Title != "Sample" {
		t.Errorf("path = %q, title = %q", out.Path, out.Title)
	}
	if !out.Complete {
		t.Error("Complete should be true once the required question is answered")
	}
	if got, want := out.Answers["name"].Answer, "quizme"; got != want {
		t.Errorf("name answer = %#v, want %q", got, want)
	}
	if got := out.Answers["notes"].Answer; got != "a note\nover two lines" {
		t.Errorf("notes answer = %#v", got)
	}
	if out.Answers["name"].Type != questionnaire.TypeText {
		t.Errorf("type not reported: %+v", out.Answers["name"])
	}
	if out.SubmittedAt == "" {
		t.Error("submitted_at should be reported")
	}
}

func TestHiddenAnswersAreNotReported(t *testing.T) {
	path := writeSample(t, `title: t
questions:
  - {id: gate, type: select, prompt: "Gate?", options: [open, shut]}
  - id: gated
    type: text
    prompt: Only when open
    show_if: {gate: open}
`)
	got := invoke(t, []string{path}, answered(questionnaire.StatusSubmitted, map[string]any{
		"gate":  "shut",
		"gated": "should not be reported",
	}))
	var out questionnaire.Output
	if err := json.Unmarshal([]byte(got.stdout), &out); err != nil {
		t.Fatal(err)
	}
	if _, present := out.Answers["gated"]; present {
		t.Errorf("a hidden question was reported: %v", out.Answers)
	}
}

func TestCommentsAreReported(t *testing.T) {
	path := writeSample(t, sample)
	got := invoke(t, []string{path}, func(doc *questionnaire.Document, _ ui.Options) (questionnaire.Status, error) {
		// A comment with no answer must still travel: it is often the most
		// precise thing the responder said.
		doc.Question("notes").Comment = "worth a conversation"
		doc.Question("name").Answer = "quizme"
		return questionnaire.StatusSubmitted, nil
	})
	var out questionnaire.Output
	if err := json.Unmarshal([]byte(got.stdout), &out); err != nil {
		t.Fatal(err)
	}
	entry, present := out.Answers["notes"]
	if !present {
		t.Fatalf("a comment without an answer was dropped: %v", out.Answers)
	}
	if entry.Comment != "worth a conversation" || entry.Answer != nil {
		t.Errorf("entry = %+v", entry)
	}
}

// The icon travels in the binary, so a build that cannot find the drawing fails
// to compile rather than shipping without one. This says the embed carries
// something, which a missing or empty file would not.
func TestTheIconIsCompiledIn(t *testing.T) {
	if len(iconSVG) == 0 {
		t.Fatal("no icon was embedded")
	}
	if !strings.Contains(string(iconSVG), "<svg") {
		t.Error("the embedded icon is not an SVG")
	}
}
