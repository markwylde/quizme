## 1. Fix the inline-versus-block decision

- [x] 1.1 Give `renderEntry` a value-shaped test for whether the marshalled body may sit after `key: `, so a collection goes beneath the key however few lines it renders to, and verify a one-element `[]string` comes out as a block sequence
- [x] 1.2 Keep an empty collection on the inline branch, since it marshals to the flow form, and verify `[]` is still written inline rather than pushed under the key
- [x] 1.3 Correct the `renderEntry` doc comment so it describes the decision as made on the value rather than on the number of lines

## 2. Hold it with tests

- [x] 2.1 Extend the list-answer coverage to a one-item `multiselect` answer, and verify it fails against the current code and passes after the fix
- [x] 2.2 Establish whether a `rank` question is a second route to a one-element list — it is not: validation requires at least two options, so a rank answer is always a permutation of two or more, and the test records that reasoning instead
- [x] 2.3 Cover a text answer that begins with `- `, and verify it round-trips as text rather than becoming a sequence
- [x] 2.4 Verify the rendered document still loads in each case — the shared `render` helper already asserts this, so confirm it is what fails before the fix

## 3. Confirm nothing else moved

- [x] 3.1 Verify multi-item lists, block scalars, scalars, and `null` answers render byte-for-byte as they did before
- [x] 3.2 Verify a flow-mapping question still renders a one-item answer as `[a]` inline, unchanged by this fix
- [x] 3.3 Run the full test suite and `gofmt`, and verify both are clean

## 4. Verify end to end

- [x] 4.1 Answer a real questionnaire with exactly one box ticked in a `multiselect`, and verify the command exits `0`, prints the answers as JSON, and leaves a file that `interrogate --validate` accepts
