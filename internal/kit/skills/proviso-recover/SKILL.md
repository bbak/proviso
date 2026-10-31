---
name: proviso-recover
description: Recover the Proviso spec tree of an existing system from what it does, one capability per amendment, for the business to confirm; and, on request, check a recovered capability against the code with acceptance tests written from the tree. Use when the engineer asks to recover, reverse-engineer or bring an existing system or one of its capabilities into a spec tree, or to check a recovered capability against the code. Writes the tree as built, raises the questions that confirm it, hands over to proviso-check, and never changes the system's code. Does not accept.
---

# Recovering a spec tree

You write the spec tree of a system that exists already: what it does
today, as built, for the business to confirm. The code decided what the
system does, whether or not anyone meant it to; you write those
decisions down, and you ask about every one nobody can point to. Nothing
you write is invented, and nothing you find is dropped.

A recovered tree is accepted like any other: one capability per
amendment, checked and accepted on its own, so that every later
amendment has a baseline to compare with. Until the business confirms
it, its behaviour is doubted, not undecided: the rules say what the
system does, and the questions say what nobody has confirmed.

## Proviso's command

Wherever this skill says `proviso`, run Proviso's command: the one
`.proviso/build.md` names, or `proviso` from the path if it names none.
Look nowhere else.

## Who answers what

- **The code** decides what a rule says: what the system does is what
  the tree states.
- **The engineer,** in this session, answers how the tree is cut into
  capabilities and features, whom to ask about each, and the technical
  questions: mechanisms, names, contracts, constraints.
- **The business** confirms what the system should do and what its
  words mean. Neither you nor the engineer answers that: every rule
  stays as built, and the question goes into `questions.md`.

Ask the engineer one question at a time, with your recommendation and
why. Write a decision into the tree only once the engineer has made it.

## Sources, and which one decides

| Source | What it gives | Where it disagrees with the code |
| --- | --- | --- |
| The code | the rules: events, conditions, both outcomes | — |
| The tests | Examples rows, from their inputs and expected results; boundaries | the rule as the code does it; a question quoting the test |
| Docs, specifications, tickets, commit messages | terms, Goals, Rationales: the reasons code never states | the rule as the code does it; a question: "the docs say X, the code does Y: which is intended?" |
| The engineer | structure, contacts, technical decisions | the engineer's answer goes into the tree |

Never write what a document says instead of what the code does, even
where the document seems more right: the rule would describe behaviour
the system does not have, and claim a decision nobody made in it.

**Behaviour a default decided** is the most important thing you find:
where the code decides nothing itself and a library, a framework, a
platform or an unset value does, such as a rounding mode, a sort order,
an encoding or a timeout. Write the rule as built, with the default's
outcome in its Examples, and ask about it by name:
`- [billing.invoice.round] Rounding is the library's default, half to even: is this intended? (ask: Billing)`.

<!-- proviso: writing.md -->

## Where recovering departs from the writing guide

The writing guide is written for bringing a request into the tree. In a
recovered tree the code has already decided, so where the two differ,
this section holds:

- **What a line traces back to:** to the code, read as the sources
  above say, as well as to the business, the engineer and the tree.
- **No `?`:** the system does something in every case, so write what it
  does. A doubt is a question in `questions.md`, never a `?`. A
  recovered capability is accepted only without one.
- **Glossary terms:** define a term as the code and the docs use it,
  with Status `defined`, and ask its contact to confirm the definition
  in `questions.md`.
- **Contacts:** the engineer names them; CODEOWNERS and the history of
  the code may suggest whom. Where nobody is known, the engineer is the
  contact until someone else is named: whom to ask is true of them.
- **Goals** say what the system, capability or feature does and for
  whom, from the code and the docs, for its contact to confirm.
- **Rationales** state the reasons a doc, a commit message or the
  engineer gives. Where no source gives one, the Rationale says so, as
  "No reason was recorded when this was built.", and a question asks
  for it. Never infer a plausible reason: it would read as a decision
  someone made.

## Questions

The questions in `questions.md` are the agenda for confirming the tree.
Each is tagged with the element it concerns and names whom to ask:

- one per feature: `- [cash.withdrawal] Do these rules state the intended behaviour? (ask: Cash Operations)`;
- one for each behaviour a default decided;
- one for each disagreement between the code and a test or a doc;
- one for each recovered term and each missing reason.

A question about every single rule would drown the agenda; the feature's
question covers its plain rules.

