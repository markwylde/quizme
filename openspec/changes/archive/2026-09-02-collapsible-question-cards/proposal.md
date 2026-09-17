## Why

Every question on the page stays at full height for the whole session, whether it has been dealt with or not. On a ten-question form that means the responder scrolls past their own finished work to reach what is left, and nothing on the page distinguishes "answered" from "still open" except reading each control and inferring it. The progress bar in the footer says *how many* are left but never *which*. Folding a question down to one line the moment it is done turns the page into a shrinking list of outstanding work, and gives each answer a visible, unambiguous mark that it is settled.

## What Changes

- **Each question card gains a collapsed and an expanded state.**
- **A questionnaire opens showing what is left.** A question that already carries an answer opens folded; anything unanswered opens in full, so nothing outstanding is hidden from a first-time responder. A comment is not an answer.
- **A card's header row becomes a toggle.** Clicking anywhere on the header folds or unfolds that question, answered or not. A chevron on the left of the prompt shows which way it will go.
- **A settled answer folds itself.** Questions answered in a single gesture — `select`, `boolean`, `scale`, and `rank` once its order is confirmed — collapse as soon as the answer is recorded.
- **Typed and multi-part answers fold on an explicit signal, never mid-keystroke.** `text`, `number`, `textarea` and `multiselect` cards carry a "Done" action, and pressing Enter in a single-line `text` or `number` field does the same thing. They do not collapse on focus loss, so a half-typed answer is never folded away under the responder.
- **A collapsed card shows the prompt, a short rendering of the answer, and a green tick on the right.** The folded page is therefore a review of everything said so far, not just a list of prompts.
- **A settled row is tinted.** A collapsed question carrying an answer sits on the card colour with a green cast, in both presentations, so the page can be read at a glance rather than tick by tick. A question folded with nothing answered keeps the ordinary colour — putting something aside must not look like finishing it.
- **A manual choice outranks the automatic one.** Once the responder has folded or unfolded a card by hand, later answers to that card do not fold it again — auto-collapse is a convenience for the first pass, not a behaviour that fights the responder.
- **A blocked submit force-expands the questions it flags.** Required-but-unanswered cards are unfolded with their warning showing, and the page scrolls to the first, so the thing that must be fixed can never be hidden inside a fold.
- **A `rank` reorder is shown happening.** Rows slide to their new places instead of the list changing under the reader, whether an arrow was pressed or a row dragged. It is brief, and the answer is recorded the moment the reorder is made — never on the animation finishing.
- **Answering is unaffected.** Collapsing is presentation only: the answer, the comment, `show_if` evaluation, progress counting, dirty tracking, and what gets written to the file all behave exactly as they do now.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `quizme-cli`: the single-scrolling-form and visual-separation requirements gain a collapsed state per question — what folds a card, what a folded card shows, that a fold never changes the recorded answer, and that a blocked submit reveals what it flags.

## Impact

- `internal/ui/form.go`: the card struct and `buildCard` grow a header/body split and collapse state; `set`, `flagMissing` and `scrollTo` learn about folds.
- A new `internal/ui/widget_collapse.go` (or equivalent) for the header row: chevron, prompt, answer summary, tick.
- `internal/ui/theme.go`: the existing `success` palette entry gets used for the tick, and each palette gains a `settled` card colour; `cardProvider` gains `SettledCard`.
- `internal/ui/widget_rank.go`: rows become long-lived and are positioned rather than laid out, so they can travel; the widget gains its own `MinSize` and `Resize`.
- Answer summary rendering needs a one-line form for every question type, including `multiselect` and `rank` lists.
- `internal/ui/form_test.go` and `polish_test.go` build cards and assert on their children; the header/body split moves things they walk.
- No change to `internal/questionnaire`, the YAML format, the CLI surface, the JSON on stdout, or the exit codes.

## Non-goals

- Collapsing on focus loss or blur. It was considered and rejected: it folds text answers the responder is still composing.
- A "collapse all" / "expand all" control. The per-card toggle plus auto-collapse covers the need; a global control can be added later if the folded page proves hard to reopen.
- Persisting fold state into the questionnaire file. Which cards are folded is a property of the session, not of the document.
- Any change to the comment field's own collapse behaviour, which already works and is specified separately.
