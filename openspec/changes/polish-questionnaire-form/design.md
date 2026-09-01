## Context

`interrogate` shipped and was reviewed by answering a questionnaire with it. See proposal.md for what that review asked for. The relevant constraints from the existing build:

- Cards are constructed once and shown or hidden, never rebuilt, so that a question appearing cannot steal focus from a control being typed in. Anything added here has to respect that.
- Fyne's widget set is thin. `scale` and `rank` are already custom widgets; a collapsible comment is a third.
- The footer already carries the outstanding-required-questions summary, which is the natural home for a progress indicator.

## Goals / Non-Goals

**Goals:**
- Give the questions the vertical space currently spent on chrome.
- Keep every comment one action away, so collapsing costs nothing but a click.
- Let an agent check a questionnaire it wrote without opening a window at anyone.

**Non-Goals:**
- Any change to the questionnaire format, the splicer, or the skill.
- Theme or palette work. The look was scored 5; this is density, not appearance.
- Structural navigation — sections, jump lists, per-question collapse. It stays one scrolling page.

## Decisions

### The header moves into the scroll, the footer stays pinned

`container.NewBorder(header, footer, ...)` becomes `NewBorder(nil, footer, ...)` with the header prepended to the scrolling list.

*Why asymmetric:* the two are read at different rates. The title and intro are context, read once before starting; keeping them on screen for the next ten minutes buys nothing. The actions and the progress indicator are referenced from anywhere in the page, so they stay put.

*Consequence:* the progress indicator must live in the footer beside the existing summary, not in the header where it would scroll away.

### Comment fields collapse behind a per-question toggle

Each card gets a small "Add comment" control. Pressing it reveals the entry and focuses it. A question that loads with a comment already on it starts expanded, and a question whose comment is non-empty shows that in the collapsed label ("Comment" rather than "Add comment") so a collapsed comment is never invisible.

*Why not remove the boxes and use a dialog:* a comment is often written while looking at the answer; a modal hides exactly the thing being commented on.

*Why not auto-expand on focus-through:* keyboard traversal would open every field in turn, which is worse than the problem.

*Implementation:* the entry is built with the card as today and simply hidden, so `q.Comment` binding, dirty tracking, and the show/hide-not-rebuild rule all keep working unchanged. Only the resting visibility differs.

### Progress counts visible questions, not all questions

The indicator reads over the same visibility pass that drives `show_if`. A hidden question is not applicable, so it is in neither the numerator nor the denominator — the total moves as conditions resolve.

*Why a moving total is right:* the alternative is counting questions the responder will never see, which would leave the form stuck short of complete with nothing to act on. A total that grows when a follow-up appears is honest about what a conditional questionnaire is.

*Alternatives considered:* a progress bar across the top (scrolls away, and duplicates the scrollbar); a per-question tick in the margin (fine, but does not answer "how much is left").

### `--validate` loads and reports, sharing the command's existing path

The flag reuses the load-and-validate step that already runs before any window opens, then returns instead of presenting. Success is exit `0`; a validation failure is exit `1`, the same code and the same stderr text the command already produces when it refuses to open.

*Why the codes do not need a new value:* validation is not a responder outcome, so it does not compete with `2` dismissed or `3` saved.

*Note on scope:* the review that asked for this described it as "validates the questionnaire before showing it to the user", which the command already does. The flag's actual value is checking **without** showing — an agent verifying a file it just wrote. It is built for that, and the usage text says so, so nobody reaches for it expecting different behaviour from a plain run.

## Risks / Trade-offs

- **A collapsed comment is a comment nobody writes.** → The affordance stays on every card rather than behind a menu, and a question carrying a comment says so when collapsed. If the review after this one says comments dropped off, the resting state is one line to change back.
- **The scrolling header loses the title once you are three questions in.** → The window's own title bar carries it, so it is never actually gone.
- **A moving progress total could read as going backwards** when an answer is withdrawn and a follow-up disappears. → It is the truthful count; the alternative is a total that includes questions the responder cannot reach.
- **`--validate` invites a caller to validate then run**, doing the work twice. → The skill should say to just run it; a bad questionnaire fails fast either way.
