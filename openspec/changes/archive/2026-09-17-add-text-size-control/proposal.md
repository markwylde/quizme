## Why

The form's text sizes are fixed in the theme, so someone who finds 14pt body text too small or too large has no way to change it. The form is read carefully before it is answered, so reading comfort matters, and a size that has to be picked again on every run is not really a setting.

## What Changes

- A text size control on the form: visible A− / A+ buttons in the form's header, plus keyboard shortcuts (Cmd on macOS, Ctrl elsewhere, with `+` / `-`, and `0` to reset).
- The size scales all text and the spacing around it together, in 10% steps from 70% to 200%, with 100% as the default.
- The chosen size is saved as soon as it changes to `~/.config/quizme/config.yaml`, and every later run starts at that size. It is saved even if the form is then dismissed, because it is a preference, not an answer.
- A new `--text-size <percent>` flag sets the size for a single run without changing the saved preference.
- The questionnaire file, the JSON on stdout, and the exit codes are unaffected.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `quizme-cli`: command invocation gains the `--text-size` flag; new requirements cover changing the text size on the form, remembering it in the user's config, and keeping it separate from the questionnaire's outcome.

## Impact

- `main.go`: parse and validate `--text-size`, update the usage text, and pass the starting size to the UI.
- `internal/ui`: the theme scales its sizes by a factor, and the form gains the header buttons and the shortcuts.
- A new small package reads and writes the user config file.
- There are no new dependencies (`gopkg.in/yaml.v3` is already used).
- `README.md`: document the control, the flag, and the config file.
