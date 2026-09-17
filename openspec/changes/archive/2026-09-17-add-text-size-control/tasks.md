## 1. User config

- [x] 1.1 Create `internal/config` with `Path()`, `Load`, and `Save` for `text_size`, and verify unit tests cover a missing file, a valid value, an out-of-range or off-step value clamped with a warning, and unparseable YAML falling back to 100 with a warning
- [x] 1.2 Make `Save` create `~/.config/quizme/`, keep other keys and comments, and write atomically, and verify with a test that saves into a temp home dir over a file with an extra key and a comment
- [x] 1.3 Verify with a test that `Save` on an unwritable path returns an error rather than panicking

## 2. Scalable theme

- [x] 2.1 Give `formTheme` an integer percent scale that multiplies every `Size` result, including fall-through sizes, and verify with a test that body text is 14 at 100% and 28 at 200%, and padding scales too
- [x] 2.2 Route `introLeading`, `promptLeading`, `promptTail`, `cardOpticalTail`, and `progressHeight` through the current scale, and verify the existing alignment and padding tests still pass at 100%
- [x] 2.3 Add a render test of `examples/demo.yaml` at 150%, and verify the header and prompt alignment tests hold at that scale

## 3. Controls on the form

- [x] 3.1 Add A− / A+ buttons to the header beside the title, and verify with a UI test that they step the size by 10 and are disabled at 70 and 200
- [x] 3.2 Register the increase (`=` and `+`), decrease, and reset shortcuts with the default shortcut modifier on the window canvas, and verify with a UI test that each changes the size, including while a text entry has focus without inserting a character
- [x] 3.3 Apply a size change by swapping the theme rather than rebuilding the form, and verify with a test that answers, comments, and fold states are unchanged after a change
- [x] 3.4 Keep size changes out of the dirty check, and verify with a test that closing after only a size change dismisses without a save-or-discard prompt
- [x] 3.5 Call an `onTextSize` callback passed to `ui.Run` on every change, and verify with a test that a recorder sees each new percent

## 4. Command line

- [x] 4.1 Add `--text-size <percent>` to `parseArgs` and the usage text, and verify with `main_test.go` cases that 130 is accepted and that 60, 210, 125, and `abc` exit with a usage error naming the range and step
- [x] 4.2 In `main.go`, pick the starting size from the flag or `config.Load`, write config warnings to stderr, and pass a callback that saves and warns on failure, and verify with tests that stdout JSON and exit codes are unchanged when the config is broken or unwritable
- [x] 4.3 Verify with a test that a flagged run does not write the config unless the size is changed on the form

## 5. Docs and end to end

- [x] 5.1 Document the buttons, shortcuts, `--text-size`, and `~/.config/quizme/config.yaml` in `README.md`, and verify `readme_test.go` still passes
- [x] 5.2 Run `go test ./...`, `go vet ./...`, and `gofmt -l .`, and verify all are clean
- [x] 5.3 Build and run `quizme examples/demo.yaml` on a scratch copy, raise the size to 130%, dismiss, then run again and verify it opens at 130% and the config file holds `text_size: 130`
