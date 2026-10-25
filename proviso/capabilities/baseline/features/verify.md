# <a id="baseline.verify"></a> Verifying the coherence check

- Contact: Bruno Baketarić

## Goal

Verifying tells whether the spec tree may become the baseline as it
stands: either nothing changed since the manifest was written and no
check or amendment note is left over, or the coherence check covers the amendment's final
state, with every candidate judged and no finding open.

## Rationale

CI and the local checks must hold every amendment to acceptance without
judging anything themselves. One command that compares the recorded
check with the tree serves them on every branch: it needs to know
neither which branch is the mainline nor who made the edit. What it
lists is what a re-run of the check needs to look at, and no more.

`verify` derives the element diff and the candidates as `impact` does,
and reads `coherence_file`, if the spec root holds one, before anything
else; one it cannot read stops it, as `baseline.judge.unreadable` says.
Of
`amendment_note`, it only asks whether the spec root holds one.

## <a id="baseline.verify.missing"></a> Rule: Without a coherence file, only an unchanged tree passes

```guarded
When      `verify` runs on a [spec root] without `coherence_file`, whose [element diff] lists the [elements] <elements>
Provided  <elements> is empty
Then      nothing is reported as "missing"
  And     `verify` reads no [judgement]
Otherwise `verify` reports "missing"
  And     `verify` looks no further
```

| elements | provided | outcome | why |
| --- | --- | --- | --- |
| none: the tree is as the manifest records it | yes | nothing reported as "missing" | the mainline after an acceptance |
| `lint.run.exit`, changed | no | "missing" | a hand edit whose check is missing |

## <a id="baseline.verify.leftover"></a> Rule: A coherence file or an amendment note beside an unchanged tree is left over

### Rationale

`manifest` removes `coherence_file` and `amendment_note` when it records
the tree, so either file beside a tree the manifest records was not
accepted: a merge that skipped acceptance, or an amendment whose edits
were all undone. A note left over would also reach the next
amendment's pack, and its agents would leave out what nobody excluded.

```guarded
When      `verify` runs on a [spec root] with `coherence_file`, `amendment_note` or both, whose [element diff] lists the [elements] <elements>
Provided  <elements> is not empty
Then      `verify` compares what `coherence_file` records with the [spec tree], where the [spec root] holds one
  And     nothing is reported for either file itself
Otherwise `verify` reports "leftover", once
  And     `verify` looks no further
```

| elements | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.run.exit`, changed, with `.coherence.json` | yes | compared | |
| `lint.run.exit`, changed, with `.amendment.md` only | yes | nothing reported for the note | the check is missing, as `baseline.verify.missing` says |
| none, with `.coherence.json` | no | "leftover" | |
| none, with `.amendment.md` only | no | "leftover" | |
| none, with `.coherence.json` and `.amendment.md` | no | "leftover", once | one problem, however many files |

## <a id="baseline.verify.unjudged"></a> Rule: A candidate without a judgement is unjudged

```guarded
When      `verify` compares what `coherence_file` records for the [candidate] <candidate>
Provided  `coherence_file` holds a [judgement] for <candidate>
Then      `verify` compares the [fingerprints] of that [judgement]
  And     nothing is reported as "unjudged" for <candidate>
Otherwise `verify` reports "unjudged" for <candidate>
  And     everything else stays as it was
```

| candidate | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.examples.cells`, judged | yes | compared | |
| `lint.examples.cells`, found through an event added after the check | no | "unjudged" | |

## <a id="baseline.verify.stale"></a> Rule: An element changed since the check is stale

### Rationale

A judgement holds for the text it was made on, and the check of a
touched element for the text it read and the baseline it read it
against. A merge of the mainline into the amendment can change the
baseline while the branch's own text stays the same.

```guarded
When      `verify` compares the [element] <element>, a [touched element] or a judged [candidate], with what `coherence_file` records for it
Provided  the [fingerprints] of the [parts] of <element> now, and for a [touched element] also in the [baseline], are the ones recorded
Then      nothing is reported as "stale" for <element>
  And     everything else stays as it was
Otherwise `verify` reports "stale" for <element>
  And     everything else stays as it was
```

| element | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.run.exit`, touched, unchanged since the check | yes | nothing reported | |
| `lint.run.exit`, touched, its `Then` step edited after the check | no | "stale" | |
| `lint.run.exit`, touched, its baseline changed by a merge of the mainline | no | "stale" | |
| `lint.run.root`, touched only after the check | no | "stale" | the check never read it |
| `lint.examples.cells`, judged, its Examples row edited by a merge of the mainline | no | "stale" | |

## <a id="baseline.verify.open"></a> Rule: A finding neither resolved nor dismissed is open

```guarded
When      `verify` takes the [finding] <finding> that `coherence_file` holds
Provided  <finding> is resolved or dismissed
Then      nothing is reported for <finding>
  And     everything else stays as it was
Otherwise `verify` reports "open" for <finding>, with its text
  And     everything else stays as it was
```

| finding | provided | outcome | why |
| --- | --- | --- | --- |
| `f1`, resolved | yes | nothing reported | |
| `f2`, dismissed with a reason | yes | nothing reported | |
| `f3`, as `judge` recorded it | no | "open" for `f3` | |

## <a id="baseline.verify.output"></a> Rule: The problems are written as text or as JSON

```guarded
Given     `verify` has found the problems <problems>
When      `verify` writes them, run with the options <options>
Provided  <options> include `verify.json`
Then      `verify` writes the `verification` of <problems> to "stdout"
  And     nothing else goes to "stdout"
Otherwise `verify` writes <problems> to "stdout" as text
  And     nothing else goes to "stdout"
```

| problems | options | provided | outcome | why |
| --- | --- | --- | --- | --- |
| one "unjudged" | `--json` | yes | a `verification` with one problem | |
| none | `--json` | yes | a `verification` with no problems | an empty list, not an empty output |
| one "unjudged" | none | no | that problem as text | |
| none | none | no | nothing written | silence when there is nothing to say |

## <a id="baseline.verify.order"></a> Rule: The problems are ordered

```guarded
When      `verify` writes the problems <problems>
Then      they appear ordered by kind, as `verification` lists the kinds
  And     the problems of one kind appear ordered by [identifier], compared byte by byte, or by the number of their [finding]
```

| problems | outcome |
| --- | --- |
| "stale" for `lint.run.exit`, "unjudged" for `lint.run.root` | "unjudged" first, then "stale" |
| "open" for `f10` and for `f2` | `f2`, then `f10` |

## <a id="baseline.verify.exit"></a> Rule: The exit status tells whether the check covers the tree

```guarded
When      `verify` exits, having found the problems <problems>
Provided  <problems> is empty
Then      `verify` exits with `verify.exit-status` 0
  And     everything else stays as it was
Otherwise `verify` exits with `verify.exit-status` 1
  And     everything else stays as it was
```

| problems | provided | outcome | why |
| --- | --- | --- | --- |
| none | yes | 0 | |
| one "leftover" | no | 1 | |
| one "open" | no | 1 | a finding left open blocks acceptance |
