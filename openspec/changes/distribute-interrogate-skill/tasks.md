## 1. Layout

- [x] 1.1 Move `SKILL.md` to `skills/interrogate/SKILL.md` as the authoritative copy, and verify the existing skill tests still pass against the new path
- [x] 1.2 Verify `npx skills` discovers the skill from that path by listing it against this repository
- [x] 1.3 Remove the Claude plugin manifests, and verify nothing left in the repository refers to them

## 2. One source of truth

- [x] 2.1 Git-ignore the generated copy under `.claude/skills/interrogate/` and stop tracking it, and verify a fresh clone carries exactly one copy of the skill
- [x] 2.2 Keep the `make skill` target that regenerates the copy, and verify running it after an edit reproduces the authoritative file byte for byte
- [x] 2.3 Make the divergence test skip when no copy has been generated and fail when a generated copy differs, and verify both behaviours
- [x] 2.4 Assert the skill's frontmatter declares a version, and verify the test fails when it is missing

## 3. Verify the real install path

- [x] 3.1 Install the skill into a scratch repository with `npx skills add`, and verify only the interrogate skill lands and it is the authoritative text
- [x] 3.2 Change the instructions, bump the version, run `npx skills update`, and verify the installed copy carries the new text and reports the new version
      → verified by re-running `skills add`, which is the update path for a local source: `skills update` skips local sources ("No project skills to update"). The git-source path cannot be verified until the repository has a remote — see 4.2.
- [x] 3.3 Remove the skill with `npx skills remove` and verify nothing of it remains

## 4. Documentation

- [x] 4.1 Document installing, updating, and removing in the README, naming `--skill interrogate`, and verify the commands work as written by following them from a clean state
- [x] 4.2 Note that the install command needs a GitHub remote, and which copy of `SKILL.md` is authoritative, and verify the note matches what the tests enforce
