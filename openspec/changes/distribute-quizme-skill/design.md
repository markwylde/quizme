## Context

See proposal.md for why. The approach changed once the packaging was tried for real, so the reasoning below records both what was rejected and why.

The first plan was a Claude Code plugin. It works — the plugin installs, reports its version, and exposes the skill — but a plugin ships a whole repository directory. Installing it copied 2.6 MB into the plugin cache: the Go source, 1.6 MB of bundled fonts, `openspec/`, every test file, and this repository's own `.claude/skills/` including seven unrelated OpenSpec skills. Restructuring into `plugins/quizme/` would have fixed the size, at the cost of a second manifest, a second version to bump, and a distribution that only ever serves Claude Code.

`npx skills` (`vercel-labs/skills`) reads a repository, finds `SKILL.md` files, and copies or symlinks the ones you pick into whichever agent's skills directory you name — `.claude/skills/` for Claude Code, `.agents/skills/` for Codex and Cursor, and so on for some seventy-odd others. Installing the same skill this way copied 8 KB.

```
  plugin                       npx skills
  ------                       ----------
  add marketplace              npx skills add <repo> --skill quizme
  install plugin
  2.6 MB, whole repo           8 KB, one file
  Claude Code                  77+ agents
```

## Goals / Non-Goals

**Goals:**
- Install into another repository in one command, update in another.
- One committed copy of the instructions, so there is nothing to drift.
- An installed copy that can be identified by version.

**Non-Goals:**
- Publishing to npm, or to any registry. The repository is the distribution.
- Supporting both a plugin and the installer. One path, documented once.
- Any change to what the skill tells an agent to do.

## Decisions

### `skills/quizme/SKILL.md` is the authoritative copy

That is the layout `npx skills` discovers, and it is the file other people receive, so it is the one to edit and review.

*Verified rather than assumed:* running the installer against this repository listed `quizme` from that path, and a real install into a scratch repository produced `.claude/skills/quizme/SKILL.md` and nothing else.

### The copy under `.claude/` is generated and ignored

`.claude/skills/quizme/SKILL.md` is what this repository's own agents read, because that is where Claude Code looks. It is produced by `make skill` and git-ignored.

*Why not commit it:* two committed copies of the same instructions is a drift problem that needs a test to police. An ignored copy cannot be shipped stale, because it is never shipped at all.

*Why not a symlink:* it would work on macOS and Linux and quietly not on a Windows checkout.

*Why keep it at all:* otherwise the one repository where the skill is developed is the one place it cannot be exercised, which is how stale instructions survive.

*Consequence:* the divergence test skips on a fresh clone where the copy does not exist yet, and asserts equality for anyone who has generated one. That is the right shape — it catches a hand-edited copy without failing a checkout that simply has not run `make skill`.

### The version lives only in the skill's frontmatter

`metadata.version` in `SKILL.md` travels with the file, so an installed copy reports the version it actually carries.

*Why one place:* the plugin plan had the version in both `plugin.json` and the frontmatter, needing a test to keep them equal. With no manifest there is nothing to disagree with.

### Discovery finds more than our skill, and that is fine

`npx skills add markwylde/quizme` with no `--skill` lists seven skills, because this repository also carries the OpenSpec skills under `.claude/`. The documented command names `--skill quizme`.

*Alternative considered:* moving the OpenSpec skills out. They are this repository's own tooling and belong where their own workflow put them; contorting the repo to tidy one listing is the wrong trade.

## Risks / Trade-offs

- **`npx skills` is third-party, and could change or go away.** → A `SKILL.md` in a conventional location is portable by hand whatever happens to the installer; the file is the artefact, not the packaging.
- **No GitHub remote yet**, so the documented install command does not resolve for anyone else. → Called out in the tasks and the README; installing from a local path works meanwhile and is what the tasks verify.
- **The generated copy gets edited by hand.** → The test names which file is authoritative and how to regenerate.
- **Someone installs the skill without the binary.** → Already handled: the skill's not-installed requirement tells the agent to give the install line and fall back to conversation.

## Open Questions

- Whether to publish the repository under a `skills.sh` listing so it is discoverable by `npx skills find`. Independent of the layout, and can be done later.
