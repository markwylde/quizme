## MODIFIED Requirements

### Requirement: Single scrolling form

All questions SHALL be presented together on one vertically scrolling page, so the responder can read ahead, answer in any order, and revise earlier answers before submitting.

A question MAY be presented collapsed to a single row rather than at full height, provided every question is still on the page in its document order and can be returned to full height in one action. Collapsing SHALL NOT remove a question from the page, reorder it, or prevent it from being answered or revised.

#### Scenario: Questions exceed the window height
- **WHEN** the questionnaire is taller than the window
- **THEN** the page scrolls and the submit and dismiss controls remain reachable

#### Scenario: Answering out of order
- **WHEN** the responder answers question five before question two
- **THEN** both answers are retained

#### Scenario: Revising an earlier answer
- **WHEN** the responder changes an answer they already gave
- **THEN** the new value replaces the old one

#### Scenario: Revising a collapsed answer
- **WHEN** the responder returns to a question that has collapsed and changes its answer
- **THEN** the new value replaces the old one, exactly as it would had the question never collapsed

#### Scenario: Collapsed questions keep their place
- **WHEN** some questions on the page are collapsed and others are not
- **THEN** every question is still present, in the order the document declares

### Requirement: Comment box on every question

Every question SHALL offer a comment field alongside its answer control, usable independently of whether the question is answered. The field MAY rest collapsed behind a visible affordance, so that a page of questions is not dominated by fields most responders will not use, but SHALL always be reachable in one action from the expanded question.

A question presented collapsed SHALL indicate that it carries a comment, so a comment cannot be hidden without trace.

#### Scenario: Comment on any type
- **WHEN** any question of any type is presented expanded
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

#### Scenario: Comment survives collapsing the whole question
- **WHEN** a responder types a comment and the question itself then collapses
- **THEN** the comment is still recorded, the collapsed row shows that the question carries a comment, and expanding the question again shows the comment text

## ADDED Requirements

### Requirement: Questions can be collapsed and expanded

Each question SHALL have an expanded presentation, showing its prompt, help, answer control and comment field, and a collapsed presentation occupying a single row.

When the questionnaire opens, a question that already carries an answer SHALL be presented collapsed, and a question with no answer SHALL be presented expanded. A questionnaire therefore opens showing what is left to do rather than what is already settled, and nothing outstanding is hidden from a responder seeing the page for the first time. A comment is not an answer for this purpose.

Each question's header SHALL act as a control that collapses it when expanded and expands it when collapsed, and SHALL show which of the two it will do. This control SHALL be available on every question, whether or not it carries an answer.

#### Scenario: A questionnaire with nothing answered
- **WHEN** a questionnaire with no answers in it is opened
- **THEN** every applicable question is presented expanded

#### Scenario: A questionnaire that already carries answers
- **WHEN** a questionnaire is opened with answers already in the file
- **THEN** each question carrying an answer is presented collapsed, showing that answer and its completion mark, and every question without one is presented expanded

#### Scenario: A question carrying only a comment
- **WHEN** a questionnaire is opened containing a question with a comment but no answer
- **THEN** that question is presented expanded, because a comment is not an answer

#### Scenario: Opening a question folded on load
- **WHEN** the responder activates the header of a question that opened collapsed
- **THEN** it expands with the answer from the file in its control, ready to be revised

#### Scenario: Collapsing by hand
- **WHEN** the responder activates the header of an expanded question
- **THEN** that question collapses to a single row and its control, help text and comment field are no longer shown

#### Scenario: Expanding by hand
- **WHEN** the responder activates the header of a collapsed question
- **THEN** that question returns to full height with its control, help text and comment field as they were

#### Scenario: Collapsing an unanswered question
- **WHEN** the responder activates the header of a question they have not answered
- **THEN** it collapses, and is still counted as outstanding

#### Scenario: A question revealed by a condition
- **WHEN** answering a question causes a further question to become applicable
- **THEN** that further question appears expanded

### Requirement: A settled answer collapses its question

A question answered in a single gesture SHALL collapse itself as soon as that answer is recorded, so the page becomes a shrinking list of outstanding work without the responder having to fold anything by hand. This SHALL apply to `select`, `boolean` and `scale` questions, and to a `rank` question once its order is confirmed.

