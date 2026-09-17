## 1. Clearing the model

- [x] 1.1 Add `clearAll` to `form`, which sets every question's answer to nil and comment to "" (hidden questions included) and rebuilds the window content from the model, and verify with a UI test that a filled form reports no answers, empty comments, and zero progress afterwards
- [x] 1.2 Verify with a test that a hidden `show_if` question's answer is gone after clearing and stays gone when its condition is met again
- [x] 1.3 Verify with a test that a reordered `rank` question is back in authored order and counts as unanswered after clearing
- [x] 1.4 Verify with a test that every question is expanded and conditional questions are hidden again after clearing

## 2. Button and confirmation

- [x] 2.1 Add a low-importance "Clear all answers" button to the footer, left of Dismiss, and verify with a UI test that it is present and reachable while the page is scrolled
- [x] 2.2 Show a confirmation dialog reading "This will wipe all answers and comments. Are you sure?" with "Cancel" and a danger "Clear all" button, and verify with tests that Cancel leaves answers, comments, and folds unchanged and that Clear all calls `clearAll`

## 3. Interaction with closing and submitting

- [x] 3.1 Verify with a test that clearing a questionnaire that opened with answers makes `dirty()` true, that saving writes a file with no answers or comments that `--validate` accepts, and that discarding leaves the file byte-for-byte unchanged
- [x] 3.2 Verify with a test that answering then clearing a questionnaire that opened blank closes without a prompt as `dismissed`
- [x] 3.3 Verify with a test that submitting after a clear with required questions outstanding is refused and flags them
- [x] 3.4 Verify with a test that confirming a clear does not write the questionnaire file

## 4. Docs and end to end

- [x] 4.1 Mention "Clear all answers" in `README.md`, and verify `readme_test.go` passes
- [x] 4.2 Run `go test ./...`, `go vet ./...`, and `gofmt -l .`, and verify all are clean
- [x] 4.3 If `add-text-size-control` is already applied, verify that clearing at 150% keeps the form at 150%
- [x] 4.4 Build and run `quizme` on a scratch copy of `examples/demo.yaml`, answer and comment on a few questions, clear, cancel once, then confirm, and verify the form is blank and closing without further changes behaves as specified
