## 1. Answer summaries

- [x] 1.1 Add `answerSummary(q *questionnaire.Question) string` in the ui package, covering every question type — option text for `select`, `Yes`/`No` for `boolean`, the number for `scale` and `number`, shortened text for `text` and `textarea`, comma-joined options for `multiselect`, numbered order for `rank`, empty for unanswered — and verify a table test asserts one line per type
- [x] 1.2 Reuse the existing `preview` helper for shortening, and verify a summary longer than the limit is truncated with an ellipsis and contains no newline

## 2. The header widget

- [x] 2.1 Add `internal/ui/widget_header.go` with a `questionHeader` widget laying out chevron, bold wrapping prompt, the `Required` marker, the answer summary, and the completion tick, and verify a test builds one and finds each part in its object tree
- [x] 2.2 Draw the tick from `palette.success` and re-read the colour in `Refresh`, and verify a test that switches the theme variant sees the tick colour change rather than keeping the one it was built with
- [x] 2.3 Show the tick only when the question carries an answer, and verify a header built for an unanswered question has no tick and one built for an answered question does
- [x] 2.4 Make the header report that its question carries a comment, and verify a header for a question with a comment shows the indicator and one without does not
- [x] 2.5 Implement `fyne.Tappable` and `desktop.Hoverable` on the header, and verify tapping it calls the toggle callback and tapping the prompt label inside it does too

## 3. Splitting the card

- [x] 3.1 Split `buildCard` into `card.header` and `card.body`, with the heading moving into the header and the help text, control, comment field and warning staying in the body, and verify the existing card and polish tests pass once updated for the new tree
- [x] 3.2 Add the three-valued `card.fold` and `card.pinned`, plus `applyFold` showing or hiding `card.body`, refreshing the `cardBox` and then `f.list`, and verify folding a card hides its control and comment field while leaving both objects in the tree
- [x] 3.3 Fold a question that arrives already answered, and verify a questionnaire loaded with answers opens with those folded and ticked, everything unanswered expanded, and a comment-only question expanded
- [x] 3.4 Wire the header's tap to toggle the fold and set `pinned`, and verify tapping folds an expanded question, tapping again expands it, and an unanswered question can be folded

## 4. Settling an answer

- [x] 4.1 Add `form.settle(q)`, folding the card only when it is not pinned and the form is not still building, and verify a settle during build leaves the card expanded
- [x] 4.2 Settle from the `select` (radio and dropdown), `boolean` and `scale` controls when the new value is non-empty, and verify each collapses on being answered and does not collapse when its answer is cleared
- [x] 4.3 Settle from the `rank` control's existing "Use this order" handler only, and verify reordering rows leaves the question expanded and confirming the order collapses it
- [x] 4.4 Add a "Done" action to the `text`, `number`, `textarea` and `multiselect` cards, on the same row as the comment toggle, and verify activating it records the answer as typed and collapses the question
- [x] 4.5 Settle `text` and `number` from `entry.OnSubmitted`, and verify Enter in a field inside the scroll shield collapses the question while ordinary typing does not
- [x] 4.6 Verify a responder who types into a question and clicks into another leaves the first expanded with its text still recorded
- [x] 4.7 Verify a card the responder toggled by hand is not folded again by a later answer to it

## 5. Interaction with the rest of the form

- [x] 5.1 Verify `show_if` still shows and hides whole cards independently of folding, and that a question revealed by a condition appears expanded
- [x] 5.2 Verify the progress indicator, `Unanswered`, and the dirty check give identical results for a folded question and an expanded one
- [x] 5.3 Expand every question flagged by a refused submit, regardless of `pinned`, refresh, then scroll to the first in document order, and verify a folded unanswered required question is expanded with its warning visible and the form moves to it
- [x] 5.4 Verify submitting after answering and folding writes the same file bytes and prints the same stdout JSON as submitting with everything expanded, and that nothing about fold state reaches the file

## 6. The settled panel

- [x] 6.1 Add a `settled` card colour to both palettes and `SettledCard` to `cardProvider`, and verify it contrasts with the page within the same bounds as the plain card, carries a green cast, and is distinguishable from the plain card in both variants
- [x] 6.2 Give `cardBox` a settled state resolved through the theme on every refresh, and verify a settled card paints the settled colour and an unsettled one the plain colour
- [x] 6.3 Set it from `applyFold` for a collapsed question that carries an answer, and verify a folded answered question reads as settled, an expanded one does not, and a question folded with no answer keeps the plain colour

## 7. Showing a reorder happen

- [x] 7.1 Make rank rows long-lived, updating their ordinal and arrow states in place, and verify a reorder swaps the same row objects rather than rebuilding them and that the numbers and arrows still match their new places
- [x] 7.2 Position the rows in a layout-less container, giving the widget its own `MinSize` and `Resize`, and verify the stack still measures rows plus gaps and rows still fill the width
- [x] 7.3 Animate every moved row to its new place in one brief eased animation, and verify the slide starts where the rows were, ends in their slots, and reports the new order immediately rather than on completion
- [x] 7.4 Continue from the rows' current positions when a reorder arrives mid-slide, and leave a slide undisturbed by a resize, and verify both
- [x] 7.5 Verify a drag moves one place per row height, animating each step, and that a refused move animates nothing

## 8. Finishing

- [x] 8.1 Refresh the preview images with `INTERROGATE_RENDER=1 go test ./internal/ui -run RenderPreview` and verify both variants show a mix of folded and expanded cards legibly, with the settled rows reading as green without competing with their content
- [x] 8.2 Update the README and `examples/demo.yaml` if either describes the form's behaviour, and verify the README tests pass
- [x] 8.3 Run `make` (or `go test ./...` and `go vet ./...`) and verify the whole suite passes
