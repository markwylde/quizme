## MODIFIED Requirements

### Requirement: Scrolling the page is never trapped by a field

Scrolling the questionnaire SHALL continue to work wherever the pointer rests, including over any text field the form presents — single-line, multi-line, number, or comment. A field that has nothing of its own to scroll SHALL NOT absorb the gesture.

#### Scenario: Scrolling across a text field
- **WHEN** the responder scrolls the page and the pointer passes over a multi-line text field
- **THEN** the page keeps scrolling, without the responder having to move the pointer aside

#### Scenario: Scrolling across a single-line text field
- **WHEN** the responder scrolls the page and the pointer passes over a single-line text field or a number field
- **THEN** the page keeps scrolling, exactly as it does over a multi-line field

#### Scenario: Scrolling down a form of short-answer questions
- **WHEN** every question on the page takes a single-line answer, so the pointer cannot avoid crossing a field
- **THEN** one continuous gesture carries the responder down the whole page

#### Scenario: Scrolling stops at the ends of the page
- **WHEN** the responder scrolls up at the top of the questionnaire, or down at the bottom
- **THEN** the page stops rather than running past its content

#### Scenario: The field still behaves as a field
- **WHEN** the responder clicks into, selects text in, or types into a shielded field
- **THEN** it behaves exactly as it would without the shield

#### Scenario: A number field still refuses an out-of-range answer
- **WHEN** the responder types a number outside the question's bounds into a shielded number field
- **THEN** the bound is shown against the field as it was before, and no answer is recorded
