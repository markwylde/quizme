## Why

When an agent needs several decisions from a human, it currently asks them one at a time in chat: slow, easy to lose, and impossible to answer out of order or revise. A questionnaire that opens as a real form lets the human see every question at once, answer in any order, attach a comment to anything, and hand back a structured record the agent can read — with the questions and answers living in one durable file inside the change.

## What Changes

- Add a YAML **questionnaire format**: a document of questions with ids, prompts, field types, options, optional conditional visibility, and a universal comment box on every question. The same file later carries the answers and a status lifecycle.
- Add **`quizme`**, a Go CLI that takes a questionnaire path, opens a native desktop form (Fyne), and on submit writes answers back into the same file and prints them as JSON on stdout.
- Answers are written by **splicing text**, not re-serializing: `quizme` owns only the `answer:` and `comment:` keys of each question and leaves every other byte — comments, spacing, key order — untouched.
- The file carries a `status` of `pending` / `submitted` / `saved` / `dismissed`, so an agent can tell what happened without a lock file or sentinel, and exit codes mirror it.
- Add an **`quizme` skill** that an opsx agent invokes: it authors the questionnaire into `openspec/changes/<change>/questionnaires/<slug>.yaml`, runs the binary in the background, and reads the answers back when it exits.

## Capabilities

### New Capabilities
- `questionnaire-format`: the YAML document schema — question ids, field types, options, conditional visibility, comments, answers, and the status lifecycle that describes how a questionnaire was completed.
- `quizme-cli`: the command-line contract and desktop form behaviour — argument handling, rendering, validation, close-without-submit semantics, in-place answer writing, stdout output, and exit codes.
- `quizme-skill`: the agent-facing contract — when to reach for a questionnaire, where the file goes, how the binary is launched and awaited, and how answers are consumed.

### Modified Capabilities

_None — this is a greenfield project._

## Impact

- New Go module at the repo root: `main.go`, questionnaire parser/splicer, Fyne UI, custom theme, and custom `rank` / `scale` widgets that Fyne does not ship.
- Dependencies: `fyne.io/fyne/v2` (UI, OpenGL-rendered, no browser engine), `gopkg.in/yaml.v3` (parsing and node positions only).
- Distribution: `go install`, so the binary is on `PATH` as `quizme`; `go run .` during development.
- The skill must be installable into *other* repos where opsx runs — this repo builds the binary but the skill has to travel separately.
- Requires a desktop session. Headless CI cannot open the form, so any non-interactive path must fail loudly rather than hang.
