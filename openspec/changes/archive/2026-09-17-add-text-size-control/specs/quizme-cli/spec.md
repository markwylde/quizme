## MODIFIED Requirements

### Requirement: Command invocation

`quizme` SHALL accept a single positional argument: the path to a questionnaire document. It SHALL open a desktop window presenting that questionnaire and SHALL run until the responder submits, saves, or dismisses it.

It SHALL also accept a `--validate` flag, which checks the questionnaire and exits without opening a window, so a questionnaire can be verified without interrupting anyone.

It SHALL also accept a `--text-size <percent>` flag, which sets the form's text size for that run. The value SHALL be a whole-number percentage from 70 to 200 in steps of 10.

#### Scenario: Valid questionnaire path
- **WHEN** `quizme path/to/questions.yaml` is run and the file is a valid questionnaire
- **THEN** a window opens presenting the questionnaire's title, intro, and questions

#### Scenario: Missing file
- **WHEN** the given path does not exist
- **THEN** no window opens, an error naming the path is written to stderr, and the exit code signals a usage error

#### Scenario: Invalid questionnaire
- **WHEN** the file exists but fails questionnaire validation
- **THEN** no window opens, the validation errors are written to stderr, and the exit code signals a usage error

#### Scenario: No argument given
- **WHEN** `quizme` is run with no path
- **THEN** usage text is written to stderr and the exit code signals a usage error

#### Scenario: No desktop session available
- **WHEN** the command runs in an environment with no display available
- **THEN** it fails immediately with an explanatory error rather than blocking

#### Scenario: Validating a good questionnaire
- **WHEN** `quizme --validate path/to/questions.yaml` is run on a valid questionnaire
- **THEN** no window opens, the exit code signals success, and the questionnaire file is left unchanged

#### Scenario: Validating a bad questionnaire
- **WHEN** `--validate` is given a questionnaire that fails validation
- **THEN** no window opens, every validation error is written to stderr, and the exit code signals a usage error

#### Scenario: Validating without a desktop session
- **WHEN** `--validate` is run in an environment with no display available
- **THEN** it still reports the questionnaire's validity, because no window was needed

#### Scenario: Text size given on the command line
- **WHEN** `quizme --text-size 130 path/to/questions.yaml` is run
- **THEN** the form opens at 130% text size, whatever size is saved in the user config

#### Scenario: Text size out of range or off-step
- **WHEN** `--text-size` is given `60`, `210`, `125`, or a value that is not a whole number
- **THEN** no window opens, an error naming the accepted range and step is written to stderr, and the exit code signals a usage error

## ADDED Requirements

### Requirement: Adjustable text size

The form SHALL let the responder make its text larger or smaller, and return it to the default, while the form is open. A size change SHALL scale all text on the form and the spacing around it together. Sizes SHALL run from 70% to 200% of the default in steps of 10%, and the default SHALL be 100%.

The control SHALL be offered as visible decrease and increase buttons in the form's header, and as keyboard shortcuts: the platform's primary modifier (Cmd on macOS, Ctrl elsewhere) with `+` (or `=`) to increase, `-` to decrease, and `0` to reset to 100%.

#### Scenario: Increasing the size
- **WHEN** the form is at 100% and the responder presses the increase button or the increase shortcut
- **THEN** the form is shown at 110%, with text and spacing both larger

#### Scenario: Decreasing the size
- **WHEN** the form is at 100% and the responder presses the decrease button or the decrease shortcut
- **THEN** the form is shown at 90%

#### Scenario: Resetting the size
- **WHEN** the form is at any size other than 100% and the responder presses the reset shortcut
- **THEN** the form is shown at 100%

#### Scenario: At the largest size
- **WHEN** the form is at 200%
- **THEN** the increase button is disabled and the increase shortcut leaves the size unchanged

#### Scenario: At the smallest size
- **WHEN** the form is at 70%
- **THEN** the decrease button is disabled and the decrease shortcut leaves the size unchanged

#### Scenario: Changing size keeps the responder's work
- **WHEN** the responder changes the text size after answering, commenting on, or folding questions
- **THEN** every answer, comment, and fold state is unchanged and the form still counts as having the same unsaved changes as before

#### Scenario: Shortcut while typing in a field
- **WHEN** focus is in a text field and the responder presses the increase shortcut
- **THEN** the size increases and no character is inserted into the field

### Requirement: Text size is remembered in the user config

The text size SHALL be stored in the user's config file at `~/.config/quizme/config.yaml` on every platform. The form SHALL open at the stored size when no `--text-size` flag is given. A change the responder makes on the form SHALL be written to the config as soon as it happens, whatever the questionnaire's eventual outcome.

A missing or unusable config SHALL never stop a questionnaire from being answered.

#### Scenario: First run with no config
- **WHEN** `quizme` runs and `~/.config/quizme/config.yaml` does not exist
- **THEN** the form opens at 100% and no config file is created until the responder changes the size

#### Scenario: Size remembered across runs
- **WHEN** the responder sets the size to 130%, closes the form, and later runs `quizme` again without `--text-size`
- **THEN** the new form opens at 130%

#### Scenario: Remembered even when dismissed
- **WHEN** the responder changes the size and then dismisses the questionnaire
- **THEN** the new size is still saved to the config

#### Scenario: Missing config directory
- **WHEN** the responder changes the size and `~/.config/quizme/` does not exist
- **THEN** the directory and file are created and the size is saved

#### Scenario: Flag does not overwrite the saved size
- **WHEN** the saved size is 120% and `quizme --text-size 150` runs and the responder changes nothing
- **THEN** the form opens at 150% and the config still says 120%

#### Scenario: Changing the size during a flagged run
- **WHEN** a run was started with `--text-size 150` and the responder increases the size to 160%
- **THEN** 160% is saved to the config

#### Scenario: Unreadable or invalid config
- **WHEN** the config file cannot be parsed, or holds a text size outside 70–200 or off the 10% step
- **THEN** the form opens at 100% (or at the nearest valid size, if the value was a number), a warning naming the file is written to stderr, and the exit code and stdout are unaffected

#### Scenario: Config cannot be written
- **WHEN** the responder changes the size and the config file cannot be written
- **THEN** the form still changes size, a warning is written to stderr, and the questionnaire can still be submitted, saved, or dismissed as normal

#### Scenario: Other keys in the config
- **WHEN** the config file holds keys other than the text size and the size is saved
- **THEN** those other keys are kept

### Requirement: Text size is not part of the questionnaire's outcome

Changing the text size SHALL NOT be written to the questionnaire file, reported on stdout, or reflected in the exit code, and SHALL NOT count as an unsaved change to the questionnaire.

#### Scenario: Only the size changed
- **WHEN** the responder changes the text size, answers nothing, and closes the window
- **THEN** no save-or-discard prompt is shown, the status becomes `dismissed`, and the questionnaire file is unchanged

#### Scenario: Submitting after a size change
- **WHEN** the responder changes the size and then submits
- **THEN** the questionnaire file and the JSON on stdout are the same as if the size had never changed
