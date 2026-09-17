# quizme-skill Specification

## Purpose
Defines the contract an agent follows when it wants a batch of decisions from a human: when to reach for a questionnaire instead of asking in chat, where the file goes, how the form is launched and awaited, and how the answers come back.

## Requirements

### Requirement: When an agent uses a questionnaire

The skill SHALL direct an agent to author a questionnaire when it has several related decisions to put to the human at once, and SHALL direct it to ask in conversation instead when there is a single question or when the next question depends on reasoning about the previous answer.

#### Scenario: Several decisions at once
- **WHEN** an agent has four open decisions blocking a proposal
- **THEN** it authors one questionnaire covering all four rather than asking them in sequence

#### Scenario: A single question
- **WHEN** an agent has one thing to clarify
- **THEN** it asks in conversation rather than opening a form

#### Scenario: Genuinely branching enquiry
- **WHEN** each follow-up question can only be written after interpreting the previous answer
- **THEN** the agent asks in conversation, rather than guessing at a branching questionnaire

### Requirement: Questionnaire location

A questionnaire belonging to an OpenSpec change SHALL be written to that change's `questionnaires/` directory under a descriptive kebab-case filename, so it is versioned and archived alongside the change it informs.

#### Scenario: Questionnaire for a change
- **WHEN** an agent working on the change `add-dark-mode` needs decisions about scope
- **THEN** it writes `openspec/changes/add-dark-mode/questionnaires/scope.yaml`

#### Scenario: Second questionnaire for the same change
- **WHEN** a later round of questions arises on the same change
- **THEN** it is written as a separate file in the same directory rather than overwriting the first

### Requirement: Authoring guidance

The skill SHALL direct the agent to write questions the human can answer without re-deriving context: each question stating what is being decided, choice questions offering the realistic options rather than an open field, and the questionnaire ordered so early answers gate later ones via conditional visibility.

#### Scenario: Options supplied for a decision with known candidates
- **WHEN** the agent knows the plausible answers to a decision
- **THEN** it offers them as options rather than asking an open text question

#### Scenario: Follow-up gated on a choice
- **WHEN** a question is only relevant given a particular earlier answer
- **THEN** it is included with a conditional-visibility rule rather than asked unconditionally

#### Scenario: Agent's own recommendation is visible
- **WHEN** the agent has a view on the right answer
- **THEN** that view is stated in the question so the human can accept or overrule it

### Requirement: Launching and awaiting the form

The agent SHALL run the installed `quizme` command against the questionnaire path without blocking on it, because a human may take many minutes to respond, and SHALL resume when the command exits.

#### Scenario: Long response time
- **WHEN** the human leaves the form open for longer than a foreground command would tolerate
- **THEN** the agent is still able to collect the answers when the form is finally submitted

#### Scenario: Agent resumes on completion
- **WHEN** the form exits
- **THEN** the agent reads the outcome and continues without the human having to prompt it

#### Scenario: Determining the outcome after an interruption
- **WHEN** the agent lost track of the running command
- **THEN** it determines the outcome by reading the questionnaire file's status

### Requirement: Acting on the outcome

The agent SHALL branch on the questionnaire's outcome: proceeding on a submission, working from partial answers on a save, and not re-opening the form unbidden on a dismissal.

#### Scenario: Submitted
- **WHEN** the questionnaire comes back submitted
- **THEN** the agent proceeds using the answers, and records decisions in the change's artifacts where they belong

#### Scenario: Saved partially
- **WHEN** the questionnaire comes back saved with some questions unanswered
- **THEN** the agent uses the answers given and raises the still-open questions with the human rather than assuming defaults

#### Scenario: Dismissed
- **WHEN** the questionnaire comes back dismissed
- **THEN** the agent does not reopen it, and asks the human how to proceed

#### Scenario: Comments outweigh the selected option
- **WHEN** a comment qualifies or contradicts the option chosen on the same question
- **THEN** the agent treats the comment as the more precise statement of intent

### Requirement: Availability of the command

The skill SHALL be usable in any repository where an agent runs, and SHALL tell the agent what to do when the `quizme` command is not installed rather than failing opaquely.

#### Scenario: Command not installed
- **WHEN** the agent attempts to run `quizme` and it is not on the path
- **THEN** the agent reports that it is not installed and how to install it, and falls back to asking in conversation
