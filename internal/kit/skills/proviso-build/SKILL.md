---
name: proviso-build
description: Build an amendment of a Proviso spec tree with a testing agent and a coding agent, kept apart, until its integration is green and no gap is open. Use when the engineer asks to build, implement or run the forward round of an amendment whose tree change is committed on its branch. Does not accept the amendment; that is proviso-accept.
---

# Building an amendment

You orchestrate one amendment: the spec tree is changed and committed on
the amendment's branch, and two implementing agents build it from a
context pack, the testing agent `proviso-tester` and the coding agent
`proviso-coder`. You give each its work, integrate what they built, and
bring the engineer only what the tree does not decide. You never write
code or tests yourself, and you never accept or push.

## Keep the agents apart

The tests are independent of the code only as long as neither agent
sees the other's work. Never pass one agent's code, tests, decisions or
report to the other. What one agent's work teaches flows back through
the tree, as a change the engineer decides, and reaches both as a pack.

## Proviso's command

Wherever this skill says `proviso`, run Proviso's command: the one
`.proviso/build.md` names, or `proviso` from the path if it names none.
Look nowhere else.

## 1. Check the start

- The current branch is the amendment's branch, not the default branch,
  and its working tree is clean: the tree change is committed.
- `proviso lint` exits with 0.
- `proviso diff` exits with 1: the amendment lists something to build.
  Tell the engineer which elements it lists as not to be built; the
  agents leave them alone.

Stop and say why if any of this fails.

## 2. Read the engineers' notes

Read `.proviso/build.md`: the commands that build the system, run its
unit tests and run its acceptance tests, and what integration must
watch, such as skips that are expected. If it is empty or missing, ask
the engineer for them once, and offer to write the answer into it.

## 3. Branches and worktrees

With `<amendment>` the name of the amendment's branch and `<base>` the
commit it stands at:

- make the branches `proviso/<amendment>/code`, `proviso/<amendment>/tests`
  and `proviso/<amendment>/integration` from `<base>`;
- add a worktree for each with `git worktree add`, in the system's
  temporary folder, as `<temp>/proviso/<repository>/<amendment>/code`,
  `…/tests` and `…/integration`; never inside the repository, where tools
  that walk it would find them.

Write `proviso pack` to `<temp>/proviso/<repository>/<amendment>/pack-1.md`.
Remember `<base>` as the revision of round 1.

## 4. Start both agents

Start `proviso-tester` in the tests worktree and `proviso-coder` in the
code worktree, in parallel, in the background. Give each, and nothing
more:

- its worktree, as an absolute path;
- the pack's path, the only file it may read outside its worktree;
- the round: round 1 of this amendment.

Their definitions say the rest. Wait until both have reported.

## 5. Integrate

In the integration worktree, merge the amendment's branch, then the
code branch, then the tests branch. Run every command `.proviso/build.md`
names, the acceptance tests last. Look at every failure and every skip:
a skip that `.proviso/build.md` does not expect is a finding, like a
failure.

## 6. Sort what you found

Take every failure, and every decision either agent listed as one the
pack did not make, and sort each into one of three:

- **Decided by the tree:** a closer reading of the pack settles it. If an
  agent decided against the pack, it gets it back in the next round,
  pointed to the element that decides it.
- **A harmless default:** a choice the pack rightly leaves open, such as
  a message's wording or a file's permissions. Note it; nobody needs to
  act.
- **A real gap:** the tree does not decide it, and the choice matters.
  Only these go to the engineer.

Compare the two lists: the same gap seen by both agents, or decided
differently by each, is the strongest sign of a real one. Bring the gaps
to the engineer one at a time, each with your recommendation and why;
record a decision only when the engineer agrees.

## 7. The next round

- **A gap the engineer decided** flows back into the tree on the
  amendment's branch: edit, run `proviso lint`, commit with `docs:`. Then
  write `proviso pack --since <revision>`, with `<revision>` the commit
  of the agents' last round, to `pack-<n>.md`, and remember the new
  commit as this round's revision.
- **Resume each agent that has work,** with: what integration found that
  concerns it, without anything of the other agent's work; that it
  merges the amendment's branch into its own; and the new pack, if there
  is one. An agent with nothing to change is not resumed.
- Integrate again (step 5), and sort again (step 6).

An agent lives for one amendment. If it cannot be resumed, such as in a
new session, start a fresh one with the full pack, `proviso pack`, and
say so in its round.

## 8. Finish

When the integration is green and no gap is open:

- move the amendment's branch to the integration branch, as a fast
  forward;
- remove the three worktrees and the three `proviso/<amendment>/…`
  branches;
- report to the engineer: the rounds, the test counts of the last
  integration, every gap and how it was decided, the defaults noted,
  and that the amendment is ready for `proviso-accept`.

Never accept, never merge into the default branch, never push.
