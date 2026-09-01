## 1. Project setup

- [x] 1.1 Initialise the Go module at the repo root and verify `go build ./...` succeeds on an empty `main`
- [x] 1.2 Add the `yaml.v3` and Fyne dependencies and verify `go mod tidy` leaves a clean tree and `go build ./...` still succeeds
- [x] 1.3 Add a `Makefile` or equivalent with `build`, `test`, and `run` targets and verify each runs from a clean checkout

## 2. Questionnaire format

- [x] 2.1 Define the questionnaire, question, and answer types covering all eight field types plus `comment`, `show_if`, and `required`, and verify a hand-written example fixture unmarshals into them
- [x] 2.2 Implement loading and validation — missing `questions`, duplicate ids, missing prompts, unknown types, choice types with fewer than two options — and verify each rejection has a table-driven test naming the offending question
- [x] 2.3 Implement `show_if` reference and cycle checking and verify dangling references and circular chains are both rejected with the ids named
- [x] 2.4 Implement per-type answer validation, including `number` bounds, `scale` range, and `rank` permutation completeness, and verify with a table-driven test per type
- [x] 2.5 Implement visibility evaluation (which questions are visible given a set of answers) and verify chained and multi-condition cases resolve correctly
- [x] 2.6 Implement required-question checking against visible questions only and verify a hidden required question does not block completion

## 3. In-place answer writing

- [x] 3.1 Implement locating each question's line range and any existing `answer` / `comment` keys from `yaml.Node` positions, and verify positions are correct for nested sequences and varied indentation
- [x] 3.2 Implement replacing an existing `answer` or `comment` value in place and verify a golden-file test shows only those lines changed
- [x] 3.3 Implement inserting a new `answer` or `comment` block at the correct indent within a question and verify a golden-file test shows surrounding content untouched
- [x] 3.4 Implement multi-line answers as block scalars and verify a `textarea` answer containing blank lines and a colon round-trips correctly
- [x] 3.5 Implement removing an `answer` or `comment` key when a value is cleared or its question becomes hidden, and verify no orphaned keys remain
- [x] 3.6 Implement top-level `status` and `submitted_at` writing, creating the keys if absent, and verify each of the four status values is written correctly
- [x] 3.7 Implement atomic write via temp file and rename and verify an injected mid-write failure leaves the original file byte-identical
- [x] 3.8 Build a golden-file suite covering YAML comments, blank lines, quoting styles, CRLF line endings, and re-answering an already-answered file, and verify every non-owned byte is preserved in each case

## 4. CLI shell

- [ ] 4.1 Implement argument parsing for a single questionnaire path plus usage text, and verify missing and excess arguments both produce a usage error
- [ ] 4.2 Wire load-and-validate to run before any window opens and verify a malformed questionnaire writes errors to stderr and opens nothing
- [ ] 4.3 Implement the four exit codes (0 submitted, 1 error, 2 dismissed, 3 saved) and verify each outcome returns the right code from a scripted run
- [ ] 4.4 Implement the JSON answers document on stdout, keyed by question id with answer and comment, and verify its shape against a schema test and that diagnostics go to stderr only
- [ ] 4.5 Implement fail-fast detection of a missing display and verify the command errors promptly rather than hanging when no display is available

## 5. Form UI

- [ ] 5.1 Build the single scrolling page shell with title, intro, and pinned Submit / Dismiss controls, and verify the controls stay reachable with a questionnaire taller than the window
- [ ] 5.2 Implement controls for `select`, `multiselect`, `text`, `textarea`, `boolean`, and `number` including bound enforcement, and verify each round-trips a value into the answer model
- [ ] 5.3 Build the custom `scale` widget and verify it emits an integer within the declared range
- [ ] 5.4 Build the custom `rank` widget with drag-to-reorder and verify its answer is always a complete permutation of the options
- [ ] 5.5 Add the comment field to every question regardless of type and verify a comment can be entered on an otherwise unanswered question
- [ ] 5.6 Implement live `show_if` re-evaluation rebuilding the visible question list on any answer change, and verify questions appear and disappear in document order without disturbing scroll position unreasonably
- [ ] 5.7 Implement submit-time required validation that refuses submission and identifies the offending question, and verify a blank required question blocks submission while a hidden one does not
- [ ] 5.8 Implement the close-with-unsaved-changes prompt offering Save or Discard, and verify closing with no changes made skips the prompt and dismisses

## 6. Presentation pass

- [ ] 6.1 Implement a custom `fyne.Theme` covering colors, sizes, and padding, and verify it applies across every control type in a demo questionnaire
- [ ] 6.2 Bundle a font and verify it renders on macOS, Linux, and Windows builds
- [ ] 6.3 Review spacing, grouping, and typography against a realistic twenty-question questionnaire and verify by screenshot review with the user

## 7. Skill

- [ ] 7.1 Write the `interrogate` skill covering when to use a questionnaire versus asking in conversation, and verify it names the single-question and branching-enquiry exceptions from the spec
- [ ] 7.2 Document the questionnaire path convention and authoring guidance, and verify the skill produces a valid questionnaire for a sample change end to end
- [ ] 7.3 Document background launch and resume-on-exit, plus status-file recovery when the launch is lost, and verify by running the flow against a real questionnaire
- [ ] 7.4 Document outcome handling for submitted, saved, and dismissed, including treating comments as more precise than the selected option, and verify each branch against a fixture questionnaire
- [ ] 7.5 Document the not-installed fallback and verify the skill's guidance holds when `interrogate` is absent from the path

## 8. Delivery

- [ ] 8.1 Verify `go install` puts `interrogate` on the path and the binary runs a questionnaire from an arbitrary working directory
- [ ] 8.2 Verify cross-platform builds compile for macOS, Linux, and Windows
- [ ] 8.3 Write the README covering install, the questionnaire format with a worked example, and the exit codes, and verify a reader can author and run a questionnaire from it alone
- [ ] 8.4 Run one real end-to-end interrogation driven by an opsx agent on an actual change and verify the answers land in the file with formatting preserved
