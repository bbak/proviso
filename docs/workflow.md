# Workflow

One amendment, from request to acceptance, as the engineer sees it.
Proviso does the mechanics, the skills run the agents, and the engineer
decides. Every decision lands in the spec tree, where the next amendment
finds it.

## Who decides what

| Who | Decides | Never |
| --- | --- | --- |
| Business requester | what the system should do, and what a rule means | sees the tree or the command line |
| Engineer | how the request enters the tree, every gap the agents find, every finding of the check | leaves a decision outside the tree |
| Implementing agents | nothing about the system; they report what the tree did not decide | see each other's work |
| Checking agent | whether each candidate stays as it is, with a reason | closes a finding, sees code, tests or reports |
| Proviso | nothing; it computes | drives the agents |

## 1. Bring the request into the tree

Make the amendment's branch, and edit the tree on it: new and changed
rules, glossary entries, invariants, constraints, contracts and
parameters. Proviso tells you what the edit does:

- `proviso lint`: errors in the notation.
- `proviso diff`: the element diff, and which elements are not to be
  built because they hold an undecided place.
- `proviso impact`: the candidates, existing rules and invariants the
  edit may affect.

Read the candidates while you edit: one that the request does affect is
better edited now than found by the check. Leave a place `?` when the
business has not decided it, and take the question back to them;
nothing with an undecided place is built.

If the amendment deliberately leaves part of the request out, list it in
`proviso/.amendment.md`, one item each, no headings. The pack carries
the list to the agents, who then neither build it nor report it as a
gap. Only what this amendment leaves out goes there; a lasting "the
system never …" is a constraint in the tree.

Commit the tree, and the note if there is one, before anything is
built. The tree comes first on the branch, then tests and code.

A skill that guides business requesters to a checkable request is in
the works. Until then, requests arrive in whatever form the business
writes them.

## 2. Build: `proviso-build`

Ask Claude Code to build the amendment. The skill:

1. checks that the branch is clean, the tree lints and the element diff
   lists something to build;
2. writes the context pack;
3. makes a branch and a worktree each for the code, the tests and their
   integration, in the system's temporary folder, never inside the
   repository;
4. starts the testing agent and the coding agent in parallel, each with
   its worktree and the pack;
5. merges their work and runs the commands of `.proviso/build.md`;
6. sorts every failure, and every decision the agents report, into
   decided by the tree, a harmless default, or a real gap.

**What you decide: the gaps.** Only real gaps reach you, one at a time,
each with a recommendation. A gap both agents report, or that they
decided differently, is the strongest sign of a real one. Your decision
becomes an edit of the tree, committed on the branch, and the agents
build again from a pack of what changed since their last round.

When the integration is green and no gap is open, the amendment's
branch moves to the integration, and the skill removes its own branches
and worktrees.

## 3. Check: `proviso-check`

Ask Claude Code to check the amendment. The check can run at any time
after the tree change, also before the build; `proviso-accept` starts it
itself when it is missing or out of date. The skill:

1. writes the pack, the candidates and what is still to check into the
   system's temporary folder;
2. starts a fresh checking agent, which reads them and only the parts of
   the tree the amendment concerns, never code, tests or the agents'
   reports;
3. records the agent's judgements and findings with `proviso judge` in
   `proviso/.coherence.json`.

**What you decide, in this order:**

- **Points the agent could not judge.** Let each go, or make it a
  finding.
- **Each open finding.** Resolve it by an edit of the tree, or dismiss
  it with a reason in your own words. An edit is checked again by a new
  agent. An edit that changes behaviour, a rule, invariant, constraint,
  contract element or parameter, has to be built: once every finding of
  the run is decided, the skill hands over to `proviso-build`.

The skill ends when `proviso verify` exits with 0, and commits the
coherence file on the branch.

## 4. Accept: `proviso-accept`

Ask Claude Code to accept the amendment, and name the target branch;
without one, the skill proposes the default branch. It checks the
branch, brings the target in, and hands over to `proviso-check` while
the check does not cover the amendment. Then, as `.proviso/build.md`
says:

- **Directly:** it merges into the target, rewrites the manifest, which
  removes the coherence file and the amendment note, runs your project's checks, and commits
  `feat: accept <amendment>`. If a check fails, it aborts the merge.
- **Through a pull request:** it rewrites the manifest on the
  amendment's branch, runs your checks, and commits as the branch's last
  commit. You push and merge the pull request.

It never pushes and never opens a pull request.

**A conflict in `proviso/manifest.sum`** means two amendments met. Never
resolve it by running `proviso manifest`: the merge is an amendment of
its own, and the check shows what the two did to each other.

## Edits outside the skills

The tree can be edited by hand, and an amendment need not be built by
agents. Acceptance still requires the check. A hand edit committed
straight to the target is caught: `proviso verify` fails there until
the edit has gone through the check and acceptance. See
[Repository setup](repository-setup.md) for running it in CI.
