package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/markwylde/interrogate/internal/questionnaire"
)

// The skill tells an agent how to write and run a questionnaire. These tests
// hold the documentation to the same standard as the code: the example it shows
// must load, and the outcomes it describes must be the ones that happen.

func skillText(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("skills", "interrogate", "SKILL.md"))
	if err != nil {
		t.Fatalf("the skill is missing: %v", err)
	}
	return string(raw)
}

// fencedYAML pulls the yaml code blocks out of the skill.
func fencedYAML(doc string) []string {
	re := regexp.MustCompile("(?s)```yaml\n(.*?)```")
	var out []string
	for _, m := range re.FindAllStringSubmatch(doc, -1) {
		out = append(out, m[1])
	}
	return out
}

func TestSkillExampleIsAValidQuestionnaire(t *testing.T) {
	blocks := fencedYAML(skillText(t))
	if len(blocks) == 0 {
		t.Fatal("the skill shows no yaml example")
	}
	for i, block := range blocks {
		doc, err := questionnaire.LoadBytes("skill-example.yaml", []byte(block))
		if err != nil {
			t.Fatalf("yaml example %d in the skill does not load:\n%v\n---\n%s", i, err, block)
		}
		if len(doc.Questions) == 0 {
			t.Errorf("yaml example %d has no questions", i)
		}
	}
}

func TestSkillExampleBehavesAsDocumented(t *testing.T) {
	block := fencedYAML(skillText(t))[0]
	path := filepath.Join(t.TempDir(), "scope.yaml")
	if err := os.WriteFile(path, []byte(block), 0o644); err != nil {
		t.Fatal(err)
	}

	// The gated question is hidden until its condition holds, exactly as the
	// skill claims.
	doc, err := questionnaire.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Evaluate().Visible("storage_why") {
		t.Error("storage_why should start hidden")
	}

	got := invoke(t, []string{path}, answered(questionnaire.StatusSubmitted, map[string]any{
		"storage": "sidecar",
	}))
	if got.code != exitSubmitted {
		t.Fatalf("exit code = %d\nstderr: %s", got.code, got.stderr)
	}

	// Answering the gate reveals the follow-up, and the file records the
	// outcome the skill's table describes.
	answered, err := questionnaire.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !answered.Evaluate().Visible("storage_why") {
		t.Error("storage_why should be visible once sidecar is chosen")
	}
	if answered.Status != questionnaire.StatusSubmitted {
		t.Errorf("status = %q, want submitted", answered.Status)
	}
	if answered.SubmittedAt == "" {
		t.Error("submitted_at should be recorded")
	}
}

func TestSkillDocumentsEveryTypeAndStatus(t *testing.T) {
	doc := skillText(t)
	for _, ty := range questionnaire.AllTypes {
		if !strings.Contains(doc, "`"+string(ty)+"`") {
			t.Errorf("the skill does not mention the %q type", ty)
		}
	}
	for _, status := range []questionnaire.Status{
		questionnaire.StatusPending,
		questionnaire.StatusSubmitted,
		questionnaire.StatusSaved,
		questionnaire.StatusDismissed,
	} {
		if !strings.Contains(doc, string(status)) {
			t.Errorf("the skill does not explain the %q status", status)
		}
	}
}

func TestSkillDocumentsTheExitCodes(t *testing.T) {
	doc := skillText(t)
	for _, want := range []string{"`0` submitted", "`1` error", "`2` dismissed", "`3` saved"} {
		if !strings.Contains(doc, want) {
			t.Errorf("the skill does not document %s", want)
		}
	}
}

func TestSkillCoversTheThingsThatGoWrong(t *testing.T) {
	doc := skillText(t)
	required := map[string]string{
		"the single-question exception":  "only one question",
		"the branching exception":        "genuinely branches",
		"the path convention":            "questionnaires/",
		"running in the background":      "background",
		"status recovery":                "lose track",
		"comments outweighing an option": "more precise statement",
		"the not-installed fallback":     "go install github.com/markwylde/interrogate",
		"no desktop session":             "desktop session",
	}
	for name, needle := range required {
		if !strings.Contains(doc, needle) {
			t.Errorf("the skill does not cover %s (looked for %q)", name, needle)
		}
	}
}

func TestSkillDocumentsValidateHonestly(t *testing.T) {
	doc := skillText(t)
	if !strings.Contains(doc, "--validate") {
		t.Error("the skill does not mention --validate")
	}
	// The obvious reading of the flag describes what a plain run already does,
	// so the skill has to say what it is actually for, or an agent will run it
	// before every invocation for no reason.
	if !strings.Contains(doc, "not a prerequisite") {
		t.Error("the skill does not say that --validate is optional before a run")
	}
	if !strings.Contains(doc, "validates first too") {
		t.Error("the skill does not explain that a plain run validates as well")
	}
}
