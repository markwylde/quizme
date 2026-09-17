## Why

The first questionnaire answered with `quizme` was also a review of it. The form works and reads well, but three quarters of the window is spent on chrome and empty comment boxes: a pinned header that repeats what the responder already read, and a two-row comment field under every question whether or not they want one. On a ten-question form that is most of the scrolling. The same review asked for a sense of progress through the page, and for a way to check a questionnaire without opening a window — which an agent needs to verify a file it has just written.

## What Changes

- **The header scrolls with the page.** Title and intro are read once, at the start; pinning them costs a fifth of the window for the rest of the session. The footer stays pinned, because the actions are needed from anywhere.
- **Comment boxes collapse.** Each question shows a quiet affordance instead of an open field; the field appears when the responder wants it, and a question that already carries a comment opens with it visible. Every question still accepts a comment — the capability is unchanged, only its resting state.
- **A progress bar** across the top edge of the footer. The responder can see how far through the questionnaire they are at a glance, without reading a count or judging it by the scrollbar.
- **Scrolling no longer stops over a text field.** Fyne gives every wrapping multi-line entry an internal scroller that consumes scroll events whether or not it has anywhere to scroll, and events do not bubble, so a trackpad scroll dies the moment the pointer crosses a comment box.
- **Each question sits on its own card.** A hairline separator turned out not to be enough to tell one question from the next when scrolling. Every question gets a panel in a single colour that contrasts with a lightened page, separated by a margin — one shade, not a cycle of them, so the page reads as a stack of cards rather than a swatch book.
- **`quizme --validate <path>`** checks a questionnaire and exits without opening a window, reporting the same errors it would refuse to open on.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities
- `quizme-cli`: adds a validate-only mode to the command's contract, and changes the form's resting layout — a scrolling header, collapsed comment fields, and a visible sense of progress.

## Impact

- `internal/ui/form.go`: header placement, a collapsible comment control, the progress bar in the footer, and each question wrapped in a card with a margin.
- `internal/ui/theme.go`: a lightened page colour and a card colour per palette.
- `main.go`: flag parsing gains `--validate`, and a path that loads and reports without constructing a form.
- The comment control is new custom widget work, like `scale` and `rank` before it.
- No change to the questionnaire format, the splicer, or the skill. A questionnaire written before this change behaves identically after it.

## Non-goals

- Rewriting the palette beyond the page and card colours. The colour changed here is for telling questions apart, not for restyling the form.
- Making the form resizable-by-question, collapsible sections, or any other structural navigation. It stays one scrolling page.
