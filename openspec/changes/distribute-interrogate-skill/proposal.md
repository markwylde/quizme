## Why

The skill lives at `.claude/skills/interrogate/SKILL.md` in this repository, which is the one place it is least needed: this is where the tool is built, not where questions get asked. Every other repository an agent works in needs its own copy, and copying a file by hand means the copies drift — a skill updated here silently leaves stale instructions behind everywhere else. Packaging it as a Claude Code plugin makes installing it one command and updating it another.

## What Changes

- Package the skill as a **Claude Code plugin**, so it can be installed into any repository or user configuration from this repository rather than copied.
- Add the plugin manifest and directory layout a plugin requires, with the existing `SKILL.md` as its payload.
- Version the plugin, so an installed copy can be identified and updated rather than guessed at.
- Document installing, updating, and removing it in the README, replacing the current implicit "it is in this repo somewhere".
- **The skill's content is unchanged.** This is packaging: what the agent is told to do stays exactly as specified.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities
- `interrogate-skill`: the skill's availability requirement gains a defined distribution and update path, rather than only describing what to do when the command is missing.

## Impact

- New plugin manifest and layout at the repository root, alongside the existing `.claude/skills/interrogate/`.
- The two copies of the skill must not diverge: whichever becomes the source, the other has to be generated from or point at it, and something has to fail when they disagree.
- README gains install, update, and remove instructions.
- No change to the Go binary, the questionnaire format, or the form.

## Non-goals

- Publishing to any marketplace or registry. Installing from this repository is enough.
- Bundling the `interrogate` binary itself into the plugin. The plugin carries the skill; the binary is still `go install`, and the skill already handles the case where it is missing.
