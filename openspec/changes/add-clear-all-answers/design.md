## Context

See proposal.md for why. The form keeps each question's answer and comment on `questionnaire.Question`. `dirty()` compares them against the `baseline` snapshot taken when the form opened, and treats a touched `rank` as a change. Controls report changes through `f.set`, which also drives `settle`, folding, visibility, and the progress summary. The `building` flag suppresses the fold side effects of programmatic updates while the page is built. The close prompt is an existing custom dialog in `requestClose`.

## Goals / Non-Goals

**Goals:**
- A clear leaves the form in the same state as a newly opened, unanswered copy of the questionnaire.
- `dirty()` keeps working without special cases, so closing still behaves correctly.

**Non-Goals:**
- Undo after confirming.
- Clearing a single question.
- Writing the cleared state to the file straight away.

## Decisions

**The button goes in the pinned footer, left of Dismiss.** The footer is reachable from anywhere on the page, which the requirement asks for. The button uses low importance so it doesn't compete with Submit, and it sits beside the other form-level actions rather than in the header, which scrolls away.

**Confirmation uses a custom dialog in the style of `requestClose`.** The message is "This will wipe all answers and comments. Are you sure?", with a "Cancel" button and a danger-importance "Clear all" button. A custom dialog rather than `dialog.ShowConfirm` keeps the wording and button styling consistent with the existing close prompt.

**Clear the model, then rebuild the page, rather than resetting each control in place.** Each control type (select, radio, checks, entry, number, boolean, scale, rank, comment) has its own empty state and callback quirks. For example, Fyne's `SetSelected` fires `OnChanged`. Resetting every widget kind in place risks leaving one showing a stale value. Instead, `clearAll` sets every `Question.Answer` to nil and `Comment` to "", then calls `win.SetContent(f.build())`. `build` already produces the correct page from the model under `building`, including folds, visibility, and progress. `baseline` is not touched, so `dirty()` naturally compares the cleared state against what was in the file. Rebuilding also drops the old rank widgets, so no stale `Touched()` state is left behind. The scroll offset returns to the top, which suits a fresh start.

**Hidden answers are cleared by walking `doc.Questions`, not the visible cards,** so a hidden question can't keep an answer it would reveal later.

**The button is always enabled.** Disabling it when nothing is answered would need another sync point on every change, for little benefit. Confirming a clear on an empty form is harmless.

## Risks / Trade-offs

- [Rebuilding loses keyboard focus and scroll position] → This is acceptable after a deliberate reset. The page starts at the top.
- [If the text size change lands first, a rebuild must keep the current scale] → The scale lives in the app theme, not the page, so a rebuild picks it up. The tasks include a check if both changes are applied.
- [State held only in a widget, not the model, would survive a rebuild incorrectly] → The rebuild creates new widgets, so nothing old survives. The `f.cards` map is rebuilt too.
