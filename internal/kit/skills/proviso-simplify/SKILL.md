---
name: proviso-simplify
description: Find where a Proviso spec tree could say the same more simply, or where its behaviour could be simpler, through the simplifying agent, which reads only the tree, and bring each proposal to the engineer. Use when the engineer asks to simplify, tidy or declutter the tree, or to find repetition, rules that restate one another, or promises held only in a Rationale. Runs outside any amendment and edits nothing; hands every proposal the engineer takes to proviso-integrate. Never pushes.
---

# Simplifying the spec tree

A spec tree grows by amendments, and every finding of a check is
settled by adding to it: each addition is coherent, and together they
make the tree say things in several places, decide cases one by one
that one element could decide once, or keep promises no rule holds.
You have the simplifying agent read the tree afresh, bring its
proposals to the engineer, and hand those the engineer takes to
`proviso-integrate`. You decide nothing and edit nothing yourself.

## Proviso's command

Wherever this skill says `proviso`, run Proviso's command: the one
`.proviso/build.md` names, or `proviso` from the path if it names none.
Look nowhere else.

## 1. Check the start

- `proviso lint` exits with 0.
- `proviso diff` exits with 0: no amendment is in progress, so the
  proposals concern the accepted tree. If one is, say so: the engineer
  may want to run this after it is accepted.

## 2. What the agent reads

The whole spec tree, unless the engineer names a capability or the
system level: then that capability's folder, or the files directly in
the spec root and in its `contracts/` folder, and the system level in
either case. Name each folder and file by its path; never name the spec
root itself, which holds the coherence file.

## 3. Start the agent

Start a fresh `proviso-simplifier` with the paths it may read, and
nothing more: no earlier proposals, no view of your own, nothing of how
the tree came to be. Wait until it has reported.

## 4. Bring the proposals to the engineer

One at a time, the most valuable first: what it proposes, the elements,
what it preserves, its kind, its cost, and your recommendation with the
reason. Before you recommend, read the elements it names with
`proviso show` and check that the tree says what the proposal claims;
a proposal resting on a misreading is dropped, and you say why.

- **Same meaning, taken:** hand it to `proviso-integrate` as a request
  of the engineer's, with the proposal as it stands; it becomes an
  amendment like any other, checked and built.
- **Changed meaning, taken:** hand it to `proviso-integrate` too, with
  its case question; who answers the question is for that skill to say.
- **Left:** say nothing more of it.

Keep the proposals not yet decided as a list in the session, so that the
engineer can come back to them. Several taken proposals are separate
amendments, unless the engineer joins them.

## 5. Report

The proposals, and for each whether it was taken, handed to
`proviso-integrate`, left or dropped, and why.

Never edit the spec tree, never push.
