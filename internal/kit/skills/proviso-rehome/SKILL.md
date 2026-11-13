---
name: proviso-rehome
description: Give the elements of a Proviso spec tree that moved to another capability or to the system, or whose identifiers name no place, the identifiers of their place, with every reference in the tree, the manifest, and the code and tests beside it, in one step outside any amendment. Use when the engineer asks to rehome, realign or tidy identifiers, after an amendment moved elements or gave a capability another identifier, or when lint asks about an identifier-unhomed. Shows the plan and waits for the engineer's approval before anything is written; does not integrate, check, build or accept, and never pushes.
---

# Rehoming identifiers

You give the elements that moved, or whose identifiers name no place,
the identifiers of their place, in one step of its own: Proviso renames
them in the spec tree and writes the manifest, and you carry the same
renames into the code, tests and documents beside the tree, which
Proviso never reads. Nothing changes
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
  the identifier and the elements planned to take it, and ask the
  engineer for another new identifier for one of them; renaming one by
  hand is an edit of the tree, an amendment of its own.
- Exit 2: say why, from what it wrote to standard error, and stop.

An identifier whose first segment names no place, such as a contract's
name alone, gets its place put in front: `okf-get` in the capability
`access` becomes `access.okf-get`. Where a derived identifier is taken
or awkward, such as `coherence.checks.run` for a feature whose
capability was once identified as `checks`, the engineer may give its
new identifier: run the plan again with
`--rename <old>=<new>` for each, such as
`--rename checks.run=coherence.run`. The new identifier starts with the
element's place. Keep every `--rename` the engineer gives, and pass the
same ones to every later run of `proviso rehome`.

## 3. What the renames reach beside the tree

Proviso renames only what the notation marks as an identifier: anchors,
hand-overs, link fragments and question tags. Search the repository's
tracked files for every old identifier of the plan, as a whole token,
so that `a.b` never matches inside `a.b.c`:

- **In code, tests, notes and documents** outside the spec root: the
  identifier as it stands, and every form derived from it where
  `.proviso/build.md` or the engineers' notes say how one is derived,
  such as a test named after its element. For each new identifier,
  check that its derived form is not already taken; if it is, ask the
  engineer for another new identifier, given with `--rename`, and run
  the plan again.
- **In the spec tree itself:** the plan's `mentions` list every line
  where an old identifier stands as a whole code span outside rule
  blocks and fenced blocks, such as in an Examples cell or a Rationale.
  Read each in its element: a reference goes stale with the rename, such
  as a rule named in a Rationale, or an Examples row whose outcome
  depends on the tree holding the element; data does not, such as an
  Examples row about the form of identifiers. Propose to rename every
  reference and to keep every piece of data. Search the tree too for
  the old identifiers where the plan cannot see them, in plain text and
  in fenced blocks: those stay as they are, since changing them is an
  edit of the tree, an amendment of its own.

Tell a reference from data by reading it, never by its form alone: a
short identifier can also be a module path or a word in another
language.

## 4. Approval

Show the engineer, and wait for the answer:

- every rename, old and new;
- every place beside the tree you will rename, file by file, and every
  match you will leave, with why;
- every derived name already taken, which stops the step until the
  engineer gives another new identifier with `--rename`, or decides
  otherwise;
- every new identifier the engineer gave with `--rename`;
- the plan's mentions, grouped by element, each with its file and
  line, and whether you propose to rename it, a reference, or to keep
  it, data;
- the mentions in plain text and fenced blocks that stay as they are,
  each with its file and line, and whether it is a reference or data.

Write nothing until the engineer agrees.

## 5. Write

1. Write the mentions the engineer approved for renaming, as a
   `mention_list` of the plan's items, to a file in the system's
   temporary folder, and run `proviso rehome --write --mentions --json`,
   with every `--rename` the engineer gave, and with that file on
   standard input; with none approved, run it without `--mentions`. It
   must exit with 0 and list the same renames and mentions as the
   plan; if not, stop and say why.
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

The renames, the files changed beside the tree, the mentions renamed
in the tree and those left, the checks that ran, and the commit. If any
mention left is a reference, recommend a small amendment that names it
by its new identifier, before the next amendment builds on it.

Never push, never edit the spec tree but through `proviso rehome`.