A question whose answer is typed or assembled over several actions SHALL NOT collapse itself while the responder is composing it. `text`, `number`, `textarea` and `multiselect` questions SHALL collapse only on an explicit signal from the responder: a "Done" action on the question, or pressing Enter in a single-line `text` or `number` field.

Clearing a question's answer SHALL NOT collapse it.

#### Scenario: Choosing an option
- **WHEN** the responder picks an option on a `select` question
- **THEN** the answer is recorded and the question collapses

#### Scenario: Answering yes or no
- **WHEN** the responder picks a side of a `boolean` question
- **THEN** the answer is recorded and the question collapses

#### Scenario: Picking a point on a scale
- **WHEN** the responder picks a value on a `scale` question
- **THEN** the answer is recorded and the question collapses

#### Scenario: Confirming a ranking
- **WHEN** the responder reorders a `rank` question's options and confirms the order
- **THEN** the order is recorded and the question collapses

#### Scenario: Reordering without confirming
- **WHEN** the responder drags a `rank` question's rows but does not confirm the order
- **THEN** the question stays expanded

#### Scenario: Typing an answer
- **WHEN** the responder types into a `text`, `number` or `textarea` question
- **THEN** the question stays expanded however much is typed

#### Scenario: Finishing a typed answer
- **WHEN** the responder activates the question's "Done" action, or presses Enter in a single-line `text` or `number` field
- **THEN** the answer as typed is recorded and the question collapses

#### Scenario: Moving on without finishing
- **WHEN** the responder types into a question and then clicks into a different question without giving the explicit signal
- **THEN** the first question stays expanded, and what was typed is still recorded

#### Scenario: Ticking several options
- **WHEN** the responder ticks an option on a `multiselect` question
- **THEN** the question stays expanded so further options can be ticked

#### Scenario: Finishing a multiselect
- **WHEN** the responder activates the "Done" action on a `multiselect` question
- **THEN** the selection is recorded and the question collapses

#### Scenario: Clearing an answer
- **WHEN** the responder removes the answer from a question, leaving it unanswered
- **THEN** the question does not collapse

#### Scenario: A responder's own choice is not overridden
- **WHEN** the responder has collapsed or expanded a question by hand and then changes that question's answer
- **THEN** the question stays as the responder left it rather than collapsing itself again

### Requirement: A collapsed question shows what was answered

A collapsed question SHALL show its prompt, a single-line rendering of the answer it carries, and a mark distinguishing an answered question from an unanswered one, so that a page of collapsed questions reads as a review of what the responder has said.

The completion mark SHALL be positioned consistently at the trailing edge of the row, SHALL be shown only for a question that carries an answer, and SHALL be drawn in a colour that reads as success in both the light and dark presentations.

A collapsed question that carries an answer SHALL additionally be presented on a panel whose colour reads as settled — the card colour with a green cast — so a page of collapsed rows can be read at a glance rather than mark by mark. That colour SHALL be distinguishable from the ordinary card colour and SHALL remain quiet enough not to compete with the row's own content, in both presentations. A collapsed question with no answer SHALL keep the ordinary card colour, so putting a question aside never looks like finishing it.

An answer too long for one row SHALL be shortened rather than wrapped or allowed to widen the page.

#### Scenario: An answered question collapsed
- **WHEN** an answered question is collapsed
- **THEN** its row shows the prompt, a rendering of the answer, and the completion mark

#### Scenario: An unanswered question collapsed
- **WHEN** an unanswered question is collapsed
- **THEN** its row shows the prompt and no completion mark

#### Scenario: Every type has a one-line answer rendering
- **WHEN** a collapsed question is of any of the supported types, including a `multiselect` with several options chosen and a `rank` with a confirmed order
- **THEN** its answer is rendered on the single row in a form a reader can recognise as that answer

#### Scenario: A long answer
- **WHEN** a collapsed question's answer is longer than the row can show
- **THEN** it is shortened to fit, and the page does not scroll sideways or grow taller for it

#### Scenario: Reviewing by scanning
- **WHEN** the responder has answered and collapsed several questions
- **THEN** scrolling the page shows each of those answers without expanding anything

#### Scenario: The mark follows the presentation
- **WHEN** the form is presented in its dark variant
- **THEN** the completion mark remains legible against the collapsed row

#### Scenario: A settled row is tinted
- **WHEN** a question that carries an answer is collapsed
- **THEN** its row is presented on the settled panel rather than the ordinary card colour

