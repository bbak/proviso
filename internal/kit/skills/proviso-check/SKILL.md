---
name: proviso-check
description: Run the coherence check of a Proviso amendment through an independent checking agent, record its judgements and findings with proviso judge, and bring every finding to the engineer until proviso verify passes. Use when the engineer asks to check an amendment, after the tree of an amendment was edited, or when proviso-accept hands over because verify fails. Does not build and does not accept.
---

# Checking an amendment

You run the coherence check of one amendment: the spec tree is edited on
the amendment's branch, and the checking agent `proviso-checker` judges
whether it still makes sense with that edit. You prepare what it reads,
record what it returns, and bring the engineer every finding. You never
judge a candidate yourself, and you never close a finding on your own.

## Proviso's command

Wherever this skill says `proviso`, run Proviso's command: the one
`.proviso/build.md` names, or `proviso` from the path if it names none.
Look nowhere else.

## Keep the check independent

The checker judges the tree, not the work around it. Never pass it the
implementing agents' reports or decisions, code, tests, how a gap was
settled, or your own view of a candidate. It is given files, and only
the files below.

## 1. Check the start

- The current branch is the amendment's branch, not the default branch,
  nor the branch the engineer names as the target.
- `proviso lint` exits with 0, and `proviso diff` exits with 1: there is
  an amendment to check.
- `proviso verify --json`: if it exits with 0, the check covers the
  amendment as it stands; say so and stop. If it exits with 2, stop and
  say why.

## 2. Prepare the run

With `<amendment>` the name of the branch and `<n>` the number of this
run, starting with 1, write into
`<temp>/proviso/<repository>/<amendment>/check-<n>/`, in the system's
temporary folder:

- `pack.md`: `proviso pack`;
- `impact.json`: `proviso impact --json`;
- `verify.json`: the output of `proviso verify --json`.

The affected capabilities are those holding an element the element diff
lists, or a candidate `verify` reports. The checker may read the folder
of each, the one holding its `capability.md`, and the system level: the
`.md` files directly in the spec root and the files in its `contracts/`
folder. Name each folder and file by its path; never name the spec root
itself, which holds the coherence file.

If the candidates to judge span more than one capability, because a
system-level frame changed, prepare one run per capability, each with
only that capability's candidates named, and the system level.

## 3. Start the checker

Start a fresh `proviso-checker` for each run, in parallel if there are
several. Give each, and nothing more:

- the paths of `pack.md`, `impact.json` and `verify.json`;
- the folders and files of the tree it may read;
- for a split run, the identifiers of the candidates it judges.

Wait until each has reported.

## 4. Record

Write each checker's JSON block to `check-<n>/result.json` (one file per
run when split) and give it to `proviso judge` on standard input.

- Exit 0: recorded.
- Exit 1: refused. Resume that checker with what `judge` wrote to
  standard error, within the same run, and record again.
- Exit 2: stop and say why.

## 5. Bring the findings to the engineer

Run `proviso verify --json`. For every `open` finding, one at a time:
what it says, the elements it names, the checker's recommendation, and
your own recommendation with the reason. The engineer either:

- **resolves it by an edit:** make the edit on the amendment's branch,
  run `proviso lint`, commit with `docs:`, and record
  `{"findings": [{"id": "<id>", "status": "resolved", "reason": "edited in <commit>"}]}`
  with `proviso judge`;
- **dismisses it:** record `"status": "dismissed"` with the engineer's
  reason, in the engineer's words.

Record a decision only when the engineer has made it.

Then bring every point the checker could not judge, one at a time:
what it says, and your recommendation. The engineer either lets it go,
or makes it a finding: record it with `proviso judge` as
`{"findings": [{"kind": "<kind>", "elements": ["<identifier>"], "text": "<text>"}]}`,
and bring it like any other finding.

## 6. Again, until verify passes

Run `proviso verify --json`. An edit leaves elements `stale` or
candidates `unjudged`: run again from step 2, with a fresh checker,
which receives only what is left. When `verify` exits with 0:

- commit `proviso/.coherence.json` on the amendment's branch, as
  `docs: record the coherence check of <amendment>`;
- report to the engineer: the runs, the judgements by reason, every
  finding and how it was closed.

Never build, never accept, never push.
