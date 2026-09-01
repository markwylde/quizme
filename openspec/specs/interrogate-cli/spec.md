# interrogate-cli Specification

## Purpose
Defines the `interrogate` command: how it is invoked, how the desktop form behaves for the person filling it in, and what it leaves behind for the caller once the window closes.

## Requirements

### Requirement: Command invocation

`interrogate` SHALL accept a single positional argument: the path to a questionnaire document. It SHALL open a desktop window presenting that questionnaire and SHALL run until the responder submits, saves, or dismisses it.

It SHALL also accept a `--validate` flag, which checks the questionnaire and exits without opening a window, so a questionnaire can be verified without interrupting anyone.

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

#### Scenario: Validating a good questionnaire
- **WHEN** `interrogate --validate path/to/questions.yaml` is run on a valid questionnaire
- **THEN** no window opens, the exit code signals success, and the questionnaire file is left unchanged

#### Scenario: Validating a bad questionnaire
- **WHEN** `--validate` is given a questionnaire that fails validation
- **THEN** no window opens, every validation error is written to stderr, and the exit code signals a usage error

#### Scenario: Validating without a desktop session
- **WHEN** `--validate` is run in an environment with no display available
- **THEN** it still reports the questionnaire's validity, because no window was needed

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

Every question SHALL offer a comment field alongside its answer control, usable independently of whether the question is answered. The field MAY rest collapsed behind a visible affordance, so that a page of questions is not dominated by fields most responders will not use, but SHALL always be reachable in one action.

#### Scenario: Comment on any type
- **WHEN** any question of any type is presented
- **THEN** a free-text comment field is reachable for it in a single action

#### Scenario: Existing comment is shown
- **WHEN** a question already carries a comment from an earlier session
- **THEN** its comment field is presented open, with the text visible, rather than hidden behind the affordance

#### Scenario: Comment survives collapsing
- **WHEN** a responder types a comment and then collapses the field
- **THEN** the comment is still recorded, and the question shows that it carries one

#### Scenario: Comment without an answer
- **WHEN** a responder opens a comment field on an unanswered question and types in it
- **THEN** the comment is recorded and the question's answer remains unanswered

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

### Requirement: Header scrolls with the page

The questionnaire's title and intro SHALL scroll away with the content rather than remaining pinned. The actions and the progress indicator SHALL remain reachable at all times.

#### Scenario: Scrolling past the header
- **WHEN** the responder scrolls down a questionnaire longer than the window
- **THEN** the title and intro scroll out of view, giving their space to the questions

#### Scenario: Actions stay reachable
- **WHEN** the responder has scrolled anywhere in the questionnaire
- **THEN** the submit and dismiss actions are still visible without scrolling back

### Requirement: Questions are visually separated

Each question SHALL be presented on its own card: a panel whose background contrasts with the page behind it, separated from its neighbours by a margin. All cards SHALL share one background colour rather than being individually tinted. The contrast SHALL be strong enough to read as a distinct panel and quiet enough not to compete with the question's own content, in both the light and dark presentations.

#### Scenario: Questions read as separate panels
- **WHEN** a questionnaire of any length is presented
- **THEN** each question sits on a card that contrasts with the page, with a visible gap to the questions above and below it

#### Scenario: Cards share one colour
- **WHEN** several questions are presented together
- **THEN** they all carry the same card background, rather than a different colour each

#### Scenario: Card colours follow the presentation
- **WHEN** the form is presented in its dark variant
- **THEN** the page and card colours are drawn from a set suited to a dark background rather than inverted light ones

#### Scenario: Card does not obscure content
- **WHEN** a question is presented on a card
- **THEN** its prompt, controls, and help text remain legible against it

### Requirement: Visible progress through the questionnaire

The form SHALL show how far through the questionnaire the responder is as a proportion that can be read at a glance, so progress is legible without counting or inferring it from the scrollbar.

#### Scenario: Progress at the start
- **WHEN** a questionnaire with no answers is opened
- **THEN** the indicator shows none of its questions answered

#### Scenario: Progress as questions are answered
- **WHEN** the responder answers a question
- **THEN** the indicator advances without any other action

#### Scenario: Progress counts only applicable questions
- **WHEN** a question is hidden because its condition does not hold
- **THEN** it is counted in neither the answered nor the outstanding total

#### Scenario: Progress when a conditional question appears
- **WHEN** answering a question reveals a further question
- **THEN** the total grows to include it, and the indicator adjusts accordingly

#### Scenario: A questionnaire with everything answered
- **WHEN** every applicable question carries an answer
- **THEN** the indicator reads as complete

### Requirement: Scrolling the page is never trapped by a field

Scrolling the questionnaire SHALL continue to work wherever the pointer rests, including over a text field. A field that has nothing of its own to scroll SHALL NOT absorb the gesture.

#### Scenario: Scrolling across a text field
- **WHEN** the responder scrolls the page and the pointer passes over a multi-line text field
- **THEN** the page keeps scrolling, without the responder having to move the pointer aside

#### Scenario: Scrolling stops at the ends of the page
- **WHEN** the responder scrolls up at the top of the questionnaire, or down at the bottom
- **THEN** the page stops rather than running past its content

#### Scenario: The field still behaves as a field
- **WHEN** the responder clicks into, selects text in, or types into a shielded field
- **THEN** it behaves exactly as it would without the shield
