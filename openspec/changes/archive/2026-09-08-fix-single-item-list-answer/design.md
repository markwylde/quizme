## Context

See proposal.md — Why. `renderEntry` formats one `key: value` entry for the splicer. It marshals the value on its own, then picks between three shapes: inline after the key, a block scalar whose indicator stays on the key line, or a block nested beneath the key. The pick is made on `len(body)`, and a single line is taken to mean a scalar.

The values that reach it have been through `NormalizeAnswer`, so each is a `string`, `[]string`, `float64`, `bool`, or `int` — plus the `status` and `submitted_at` strings. `[]string` is the only collection, and `yaml.Encoder` renders it as a block sequence: `["a", "b"]` becomes two lines, `["a"]` becomes one. So the line count separates scalars from collections for every list but the one-element one.

The flow-style path is separate and unaffected: a question written as `{id: q, ...}` has its answer rendered by `marshalFlowValue`, which forces flow style and yields `[a]`.

## Goals / Non-Goals

**Goals:**
- Every answer `quizme` writes produces a document that loads again.
- The decision is made on grounds that stay true as values change — the value's shape, not the size of its rendering.

**Non-Goals:**
- Any change to the block-scalar branch, the flow branch, quoting, or indentation.
- A general-purpose emitter. `renderEntry` serves this splicer only.

## Decisions

**Decide from the value, not from the line count.** A collection's rendering goes beneath the key; a scalar's may sit after it. That is the actual YAML rule — a block sequence entry or a mapping key cannot follow `key: ` on one line — so stating it directly removes the class of bug rather than the instance. The alternative, testing whether `body[0]` starts with `- `, fixes the same case with a rule about text rather than about YAML: it reads as a string hack, and it says nothing about a single-key mapping, which would break identically if one ever reached here.

**Keep the empty collection inline.** An empty `[]string` marshals to the flow form `[]`, which is a legal inline value, and forcing it under the key would render `answer:` with a bare `[]` on the next line for no gain. So a collection goes inline when its single line is already flow — it opens with `[` or `{`. In practice `NormalizeAnswer` turns a blank answer into `nil` and the splicer removes the key entirely, so this guards a path rather than serving one.

**Classify with `reflect.Kind`, not a type switch on `[]string`.** The narrower switch would work today and break silently the day an answer type is added — a `[]any` from a hand-edited file, say. Kind covers slice, array, map, and struct in one clause, and `reflect.ValueOf(nil).Kind()` is `Invalid`, which falls through to the scalar branch and keeps `null` inline where it already is.

**Test the one-item case where the list case already lives.** `TestSpliceListAnswers` covers `multiselect` and `rank` together; the one-item case belongs beside it, including a single-option `rank`, which is the other way to produce a one-element list. The shared `render` helper already re-loads what it renders and fails if it will not parse, so these tests fail against the current code without any new assertion machinery.

## Risks / Trade-offs

- **`reflect` on a hot path** → the splicer runs once per save over a handful of keys; a `Kind()` call is not worth avoiding.
- **A string that looks like a sequence entry** — an answer of literally `- TypeScript` → it is a scalar, so it stays on the inline branch, and the encoder quotes it as `'- TypeScript'`. Worth a test rather than an argument, since it is the mirror image of the bug.
- **The corrupted-file case is left to the user** → the write is atomic, so the file is whole; it is just unparseable, and it is one line to fix by hand. Detecting and repairing it would mean a tool that edits documents it did not write, which is a bigger commitment than the bug deserves.
