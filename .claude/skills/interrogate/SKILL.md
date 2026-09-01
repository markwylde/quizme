---
name: interrogate
description: Ask the user several related questions at once by opening a desktop form instead of asking in chat. Use when a handful of decisions are blocking progress on a change - scope, approach, trade-offs - and the user would rather see them all together and answer in any order. Not for a single question, and not when each follow-up depends on interpreting the previous answer.
allowed-tools: Bash(interrogate:*), Read, Write, Edit
license: MIT
compatibility: Requires the interrogate CLI on PATH (go install github.com/markwylde/interrogate@latest).
metadata:
  author: markwylde
  version: "1.0"
---

# Interrogate

Put a batch of decisions to the user as a form rather than a conversation.

You write a YAML questionnaire, run `interrogate` against it, and the user gets
a window with every question on one scrolling page. They answer in any order,
attach a comment to anything, and submit. The answers are written back into the
same file and printed to stdout as JSON.

## When to reach for this

Use a questionnaire when **several related decisions** are open at once and you
could write them all down now. Four open questions blocking a proposal is the
case this exists for.

Ask in conversation instead when:

- **There is only one question.** Opening a window for it is theatre.
- **The enquiry genuinely branches.** If you cannot write question three until
  you have read and thought about the answer to question two, a form is the
  wrong shape - you would be guessing at the branches. `show_if` covers
  *mechanical* branching ("only if they picked sidecar"), not the kind that
  needs your judgement in between.
- **You are checking something quick.** A form is a context switch; do not spend
  one on a detail.

Do not open a questionnaire the user did not ask for in the middle of unrelated
work. Say what you are about to ask about, then open it.

## Where the file goes

For an OpenSpec change:

```
openspec/changes/<change-name>/questionnaires/<slug>.yaml
```

Use a descriptive kebab-case slug for the round of questions - `scope.yaml`,
`storage-decisions.yaml`. A later round is a **new file** in the same directory,
never an overwrite: the answered one is a record of what was decided and when.

Outside a change, put it somewhere equally durable and version-controlled.

## Writing questions worth answering

The user should be able to answer without reconstructing your reasoning first.

- **Say what is being decided**, not what you are uncertain about. "Where do
  answers live?" beats "I wasn't sure about storage."
- **Offer the real options.** If you know the plausible answers, make it a
  `select` rather than an open text box. An open box makes the user do work you
  have already done.
- **Give your view.** If you have a recommendation, put it in the prompt or the
  `help`. The user can then agree in one click or overrule you, which is faster
  than reverse-engineering what you think.
- **Gate the follow-ups.** A question that only matters given a particular
  answer gets a `show_if`, so the user never reads a question that does not
  apply to them.
- **Mark only what truly blocks you** as `required`. Everything else the user
  can skip, and you can ask about later.
- **Keep it short.** Ten questions is a lot to sit down to. If you have thirty,
  you have not finished thinking.

## The format

```yaml
title: Storage decisions
intro: |
  A few things to settle before I write the proposal.

status: pending
submitted_at: null

questions:
  - id: storage                  # unique, stable, used as the answer key
    type: select
    prompt: Where do answers live?
    help: I'd lean towards in-place; two files drift apart.
    required: true
    options: [in-place, sidecar, both]

  - id: storage_why
    type: textarea
    prompt: What makes sidecar the right call?
    show_if:
      storage: sidecar           # only shown when that answer is chosen
```

Types: `select`, `multiselect`, `text`, `textarea`, `number` (with `min`/`max`),
`boolean`, `scale` (`min`/`max`, default 1-5), `rank` (orders its `options`).

Every question also accepts a `comment`, whatever its type - the user always has
somewhere to qualify an answer.

`show_if` takes question ids to expected values, and all clauses must hold. A
list-valued answer matches a single expected value it contains; a list-valued
expectation means "any of these".

You own the questions. `interrogate` owns only `answer`, `comment`, `status`,
and `submitted_at` - it writes those and leaves every other byte of the file
alone, so your comments and formatting survive being answered.

## Checking a questionnaire before you run it

```bash
interrogate --validate <path>
```

Checks the file and exits without opening a window or writing anything: exit `0`
if it is sound, `1` with the errors on stderr if it is not. Use it after writing
a questionnaire, if you want to know it is well formed before putting it in
front of anyone.

It is not a prerequisite. A plain run validates first too, and refuses to open
on a bad file with the same errors — so if you are about to run it anyway, just
run it.

## Running it

The user may take ten minutes over the form, which is longer than a foreground
command should ever wait. **Run it in the background** and pick the answers up
when it exits:

```bash
interrogate openspec/changes/add-dark-mode/questionnaires/scope.yaml
```

Launch that with your tool's background option so the session is notified when
the process exits, then read its stdout.

If you lose track of the run - a restart, an interrupted session - the file
itself tells you what happened. Read its `status`:

| `status`    | What happened                                          |
|-------------|--------------------------------------------------------|
| `pending`   | Not answered yet; the form may still be open           |
| `submitted` | Answered in full                                       |
| `saved`     | Answered in part, deliberately kept                    |
| `dismissed` | The user discarded this session's answers              |

There is no lock file and no sidecar to check. Exit codes match: `0` submitted,
`1` error, `2` dismissed, `3` saved.

## Acting on the answers

**Submitted.** Proceed. Record the decisions where they belong - a design
decision in `design.md`, a scope change in `proposal.md`, a new requirement in
the specs - rather than leaving them only in the questionnaire.

**Saved.** Use what you were given. For the questions still open, ask the user
directly rather than assuming a default; they left them blank on purpose.

**Dismissed.** Do not reopen the form. Ask the user how they want to proceed -
they may have decided the questions were wrong.

**Read the comments carefully.** When a comment qualifies or contradicts the
option chosen on the same question, the comment is the more precise statement of
what the user meant. "sidecar" with a comment saying "only if in-place turns out
to be a pain" is not a decision for sidecar.

A hidden question's answer is never recorded or reported. If you need to know
something a `show_if` hid, ask.

## If the command is missing

`interrogate` has to be installed on the machine:

```bash
go install github.com/markwylde/interrogate@latest
```

If it is not on `PATH`, say so, give that line, and **ask your questions in
conversation instead** - do not stall waiting for an install. It also needs a
desktop session; over a bare SSH connection it exits immediately with an
explanation rather than hanging, and the same fallback applies.
