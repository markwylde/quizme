// Package config reads and writes the responder's own preferences, which belong
// to the person at the keyboard rather than to any one questionnaire.
//
// The file is ~/.config/quizme/config.yaml on every platform. It is a
// convenience: nothing in it is ever allowed to stop a questionnaire from being
// answered, so a bad file is reported and worked around rather than refused.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Text sizes are whole percentages of the form's default size. An integer keeps
// the limits exact, and it is the same number the responder sees and types.
const (
	MinTextSize     = 70
	MaxTextSize     = 200
	TextSizeStep    = 10
	DefaultTextSize = 100
)

// textSizeKey is the text size's key in the file.
const textSizeKey = "text_size"

// Config is the responder's preferences.
type Config struct {
	TextSize int
}

// Default is what a responder with no config file gets.
func Default() Config { return Config{TextSize: DefaultTextSize} }

// ValidTextSize reports whether n is a size the form offers.
func ValidTextSize(n int) bool {
	return n >= MinTextSize && n <= MaxTextSize && n%TextSizeStep == 0
}

// ClampTextSize returns the offered size nearest to v.
func ClampTextSize(v float64) int {
	n := int(math.Round(v/TextSizeStep)) * TextSizeStep
	return min(max(n, MinTextSize), MaxTextSize)
}

// Path is where the config lives: ~/.config/quizme/config.yaml.
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not find the home directory: %w", err)
	}
	return filepath.Join(home, ".config", "quizme", "config.yaml"), nil
}

// Load reads the config at path.
//
// A missing file is the ordinary first run and is not worth mentioning. Anything
// else wrong with the file comes back as warnings alongside a usable config, so
// the caller can say what happened and carry on.
func Load(path string) (Config, []error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, []error{fmt.Errorf("%s could not be read, so the default text size is used: %w", path, err)}
	}

	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return cfg, []error{fmt.Errorf("%s is not valid YAML, so the default text size is used: %w", path, err)}
	}

	v, ok := raw[textSizeKey]
	if !ok || v == nil {
		return cfg, nil
	}
	var n float64
	switch t := v.(type) {
	case int:
		n = float64(t)
	case float64:
		n = t
	default:
		return cfg, []error{fmt.Errorf("%s: %s should be a number from %d to %d, got %v, so the default is used",
			path, textSizeKey, MinTextSize, MaxTextSize, v)}
	}
	if n == math.Trunc(n) && ValidTextSize(int(n)) {
		cfg.TextSize = int(n)
		return cfg, nil
	}
	cfg.TextSize = ClampTextSize(n)
	return cfg, []error{fmt.Errorf("%s: %s %v is not one of the offered sizes (%d to %d in steps of %d), so %d is used",
		path, textSizeKey, v, MinTextSize, MaxTextSize, TextSizeStep, cfg.TextSize)}
}

// Save writes cfg to path, creating the directory if it has to.
//
// Only the keys this package owns are touched: anything else already in the
// file, comments included, is kept. The file is replaced in one step, so a
// crash mid-write never leaves half a config behind.
func Save(path string, cfg Config) error {
	var doc yaml.Node
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return fmt.Errorf("%s is not valid YAML, so it was left alone: %w", path, err)
		}
	case errors.Is(err, fs.ErrNotExist):
	default:
		return fmt.Errorf("could not read %s: %w", path, err)
	}

	if doc.Kind == 0 {
		doc = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode}}}
	}
	body := doc.Content[0]
	if body.Kind != yaml.MappingNode {
		return fmt.Errorf("%s does not hold a mapping, so it was left alone", path)
	}
	setInt(body, textSizeKey, cfg.TextSize)

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&doc); err != nil {
		return fmt.Errorf("could not encode %s: %w", path, err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("could not encode %s: %w", path, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("could not create %s: %w", filepath.Dir(path), err)
	}
	return atomicWrite(path, buf.Bytes())
}

// setInt sets key to n in a mapping, in place if the key is already there.
func setInt(mapping *yaml.Node, key string, n int) {
	value := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: fmt.Sprint(n)}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			value.LineComment = mapping.Content[i+1].LineComment
			mapping.Content[i+1] = value
			return
		}
	}
	mapping.Content = append(mapping.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, value)
}

func atomicWrite(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".config-*.tmp")
	if err != nil {
		return fmt.Errorf("could not write %s: %w", path, err)
	}
	name := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(name)
		return fmt.Errorf("could not write %s: %w", path, err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return fmt.Errorf("could not write %s: %w", path, err)
	}
	if err := os.Rename(name, path); err != nil {
		os.Remove(name)
		return fmt.Errorf("could not replace %s: %w", path, err)
	}
	return nil
}
