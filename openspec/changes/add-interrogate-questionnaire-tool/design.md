## Context

Greenfield repo — it currently holds only OpenSpec scaffolding. See proposal.md for motivation. The shape of the problem is set by three constraints the user fixed during exploration:

- The UI must be a real desktop window, not a browser or Electron shell.
- Questions and answers live in **one** file. A separate answers file was rejected outright: two files drift apart the moment the questions are edited.
- The agent that launches the form cannot sit blocked on it, because a human may take ten minutes to respond and a foreground command call is capped well below that.

```
  agent authors                         reads back
  questionnaire  --------+          +------------------+
        |                |          |                  |
        v                v          |                  |
  scope.yaml  --> interrogate --> splice answers  --> scope.yaml
  status: pending    (window)     + status + stdout    status: submitted
```

## Goals / Non-Goals

**Goals:**
- A questionnaire file that is pleasant to read and edit by hand, before and after it is answered.
- Byte-level preservation of everything in that file the tool does not own.
- An outcome an agent can determine from the file alone, with no lock file, sidecar, or sentinel.
- A form that looks considered rather than like a default widget toolkit demo.

**Non-Goals:**
- Multi-page or wizard navigation — one scrolling page, decided during exploration.
- Concurrent or remote responders. One person, one machine, one questionnaire at a time.
- Editing questions from within the form. The agent authors, the human answers.
- Headless or terminal-based operation. If there is no display, the command fails fast.

## Decisions

### Splice text, do not round-trip YAML

`interrogate` parses the document to locate nodes, then edits the original bytes by line range — replacing or inserting only `answer:`, `comment:`, `status:`, and `submitted_at:` — and writes the result. It never re-serializes the whole document.

*Why:* Go has no round-trip-fidelity YAML library. `yaml.v3` via `yaml.Node` keeps comments and key order but reflows indentation, quoting, and line breaks; `goccy/go-yaml` behaves similarly. Since the file is authored by an agent, read by a human, and rewritten by a tool, formatting churn on every submit would be constant and ugly.

*How:* `yaml.Node` carries `Line` and `Column` for every node. That is enough to find each question's block, find an existing `answer` key within it, and compute the line range to replace — or, absent one, the insertion point and indent.

*Alternatives considered:* Naive marshal (fifteen minutes' work, destroys the file's shape); a sidecar answers file (rejected by the user — drift); JSON instead of YAML (no comments, worse to hand-edit).

*Cost:* This is the fiddliest part of the build. Block-scalar answers (a multi-line `textarea`) and re-indentation of inserted blocks are where the bugs will be, so it needs a golden-file test suite of its own.

### Status in the document, not a lock file

The top-level `status` field is the coordination channel between tool and agent. `pending` on authoring; `submitted`, `saved`, or `dismissed` when the window closes.

*Why:* Considered and rejected: polling the file's modification time. The agent writes the file itself, so its mtime is already fresh at launch, and any autosave would produce false positives. A lock file adds a second artifact to keep in sync — the same objection that killed the sidecar. Status is self-describing, survives an agent restart, and doubles as the record of what happened.

The primary path is not polling at all: the agent launches `interrogate` in the background and is notified on exit, reading stdout and the exit code. Status is what makes recovery possible when that link is lost.

### Exit codes

`0` submitted, `1` error, `2` dismissed, `3` saved-partial. Distinct codes for the three human outcomes mean the agent branches without parsing anything, and `1` stays conventional for failure.

### Fyne, with a custom theme

*Why:* Cross-platform from one codebase, a single static binary, no cgo pain on macOS, and no browser engine. Gio was the alternative — more control, considerably more layout code for a form this conventional. `andlabs/ui` wraps genuinely native widgets but is effectively unmaintained.

*Consequence to accept:* Fyne draws its own widgets via OpenGL, so this will not look like a stock Mac app. Given the user's "not web-based, otherwise flexible", that trade is fine — but it means the default Material-ish theme is what people will judge, so a custom `fyne.Theme` (colors, spacing, sizes) plus a bundled font is part of the build, not a polish item afterwards.

*Consequence to budget for:* Fyne's widget set is thin. `rank` (drag to reorder) and a decent segmented `scale` control do not exist and must be written. These are the second-largest chunk of work after the splicer.

### Conditional visibility recomputed on every change

`show_if` is evaluated after any answer changes, and the question list is rebuilt. Hidden questions are never required and never written to the output, even if they were answered before being hidden.

*Why not keep an orphaned answer:* an answer to a question the human can no longer see is a trap for whoever reads the file later. Cycles and dangling references are rejected at load, so evaluation cannot loop.

### Distribution splits in two

The binary ships via `go install` and lives on `PATH` as `interrogate`; `go run .` during development. The skill has to reach the *other* repos where opsx runs, and this repo cannot put it there.

## Risks / Trade-offs

- **The splicer corrupts a file.** → Write atomically via a temp file and rename, so a failure leaves the original intact; golden-file tests covering comments, block scalars, nested sequences, CRLF, and re-answering an already-answered file.
- **The custom theme and widgets eat the schedule.** → Get the format, splicer, and status lifecycle correct behind stock widgets first; theme and `rank`/`scale` are a separable later pass against a working tool.
- **Fyne doesn't look good enough and the user rejects it.** → The UI layer is the only Fyne-dependent code; format, splicer, and CLI contract are UI-agnostic, so a swap to Gio would not restart the project.
- **The agent forgets it launched a form, or is interrupted.** → Status in the file makes the outcome recoverable at any later point.
- **No display available (CI, SSH).** → Fail immediately with a clear message. Hanging forever would be far worse than an error.
- **An agent overuses questionnaires and makes them annoying.** → The skill spec makes "one question, or genuinely branching" an explicit conversation case rather than a form.

## Open Questions

- Whether `interrogate` should accept a `--print-only` or `--validate` mode for authoring-time checks without opening a window. Useful, but does not change the format, the approach, or the task breakdown.
- Exactly how the skill is distributed to other repos (plugin, or copied into `.claude/skills`). The skill's content is unaffected either way.
