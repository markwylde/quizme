package main

import (
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/markwylde/interrogate/internal/questionnaire"
)

// The README shows a questionnaire before and after it is answered, and claims
// that everything outside the four owned keys is untouched. That claim is the
// whole point of the tool, so it is checked rather than trusted.

func readmeBlocks(t *testing.T) []string {
	t.Helper()
	raw, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile("(?s)```yaml\n(.*?)```")
	var out []string
	for _, m := range re.FindAllStringSubmatch(string(raw), -1) {
		out = append(out, m[1])
	}
	return out
}

func TestReadmeBeforeAndAfterAreExact(t *testing.T) {
	blocks := readmeBlocks(t)
	if len(blocks) < 2 {
		t.Fatalf("expected a before and an after example, found %d yaml blocks", len(blocks))
	}
	before, after := blocks[0], blocks[1]

	doc, err := questionnaire.LoadBytes("README.yaml", []byte(before))
	if err != nil {
		t.Fatalf("the README's example does not load: %v", err)
	}

	set(t, doc, "storage", "in-place")
	doc.Question("storage").Comment = "agreed, drift is the real risk"
	set(t, doc, "priorities", []any{"correctness", "looks", "speed"})

	at := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	got, err := doc.Render(questionnaire.StatusSubmitted, at)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if string(got) != after {
		t.Errorf("the README's answered example is not what the tool produces\n--- README says ---\n%s\n--- tool produces ---\n%s", after, got)
	}
}

func TestReadmeExampleLeavesHiddenQuestionsAlone(t *testing.T) {
	// The README points out that storage_why records nothing because it was
	// never shown.
	after := readmeBlocks(t)[1]
	doc, err := questionnaire.LoadBytes("README.yaml", []byte(after))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Question("storage_why").HasAnswer() {
		t.Error("the hidden question should have no answer in the answered example")
	}
	if !strings.Contains(after, "storage_why") {
		t.Error("the hidden question should still be present in the file")
	}
}

func TestReadmeDocumentsEveryTypeAndExitCode(t *testing.T) {
	raw, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	doc := string(raw)
	for _, ty := range questionnaire.AllTypes {
		if !strings.Contains(doc, "`"+string(ty)+"`") {
			t.Errorf("the README does not document the %q type", ty)
		}
	}
	for _, want := range []string{"`0`", "`1`", "`2`", "`3`"} {
		if !strings.Contains(doc, want) {
			t.Errorf("the README does not document exit code %s", want)
		}
	}
}

func set(t *testing.T, doc *questionnaire.Document, id string, v any) {
	t.Helper()
	q := doc.Question(id)
	if q == nil {
		t.Fatalf("no question %q", id)
	}
	normalized, err := questionnaire.NormalizeAnswer(q, v)
	if err != nil {
		t.Fatalf("%s: %v", id, err)
	}
	q.Answer = normalized
}
