## Context

See proposal.md — Why. The relevant current state:

- `internal/ui/form.go` builds one `card` per question in `buildCard`: a padded `VBox` of heading, help, control, comment field and warning, wrapped in a `cardBox` (which paints the panel) and then in a padding container held as `card.root`. `form.cards` maps question id → `*card`.
- Every answer change funnels through one method, `form.set(q, v)`, which records the answer and calls `syncVisibility`. `syncVisibility` shows and hides whole cards for `show_if`, by calling `Show`/`Hide` on `card.root`.
- Controls are wired in `form.control`, one constructor per type. `rank` already has a two-stage shape — drag or arrow to reorder, then a "Use this order" button that calls `rank.confirm()` — which is exactly the settle signal this change needs elsewhere.
- `commentField` (`widget_comment.go`) is the precedent for a collapsed-by-default region: built once, then `Show`/`Hide` on the wrapped entry, with the toggle's label reporting what is inside. Follow its shape rather than inventing a second one.
- `theme.go` already defines a `success` colour in both palettes (`0x2B8A3E` light, and the dark equivalent), currently unused. That is the tick's colour; no palette change is needed.
- `cardBox` deliberately re-reads its colour on `Refresh` so a form open across a light/dark switch keeps up. Anything new that carries a colour must do the same.

## Goals / Non-Goals

**Goals:**

- Collapse is a presentation layer over the existing card. The question, its control and its comment field are built exactly once and are the same objects whether folded or not, so answers, `show_if`, dirty tracking and focus behave as they do today.
- One place decides whether a card is folded, and one place applies it, so the interaction between auto-collapse, manual toggling, `show_if` hiding and submit-time expansion stays legible.
- The settle signal is explicit per control type, not inferred from the answer's shape.

**Non-Goals:**

- Animating the fold. Fyne gives no cheap height animation, and a form is not the place to spend one.
- Reworking `commentField`. The question's fold sits above it and hides it wholesale; the comment's own toggle is untouched.
- Any change to `internal/questionnaire`. Fold state never reaches the document.

## Decisions

### The card splits into a header widget and a body container

`buildCard` produces `card.header` (a new widget) and `card.body` (the existing `VBox`, minus the heading), stacked inside the same `cardBox`. Folding is `card.body.Hide()`; unfolding is `Show()`. `card.root` stays the outermost object so `syncVisibility` and `scrollTo` keep working unchanged.

The heading moves *into* the header widget, because the prompt is the thing the responder clicks. Alternative considered: keep the heading in the body and put a bare chevron button above it. Rejected — it makes an unfolded card show its prompt in a different place from a folded one, and gives a 20px target for the commonest interaction on the page.

### The header is a custom tappable widget, not a Button

A new `internal/ui/widget_header.go` defines `questionHeader`: a widget implementing `fyne.Tappable` (and `desktop.Hoverable`, for a hover tint and pointer cursor) laying out chevron · prompt · required marker · answer summary · tick. A `widget.Button` cannot hold a wrapping bold prompt beside a caption and a coloured icon, and Fyne's hit test walks up to the nearest `Tappable` parent, so plain labels inside it still deliver the tap.

The tick is a `canvas.Text`/icon coloured from `palette.success`, re-read in `Refresh` for the same reason `cardBox` does.

### The settled tint is a second card colour, not a decoration on the header

Each palette gains `settled`, and `cardProvider` gains `SettledCard(variant)` beside `QuestionCard(variant)`. `cardBox` carries a `settled bool` and resolves one or the other in `colour()`, which it already re-reads on every `Refresh` — so the tint tracks a light/dark switch for free. `applyFold` is the single caller: `c.card.SetSettled(c.collapsed() && c.question.HasAnswer())`.

The tint goes on the card rather than behind the header row because a folded card *is* its header row, and painting the header would leave a hairline of the old colour around it. Alternatives considered: a coloured left edge (harder to see at a glance on a page of rows, and it fights the chevron) and tinting the tick's area only (which is what the tick already does).

Folded-and-unanswered deliberately keeps the plain colour. Green here means settled; a question the responder put aside is not, and the page would otherwise claim work that has not happened.

### Answer summaries are a single function over the answer value

`answerSummary(q *questionnaire.Question) string` in the ui package, switching on the question's type and answer: the option text for `select`, "Yes"/"No" for `boolean`, the number for `scale`/`number`, the text (shortened) for `text`/`textarea`, comma-joined options for `multiselect`, numbered order for `rank`. It reuses the existing `preview` helper from `widget_comment.go` for shortening, so the collapsed row and the collapsed comment truncate identically.

The tick is driven by `q.HasAnswer()`, which already exists and already means what the spec means by "carries an answer".

### The settle signal is passed by the control, not deduced in `set`

`form.set` stays as it is. Each control constructor decides whether its gesture settles the question and calls `f.settle(q)` alongside `f.set(q, v)`:

