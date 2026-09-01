package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The skill exists twice: once as the plugin payload other repositories install,
// and once under .claude/ so this repository's own agents see it without
// installing the plugin into itself. The payload is authoritative. These tests
// are what stops the two drifting, which would mean stale instructions in every
// repository that installed it and nothing to notice by.

const (
	skillSource = "skills/interrogate/SKILL.md"
	skillCopy   = ".claude/skills/interrogate/SKILL.md"
	pluginJSON  = ".claude-plugin/plugin.json"
	marketJSON  = ".claude-plugin/marketplace.json"
)

func TestSkillCopiesAgree(t *testing.T) {
	source, err := os.ReadFile(skillSource)
	if err != nil {
		t.Fatalf("the authoritative skill is missing: %v", err)
	}
	copied, err := os.ReadFile(skillCopy)
	if err != nil {
		t.Fatalf("the local copy is missing; run `make skill`: %v", err)
	}
	if string(source) != string(copied) {
		t.Errorf("%s and %s have diverged.\n%s is authoritative — run `make skill` to regenerate the copy.",
			skillSource, skillCopy, skillSource)
	}
}

type pluginManifest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Author      struct {
		Name string `json:"name"`
	} `json:"author"`
	Homepage string `json:"homepage"`
}

func readPlugin(t *testing.T) pluginManifest {
	t.Helper()
	raw, err := os.ReadFile(pluginJSON)
	if err != nil {
		t.Fatalf("the plugin manifest is missing: %v", err)
	}
	var m pluginManifest
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("%s is not valid JSON: %v", pluginJSON, err)
	}
	return m
}

func TestPluginManifestIsComplete(t *testing.T) {
	m := readPlugin(t)
	if m.Name != "interrogate" {
		t.Errorf("name = %q, want %q", m.Name, "interrogate")
	}
	if m.Description == "" {
		t.Error("the manifest has no description; it is what someone reads before installing")
	}
	if m.Version == "" {
		t.Error("the manifest has no version, so an installed copy cannot be identified")
	}
	if m.Author.Name == "" {
		t.Error("the manifest has no author")
	}
}

// frontmatterVersion pulls metadata.version out of the skill's frontmatter.
var frontmatterVersion = regexp.MustCompile(`(?m)^\s+version:\s*"?([0-9][0-9A-Za-z.\-]*)"?\s*$`)

func TestSkillVersionMatchesTheManifest(t *testing.T) {
	// Bumping one without the other would have an installed copy reporting a
	// version that does not match what it actually carries.
	raw, err := os.ReadFile(skillSource)
	if err != nil {
		t.Fatal(err)
	}
	match := frontmatterVersion.FindStringSubmatch(string(raw))
	if match == nil {
		t.Fatalf("%s has no metadata.version in its frontmatter", skillSource)
	}
	if got, want := match[1], readPlugin(t).Version; got != want {
		t.Errorf("the skill says version %q and %s says %q; bump both", got, pluginJSON, want)
	}
}

func TestMarketplaceListsThePlugin(t *testing.T) {
	raw, err := os.ReadFile(marketJSON)
	if err != nil {
		t.Fatalf("the marketplace manifest is missing: %v", err)
	}
	var m struct {
		Name    string                `json:"name"`
		Owner   struct{ Name string } `json:"owner"`
		Plugins []struct {
			Name        string `json:"name"`
			Source      string `json:"source"`
			Description string `json:"description"`
		} `json:"plugins"`
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("%s is not valid JSON: %v", marketJSON, err)
	}
	if m.Name == "" || m.Owner.Name == "" {
		t.Errorf("the marketplace needs a name and an owner, got %+v", m)
	}
	if len(m.Plugins) != 1 {
		t.Fatalf("the marketplace lists %d plugins, want exactly 1", len(m.Plugins))
	}
	if got, want := m.Plugins[0].Name, readPlugin(t).Name; got != want {
		t.Errorf("the marketplace lists %q but the manifest is for %q", got, want)
	}
	// The plugin's files sit at the repository root, so the source points there.
	if got := m.Plugins[0].Source; got != "./" {
		t.Errorf("source = %q, want %q for a single-plugin repository", got, "./")
	}
}

func TestSkillIsWhereThePluginLoaderLooks(t *testing.T) {
	// A plugin's skills live at skills/<name>/SKILL.md relative to its root.
	if _, err := os.Stat(filepath.Join("skills", "interrogate", "SKILL.md")); err != nil {
		t.Errorf("the payload is not where the loader looks: %v", err)
	}
}

func TestSkillFrontmatterNamesTheSkill(t *testing.T) {
	raw, err := os.ReadFile(skillSource)
	if err != nil {
		t.Fatal(err)
	}
	head, _, found := strings.Cut(strings.TrimPrefix(string(raw), "---\n"), "\n---")
	if !found {
		t.Fatal("the skill has no frontmatter block")
	}
	if !strings.Contains(head, "name: interrogate") {
		t.Errorf("the skill's frontmatter does not name it `interrogate`:\n%s", head)
	}
	if !strings.Contains(head, "description:") {
		t.Error("the skill's frontmatter has no description, so nothing will know when to use it")
	}
}
