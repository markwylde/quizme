## Why

Starting a questionnaire over means clearing every field and comment by hand, or editing the YAML. That is slow on a long form, and easy to get wrong when someone reopens a questionnaire that was already answered and wants a clean pass.

## What Changes

- A "Clear all answers" button in the form's pinned footer.
- Clicking it asks for confirmation: "This will wipe all answers and comments. Are you sure?" Nothing is cleared unless the responder confirms.
- Confirming returns every question to unanswered, empties every comment, and puts every question back to its default, expanded fold state. This includes answers on questions that `show_if` currently hides.
- Clearing only changes the form. Nothing is written until the responder submits or saves, and closing afterwards goes through the usual save-or-discard prompt.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `quizme-cli`: a new requirement for clearing all answers and comments from the form, with confirmation.

## Impact

- `internal/ui/form.go`: the footer button, the confirmation dialog, and a reset pass over the cards that goes through the existing control and visibility paths.
- `internal/ui/widget_rank.go`, `widget_comment.go`, and the other controls may each need a way to return to their empty state.
- There are no changes to the questionnaire format, the file writer, stdout, or exit codes.
- `README.md`: mention the button.
