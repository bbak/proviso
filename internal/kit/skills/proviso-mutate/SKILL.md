---
name: proviso-mutate
description: Measure how many slips the tests of a system specified in a Proviso spec tree catch, and where they miss, with mutants written from the tree's rules by the mutating agent, which never sees the tests. Use when the engineer asks for mutation testing, or to measure, check or prove that the tests have teeth. Runs outside any amendment; brings every surviving mutant to the engineer, hands holes in the tree to proviso-integrate and holes in the tests to the testing agent, and never edits the spec tree. Never pushes.
---

# Measuring the tests with mutants

You measure whether the tests have teeth: the mutating agent writes
small slips into the system's code, one per rule, and you run the tests
against each. A slip the tests catch is killed; one they miss
survived, and says where the tests, or the tree they are written from,
fail to decide. Nothing changes the engineer's checkout: every mutant
is applied in a worktree of its own and reverted.

## Proviso's command

Wherever this skill says `proviso`, run Proviso's command: the one
`.proviso/build.md` names, or `proviso` from the path if it names none.
Look nowhere else.

## 1. Check the start

- The working tree is clean.
- `proviso lint` exits with 0.
- `proviso diff` exits with 0: the tree is what the manifest records,
  so the code carries out the tree as it stands. If it lists elements,
  an amendment is in progress and the code may not carry it out yet,
  so a survivor may only be unbuilt work: say so, and go on only if the
  engineer wants to.

Stop and say why if the working tree is not clean or lint fails.

## 2. How to run the tests

Read `.proviso/build.md`. You need from it, or from the engineer:

- **the commands** that build the system and run its tests;
- **how to run the tests of one element,** by its identifier, if they
  can be selected that way;
- **where the tests are:** their folders, files or name patterns, which
  the mutating agent must never read;
- **the results folder,** where the records of runs are kept, if any.

What the notes do not say, find in the repository, such as a
`Taskfile.yml`, a `Makefile`, a `package.json` or a `pyproject.toml`,
then confirm it with the engineer. For the results folder, ask: keep the
record in a folder, which then belongs in `.gitignore`, or keep nothing
and only report. With each answer, tell the engineer that a line in
`.proviso/build.md` saves the question next time; never write the notes
yourself.

## 3. Unmutated, first

Make a worktree from `HEAD`, detached, in the system's temporary
folder, as `<temp>/proviso/<repository>/mutate-<n>/`, with `<n>` the
first number not in use; never inside the repository. Build the system
there and run the whole suite, and note how long it takes. It must
pass, but for the failures and skips the notes expect: a mutant can
only be judged against tests that pass without it. If it fails, stop
and say why.

## 4. Scope and cost

Count the rules: the headings `Rule:` in the tree, per capability. A
run probes every rule in its scope once, with one mutant each, and adds
about one unanchored mutant per five rules.

Estimate what a run costs: with tests selectable by identifier, a
mutant first runs only its rule's tests, and only a survivor runs the
whole suite; without, every mutant runs the whole suite. Use the
duration of step 3. Then ask the engineer, with the counts and the
estimates: the whole system, or a part, such as a capability or named
rules. A run the engineer starts deliberately should cover a
meaningful share; recommend the whole system where its cost allows.

## 5. The mutants

Start one `proviso-mutator` per capability in scope, in parallel. Give
each, and nothing more:

- the folder of its capability and the system level of the tree: the
  `.md` files directly in the spec root and its `contracts/` folder,
  by path;
- the identifiers of the rules in scope of that capability;
- the number of unanchored mutants it writes;
- the folders or files holding the system's code, by path;
- where the tests are, which it must never read;
- the file to write to: `<temp>/proviso/<repository>/mutate-<n>/mutants-<capability>.json`.

Wait until each has reported. Never give it the tests, a test's name, a
test result or anything from an earlier run.

## 6. Run each mutant

Number the mutants `m1`, `m2` and so on, across all capabilities. For
each, in the worktree of step 3:

1. **Apply it:** replace `find` with `replace` in `file`. If `find` does
   not occur exactly once, the mutant is `unapplied`: record it, never
   guess.
2. **Build.** If the build fails, the mutant is `invalid`: no test can
   judge it.
