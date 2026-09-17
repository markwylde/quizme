## MODIFIED Requirements

### Requirement: Availability of the command

The skill SHALL be usable in any repository where an agent runs, and SHALL tell the agent what to do when the `quizme` command is not installed rather than failing opaquely.

The skill SHALL be distributable as an installable package rather than only as a file to copy, so that a repository can obtain it in one step and update it in another. The packaged skill SHALL carry a version, so an installed copy can be identified and compared with the source.

#### Scenario: Command not installed
- **WHEN** the agent attempts to run `quizme` and it is not on the path
- **THEN** the agent reports that it is not installed and how to install it, and falls back to asking in conversation

#### Scenario: Installing into a repository
- **WHEN** someone installs the packaged skill into a repository that does not have it
- **THEN** an agent working in that repository can use the skill without any file being copied by hand

#### Scenario: Updating an installed copy
- **WHEN** the skill's instructions change at the source and an installed copy is updated
- **THEN** the installed copy carries the new instructions and reports the new version

#### Scenario: Identifying what is installed
- **WHEN** someone inspects an installed copy
- **THEN** its version is visible and can be compared with the source

#### Scenario: Removing it
- **WHEN** someone removes the installed skill from a repository
- **THEN** an agent in that repository no longer offers to open questionnaires, and nothing of the skill is left behind

## ADDED Requirements

### Requirement: One source of truth for the skill's content

The skill's instructions SHALL exist in exactly one authoritative place. Where a second copy is required by the packaging layout, it SHALL be derived from the authoritative one, and a divergence between them SHALL be detected rather than shipped.

#### Scenario: Copies agree
- **WHEN** the packaged skill is built from the source
- **THEN** its instructions are identical to the authoritative copy

#### Scenario: Copies disagree
- **WHEN** one copy is edited and the other is not
- **THEN** the divergence is reported by the project's own checks, and is not published

#### Scenario: Instructions change
- **WHEN** the skill's instructions are revised
- **THEN** only the authoritative copy needs editing for the package to carry the revision
