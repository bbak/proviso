---
title: Tests with teeth
nav_order: 8
---

# Tests with teeth: the first mutation run

Proviso promises tests with teeth: tests that fail when the code is
wrong, not tests that confirm whatever the code does. This page reports
how well Proviso's own acceptance tests keep that promise, measured with
`proviso-mutate` on Proviso itself. It is one measurement of one system,
not a proof for every system; see [What it does not show](#what-it-does-not-show).

## What was measured

Proviso specifies itself: its spec tree is
[`proviso/`](https://github.com/bbak/proviso/tree/main/proviso), and its
acceptance tests were written from that tree by the testing agent, which
never saw the code. The run measured those tests.

- **Revision:** `cb834c0`, 3 November 2026, the state released as
  v0.4.0.
- **Scope:** the capabilities `cli`, `lint`, `baseline`, `coherence` and
  `install`, 149 rules. The capability `upkeep`, 8 rules, was measured
  the day before in a trial run on `f1eeebf`, and is shown apart.
- **The mutants:** one mutating agent per capability read the tree and
  the code, never the tests, and wrote one slip per rule, the kind a
  coding agent could make: a guard inverted, a refusal skipped, a
  boundary moved by one, a wrong message or exit status. It added about
  one slip per five rules in code no rule covers, called unanchored.
- **The run:** each mutant was applied alone, in a worktree of its own,
  and the tests of its rule were run; a mutant no test of its rule
  caught was run once more against the whole acceptance suite. A mutant
  is killed when any test fails, and survived when none does. The run
  took about four and a half hours on one Windows laptop.

## Result

| Capability | Rules | Mutants | Killed / run | Kill share | Survived | Not applied |
| --- | --- | --- | --- | --- | --- | --- |
| `cli` | 4 | 5 | 5 / 5 | 100% | 0 | 0 |
| `lint` | 68 | 82 | 69 / 82 | 84.1% | 13 | 0 |
| `baseline` | 27 | 32 | 28 / 30 | 93.3% | 2 | 2 |
| `coherence` | 25 | 30 | 28 / 29 | 96.6% | 1 | 1 |
| `install` | 25 | 30 | 23 / 30 | 76.7% | 7 | 0 |
| **All five** | **149** | **179** | **153 / 176** | **86.9%** | **23** | **3** |
| `upkeep`, trial | 8 | 10 | 10 / 10 | 100% | 0 | 0 |

Every rule in scope got a mutant. No mutant failed to build, none ran
into the time limit, and none escaped the tests of its own rule only to
be caught by another capability's tests. Three mutants could not be
applied, because the text they were to replace did not occur exactly
once in the code: the mutating agent's mistake, not the tests'.

### Where the tree decides, and where it does not

The total hides the finding that matters. Split by what the mutant
touched:

| Mutants | Killed / run | Kill share |
| --- | --- | --- |
| A slip in a rule's behaviour | 143 / 146 | 97.9% |
| A change no rule covers (unanchored) | 10 / 30 | 33.3% |

Where the tree decides the behaviour, the tests caught 143 of 146 slips.
Where the tree is silent, they caught one change in three. That is the
design working as intended, not a weakness hidden in an average: the
testing agent writes tests only from the tree, so what the tree does not
decide, nothing tests. The surviving unanchored mutants are not
failures of the tests; they show the places where the code does
something nobody decided.

## The survivors

Each survivor says one of three things. A **hole in the tests**: an
Examples row covers the case, but its test does not check it. A **hole
in the tree**: no rule or row decides the case, so no test could be
derived from it. Or **equivalent**: nothing anyone could observe
changes. Each reading was proposed, then decided by the engineer, one
survivor at a time; below, each carries its verdict and what came of it,
as of 4 November 2026.

### Holes in the tests: 3

All three rule slips that survived have a row that covers them, and a
test for that row that checks too little:

- **`lint.run.model-other`:** a manifest starting with `model 00` is
  reported as needing model version 00, where the row says it names no
  model version. The test of that row checks the exit status and the
  empty standard output, not the message.
- **`install.version.unstamped`:** a skill stamped `v0.1.0`, `[1]` or
  `0.1` is reported as installed by that version, where the rows say it
  names no Proviso version. The test checks that the file is named, not
  how.
- **`install.version.manifest-unreadable`:** an unreadable manifest is
  reported as naming no model version, where the rows say it cannot be
  read. The test checks only that the line names `manifest.sum`, which
  both messages do.

**Fixed.** The testing agent got the pack of each rule, `proviso pack
--element`, and only the row and the slip in plain words, never the code
or the mutant. It strengthened the three tests; applied again, each
mutant is now killed by exactly the rows it breaks, and the unmutated
suite still passes.

### The tree decides, but no row shows it: 3

- **`baseline`, 2:** a parameter whose only change is a reworded
  Rationale must not pass the change on to the rules that refer to it,
  and a rule whose only change is an Examples row must pass it on to the
  rules that hand over to it. The parameter `change_effects` decides
  both, but no Examples row of the rule that applies it shows either
  case.
- **`coherence`, 1:** a coherence file holding two findings with the
  same id must stop `judge` and `verify`, since `judge` never writes
  one. The rule decides it, but no row shows that case.

**Fixed** by one amendment that adds the three rows. The coherence check
found nothing to object to, the coding agent found nothing to change,
and the testing agent derived a test from each row. All three mutants
are now killed.

### Holes in the tree: 13

- **`lint`, 12:** how the Markdown of a spec tree is read in corner
  cases no rule decides: a byte order mark at the start of a file, an
  anchor written as `ID=`, words after `guarded` on a fence, a list
  marked with `*` or numbers right after a heading, an escaped `\|` in a
  table cell, a Markdown link read as a term mark, a placeholder with an
  uppercase letter, a hand-over target ending in a dot, a rule followed
  by two tables, the order of several findings on one line, and two
  cases of findings reported twice under a heading without an element
  type. **Decided, waiting for one amendment.** Most are settled by
  naming CommonMark, with GitHub's table extension, as the system's
  reference for reading Markdown; the rest by one row or one sentence
  each, such as the first table after a rule block being its Examples
  table.
- **`install`, 1:** updating a skill or agent installed as a symbolic
  link wrote through the link. **Fixed** by an amendment: `update` reads
  a stamp through a link at the file's own path, but never writes
  through it or replaces it. The mutant is now killed by two tests.

### Let go: 4

- **`install`, 2:** the permissions of the files and folders `init`
  writes. Git does not record them and Windows barely knows them; a
  rule for them would decide what nobody is likely to get wrong.
- **`install`, 2:** `init` writing the wrong kit file as
  `.markdownlint.jsonc` or as the coding agent's notes stub. The
  capability's rationale leaves the wording of installed files to
  Proviso's developers, since checking it would mean keeping it twice.

## Closing the survivors

| Survivors | Count | Status |
| --- | --- | --- |
| Holes in the tests | 3 | fixed in the tests |
| The tree decides, no row shows it | 3 | fixed by an amendment |
| A link at an installed skill's path | 1 | fixed by an amendment |
| How lint reads Markdown | 12 | decided, one amendment to come |
| Let go | 4 | decided not to fix |

Every fix ends the same way: the mutant it answers is applied again,
and must now be killed. All seven were.

The amendment about the symbolic link shows what a coherence check
does with a decision made at the wrong level. Its first edit decided
one case, a link at the skill's own path; the checking agent then found
the next case beside it, a link to the skill's folder, then one
pointing nowhere, then one pointing into the repository, nine runs and
sixteen findings in all. It converged only once the border was drawn
once, in the system's design boundary: links on the way to an
installed skill are the engineers' setup, and a link at the skill's
own path, or inside the spec tree, is the rules' to decide. The
amendment that stayed is three rows and one line of that boundary, and
the code did not change.

## What it shows

- **The tests have teeth where the tree has rules.** Nearly every slip a
  coding agent could make in specified behaviour, in five capabilities,
  made some acceptance test fail.
- **The misses point somewhere.** The three slips in rules that survived
  were tests that checked less than their row says, each fixed from the
  pack of its one rule. The rest are places the tree does not decide,
  which is what Proviso exists to bring to light: each became a
  decision, to fix or to let go.
- **The test writer and the mutant writer never met.** The tests were
  written without the code, the mutants without the tests. A high kill
  share is therefore not the result of one agent writing slips that its
  own tests were bound to catch.

## What it does not show

- **One system.** Proviso is a command-line application written in Go.
  How the result carries over to other languages, or to systems heavy
  in user interface or business logic, is untested.
- **One model family.** The testing, coding and mutating agents all ran
  on Claude. Agents of the same model may share blind spots, so slips
  that none of them thinks of stay unmeasured.
- **One slip per rule.** A rule has more ways to go wrong than one
  mutant probes. The kill share estimates how well the tests catch
  slips; it does not count every slip they would miss.
- **Small unanchored sample.** Thirty unanchored mutants show the
  pattern, not a precise rate.

To measure your own system's tests the same way, see
[Mutate](workflow.md#mutate-proviso-mutate) in the workflow.
