## Why

The first questionnaire answered with `interrogate` was also a review of it. The form works and reads well, but three quarters of the window is spent on chrome and empty comment boxes: a pinned header that repeats what the responder already read, and a two-row comment field under every question whether or not they want one. On a ten-question form that is most of the scrolling. The same review asked for a sense of progress through the page, and for a way to check a questionnaire without opening a window — which an agent needs to verify a file it has just written.

## What Changes

- **The header scrolls with the page.** Title and intro are read once, at the start; pinning them costs a fifth of the window for the rest of the session. The footer stays pinned, because the actions are needed from anywhere.
- **Comment boxes collapse.** Each question shows a quiet affordance instead of an open field; the field appears when the responder wants it, and a question that already carries a comment opens with it visible. Every question still accepts a comment — the capability is unchanged, only its resting state.
- **A progress indicator.** The responder can see how far through the questionnaire they are without judging it by the scrollbar.
- **Each question sits on its own tinted band.** A hairline separator turned out not to be enough to tell one question from the next when scrolling; a quiet background tint, cycling so no two adjacent questions share one, does the separating that the rule was failing to do.
- **`interrogate --validate <path>`** checks a questionnaire and exits without opening a window, reporting the same errors it would refuse to open on.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities
- `interrogate-cli`: adds a validate-only mode to the command's contract, and changes the form's resting layout — a scrolling header, collapsed comment fields, and a visible sense of progress.

## Impact

- `internal/ui/form.go`: header placement, a collapsible comment control, the progress indicator in the footer, and each card wrapped in its tinted band.
- `internal/ui/theme.go`: a cycle of question tints per palette, so light and dark each get their own.
- `main.go`: flag parsing gains `--validate`, and a path that loads and reports without constructing a form.
- The comment control is new custom widget work, like `scale` and `rank` before it.
- No change to the questionnaire format, the splicer, or the skill. A questionnaire written before this change behaves identically after it.

## Non-goals

- Rewriting the theme or the palette beyond the per-question tints. The review scored the look 5; the colour added here is for telling questions apart, not for restyling the form.
- Making the form resizable-by-question, collapsible sections, or any other structural navigation. It stays one scrolling page.
