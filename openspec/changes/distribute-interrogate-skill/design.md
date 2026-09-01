## Context

See proposal.md for why. The shape of the work is set by how Claude Code plugins actually load, confirmed against the official marketplace already on this machine:

```
  <repo>/
    .claude-plugin/
      plugin.json          name, description, version, author
      marketplace.json     makes the repo installable by name
    skills/
      interrogate/SKILL.md the payload
```

A plugin is installed by adding its repository as a marketplace and then installing the plugin from it. Nothing is published to a registry; the repository is the distribution.

The complication is that this repository already has the skill at `.claude/skills/interrogate/SKILL.md`, where its own agents pick it up. A plugin payload at `skills/interrogate/SKILL.md` would be a second copy of the same file.

## Goals / Non-Goals

**Goals:**
- Install into another repository in one command, update in another.
- One authoritative copy of the instructions, with divergence caught by the build rather than discovered in use.
- An installed copy that can be identified by version.

**Non-Goals:**
- Publishing to any marketplace or registry beyond this repository.
- Bundling the binary. The plugin carries the skill; `interrogate` is still `go install`, and the skill already covers it being missing.
- Any change to what the skill tells an agent to do.

## Decisions

### The plugin payload is authoritative; the repo's own copy is generated

`skills/interrogate/SKILL.md` becomes the source. `.claude/skills/interrogate/SKILL.md` becomes a copy generated from it, with a Go test asserting the two are byte-identical.

*Why this direction:* the plugin is the artefact other people consume, so it should be the thing that is edited and reviewed. The local copy exists only so this repository's own agents see the skill without installing the plugin into it.

*Why not a symlink:* it would work on macOS and Linux and quietly not on Windows checkouts, and a plugin payload that is a symlink is a poor thing to ship.

*Why not drop the local copy:* then the one repository where the skill is developed is the one place it cannot be exercised, which is how stale instructions survive.

*Enforcement:* a test in the existing suite reads both files and fails on any difference, naming which to regenerate. The existing skill tests already parse `SKILL.md` and check its examples against the tool; they move to the authoritative path and keep working.

### The repository is its own single-plugin marketplace

`.claude-plugin/marketplace.json` lists one plugin, whose files sit at the repository root. Installing is `/plugin marketplace add markwylde/interrogate` then `/plugin install interrogate@interrogate`.

*Why root rather than `plugins/interrogate/`:* a nested layout is for a marketplace carrying several plugins. This repository carries one, and the flatter layout means the payload path is short enough to be obvious.

*Trade-off:* if a second plugin is ever wanted here, the layout has to move. That is a rename, and the version number is what makes it survivable.

### The version is set by hand, in one place

`plugin.json` carries the version, and the `SKILL.md` frontmatter's `metadata.version` is checked against it by the same test that checks the copies agree.

*Why not derive it from git tags:* the plugin is installed from a checkout of a branch, not from a release artefact, so a tag is not reliably present. A literal in the manifest is what an installed copy actually reports.

*Consequence:* bumping the version is a step in the release routine, and the test catches a `SKILL.md` bumped without the manifest.

## Risks / Trade-offs

- **Two copies drift anyway**, because the test is only run when tests are run. → It runs in the same `go test ./...` as everything else, which the tasks already require before delivery.
- **The generated copy gets edited by hand**, since nothing physically prevents it. → The failure message names which file is authoritative and how to regenerate, so the mistake costs one command.
- **Plugin manifest fields change under us.** → The layout was read from the installed official marketplace rather than from memory, and the install path is verified for real in the tasks rather than assumed.
- **Someone installs the plugin without the binary.** → Already handled: the skill's not-installed requirement tells the agent to give the install line and fall back to conversation.

## Open Questions

- Whether to also expose a `/interrogate` slash command alongside the skill. It would be a thin wrapper over the same instructions, and can be added later without changing the layout or the specs.
