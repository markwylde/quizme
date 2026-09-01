## Why

The skill lives at `.claude/skills/interrogate/SKILL.md` in this repository, which is the one place it is least needed: this is where the tool is built, not where questions get asked. Every other repository an agent works in needs its own copy, and copying a file by hand means the copies drift — a skill updated here silently leaves stale instructions behind everywhere else. It needs an install command and an update command.

## What Changes

- Distribute the skill through **`npx skills`** (`vercel-labs/skills`), the cross-agent skill installer: `npx skills add markwylde/interrogate --skill interrogate` puts it wherever the reader's agent looks for skills, and `npx skills update` refreshes it.
- Move the skill to the layout that installer discovers — `skills/<name>/SKILL.md` at the repository root — and make that copy the authoritative one.
- Stop committing the copy under `.claude/`. It exists only so this repository's own agents see the skill without installing it, so it becomes a generated, ignored artefact rather than a second file to keep in step.
- Document installing, updating, and removing in the README.
- **The skill's content is unchanged.** This is packaging: what the agent is told to do stays exactly as specified.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities
- `interrogate-skill`: the skill's availability requirement gains a defined distribution and update path, rather than only describing what to do when the command is missing.

## Impact

- `skills/interrogate/SKILL.md` becomes the authoritative copy; `.claude/skills/interrogate/` becomes generated and git-ignored.
- README gains install, update, and remove instructions.
- The repository needs a GitHub remote before `npx skills add markwylde/interrogate` resolves for anyone else.
- No change to the Go binary, the questionnaire format, or the form.

## Non-goals

- Publishing to npm. The installer reads the repository directly; there is nothing to publish.
- A Claude Code plugin. It was the original plan, but it can only ship a whole repository directory — 2.6 MB of fonts, Go source and unrelated skills for an 8 KB file — and only serves one agent.
- Bundling the `interrogate` binary. The skill travels; the binary is still `go install`, and the skill already handles the case where it is missing.
