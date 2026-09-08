## Why

Tick one box in a multiselect and `interrogate` writes a file it cannot read back. The answers reach the disk, then the re-read fails, the command exits `1`, and the caller is handed an error for a questionnaire the responder submitted perfectly well — with the file left in a state no tool can parse, so the answers are effectively lost.

The cause is one line in `renderEntry`: a value that marshals to a single line is placed inline after `key: `. That is right for a scalar, but a one-element sequence also marshals to a single line — `- TypeScript` — and `answer: - TypeScript` is not YAML. Two selections come out correctly, because two lines take the block branch, which is why the existing tests never saw it: they only ever use two-element lists.

## What Changes

- **A one-item list answer is written as a block sequence**, exactly as a two-item one is, so the document it produces still loads.
- **The inline-versus-block decision stops being a line count.** `renderEntry` decides from the value it was handed: only a scalar may share the key's line.
- **The list tests cover the one-item case** for `multiselect` — the only question type that can produce one, since a `rank` answer is always a permutation of two or more options — so the branch that broke is held by a test rather than by the happy accident of every fixture having two entries.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities
- `interrogate-cli`: the "Writing answers without disturbing the document" requirement now states that the written document must still load, whatever the shape of the answers, with scenarios for a single-item list and for text that reads like one.

## Impact

- `internal/questionnaire/splice.go`: `renderEntry` gains a value-shaped test for whether the body may sit inline.
- `internal/questionnaire/splice_test.go`: `TestSpliceListAnswers` extended, or joined by a sibling, covering one-item lists.
- No change to the questionnaire format, the form, the flow-style splicer, or the JSON on stdout — a flow-mapping question already renders its answer as `[TypeScript]` and was never affected.
- Files already written by the broken path stay broken; they are hand-fixable by moving the item under the key, and no migration is warranted for a tool this new.

## Non-goals

- Changing how multi-line or scalar answers are written. Block scalars, quoting, and the flow-mapping path are all untouched.
- Rewriting `renderEntry` into a general YAML emitter. It formats one entry for this document and stays that narrow.
- Recovering a questionnaire that a previous run has already corrupted.