- `select` (both the radio and the dropdown form), `boolean`, `scale`: settle on change, but only when the new value is non-empty — clearing an answer must not fold the question the responder is clearing.
- `rank`: settle inside the existing "Use this order" handler, not on reorder.
- `text`, `number`: settle from `entry.OnSubmitted` (Fyne's Enter hook on a single-line `Entry`) and from a new "Done" button on the card.
- `textarea`, `multiselect`: settle from the "Done" button only.

Alternative considered: infer it in `set` from the question type. Rejected — `set` cannot tell a `text` question's Enter from its third keystroke, which is the whole distinction the spec draws.

The "Done" button lives at the foot of the body for the four types that need it, on the same row as the comment toggle, so it does not add a row of its own.

### Fold state is three-valued, and manual beats automatic

```go
type foldState int // foldAuto, foldOpen, foldClosed
```

`card.fold` starts `foldAuto`, and `card.pinned()` is the derived question "has the responder decided this one themselves" — `fold != foldAuto`. `form.settle(q)` folds the card only when it is not pinned, and the header's tap sets `foldOpen` or `foldClosed`, either of which pins it. Keeping provenance in the one enum rather than in a second boolean means the two can never disagree, and it makes "a responder's own choice is not overridden" fall out of one condition rather than a pile of special cases.

`form.expand(c)` (used by submit-time flagging) forces a folded card back to `foldAuto`, since the spec says a flagged question is never left folded — and returning it to automatic means answering it folds it away again. A card the responder pinned *open* is left alone.

### Fold-on-load is decided in one place, not left to restore callbacks

Fyne's `RadioGroup.SetSelected` and `CheckGroup.SetSelected` fire `OnChanged`, and `buildCard` uses them to restore answers already in the file. Letting `settle` run from those callbacks would arrive at roughly the right result by accident, and at the wrong one for a control that reports itself more than once on the way up. So `form` gains a `building bool`, set for the duration of `build`, which makes `settle` a no-op during construction; `buildCard` then ends with an explicit `if q.HasAnswer() { c.fold = foldClosed }`. One line, one place, and it reads as the decision it is.

The state it sets is `foldClosed` rather than anything new: a question folded on load is in exactly the position of one the responder answered and let fold. Opening it pins it (`foldOpen`), so revising an answer restored from the file behaves like revising one given in the session.

### Rank rows are positioned, not laid out, so they can travel

`rankWidget` rebuilt every row on each move — a row's index drove its ordinal and its disabled arrows, so recreating them was the cheap way to keep those honest. A row that is destroyed and recreated cannot travel anywhere, so the rows become long-lived: built once, holding their own label, ordinal and arrow buttons, with `setIndex` updating what the position drives.

Their container becomes `container.NewWithoutLayout`, which leaves positions under the widget's control. That costs two methods a box layout was providing: `MinSize` (the rows' heights plus the gaps — a layout-less container would report one row's worth) and `Resize` (widths for the rows). `Resize` leaves positions alone while an animation is in flight, or a resize mid-travel would snap every row to its destination.

One `fyne.Animation` carries every row that has moved, capturing each row's *current* position as its start. So a second press mid-flight stops the first animation and continues from wherever the rows have reached, rather than completing the previous move first. `move` records the answer and calls `notify` before starting the animation: a questionnaire must never wait on a slide.

`rankSlide` is 110ms with `AnimationEaseInOut`. Fyne's `canvas.DurationShort` (150ms) was the obvious alternative and felt draggy under repeated presses.

### Collapsing and `show_if` hiding stay separate

`show_if` hides `card.root`; folding hides `card.body`. They never touch the same object, so a question hidden by a condition and then revealed comes back with whatever fold state it had — and since a newly applicable question has never been answered or toggled, that is `foldAuto`, i.e. expanded, as the spec requires.

### Height changes need an explicit refresh

Hiding `card.body` changes the card's minimum size. `container.VBox` re-lays out on `Refresh`, so `applyFold` refreshes the `cardBox` and then `f.list`, the same call `syncVisibility` already makes after a visibility change. `scrollTo` reads `c.root.Position().Y`, so it must be called *after* that refresh, not before — the submit-time path expands first, refreshes, then scrolls.

## Risks / Trade-offs

- **A `select` question can no longer be cleared in one gesture.** Picking an option folds the card, so clearing it means expanding first — and expanding pins the card open. → Accepted: clearing a radio answer is rare, and the pin is the behaviour the responder asked for. The spec's "clearing an answer does not collapse" keeps the second half of that interaction sane.
- **The "Done" button is a new control on four question types**, and a responder who ignores it simply never sees those cards fold. → Accepted deliberately over collapsing on focus loss, which folds half-typed answers. Enter on single-line fields covers the common case without the button being noticed.
- **`OnSubmitted` also fires for a `text` entry inside the scroll shield.** The shield (`scrollThrough`) wraps the entry for scrolling only and does not intercept key events, but this needs confirming with a test rather than assumed.
- **The preview images in `internal/ui/testdata` will no longer match the form.** → Regenerate with `QUIZME_RENDER=1 go test ./internal/ui -run RenderPreview`, and check whether the README references them.
- **Existing tests walk the card's object tree** (`card_test.go`, `polish_test.go`, `form_test.go`) and will break where the heading moves out of the body. → Expected and cheap; the header is reachable through `card.header`, and `cardBox.Content()` exists precisely so tests can walk inside a card.
- **A page of folded cards could be hard to reopen** if a responder folds everything and wants to re-read it. → Noted; a global expand-all is out of scope for this change (proposal — Non-goals) and can follow if it bites.
