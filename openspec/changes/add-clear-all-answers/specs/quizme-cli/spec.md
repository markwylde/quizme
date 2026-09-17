## ADDED Requirements

### Requirement: Clearing all answers

The form SHALL offer a "Clear all answers" action that stays reachable from anywhere in the questionnaire. Choosing it SHALL first ask for confirmation with the message "This will wipe all answers and comments. Are you sure?", and SHALL clear nothing unless the responder confirms.

On confirmation, every question SHALL become unanswered and every comment SHALL become empty, including those on questions currently hidden by `show_if`. Questions SHALL return to the fold state they would have on a newly opened, unanswered questionnaire. Clearing SHALL NOT write the questionnaire file. The cleared state is written only if the responder later submits or saves.

#### Scenario: Confirming the clear
- **WHEN** the responder has answered and commented on several questions, chooses "Clear all answers", and confirms
- **THEN** every control shows no answer, every comment box is empty, every question is expanded, and the progress indicator shows nothing answered

#### Scenario: Cancelling the clear
- **WHEN** the responder chooses "Clear all answers" and cancels the confirmation
- **THEN** every answer, comment, and fold state is exactly as it was

#### Scenario: Confirmation wording
- **WHEN** the responder chooses "Clear all answers"
- **THEN** a confirmation is shown reading "This will wipe all answers and comments. Are you sure?"

#### Scenario: Hidden answers are cleared too
- **WHEN** a question hidden by `show_if` holds an answer and the responder confirms a clear
- **THEN** that answer is gone, and it is still gone if the question later becomes visible

#### Scenario: Conditional questions re-hide
- **WHEN** a question is visible only because of another question's answer, and the responder confirms a clear
- **THEN** the conditional question is hidden again

#### Scenario: A ranking returns to its authored order
- **WHEN** the responder has reordered a `rank` question and confirms a clear
- **THEN** the options are back in authored order and the ranking counts as unanswered

#### Scenario: Clearing does not write the file
- **WHEN** the responder confirms a clear and nothing else happens
- **THEN** the questionnaire file on disk is unchanged

#### Scenario: Closing after clearing a previously answered questionnaire
- **WHEN** the questionnaire opened with answers in the file, the responder clears them, and closes the window
- **THEN** they are asked to save or discard. Saving writes the questionnaire with no answers or comments, and discarding leaves the file as it was

#### Scenario: Closing after clearing answers made this session
- **WHEN** the questionnaire opened unanswered, the responder answers some questions, clears them, and closes the window
- **THEN** no save-or-discard prompt is shown and the status becomes `dismissed`

#### Scenario: Submitting after clearing
- **WHEN** the responder clears the form and then submits with required questions unanswered
- **THEN** submission is refused and the outstanding required questions are shown, as for any incomplete questionnaire
