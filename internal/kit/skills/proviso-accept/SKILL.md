---
name: proviso-accept
description: Accept a Proviso amendment into its target branch, directly or through a pull request, once its coherence check covers it, rewriting the manifest in the same step. Use when the engineer asks to accept, merge or finish an amendment. Hands over to proviso-check while proviso verify fails; does not build and never pushes.
---

# Accepting an amendment

You accept one amendment: its tree change, and whatever was built for it,
become the baseline of the target, the branch it is accepted into. The
manifest is rewritten in the same step, so that no tree reaches the
target without its check. You decide nothing about the system; the
check's judgements and the engineer's decisions are already recorded.

## Proviso's command

Wherever this skill says `proviso`, run Proviso's command: the one
`.proviso/build.md` names, or `proviso` from the path if it names none.
Look nowhere else.

## 1. The target

The target is the branch the engineer names. If none is named, propose
the default branch and wait for the engineer to confirm it, or to name
another.

## 2. Check the start

- The current branch is the amendment's branch, not the target, and its
  working tree is clean.
- `proviso lint` exits with 0.
- Read `.proviso/build.md`: whether acceptance goes through a pull
  request (a line such as `Acceptance: pull request`; without one, it
  merges directly), the commands of the project's checks, and a
  convention for commit messages, if it names one; then the commit
  below follows that convention instead of `feat: accept <amendment>`.

Stop and say why if any of this fails.

## 3. Bring the target in

If the target has commits the amendment's branch does not hold, merge
the target into the amendment's branch. A conflict is the engineer's to
resolve. A conflict in `manifest.sum` is never resolved by running
`proviso manifest`: take the branch's side, since the manifest is
rewritten below, and let the check show what the two amendments did to
each other.

## 4. The check covers the amendment

Run `proviso verify`. If it reports `undecided`, stop: name each such
element; the amendment waits until each is decided in the tree, and is
then checked again. If it reports any other problem, hand over to
`proviso-check`, and go on when it has finished with `verify` exiting
with 0. If it exits with 2, stop and say why.

## 5. Accept

**Directly**, without a pull request:

1. switch to the target and run `git merge --no-ff --no-commit` of the
   amendment's branch;
2. `proviso verify` exits with 0: the check still covers the merged
   tree;
3. `proviso manifest`, which also removes `proviso/.coherence.json` and
   the amendment note `proviso/.amendment.md`;
4. `proviso verify` exits with 0 again: the tree is what the manifest
   records, and no check is left over;
5. stage `proviso/manifest.sum` and the removal of
   `proviso/.coherence.json` and of the amendment note, if there was
   one, and run every command of the project's
   checks. If one fails, `git merge --abort`, switch back to the
   amendment's branch, and report the failure;
6. commit as `feat: accept <amendment>`, with a body that says what the
   amendment does, how many elements it touched, how many candidates
   were judged, and every finding and how it was closed;
7. remove the amendment's branch with `git branch -d`.

**Through a pull request:**

1. on the amendment's branch, run `proviso manifest`, which also
   removes `proviso/.coherence.json` and the amendment note
   `proviso/.amendment.md`;
2. `proviso verify` exits with 0;
3. stage the manifest and the removals, run every command of the project's checks, and commit as
   `feat: accept <amendment>`, with the body as above, as the branch's
   last commit;
4. tell the engineer to push the branch and merge the pull request into
   the target before anything else reaches the target; if something
   does, accept again from step 3.

## 6. Report

The target, the commit, how many elements and candidates, the findings
and how they were closed, and the checks that ran.

Never push, never open a pull request, never edit the spec tree.
