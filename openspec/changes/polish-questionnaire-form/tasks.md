## 1. Validate mode

- [x] 1.1 Add `--validate` to the command's flag parsing and usage text, stating that it checks without opening a window, and verify the usage output names the flag
- [x] 1.2 Return after the existing load-and-validate step when the flag is set, and verify a valid questionnaire exits 0 having opened nothing and left the file byte-identical
- [x] 1.3 Verify an invalid questionnaire under `--validate` exits 1 with every error on stderr and nothing on stdout
- [x] 1.4 Verify `--validate` reports correctly with no display available, since no window is needed

## 2. Scrolling header

- [x] 2.1 Move the title and intro from the pinned header into the top of the scrolling list, and verify by rendered preview that they scroll away while the footer stays put
- [x] 2.2 Verify the submit and dismiss actions remain reachable from any scroll position

## 3. Collapsible comments

- [x] 3.1 Build the collapsible comment control, keeping the entry constructed-and-hidden so answer binding and dirty tracking are untouched, and verify a hidden comment still records to `q.Comment`
- [x] 3.2 Expand and focus the entry when the affordance is pressed, and verify focus lands in the field
- [x] 3.3 Start expanded when the question loads with a comment already on it, and verify against a questionnaire answered in an earlier session
- [x] 3.4 Label the collapsed affordance differently when the question carries a comment, and verify a collapsed comment is never invisible
- [x] 3.5 Verify a comment typed then collapsed is still written to the file on submit

## 4. Progress indicator

- [x] 4.1 Add the indicator to the footer beside the existing summary, reading from the same visibility pass as `show_if`, and verify it shows none answered on a fresh questionnaire
- [x] 4.2 Verify the count updates as questions are answered, with no other action
- [x] 4.3 Verify hidden questions are counted in neither the answered nor the outstanding total
- [x] 4.4 Verify the total grows when answering a question reveals a further one

## 5. Question tints

- [x] 5.1 Add a cycle of question tints to each palette in the theme, and verify every tint is within a few points of its palette's background so none competes with the content
- [x] 5.2 Wrap each card in a band that paints its tint, and verify by rendered preview that adjacent questions differ in both light and dark
- [x] 5.3 Assign tints over the visible questions so a hidden one does not leave two neighbours sharing a tint, and verify against a questionnaire with a gated question in the middle
- [x] 5.4 Have the band re-read its tint on refresh rather than capturing it once, and verify the tints follow a change of theme variant

## 6. Delivery

- [x] 6.1 Refresh the rendered previews and verify the density improvement by screenshot review with the user
- [x] 6.2 Update the README's form screenshot and document `--validate`, and verify the README's examples still pass their own tests
- [x] 6.3 Note in the skill that a questionnaire can be checked with `--validate`, and verify the skill's tests still pass
- [ ] 6.4 Run the full suite and one real interrogation, and verify formatting preservation still holds end to end
