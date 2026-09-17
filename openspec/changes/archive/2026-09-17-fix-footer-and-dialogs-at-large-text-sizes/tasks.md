## 1. Footer

- [x] 1.1 Combine the progress count and the outstanding-required count into one wrapping status label, and verify the summary tests check the combined text
- [x] 1.2 Add a footer layout that keeps the status beside the actions only while it fits and otherwise stacks the actions beneath it, and verify with footer tests at 70–200% that nothing overlaps
- [x] 1.3 Inset the footer to the cards' edges, and verify by rendering at 100%, 130% and 200%

## 2. Dialogs

- [x] 2.1 Set every dialog's content at a reading width capped by the window, and verify with a test that fails on the old narrow dialog
- [x] 2.2 Right-align the unsaved-answers dialog's buttons, and verify by rendering

## 3. Verify

- [x] 3.1 Run `go test ./...`, `go vet ./...` and `gofmt -l .`, and verify all are clean
- [x] 3.2 Open the demo at larger sizes, and confirm the footer and dialogs look right