#### Scenario: A question put aside is not tinted
- **WHEN** the responder collapses a question they have not answered
- **THEN** its row keeps the ordinary card colour

#### Scenario: Reopening a settled question
- **WHEN** the responder expands a question that was presented as settled
- **THEN** it returns to the ordinary card colour, because the row is no longer standing in for the whole question

#### Scenario: The settled panel follows the presentation
- **WHEN** the form is presented in either its light or its dark variant
- **THEN** the settled panel is drawn from a colour suited to that presentation, and the row's prompt, answer and mark remain legible against it

### Requirement: A reorder is shown happening

When a `rank` question's options change places, the rows SHALL be presented travelling to their new positions rather than appearing in them. The travel SHALL be brief enough not to delay a responder reordering repeatedly, and SHALL apply however the reorder was asked for — an arrow pressed or a row dragged.

The recorded answer SHALL NOT wait on the travel: the new order is the answer from the moment the reorder is made. A further reorder while rows are still travelling SHALL continue from where those rows have reached, rather than completing the previous travel first. A reorder that is refused, because the row is already at the end it was sent towards, SHALL move nothing.

#### Scenario: Moving a row with an arrow
- **WHEN** the responder presses a row's move-up or move-down arrow
- **THEN** that row and the row it displaces are shown travelling to each other's positions, and end in them

#### Scenario: Dragging a row
- **WHEN** the responder drags a row far enough to change its place
- **THEN** the change is shown the same way as a pressed arrow, for each place the row moves

#### Scenario: The answer does not wait
- **WHEN** a reorder is made
- **THEN** the new order is recorded immediately, whether or not the rows have finished travelling

#### Scenario: Reordering again mid-travel
- **WHEN** the responder reorders again while rows are still moving
- **THEN** the rows continue from where they have reached, without first jumping to where they were going

#### Scenario: A refused reorder
- **WHEN** the responder tries to move the top row up, or the bottom row down
- **THEN** nothing moves and nothing is recorded

#### Scenario: The ranking stays readable while it moves
- **WHEN** rows have changed places
- **THEN** each row's position number and the arrows it offers match its new place

### Requirement: Collapsing changes only what is shown

Whether a question is collapsed SHALL have no effect on anything the questionnaire records or reports. The recorded answer and comment, the evaluation of `show_if` conditions, the progress indicator, whether the form counts as having unsaved changes, the file written on submit or save, and the JSON printed on stdout SHALL all be identical for a collapsed question and an expanded one.

Which questions are collapsed SHALL NOT be written to the questionnaire file.

#### Scenario: A collapsed answer is submitted
- **WHEN** the responder answers questions, lets them collapse, and submits
- **THEN** every answer is written to the file and printed on stdout exactly as it would have been with every question expanded

#### Scenario: Progress counts collapsed questions
- **WHEN** questions are collapsed
- **THEN** the progress indicator counts them the same as it would expanded, both in the answered figure and in the total

#### Scenario: A condition driven by a collapsed answer
- **WHEN** a question that another question's `show_if` depends on is answered and collapses
- **THEN** the dependent question appears or disappears exactly as it would have

#### Scenario: Unsaved changes are still noticed
- **WHEN** the responder answers a question, it collapses, and they then close the window
- **THEN** they are asked to save or discard, as they would be for an expanded question

#### Scenario: Fold state is not persisted
- **WHEN** the responder collapses questions, answered and unanswered alike, and submits
- **THEN** the written file records nothing about which questions were collapsed, and reopening it folds each question according to whether it carries an answer and nothing else

### Requirement: Outstanding required questions are never hidden by a fold

When a submit is refused because a visible required question is unanswered, every question flagged SHALL be expanded, with its warning shown, and the form SHALL move to the first of them. A responder SHALL never be shown a warning about a question they cannot see the control for.

#### Scenario: Submit blocked by a collapsed question
- **WHEN** the responder collapses an unanswered required question and then submits
- **THEN** that question is expanded, its warning is shown, and the form moves to it

#### Scenario: Several flagged questions
- **WHEN** submit is refused with more than one required question outstanding, some of them collapsed
- **THEN** all of them are expanded and flagged, and the form moves to the first in document order

#### Scenario: Answering a flagged question
- **WHEN** the responder answers a question that was expanded by a refused submit
- **THEN** its warning clears and the answer is recorded as normal