3. **Run** the tests of its rule, where they can be selected by
   identifier; otherwise, and for an unanchored mutant, the whole suite.
   A run that takes more than three times as long as it does unmutated
   is stopped, and the mutant is `killed` by `timeout`.
4. **Record** the tests that failed. Any failure kills it: `killed`, by
   those tests. None: `survived`.
5. **Revert** the file, and check that the worktree is clean again.

Run every survivor once more against the whole suite: a test of
another element may catch it. Only a mutant the whole suite misses
survived.

## 7. The result

For each capability and for the run:

- **the kill share:** killed mutants among those applied, built and
  run, as a fraction and a percentage;
- **the coverage:** rules with a mutant among the rules in scope;
- **the survivors,** each with its rule, the slip and the code it
  changed.

## 8. Each survivor

Read each survivor's rule and its Examples in the tree, the tests of
the rule, and the code the mutant changed. Propose one of:

- **a hole in the tests:** a row of the rule's Examples covers the side
  the slip breaks, yet no test failed: the test of that row is missing
  or does not check what the row says;
- **a hole in the tree:** no row covers that side, or no rule decides
  the behaviour, as for most unanchored survivors: the testing agent
  had nothing to derive a test from;
- **equivalent:** nothing anyone could observe changes; not a finding.

Bring them to the engineer one at a time, each with your proposal and
why. The engineer decides each; record the verdict only then.

## 9. Fixes

Nothing is fixed without the engineer's yes.

- **A hole in the tree** becomes a request for `proviso-integrate`: what
  the code does in the case, and that the tree does not decide it.
  Several such holes may be one amendment; the engineer decides.
- **A hole in the tests** goes to `proviso-tester`. Make the branch
  `proviso/mutate-<n>/tests` from `HEAD` and a worktree for it in the
  system's temporary folder. Write `proviso pack --element <identifier>`
  of the rule to `<temp>/proviso/<repository>/mutate-<n>/pack-<identifier>.md`.
  Start the tester in that worktree, with the pack's path and, as the
  run's context: that the tests of this rule miss a case its Examples
  cover, the row, and the slip in behavioural terms, never the code,
  the mutant or its file. For the next rule, resume it with that rule's
  pack. Then apply each of those mutants in the tests worktree as in
  step 6: each must now be killed, and the unmutated suite must still
  pass. Show the engineer the result; on a yes, move the current branch
  to the tests branch as a fast forward, then remove the worktree and
  the branch.

## 10. Keep the record

If there is a results folder, write the run's record to it, as
`<date>-<commit>.json`, with `<date>` today in the form `YYYY-MM-DD`
and `<commit>` the short hash of `HEAD` at step 3:

```json
{
  "format": 1,
  "date": "YYYY-MM-DD",
  "commit": "<full hash>",
  "scope": ["<capability or rule identifier>"],
  "rules_in_scope": 0,
  "rules_probed": 0,
  "mutants": [
    {
      "id": "m1",
      "capability": "<identifier>",
      "rule": "<identifier, or null>",
      "part": "guard | otherwise | boundary | parameter | outcome | unanchored",
      "file": "<path>",
      "find": "<text>",
      "replace": "<text>",
      "slip": "<one line>",
      "status": "killed | survived | invalid | unapplied",
      "killed_by": ["<test name>"],
      "verdict": "tests | tree | equivalent | null",
      "fixed": false
    }
  ]
}
```

`killed_by` names the tests that failed, or `timeout`. A survivor's
`verdict` is the engineer's; `fixed` turns true once its fix is in.

Read the earlier records in the folder, and show beside this run's
result how the kill share of each capability moved, and which earlier
survivors are still open. Earlier survivors not yet fixed can be
applied again, to see whether a later change killed them.

From the `killed_by` lists, note tests that killed nothing a mutant of
their rule reached, and mutants every test of a rule killed together.
These hint at tests that check little, but one run's sample proves
nothing about a single test; say so.

## 11. Finish

Remove the worktree of step 3. Report: the scope, the kill share and
coverage per capability and for the run, every survivor and its
verdict, what was fixed and how, and where the record is, if kept.

Never edit the spec tree, never accept, never push.
