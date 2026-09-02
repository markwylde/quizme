## Context

See proposal.md — Why. The machinery already exists: `scrollThrough` stacks a transparent `scrollShield` over a field, the shield implements `fyne.Scrollable` and nothing else, and `form.scrollPage` moves the page scroller by the event's delta. `textareaControl` and every comment field already go through `form.scrollable`; the single-line controls do not.

The reason they need to is the same one written up on `scrollThrough`, and it is not specific to multi-line entries. A default Fyne `Entry` wraps with `TextTruncateClip`, which leaves its internal `widget.Scroll` live with `ScrollBoth` so the content can be dragged sideways when the text overruns the box. That scroller consumes the wheel event whether or not it has anywhere to go, and events do not bubble.

## Goals / Non-Goals

**Goals:**
- The `text` and `number` controls scroll the page like the rest of it.
- The fix is the existing shield, applied at one more call site each — no second mechanism for the same problem.

**Non-Goals:**
- Changing how the shield forwards events. Sideways deltas keep being dropped, for single-line fields as for multi-line ones.
- Auditing controls with no internal scroller. Radio groups, check groups, the scale, the rank list, and the select popup do not implement `Scrollable` and are left alone.

## Decisions

**Shield the field rather than switch off its scroller.** The alternative is to set `entry.Scroll = widget.ScrollNone` and `entry.Wrapping = fyne.TextWrapOff` on the single-line controls, which makes Fyne drop the internal scroller entirely (`entry.go` only attaches it when one of those is set). That is a smaller diff, but it takes away the field's ability to show text longer than its box: the content can no longer be moved sideways at all, so a long answer becomes unreadable past the right edge. It also diverges from how the multi-line fields are handled, leaving two explanations for one behaviour. The shield keeps the field intact and keeps the reasoning in one place.

**Wrap inside the control constructors, not around `f.control`.** Wrapping at the `control` call site would shield every question type, including those that never swallow a gesture, and would put a layer between the card layout and widgets the form reaches into later — `rankWidget` in particular is held on the card and driven directly. Shielding the two entry constructors keeps the wrapper where the problem is.

**No change to `scrollThrough`.** It takes a `fyne.CanvasObject` and stacks it; it never assumed multi-line. Only its doc comment mentions multi-line entries specifically, and that wording is corrected to say every entry.

## Risks / Trade-offs

- **A shielded number field could lose its validation feedback** → the validator and its inline error are drawn by the entry itself, inside the shielded object; the shield draws nothing. Covered by a scenario in the spec delta and a task that checks an out-of-range entry still refuses.
- **The stack could change the field's height, since `scrollThrough` reports the stack's minimum size** → the shield's renderer is an empty container with a zero minimum, so the stack's minimum is the field's own. Worth confirming on a rendered preview rather than by argument, because a single-line entry has much less slack than a four-row textarea.
- **A control added later gets forgotten again, exactly as these two were** → the shielding test stops keying off `TypeTextarea` and instead asserts over every question type that renders an entry, so a new entry-backed control fails the test until it is shielded.
