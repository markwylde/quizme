# interrogate-cli Specification

## Purpose
Defines the `interrogate` command: how it is invoked, how the desktop form behaves for the person filling it in, and what it leaves behind for the caller once the window closes.

## Requirements

### Requirement: Command invocation

`interrogate` SHALL accept a single positional argument: the path to a questionnaire document. It SHALL open a desktop window presenting that questionnaire and SHALL run until the responder submits, saves, or dismisses it.

#### Scenario: Valid questionnaire path
- **WHEN** `interrogate path/to/questions.yaml` is run and the file is a valid questionnaire
- **THEN** a window opens presenting the questionnaire's title, intro, and questions

#### Scenario: Missing file
- **WHEN** the given path does not exist
- **THEN** no window opens, an error naming the path is written to stderr, and the exit code signals a usage error

#### Scenario: Invalid questionnaire
- **WHEN** the file exists but fails questionnaire validation
- **THEN** no window opens, the validation errors are written to stderr, and the exit code signals a usage error

#### Scenario: No argument given
- **WHEN** `interrogate` is run with no path
- **THEN** usage text is written to stderr and the exit code signals a usage error

#### Scenario: No desktop session available
- **WHEN** the command runs in an environment with no display available
- **THEN** it fails immediately with an explanatory error rather than blocking

### Requirement: Single scrolling form

All questions SHALL be presented together on one vertically scrolling page, so the responder can read ahead, answer in any order, and revise earlier answers before submitting.

#### Scenario: Questions exceed the window height
- **WHEN** the questionnaire is taller than the window
- **THEN** the page scrolls and the submit and dismiss controls remain reachable

#### Scenario: Answering out of order
- **WHEN** the responder answers question five before question two
- **THEN** both answers are retained

#### Scenario: Revising an earlier answer
- **WHEN** the responder changes an answer they already gave
- **THEN** the new value replaces the old one

### Requirement: Type-appropriate controls

Each question SHALL be rendered with a control suited to its declared type, and the control SHALL only permit values valid for that type.

#### Scenario: Select renders single choice
- **WHEN** a `select` question is presented
- **THEN** exactly one of its options can be chosen at a time

#### Scenario: Multiselect renders multiple choice
- **WHEN** a `multiselect` question is presented
- **THEN** any number of its options, including none, can be chosen

#### Scenario: Number respects bounds
- **WHEN** a `number` question declares `min` and `max` and the responder enters a value outside them
- **THEN** the value is rejected and the bound is shown to the responder

#### Scenario: Rank reorders options
- **WHEN** a `rank` question is presented
- **THEN** the responder can reorder its options and the resulting order is the answer

### Requirement: Comment box on every question

Every question SHALL present a comment field alongside its answer control, usable independently of whether the question is answered.

#### Scenario: Comment on any type
- **WHEN** any question of any type is presented
- **THEN** a free-text comment field is available for it

### Requirement: Live conditional visibility

Questions gated by `show_if` SHALL appear and disappear as the answers they depend on change, without the responder taking any other action.

#### Scenario: Condition becomes satisfied
- **WHEN** the responder gives the answer a hidden question depends on
- **THEN** that question becomes visible in place, in its document order

#### Scenario: Condition becomes unsatisfied
- **WHEN** the responder changes an answer so a visible dependent question no longer applies
- **THEN** that question is removed from the form

### Requirement: Writing answers without disturbing the document

On submit or save, `interrogate` SHALL write the answers into the questionnaire file it was given, changing only the `answer` and `comment` keys of each question and the top-level `status` and `submitted_at`. All other bytes of the file — comments, blank lines, key order, quoting style, and indentation — SHALL be preserved exactly.

#### Scenario: Hand-written comments survive
- **WHEN** the questionnaire file contains YAML comments and the responder submits
- **THEN** those comments are present and unchanged in the written file

#### Scenario: Existing answers replaced in place
- **WHEN** a question already has an `answer` key and the responder gives a different answer
- **THEN** that key's value is updated where it already sits, rather than being appended elsewhere

#### Scenario: New answers inserted on the question
- **WHEN** a question has no `answer` key yet and the responder answers it
- **THEN** an `answer` key is added within that question's own block

#### Scenario: Unanswered questions stay unanswered
- **WHEN** the responder submits with a question left blank
- **THEN** no `answer` key is written for that question

#### Scenario: Write is atomic
- **WHEN** the write fails partway through for any reason
- **THEN** the original file is left intact rather than truncated or partially written

### Requirement: Answers reported on stdout

On submit or save, `interrogate` SHALL print the collected answers to stdout as JSON, keyed by question id, including each question's answer and comment, so a caller can consume them without re-reading the file.

#### Scenario: Submitted answers printed
- **WHEN** the responder submits
- **THEN** stdout carries a JSON document containing the status and every answered question's id, answer, and comment

#### Scenario: Dismissed prints no answers
- **WHEN** the responder dismisses the questionnaire
- **THEN** stdout carries a JSON document reporting the dismissed status and no answers

#### Scenario: Diagnostics kept off stdout
- **WHEN** the command emits warnings or errors
- **THEN** they are written to stderr so stdout remains parseable

### Requirement: Closing without submitting

If the responder closes the window with unsaved changes, `interrogate` SHALL ask whether to save or discard them, and SHALL not lose work silently.

#### Scenario: Close with unsaved changes
- **WHEN** the responder closes the window after answering something
- **THEN** they are asked to save or discard before the window closes

#### Scenario: Choosing save
- **WHEN** the responder chooses to save
- **THEN** the answers given so far are written and the status becomes `saved`

#### Scenario: Choosing discard
- **WHEN** the responder chooses to discard
- **THEN** no answers are written and the status becomes `dismissed`

#### Scenario: Close with nothing entered
- **WHEN** the responder closes the window having changed nothing
- **THEN** no prompt is shown and the status becomes `dismissed`

### Requirement: Exit codes reflect the outcome

The exit code SHALL distinguish a completed submission, a partial save, a dismissal, and a failure, so a caller can branch on the outcome without parsing output.

#### Scenario: Submitted
- **WHEN** the responder submits a complete questionnaire
- **THEN** the exit code indicates success

#### Scenario: Saved partially
- **WHEN** the responder saves an incomplete questionnaire
- **THEN** the exit code is distinct from both a full submission and a dismissal

#### Scenario: Dismissed
- **WHEN** the responder discards the questionnaire
- **THEN** the exit code indicates dismissal rather than success or error

#### Scenario: Failure
- **WHEN** the questionnaire cannot be loaded or the answers cannot be written
- **THEN** the exit code indicates an error and is distinct from any responder outcome
