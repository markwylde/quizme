package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadMissingFileIsTheDefaultWithoutWarning(t *testing.T) {
	cfg, warnings := Load(filepath.Join(t.TempDir(), "nope", "config.yaml"))
	if cfg.TextSize != DefaultTextSize || len(warnings) != 0 {
		t.Errorf("got %+v, %v; want the default and no warnings", cfg, warnings)
	}
}

func TestLoadValidSize(t *testing.T) {
	cfg, warnings := Load(write(t, "text_size: 130\n"))
	if cfg.TextSize != 130 || len(warnings) != 0 {
		t.Errorf("got %+v, %v", cfg, warnings)
	}
}

func TestLoadFileWithoutTheKey(t *testing.T) {
	cfg, warnings := Load(write(t, "other: true\n"))
	if cfg.TextSize != DefaultTextSize || len(warnings) != 0 {
		t.Errorf("got %+v, %v", cfg, warnings)
	}
}

func TestLoadClampsAnUnofferedSize(t *testing.T) {
	for body, want := range map[string]int{
		"text_size: 500\n":   200,
		"text_size: 10\n":    70,
		"text_size: 124\n":   120,
		"text_size: 126.5\n": 130,
	} {
		path := write(t, body)
		cfg, warnings := Load(path)
		if cfg.TextSize != want {
			t.Errorf("%q: size %d, want %d", body, cfg.TextSize, want)
		}
		if len(warnings) != 1 || !strings.Contains(warnings[0].Error(), path) {
			t.Errorf("%q: want one warning naming the file, got %v", body, warnings)
		}
	}
}

func TestLoadUnparseableFallsBackToDefault(t *testing.T) {
	for _, body := range []string{"text_size: [\n", "text_size: big\n", "- a list\n"} {
		path := write(t, body)
		cfg, warnings := Load(path)
		if cfg.TextSize != DefaultTextSize {
			t.Errorf("%q: size %d, want the default", body, cfg.TextSize)
		}
		if len(warnings) != 1 || !strings.Contains(warnings[0].Error(), path) {
			t.Errorf("%q: want one warning naming the file, got %v", body, warnings)
		}
	}
}

func TestSaveCreatesTheDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".config", "quizme", "config.yaml")
	if err := Save(path, Config{TextSize: 130}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "text_size: 130\n" {
		t.Errorf("file = %q", got)
	}
	if cfg, _ := Load(path); cfg.TextSize != 130 {
		t.Errorf("reloaded size %d", cfg.TextSize)
	}
}

func TestSaveKeepsOtherKeysAndComments(t *testing.T) {
	path := write(t, "# my settings\nother: keep me # really\ntext_size: 90\n")
	if err := Save(path, Config{TextSize: 150}); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	for _, want := range []string{"# my settings", "other: keep me # really", "text_size: 150"} {
		if !strings.Contains(string(got), want) {
			t.Errorf("file lost %q:\n%s", want, got)
		}
	}
	if strings.Contains(string(got), "90") {
		t.Errorf("old size still present:\n%s", got)
	}
}

func TestSaveUnwritableReturnsAnError(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "quizme")
	if err := os.WriteFile(blocker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	// A file where the directory should be cannot be written through.
	if err := Save(filepath.Join(blocker, "config.yaml"), Config{TextSize: 120}); err == nil {
		t.Error("want an error saving beneath a file")
	}
}

func TestSaveLeavesAnUnparseableFileAlone(t *testing.T) {
	path := write(t, "text_size: [\n")
	if err := Save(path, Config{TextSize: 120}); err == nil {
		t.Error("want an error rather than overwriting a file that could not be read")
	}
	got, _ := os.ReadFile(path)
	if string(got) != "text_size: [\n" {
		t.Errorf("file changed to %q", got)
	}
}

func TestValidTextSize(t *testing.T) {
	for n, want := range map[int]bool{60: false, 70: true, 125: false, 130: true, 200: true, 210: false} {
		if ValidTextSize(n) != want {
			t.Errorf("ValidTextSize(%d) = %v", n, !want)
		}
	}
}
