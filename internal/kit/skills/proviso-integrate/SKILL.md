---
name: proviso-integrate
description: Bring a request into a Proviso spec tree in one session with the engineer, on the amendment's branch, as checked edits of the tree. Use when the engineer brings a request, a ticket, a user story, acceptance criteria or a change to specify, in any form. Asks the tough questions, decides nothing the business has not decided, hands over to proviso-check, and offers a business view of what was integrated. Does not build and does not accept.
---

# Integrating a request

You bring one request into the spec tree, with the engineer: as edits
of the tree on the amendment's branch, in the tree's own elements. The
business brings the substance: what it wants and why, the rules as far
as it can state them, and what it means by its words. That must make
sense, and the business is held to that, not to the notation. The
rigour the notation needs is your work and the engineer's.

You are stricter than the skill that helps business requesters write a
request. It rations its questions so as not to tire them; you do not
let a hole go. Every hole you find ends either decided in the tree or
marked `?` with a question naming whom to ask. Nothing is mentioned once
and dropped.

## Proviso's command

Wherever this skill says `proviso`, run Proviso's command: the one
`.proviso/build.md` names, or `proviso` from the path if it names none.
Look nowhere else.

## Who answers what

- **Engineering questions,** such as mechanisms, values, names,
  contracts and constraints, are answered by the engineer in this
  session.
- **Business questions,** what the system should do and what a word
  means, are answered by neither you nor the engineer. They go back to
  the request's owner, each as one concrete case question, such as
  "Would $80 from one $50 and three $20 notes count as *composable*?",
  and the place stays `?` until the owner answers.
- **A serious conflict,** an edit that contradicts an existing rule,
  invariant or constraint, or uses a glossary term with another
  meaning, changes agreed behaviour. It always goes back to the
  business; say so to the engineer.

Ask the engineer one question at a time, with your recommendation and
why. Write a decision into the tree only once the engineer has made it.

<!-- proviso: writing.md -->

## 1. The branch

The amendment's branch is the engineer's, or one you make from the
target, usually the default branch, named after the request in
kebab-case, once the engineer confirms the name. Its working tree is
clean, and `proviso lint` exits with 0.

## 2. Read the request

The request may come in any form: a request written with the authoring
skill, which is the best prepared, a sentence, a ticket, prose or
existing scenarios. Start from what is there; do not ask for it again.
Find in it:

- what is wanted, and why; who decides what its rules mean, the
  request's owner;
- how things work today;
- what is in, and what is explicitly out;
- the rules: events, conditions, both outcomes, examples;
- the words that need pinning down, and the open questions.

What it does not say is a question: for the owner, or for the engineer
where it is technical.

## 3. Read the tree

Read `system.md`, every capability the request concerns with its
features, and the glossaries, invariants, constraints, parameters and
contracts of those capabilities and of the system. Then:

- **Today against the tree:** every disagreement between what the
  request says happens today and what the tree says is a finding;
  bring each to the engineer.
- **The request's words against the glossaries:** a word the tree
  defines differently is a serious conflict; a word the tree already
  defines is reused.

## 4. Does it make sense

Before you write anything, check what the business brought, on its own
and against the tree:

- the Goal holds together, and fits its capability's Goal;
- the reason explains the rules;
- the rules serve the Goal, and contradict neither each other nor the
  tree;
- the examples agree with the rules.

What does not make sense goes back to the owner as a concrete case
question. Never fix it quietly.

## 5. Settle it with the engineer

- **Each `Out:` line,** or anything the request leaves out on purpose:
  only this amendment leaves it out, or the system never does it. The
  first goes into the amendment note, `proviso/.amendment.md`, a list
  with one item each and no headings; the second becomes a constraint
  or the system's Out of scope, once confirmed.
- **Each possible invariant:** a statement that may hold beyond one
  event. Ask about it with a case on another event; the business says
  what it means, the engineer what it ranges over.
- **Each hole** the writing guide lists, and every engineering question
  the request raises.

## 6. Edit the tree

Write the request into the tree as the writing guide says:

- a reason that will still matter later goes into a Rationale; why the
  amendment is made now goes into the commit message;
- what happens today goes nowhere: the tree is today;
- a rule number or a decisions log of the request goes nowhere;
- the request itself is never stored in the tree.

Before you commit:

- `proviso lint` exits with 0;
- every Examples row of every rule you wrote or changed is checked
  against its rule;
- every possible invariant is asked about;
- every candidate `proviso impact` lists is read, and either edited or
  left for the check;
- every hole is decided in the tree, or marked `?` with its question in
  `questions.md`.

Commit the tree, and the amendment note if there is one, as the writing
guide says.

## 7. Check it

Hand over to `proviso-check`, and go on when it has finished with
`proviso verify` exiting with 0, or reporting no problem but
`undecided`: the places still waiting for an answer. The findings it brings are the
engineer's to decide, as it says.

## 8. The business view

Offer the engineer a business view: what was integrated, in the
owner's words, for the owner to check. If the engineer wants one, with
`<amendment>` the branch and `<n>` the number of this view, starting
with 1, write into `<temp>/proviso/<repository>/<amendment>/view-<n>/`,
in the system's temporary folder:

- `pack.md`: `proviso pack`;
- `impact.md`: `proviso impact`;
- `request.md`: the request, as it came.

Start a fresh agent and give it these three paths and nothing more,
with these instructions:

> Write a business view of an amendment for the request's owner, who
> never reads the spec tree. Read `pack.md`, `impact.md` and
> `request.md`, and nothing else. In plain sentences, in the language of
> the request, and with no notation: say what the system will do after
> the amendment, rule by rule, mapped to what the request asked;
> summarise technical elements rather than hide them; show every change
> the request did not ask for, such as an existing rule changed because
> of it; and list every place marked `?` as one concrete case question,
> with what it holds up. Say only what the pack says; where it is silent,
> say so. Return the view as Markdown.

Write what it returns to `business-view.md` in the same folder, and give
the engineer its path. The engineer decides whether the owner gets it;
never send it yourself. Render it again after any later edit of the
tree.

## 9. Report

The branch and its commits, what the amendment touches, the amendment
note, every open question and whom it is for, how the check ended, and
the business view's path. Building is the engineer's call: say that the
amendment is ready for `proviso-build` once the engineer wants it.

Never build, never accept, never push.
