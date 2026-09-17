## Context

See proposal.md for why. Today `formTheme.Size` returns fixed values (14pt body text, 23pt heading, 6/8 padding, and so on), and a few layout constants in `internal/ui` are hard-coded pixel values: `introLeading`, `promptLeading`, `promptTail`, `cardOpticalTail`, and `progressHeight`. The form's title and intro sit in a header that scrolls away with the page. The progress bar and actions sit in a pinned footer. Decisions below come from `questionnaires/text-size.yaml`.

## Goals / Non-Goals

**Goals:**
- One scale factor drives every size on the form, so text and spacing grow together.
- The config file lives in a small package that can be tested without a window.
- A config problem only ever produces a warning.

**Non-Goals:**
- Resizing the window when the scale changes. The page already scrolls.
- Pinch-to-zoom or Ctrl+scroll.
- Any config setting other than text size, though the file format leaves room for more.
- Honouring `XDG_CONFIG_HOME` or the OS-specific config directory. The path is `~/.config/quizme/config.yaml` everywhere, as decided.

## Decisions

**Scale in the theme, not per widget.** `formTheme` gains a `scale float32`, and `Size` multiplies every value it returns, including those it falls through to from the default theme. Fyne reads sizes from the theme on every layout and refresh, so swapping the theme with `app.Settings().SetTheme(newTheme(scale))` rescales the whole window without rebuilding the form. Rebuilding it could lose fold state and focus. The alternative, a `SizeName` on each label, would miss spacing and icons, and every future widget would have to remember it.

**Hard-coded layout constants go through the same factor.** The pixel constants above become values read through a `scaled(px)` helper that uses the current theme's scale. Otherwise leading and tails would stay fixed while text grew, and the careful alignment in `widget_header.go` and `widget_prompt.go` would drift at 200%.

**Size held as an integer percentage.** 70–200 in steps of 10 keeps equality exact, so there are no float comparisons at the limits. The flag and the config file use the same number people see: `text_size: 130`.

**Config package: `internal/config`.** `Load(path) (Config, []warning)` and `Save(path, Config) error`. `Path()` resolves `os.UserHomeDir()/.config/quizme/config.yaml`. Saving reads the existing file into a `yaml.Node`, sets only `text_size`, and writes through a temp file and rename, the same way `questionnaire/splice.go` writes answers, so any other keys and comments survive. A number that is out of range or off-step is clamped to the nearest valid step with a warning. Anything unparseable falls back to 100 with a warning.

**Where the controls live.** The buttons go in the header beside the title, as chosen, so they scroll away with it. The shortcuts work from anywhere on the page, so the header doesn't have to be in view. The buttons are small low-importance "A−" and "A+" buttons, disabled at the limits. There is no reset button, only the reset shortcut, to keep the header quiet. The reset requirement is met by the shortcut.

**Shortcuts registered on the window canvas** as `desktop.CustomShortcut` values for `KeyEqual`, `KeyPlus`, `KeyMinus`, and `Key0` with `fyne.KeyModifierShortcutDefault`, which is Cmd on macOS and Ctrl elsewhere. Canvas shortcuts fire even when an `Entry` has focus, and Fyne does not type a character for a modified key.

**Save on change, from the UI.** `ui.Run` receives the starting percent and an `onTextSize func(int)` callback. `main.go` supplies a callback that calls `config.Save` and writes any error to stderr as a warning. This keeps file I/O out of `internal/ui`, and the tests can pass a recorder. A size change never touches `baseline` or the dirty check, so closing after only a size change still counts as closing with nothing entered.

**Flag precedence.** When `--text-size` is given, it sets the starting size and the config is not read for size, so a broken config gives no warning on a flagged run. Changes made on the form are still saved, because they are the responder's explicit choice.

## Risks / Trade-offs

- [The buttons scroll out of view with the header] → The shortcuts work everywhere. If this proves annoying, moving the buttons to the pinned footer is a small follow-up that doesn't change any requirement except where the buttons are.
- [At 200% a long `select` or rank row may be wider than the 780px window] → Check against `examples/demo.yaml` at 200% during implementation, and fix any clipping with wrapping rather than by capping the scale.
- [The screenshot tests in `internal/ui/testdata` are rendered at the default scale] → They stay at 100%. Add one render test at 150% to catch a constant that doesn't scale.
- [Keyboard layouts where `+` needs Shift] → Accept both `=` and `+` for increase.
