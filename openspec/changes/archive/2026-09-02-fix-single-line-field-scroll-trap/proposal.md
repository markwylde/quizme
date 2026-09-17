## Why

Shielding the multi-line fields fixed half the problem. A trackpad scroll still dies the moment the pointer crosses a single-line text field or a number field, because Fyne gives those an internal scroller too — a default `Entry` truncates and scrolls its content sideways, so its scroller is live and consumes the gesture whether or not it has anywhere to go. On a questionnaire whose questions are mostly short answers, that is most of the page: the responder has to steer the pointer into the margin to get down the form, which is exactly the complaint the last fix was meant to end.

## What Changes

- **Single-line text fields stop swallowing page scrolling.** The `text` control gets the same transparent scroll shield the textarea and comment fields already use.
- **Number fields stop swallowing page scrolling.** The `number` control is the same `Entry` widget, and behaves the same way; it keeps its validator, its bounds hint, and its inline error.
- **The requirement stops saying "multi-line".** The spec's scroll-trap scenarios are broadened to cover every text field the form presents, so a control added later is held to the same rule rather than to the type it happens to be.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities
- `quizme-cli`: the "Scrolling the page is never trapped by a field" requirement now covers single-line text and number fields, not only multi-line ones.

## Impact

- `internal/ui/form.go`: `textControl` and `numberControl` return their entry wrapped by the existing `f.scrollable`.
- `internal/ui/polish_test.go`: the shielding test broadens from textareas to every question type that renders an `Entry`.
- No new widget work — `scrollThrough` and `scrollShield` are unchanged and already carry the reasoning.
- No change to the questionnaire format, the splicer, the skill, or any answer binding: the shield sits over the field and answers only to scrolling.

## Non-goals

- Letting a single-line field scroll its own content sideways under a horizontal gesture. The shield drops sideways deltas, as it already does for multi-line fields; a long answer is still reached with the caret and with text selection.
- Shielding controls that have no internal scroller — radio groups, check groups, the scale, the rank list, and the select popup are all unaffected today, and blanket-wrapping them would add layers for nothing.
