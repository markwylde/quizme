## 1. Shield the single-line fields

- [x] 1.1 Return the `text` control's entry through `f.scrollable`, and verify by scrolling the shield over a `text` question that the page offset moves
- [x] 1.2 Return the `number` control's entry through `f.scrollable`, and verify the same over a `number` question
- [x] 1.3 Correct the `scrollThrough` doc comment so it describes every entry rather than only wrapping multi-line ones

## 2. Hold the fields to the same test

- [x] 2.1 Broaden the shielding test from `TypeTextarea` to every question type that renders an entry — `text`, `textarea`, `number` — and verify it fails against the unshielded controls before the fix and passes after
- [x] 2.2 Verify the ends-of-page and no-sideways-scrolling behaviour holds when the gesture starts over a single-line field, not just a textarea

## 3. Confirm the fields still behave as fields

- [x] 3.1 Verify a shielded `text` field still records what is typed to the answer and still counts toward progress
- [x] 3.2 Verify a shielded `number` field still refuses an out-of-range value with the bound shown, and still records one inside the bounds
- [x] 3.3 Verify focus, tapping into a field, and text selection are unaffected, since the shield implements only `Scrollable`

## 4. Check the layout did not move

- [x] 4.1 Regenerate the light and dark previews and verify the single-line and number fields are the same height as before, with no gap introduced by the stack
- [x] 4.2 Run the full test suite and `gofmt`, and verify both are clean

## 5. Verify against a real questionnaire

- [x] 5.1 Answer a questionnaire made mostly of `text` and `number` questions, long enough to need scrolling, and verify one continuous trackpad gesture carries the page from top to bottom with the pointer resting over the fields
