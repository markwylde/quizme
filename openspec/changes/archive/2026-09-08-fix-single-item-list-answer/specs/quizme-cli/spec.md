## MODIFIED Requirements

### Requirement: Writing answers without disturbing the document

On submit or save, `quizme` SHALL write the answers into the questionnaire file it was given, changing only the `answer` and `comment` keys of each question and the top-level `status` and `submitted_at`. All other bytes of the file — comments, blank lines, key order, quoting style, and indentation — SHALL be preserved exactly. The document it writes SHALL be a loadable questionnaire, whatever the shape of the answers it records.

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

#### Scenario: A single selection is written as a list
- **WHEN** the responder ticks exactly one option in a `multiselect`
- **THEN** the answer is written as a one-item block sequence beneath the `answer` key, in the same shape a multi-item answer takes

#### Scenario: The written document loads again
- **WHEN** the responder submits any combination of answers, one-item lists included
- **THEN** the file that is written parses as a questionnaire, and the command reports the submission rather than an error

#### Scenario: An answer that reads like a list is still text
- **WHEN** the responder types a text answer beginning with `- `
- **THEN** it is written as a quoted scalar and loads back as that same text, not as a sequence

#### Scenario: Write is atomic
- **WHEN** the write fails partway through for any reason
- **THEN** the original file is left intact rather than truncated or partially written
