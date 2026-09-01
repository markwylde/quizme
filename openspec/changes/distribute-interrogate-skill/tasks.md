## 1. Plugin layout

- [ ] 1.1 Move `SKILL.md` to `skills/interrogate/SKILL.md` as the authoritative copy, and verify the existing skill tests still pass against the new path
- [ ] 1.2 Add `.claude-plugin/plugin.json` with the plugin's name, description, version, and author, and verify it parses as JSON and matches the fields a real installed plugin carries
- [ ] 1.3 Add `.claude-plugin/marketplace.json` declaring this repository as a single-plugin marketplace, and verify it parses and names the plugin at the root

## 2. One source of truth

- [ ] 2.1 Generate `.claude/skills/interrogate/SKILL.md` from the authoritative copy and verify the two files are byte-identical
- [ ] 2.2 Add a test asserting the copies agree, whose failure names the authoritative file and how to regenerate, and verify it fails when one copy is edited
- [ ] 2.3 Add a test asserting the `SKILL.md` frontmatter version matches `plugin.json`, and verify it fails when only one is bumped
- [ ] 2.4 Add a make target that regenerates the copy, and verify running it after an edit makes the tests pass again

## 3. Verify the real install path

- [ ] 3.1 Install the plugin into a scratch repository from this one and verify an agent there is offered the interrogate skill
- [ ] 3.2 Change the instructions, bump the version, update the installed copy, and verify it reports the new version and carries the new text
- [ ] 3.3 Remove the plugin from the scratch repository and verify nothing of the skill remains

## 4. Documentation

- [ ] 4.1 Document installing, updating, and removing the plugin in the README, and verify the commands work as written by following them from a clean state
- [ ] 4.2 Note in the README which copy of `SKILL.md` is authoritative, and verify the note matches what the tests enforce
