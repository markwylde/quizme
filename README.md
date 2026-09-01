# interrogate

Ask someone a batch of questions in a desktop form, and get the answers back in
the same YAML file you asked them in.

It exists for agents. When a coding agent has four decisions blocking it, asking
them one at a time in chat is slow, easy to lose, and impossible to answer out
of order. `interrogate` turns them into a form: one scrolling page, answer in
any order, attach a comment to anything, submit once.

![The form](internal/ui/testdata/preview-light.png)

## Install

```bash
go install github.com/markwylde/interrogate@latest
```

Needs a desktop session. It is a native window — Fyne, not a browser — and a
single static binary with its font compiled in, so it looks the same on macOS,
Linux and Windows.

## Use

```bash
interrogate openspec/changes/add-dark-mode/questionnaires/scope.yaml
```

The form opens. On submit, the answers are written back into that same file and
printed to stdout as JSON.

## The questionnaire

```yaml
# Questions are yours. Comments and formatting survive being answered.
title: Storage decisions
intro: |
  A few things to settle before I write the proposal.

status: pending
submitted_at: null

questions:
  - id: storage                   # unique; used as the answer key
    type: select
    prompt: Where do answers live?
    help: I'd lean towards in-place; two files drift apart.
    required: true
    options: [in-place, sidecar, both]

  - id: storage_why
    type: textarea
    prompt: What makes sidecar the right call?
    show_if:
      storage: sidecar            # only shown when that answer is chosen

  - id: priorities
    type: rank
    prompt: Rank these by importance.
    options: [correctness, speed, looks]
```

Answered, that same file becomes:

```yaml
# Questions are yours. Comments and formatting survive being answered.
title: Storage decisions
intro: |
  A few things to settle before I write the proposal.

status: submitted
submitted_at: "2026-09-01T12:00:00Z"

questions:
  - id: storage                   # unique; used as the answer key
    type: select
    prompt: Where do answers live?
    help: I'd lean towards in-place; two files drift apart.
    required: true
    options: [in-place, sidecar, both]
    answer: in-place
    comment: agreed, drift is the real risk

  - id: storage_why
    type: textarea
    prompt: What makes sidecar the right call?
    show_if:
      storage: sidecar            # only shown when that answer is chosen

  - id: priorities
    type: rank
    prompt: Rank these by importance.
    options: [correctness, speed, looks]
    answer:
      - correctness
      - looks
      - speed
```

Note what did **not** change: the comment at the top, the blank lines, the
inline note beside `id`, the quoting, the indentation. `interrogate` owns four
keys — `answer` and `comment` on each question, `status` and `submitted_at` at
the top — and rewrites nothing else. It edits the bytes in place rather than
re-serialising the document, because no Go YAML library round-trips with that
kind of fidelity.

`storage_why` was never shown, so it has no answer. A hidden question records
nothing and is never reported.

### Question types

| Type          | Answer                                    | Notes                             |
|---------------|-------------------------------------------|-----------------------------------|
| `select`      | one of `options`                          | radio buttons, or a dropdown past seven options |
| `multiselect` | any number of `options`                   | checkboxes                        |
| `text`        | a single line                             |                                   |
| `textarea`    | several lines                             | written back as a block scalar    |
| `number`      | a number                                  | `min` / `max` enforced as you type |
| `boolean`     | `true` or `false`                         | two choices, so "no" and "unanswered" stay distinct |
| `scale`       | a whole number                            | `min` / `max`, default 1–5        |
| `rank`        | all of `options`, reordered               | drag, or use the arrows           |

Every question also takes a **comment**, whatever its type, answered or not.

### Conditional questions

`show_if` maps question ids to expected values; every clause must hold. A
list-valued answer matches a single expected value it contains, and a
list-valued expectation means "any of these" — so two lists match when they
overlap.

Questions appear and disappear as you answer. Circular and dangling references
are rejected when the file is read, before any window opens.

## Outcomes

The `status` field is the whole coordination mechanism — no lock file, no
sidecar to keep in sync.

| `status`    | Exit | What happened                                     |
|-------------|------|---------------------------------------------------|
| `pending`   | —    | Authored, not answered yet                        |
| `submitted` | `0`  | Every visible required question was answered      |
| `saved`     | `3`  | Kept deliberately, with required questions open   |
| `dismissed` | `2`  | The answers from that session were discarded      |
| —           | `1`  | Could not be read, shown, or written              |

Closing the window with unsaved answers asks first. Dismissing writes only the
status — answers from an earlier session are left alone.

## Output

```json
{
  "path": "questionnaires/scope.yaml",
  "title": "Storage decisions",
  "status": "submitted",
  "submitted_at": "2026-09-01T12:00:00Z",
  "complete": true,
  "answers": {
    "storage": {
      "prompt": "Where do answers live?",
      "type": "select",
      "answer": "in-place",
      "comment": "agreed, drift is the real risk"
    }
  }
}
```

Diagnostics go to stderr, so stdout is always parseable. The file is written
before anything reaches stdout, so a reader is never ahead of it.

## For agents

A person may take ten minutes over the form — longer than a foreground command
should wait. Run it in the background and read the result when the process
exits; if you lose the run, the file's `status` tells you what happened.

The `interrogate` skill in `.claude/skills/` covers when to use a questionnaire
rather than asking in conversation, where to put the file, and how to act on
each outcome.

## Development

```bash
make build                                  # ./interrogate
make test                                   # everything
make run FILE=examples/demo.yaml            # try the form

INTERROGATE_RENDER=1 go test ./internal/ui/ -run RenderPreview   # refresh the screenshots
```

`examples/demo.yaml` is a realistic questionnaire to poke at.

## Licence

Inter is bundled under the SIL Open Font License; see
`internal/ui/fonts/LICENSE-Inter.txt`.
