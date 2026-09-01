## Purpose

Defines the YAML document that carries a set of questions from an agent to a human and the human's answers back again, in a single file that is readable and editable by hand at every stage.

## ADDED Requirements

### Requirement: Document structure

A questionnaire document SHALL be a YAML mapping with a `title`, an optional `intro`, a `status`, and a `questions` sequence. Every other top-level key SHALL be preserved untouched by any tool that writes to the document.

#### Scenario: Minimal valid questionnaire
- **WHEN** a document contains a `title` string and a `questions` sequence with at least one valid question
- **THEN** it is accepted as a valid questionnaire

#### Scenario: Missing questions
- **WHEN** a document has no `questions` key, or `questions` is empty
- **THEN** it is rejected as invalid with a message naming the missing or empty key

#### Scenario: Unknown top-level keys are preserved
- **WHEN** a document carries additional top-level keys not defined by this format
- **THEN** the document is still valid and those keys are unchanged after answers are written

### Requirement: Question identity

Every question SHALL carry an `id` that is unique within the document and a `prompt` string. Ids SHALL be stable identifiers usable as keys in the answer output.

#### Scenario: Duplicate ids
- **WHEN** two questions share the same `id`
- **THEN** the document is rejected as invalid, naming the duplicated id

#### Scenario: Missing prompt
- **WHEN** a question has an `id` but no `prompt`
- **THEN** the document is rejected as invalid, naming the offending question id

### Requirement: Field types

Each question SHALL declare a `type` from the set `select`, `multiselect`, `text`, `textarea`, `number`, `boolean`, `scale`, and `rank`. Each type SHALL define what a valid answer value looks like:

- `select` — one value drawn from `options`
- `multiselect` — a sequence of zero or more values drawn from `options`
- `text` — a single-line string
- `textarea` — a multi-line string
- `number` — a numeric value, optionally bounded by `min` and `max`
- `boolean` — `true` or `false`
- `scale` — an integer between `min` and `max` inclusive, defaulting to 1 and 5
- `rank` — a permutation of `options` expressing the responder's ordering

#### Scenario: Unknown type
- **WHEN** a question declares a `type` outside the defined set
- **THEN** the document is rejected as invalid, naming the question and the unrecognised type

#### Scenario: Choice type without options
- **WHEN** a `select`, `multiselect`, or `rank` question has no `options`, or fewer than two
- **THEN** the document is rejected as invalid, naming the question

#### Scenario: Rank answer is a permutation
- **WHEN** a `rank` question is answered
- **THEN** the recorded answer contains every one of its `options` exactly once

### Requirement: Universal comment field

Every question, regardless of type, SHALL accept a free-text `comment` alongside its answer. A comment SHALL be recordable whether or not the question itself was answered.

#### Scenario: Comment without an answer
- **WHEN** a responder types a comment on a question but leaves the answer blank
- **THEN** the comment is recorded and the question's answer remains unanswered

#### Scenario: Empty comment omitted
- **WHEN** a responder leaves a question's comment blank
- **THEN** no `comment` key is written for that question

### Requirement: Conditional visibility

A question MAY declare `show_if` as a mapping of question ids to expected values. A question SHALL be presented to the responder only when every referenced question's current answer matches its expected value, and SHALL be treated as not applicable otherwise.

#### Scenario: Condition satisfied
- **WHEN** a question declares `show_if: {storage: sidecar}` and the `storage` answer is `sidecar`
- **THEN** the question is presented and may be answered

#### Scenario: Condition not satisfied
- **WHEN** the `storage` answer is anything other than `sidecar`
- **THEN** the dependent question is not presented, is not required, and records no answer

#### Scenario: Condition references an unknown question
- **WHEN** `show_if` names an id that does not exist in the document
- **THEN** the document is rejected as invalid, naming the dangling reference

#### Scenario: Circular conditions
- **WHEN** a chain of `show_if` references forms a cycle
- **THEN** the document is rejected as invalid, naming the questions in the cycle

#### Scenario: Answer withdrawn after answering a dependent question
- **WHEN** a dependent question has been answered and the controlling answer then changes so the condition no longer holds
- **THEN** the dependent question is no longer presented and its answer is not recorded in the output

### Requirement: Required questions

A question MAY declare `required: true`. A questionnaire SHALL be submittable only when every currently visible required question holds an answer.

#### Scenario: Required question left blank
- **WHEN** a visible required question has no answer and the responder attempts to submit
- **THEN** submission is refused and the offending question is identified to the responder

#### Scenario: Hidden required question
- **WHEN** a required question's `show_if` condition is not satisfied
- **THEN** it does not block submission

### Requirement: Answers live in the same document

Answers SHALL be recorded on the question they belong to, under the reserved keys `answer` and `comment`. These two keys per question SHALL be the only content a writing tool may add, change, or remove.

#### Scenario: Answer recorded in place
- **WHEN** a responder answers the question with id `storage`
- **THEN** an `answer` key holding that value appears on the `storage` question in the same file

#### Scenario: Re-running an answered questionnaire
- **WHEN** a document that already holds answers is presented again
- **THEN** the existing answers and comments are shown as the current values and may be changed

### Requirement: Status lifecycle

A questionnaire SHALL carry a `status` of `pending`, `submitted`, `saved`, or `dismissed`, and a `submitted_at` timestamp set whenever the status moves away from `pending`. A reader SHALL be able to determine the outcome of a questionnaire from these fields alone, without any companion or lock file.

#### Scenario: Newly authored questionnaire
- **WHEN** a questionnaire is created and not yet answered
- **THEN** its status is `pending` and `submitted_at` is absent or null

#### Scenario: Completed and submitted
- **WHEN** the responder submits with all visible required questions answered
- **THEN** status becomes `submitted` and `submitted_at` records the time

#### Scenario: Partially answered and saved
- **WHEN** the responder saves without completing every visible required question
- **THEN** status becomes `saved`, `submitted_at` records the time, and the answers given so far are present

#### Scenario: Abandoned
- **WHEN** the responder discards their work
- **THEN** status becomes `dismissed` and no answers or comments from that session are recorded
