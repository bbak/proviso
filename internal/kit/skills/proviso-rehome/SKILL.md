---
name: proviso-rehome
description: Give the elements of a Proviso spec tree that moved to another capability or to the system the identifiers of their place, with every reference in the tree, the manifest, and the code and tests beside it, in one step outside any amendment. Use when the engineer asks to rehome, realign or tidy identifiers, or after an amendment moved elements. Shows the plan and waits for the engineer's approval before anything is written; does not integrate, check, build or accept, and never pushes.
---

# Rehoming identifiers

You give the elements that moved the identifiers of their place, in one
step of its own: Proviso renames them in the spec tree and writes the
manifest, and you carry the same renames into the code, tests and
documents beside the tree, which Proviso never reads. Nothing changes
what anything means, and nothing is written before the engineer has seen
every rename.

## Proviso's command

Wherever this skill says `proviso`, run Proviso's command: the one
`.proviso/build.md` names, or `proviso` from the path if it names none.
Look nowhere else.

## 1. Check the start

- The working tree is clean.
- `proviso lint` exits with 0.
- `proviso verify` exits with 0 and `proviso diff` with 0: the tree is
  what the manifest records, with no amendment in progress. A rename
  never mixes with an amendment; if one is in progress, say so and stop:
  rehome after it is accepted.
- Read `.proviso/build.md`: the commands of the project's checks, how
  the acceptance tests are named after identifiers, if it says, and a
  convention for commit messages, if it names one.

Stop and say why if any of this fails.

## 2. The plan

Run `proviso rehome --json`, which writes nothing.

- Exit 0 with no renames: every identifier names its place; say so and
  stop.
- Exit 1: some renames would take an identifier. Show each collision:
  the identifier and the elements planned to take it. Renaming one of
  them by hand is an edit of the tree, an amendment of its own; say so
  and stop.
- Exit 2: say why, from what it wrote to standard error, and stop.

## 3. What the renames reach beside the tree

Proviso renames only what the notation marks as an identifier: anchors,
hand-overs, link fragments and question tags. Search the repository's
tracked files for every old identifier of the plan, as a whole token,
so that `a.b` never matches inside `a.b.c`:

- **In code, tests, notes and documents** outside the spec root: the
  identifier as it stands, and every form derived from it where
  `.proviso/build.md` or the engineers' notes say how one is derived,
  such as a test named after its element. For each new identifier,
  check that its derived form is not already taken.
- **In the spec tree itself,** wherever the notation does not mark it:
  plain text, Examples cells, code spans, fenced blocks. Such a mention
  may be a reference or data, such as an Examples row about
  identifiers. `rehome` leaves it as it is, and so do you: changing it
  is an edit of the tree, an amendment of its own.

Tell a reference from data by reading it, never by its form alone: a
short identifier can also be a module path or a word in another
language.

## 4. Approval

Show the engineer, and wait for the answer:

- every rename, old and new;
- every place beside the tree you will rename, file by file, and every
  match you will leave, with why;
- every derived name already taken, which stops the step until the
  engineer decides;
- the mentions in the tree that stay as they are, for a later
  amendment if the engineer wants them changed.

Write nothing until the engineer agrees.

## 5. Write

1. Run `proviso rehome --write --json`. It must exit with 0 and list
   the same renames as the plan; if not, stop and say why.
2. Rename every approved place beside the tree. Make each edit fail if
   it matches nothing, and read each one back: a rename that never
   landed is the error this step most easily hides.
3. `proviso lint` and `proviso verify` exit with 0, and `proviso diff`
   with 0: the tree is what the new manifest records.
4. Run every command of the project's checks. If one fails, and the
   failure is a name you renamed or missed, fix that name; anything
   else, report and stop, leaving the changes uncommitted.
5. Commit everything as `refactor: rehome identifiers`, or as the
   convention `.proviso/build.md` names, with every rename in the body.

## 6. Report

The renames, the files changed beside the tree, the mentions left in
the tree, the checks that ran, and the commit.

Never push, never edit the spec tree but through `proviso rehome`.
