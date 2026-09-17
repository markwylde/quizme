## Why

At larger text sizes the footer broke: the progress text ran under the action buttons, and its two labels no longer shared a baseline. The confirmation dialogs also shrank to the narrowest width their words allowed, so a one-sentence warning wrapped to a word or two per line.

## What Changes

- The footer's progress and outstanding-required text become one status line that wraps.
- The actions share the status line's row only while all of it fits. Otherwise they move to their own row beneath it, at the trailing edge.
- The footer's edges line up with the question cards.
- Every dialog sets its text at a reading width that follows the text size and never exceeds the window. The unsaved-answers dialog's buttons move to the trailing edge, like the others.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `quizme-cli`: new requirements that the footer and dialogs stay legible at every text size.

## Impact

- `internal/ui/form.go`: footer assembly, status text, and dialog bodies.
- `internal/ui/widget_footer.go`: the new footer layout.
- Tests: `footer_test.go`, and a dialog-width test in `clear_test.go`. Footer text assertions in `form_test.go` are updated for the combined status line.
