---
name: proviso-explain
description: Answer an engineer's question about the system specified in a Proviso spec tree, or about what a term, identifier or name in backticks in the tree stands for, only from what the tree says, looking every element up with proviso show. Use when the engineer asks what a word of the tree means, what an identifier or reference stands for, or what the system does in a case, such as "What does 'spec tree' mean?" or "What happens when …?". Edits nothing and decides nothing; hands what the tree does not decide to proviso-integrate.
---

# Explaining the spec tree

You answer an engineer's question about the system from its spec tree,
and from nothing else. A word of the tree often means something
narrower than the everyday word, and the tree's meaning is the one the
system is built from; an answer from general knowledge, from the code or
from a library is a guess about what the tree says, and you never make
it.

## Proviso's command

Wherever this skill says `proviso`, run Proviso's command: the one
`.proviso/build.md` names, or `proviso` from the path if it names none.
Look nowhere else.

## 1. Find what the question is about

- **An identifier, a term or a name in backticks:** run
  `proviso show --json "<query>"`, the query as the tree writes it.
  Where the question comes from a file of a capability, add
  `--in <capability>`, the name of the capability's folder, so that a
  name stands for what it stands for there. If that finds nothing, run
  it once without `--in`: the name may belong to another capability,
  and you say so.
- **A case,** such as "what happens when …": search the rule blocks of
  the feature files for the rules whose `When` is that event, then show
  each with `proviso show --json <identifier>`.
- **Exit status 1:** the tree has no element of that identifier or
  name. Say so, and never answer with the nearest name instead; offer a
  search of the tree's text if the engineer wants one.
- **Exit status 2:** pass on what `show` wrote to standard error, and
  stop.

`show` reads the tree as it stands on the checked-out branch, also in
the middle of an amendment; if that branch is an amendment's, say that
the answer is the amendment's tree, not the accepted one.

## 2. Read what the answer rests on

- Look up every term mark, name in backticks and hand-over in the
  element's text that the answer depends on, with `--in` the capability
  whose folder holds the element's file, as its `file` says. Stop when
  the answer stands; never walk the whole tree.
- Where the result holds several elements, each is a meaning of its
  own, in its own capability: give every one, with where it stands, and
  never merge them into one.

## 3. Answer

- Say what the tree says, in plain words, quoting it where the wording
  decides.
- Name each element you rest on by its identifier and its `file:line`,
  from `show`'s result, so that the engineer can read it there.
- A place marked `?`, or a term whose status is open, is not decided:
  say so, with its question from `questions.md` where there is one.
- Where the tree does not decide the case, say so plainly and name what
  is missing. Never fill the gap; offer to bring it into the tree with
  `proviso-integrate`.

Never edit a file, and run no command of Proviso but `show`; reading
the tree's files and searching them is all else you do.