## 1. The branch

Each run is one amendment on a branch of its own, made from the target,
usually the default branch, and named after what it recovers, such as
`recover-system` or `recover-billing`, once the engineer confirms the
name. Its working tree is clean, and `proviso lint` exits with 0.

Run `proviso version` in the repository's root. Pass every warning it
writes on to the engineer: a skill or agent definition that another
version of Proviso installed works from other instructions than the
command expects, and `proviso update` brings them in line. Whether to
update first is the engineer's call; go on either way.

## 2. The system and its capabilities

When the tree holds no capability yet, the first run recovers the frame
only:

- read the docs, the entry points and the layout of the code, and what
  the system's interfaces offer;
- write `system.md`: its Goal, Rationale, Scope, Out of scope, Audience
  and Design boundary, in place of the skeleton's `?`;
- propose to the engineer how the system falls into capabilities, by
  what it does, not by how its code is packaged; for each capability
  the engineer agrees to, write its `capability.md`, Goal and Rationale
  only;
- add the system's questions to its `questions.md`.

Then go on with step 4.

## 3. One capability

When the engineer names a capability to recover:

- **Read the tree:** `system.md`, the capability's `capability.md`, and
  the glossaries, parameters, invariants, constraints and contracts of
  the system and of any capability recovered before. Reuse their terms;
  a word they define differently from how this capability's code uses it
  is a question, never a second definition.
- **Read the system:** the code that does what the capability does, its
  tests, and every doc about it. Note each disagreement and each default
  as you go.
- **Agree the features** with the engineer: their names, Goals and
  contacts.
- **Write the capability** as the writing guide says, with the departures
  above: features and their rules, Examples from the tests and the code,
  glossary, parameters, invariants, constraints, and contracts for every
  interface a rule refers to, at the seam the system's tests can call.
- **Write its questions.**

Before you commit:

- `proviso lint` exits with 0, and no element holds a `?`;
- every Examples row is checked against its rule and against the code;
- every candidate `proviso impact` lists is read, and either edited or
  left for the check.

## 4. Check it

Commit the tree as the writing guide says, with what was recovered and
from where as the message's body. Hand over to `proviso-check`, and go
on when it has finished with `proviso verify` exiting with 0. Accepting
is the engineer's call: say that the amendment is ready for
`proviso-accept`.

## 5. Check a recovered capability against the code

Only when the engineer asks for it: once a capability is accepted, and
its tree is solid enough that tests written from it say something.

- **The notes:** read `.proviso/build.md`, the commands that build the
  system and run its acceptance tests, and `.proviso/tester.md`. If they
  do not say how acceptance tests are built and run, stop and say so:
  setting up the system's acceptance tests is the engineer's work.
- **The branch:** one from the target, named such as
  `recover-billing-tests`, once the engineer confirms the name. With
  `<revision>` the commit on the target before the capability's
  recovery was accepted, the first parent of its merge, write
  `proviso pack --since <revision>` to
  `<temp>/proviso/<repository>/<branch>/pack-1.md`, in the system's
  temporary folder.
- **The testing agent:** add a worktree of a branch
  `proviso/<branch>/tests` at `<temp>/proviso/<repository>/<branch>/tests`,
  and start `proviso-tester` in it with the worktree, the pack's path,
  and the round: round 1, the code exists already, and the tests must
  pass against it. Never start a coding agent, and never give the
  testing agent anything of the code.
- **Run the tests** as `.proviso/build.md` says, against the unchanged
  code.

A failing test means the tree and the code disagree, and the fix goes
into the tree, never into the code:

- **The tree misstates what the code does:** correct the rule from the
  code.
- **The testing agent read the tree otherwise than it was meant:** the
  rule is not precise enough; sharpen it.
- **The behaviour cannot be tested reliably,** such as one that depends
  on the time or on chance: say so in the tree, or ask about it.

Commit each edit on the branch, write `proviso pack --since <revision>`
again, as the next pack, and resume the testing agent with it and what
failed. Never decide what the system should do: a failure that shows
the code doing something nobody would want is a question, and the rule
stays as built.

When every test passes, merge the tests branch into the branch, remove
the worktree and the tests branch, and hand over to `proviso-check` if
the tree changed. If it did not, the tests are a change of the system's
code like any other, for the engineer to merge as such. Report the
rounds, the test counts, every edit of the tree and every question it
raised; accepting is the engineer's call.

Never change the system's code, never accept, never push.
