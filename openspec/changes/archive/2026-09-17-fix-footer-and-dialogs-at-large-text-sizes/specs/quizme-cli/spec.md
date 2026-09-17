## ADDED Requirements

### Requirement: The footer stays legible at every text size

The footer's status text and its actions SHALL never overlap. While the whole status fits on one row beside the actions, the two SHALL share that row, vertically centred on each other. When it does not fit, the actions SHALL move to a row of their own beneath the status, at the trailing edge, and the status SHALL take the full width.

#### Scenario: Default size in the default window
- **WHEN** the form is shown at 100% in its default window
- **THEN** the status and the actions share one row, and the status is on a single line

#### Scenario: Larger size in the default window
- **WHEN** the form is shown at 130% in its default window, and the status and actions together are wider than the footer
- **THEN** the actions sit on their own row below the status, and neither overlaps the other

#### Scenario: Any offered size
- **WHEN** the form is shown at any text size from 70% to 200%
- **THEN** no part of the status text is drawn beneath an action

### Requirement: Dialogs are wide enough to read

A dialog's message SHALL be set at a reading width that grows with the text size, capped at the space the window leaves, rather than at the narrowest width its words allow.

#### Scenario: A one-sentence warning
- **WHEN** the clear-all confirmation is shown at 100% or 130% in the default window
- **THEN** its message is at most two lines

#### Scenario: A narrow window
- **WHEN** the window is narrower than the reading width
- **THEN** the dialog fits inside the window and its message wraps to that width
