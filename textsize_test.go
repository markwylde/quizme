package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/markwylde/quizme/internal/questionnaire"
	"github.com/markwylde/quizme/internal/ui"
)

func TestMain(m *testing.M) {
	// Every run reads the responder's config, so no test may see or touch the
	// real one.
	home, err := os.MkdirTemp("", "quizme-home-")
	if err != nil {
		panic(err)
	}
	os.Setenv("HOME", home)
	os.Setenv("USERPROFILE", home)
	code := m.Run()
	os.RemoveAll(home)
	os.Exit(code)
}

// withHome gives a test a home directory of its own, and returns where its
// config file would be.
func withHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	return filepath.Join(home, ".config", "quizme", "config.yaml")
}

func writeConfig(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// sized is a presenter that records the size the form was opened at, and
// optionally changes it the way a responder would, before submitting.
func sized(opened *int, changeTo ...int) presenter {
	return func(doc *questionnaire.Document, opts ui.Options) (questionnaire.Status, error) {
		*opened = opts.TextSize
		for _, n := range changeTo {
			opts.OnTextSize(n)
		}
		doc.Question("name").Answer = "quizme"
		return questionnaire.StatusSubmitted, nil
	}
}

func TestTextSizeFlagIsAccepted(t *testing.T) {
	withHome(t)
	var opened int
	got := invoke(t, []string{"--text-size", "130", writeSample(t, sample)}, sized(&opened))
	if got.code != exitSubmitted || opened != 130 {
		t.Errorf("exit %d, opened at %d; want 0 at 130\nstderr: %s", got.code, opened, got.stderr)
	}
}

func TestTextSizeFlagRejectsUnofferedSizes(t *testing.T) {
	for _, v := range []string{"60", "210", "125", "abc", "130.0"} {
		t.Run(v, func(t *testing.T) {
			got := invoke(t, []string{"--text-size", v, "q.yaml"}, neverCalled(t))
			if got.code != exitError {
				t.Errorf("exit code = %d, want %d", got.code, exitError)
			}
			if !strings.Contains(got.stderr, "70 to 200 in steps of 10") {
				t.Errorf("stderr should name the range and step: %s", got.stderr)
			}
			if got.stdout != "" {
				t.Errorf("stdout = %q", got.stdout)
			}
		})
	}
}

func TestTheSavedSizeIsUsedWithoutTheFlag(t *testing.T) {
	cfg := withHome(t)
	writeConfig(t, cfg, "text_size: 120\n")
	var opened int
	got := invoke(t, []string{writeSample(t, sample)}, sized(&opened))
	if opened != 120 || got.stderr != "" {
		t.Errorf("opened at %d with stderr %q, want 120 and no warnings", opened, got.stderr)
	}
}

func TestNoConfigOpensAtTheDefaultAndWritesNothing(t *testing.T) {
	cfg := withHome(t)
	var opened int
	invoke(t, []string{writeSample(t, sample)}, sized(&opened))
	if opened != 100 {
		t.Errorf("opened at %d, want 100", opened)
	}
	if _, err := os.Stat(cfg); !os.IsNotExist(err) {
		t.Errorf("a config was created though the size never changed (stat: %v)", err)
	}
}

func TestAChangedSizeIsRemembered(t *testing.T) {
	cfg := withHome(t)
	var opened int
	invoke(t, []string{writeSample(t, sample)}, sized(&opened, 110, 130))
	raw, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatalf("the size was not saved: %v", err)
	}
	if string(raw) != "text_size: 130\n" {
		t.Errorf("config = %q", raw)
	}

	invoke(t, []string{writeSample(t, sample)}, sized(&opened))
	if opened != 130 {
		t.Errorf("the next run opened at %d, want 130", opened)
	}
}

func TestASizeChangedThenDismissedIsStillRemembered(t *testing.T) {
	cfg := withHome(t)
	invoke(t, []string{writeSample(t, sample)}, func(_ *questionnaire.Document, opts ui.Options) (questionnaire.Status, error) {
		opts.OnTextSize(150)
		return questionnaire.StatusDismissed, nil
	})
	if raw, _ := os.ReadFile(cfg); string(raw) != "text_size: 150\n" {
		t.Errorf("config = %q", raw)
	}
}

func TestAFlaggedRunLeavesTheSavedSizeAlone(t *testing.T) {
	cfg := withHome(t)
	writeConfig(t, cfg, "text_size: 120\n")
	var opened int
	invoke(t, []string{"--text-size", "150", writeSample(t, sample)}, sized(&opened))
	if opened != 150 {
		t.Errorf("opened at %d, want 150", opened)
	}
	if raw, _ := os.ReadFile(cfg); string(raw) != "text_size: 120\n" {
		t.Errorf("a flagged run rewrote the config: %q", raw)
	}

	// A change made on the form is the responder's own choice, flag or not.
	invoke(t, []string{"--text-size", "150", writeSample(t, sample)}, sized(&opened, 160))
	if raw, _ := os.ReadFile(cfg); string(raw) != "text_size: 160\n" {
		t.Errorf("config = %q, want the change made on the form", raw)
	}
}

func TestABrokenConfigOnlyWarns(t *testing.T) {
	for body, want := range map[string]int{"text_size: [\n": 100, "text_size: 125\n": 130} {
		cfg := withHome(t)
		writeConfig(t, cfg, body)
		var opened int
		path := writeSample(t, sample)
		got := invoke(t, []string{path}, sized(&opened))
		if opened != want {
			t.Errorf("%q: opened at %d, want %d", body, opened, want)
		}
		if !strings.Contains(got.stderr, "warning") || !strings.Contains(got.stderr, cfg) {
			t.Errorf("%q: stderr should warn and name the file: %q", body, got.stderr)
		}
		clean := invoke(t, []string{writeSample(t, sample)}, answered(questionnaire.StatusSubmitted, map[string]any{"name": "quizme"}))
		if got.code != clean.code || !strings.HasPrefix(got.stdout, "{") {
			t.Errorf("%q: exit %d stdout %q; a config problem changed the outcome", body, got.code, got.stdout)
		}
	}
}

func TestAnUnwritableConfigOnlyWarns(t *testing.T) {
	cfg := withHome(t)
	// A file where the config directory should be.
	writeConfig(t, filepath.Dir(filepath.Dir(cfg))+"/quizme", "")
	var opened int
	got := invoke(t, []string{writeSample(t, sample)}, sized(&opened, 140))
	if got.code != exitSubmitted {
		t.Errorf("exit code = %d, want %d", got.code, exitSubmitted)
	}
	if !strings.Contains(got.stderr, "could not be remembered") {
		t.Errorf("stderr = %q", got.stderr)
	}
	if !strings.Contains(got.stdout, `"status": "submitted"`) {
		t.Errorf("stdout = %q", got.stdout)
	}
}
