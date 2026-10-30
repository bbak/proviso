# <a id="baseline.judge"></a> Recording the coherence check

- Contact: Bruno Baketarić

## Goal

Recording the coherence check keeps its judgements and findings beside
the amendment, together with what they covered, so that acceptance can
tell whether the check covers the amendment as it stands.

## Rationale

Judging is semantic and belongs to an agent; what the judgement covered
is mechanical and belongs to the command line. An agent cannot compute
fingerprints, nor count findings reliably across sessions, so `judge`
adds both. A judgement says why a candidate stays as it is: a candidate
the amendment does affect is edited, and the element diff lists it, or
it is reported as a finding until it can be. A judgement or a finding
that names an element the tree does not hold is refused rather than
recorded, since it would cover nothing while seeming to.

`judge` derives the element diff and the candidates as `impact` does,
and records them in `coherence_file`, which only `judge` writes.
`judge` reads `coherence_file`, if the spec root holds one, before it
reads its input or derives the element diff, once lint has found no
error; one it cannot read stops it, as `baseline.judge.unreadable` says.

## <a id="baseline.judge.input"></a> Rule: Judge reads a check result

```guarded
When      someone runs `judge`, with arguments it takes, giving it the input <input> on "stdin"
Provided  <input> is a `check_result`
Then      `judge` reads the [judgements] and [findings] of <input>
  And     nothing goes to "stdout"
Otherwise `judge` writes to "stderr" why <input> is no `check_result`
  And     `judge` exits with `judge.exit-status` 1
  And     `coherence_file` stays as it was
```

| input | provided | outcome | why |
| --- | --- | --- | --- |
| one judgement naming two elements, and one new finding | yes | read | |
| no judgements and no findings | yes | read | an amendment without candidates still needs its check |
| not JSON | no | exit status 1 | |
| a finding with both a `kind` and an `id` | no | exit status 1 | it is new or it changes one, not both |
| one element in two judgements | no | exit status 1 | which reason holds would be a guess |
| a judgement and a finding with an `id` | no | exit status 1 | a check and a decision are recorded apart |

## <a id="baseline.judge.amendment"></a> Rule: Judge needs an amendment to judge

### Rationale

A check of a tree the manifest records judges nothing, and its file
would be left over at once.

```guarded
When      `judge` has read the input of a run on the [spec tree] whose [element diff] lists the [elements] <elements>
Provided  <elements> is not empty
Then      `judge` takes the [judgements] and [findings] of the input one by one
  And     nothing goes to "stdout"
Otherwise `judge` writes to "stderr" that there is no [amendment] to judge
  And     `judge` exits with `judge.exit-status` 1
  And     `coherence_file` stays as it was
```

| elements | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.run.exit`, changed | yes | taken one by one | |
| none: the tree is as the manifest records it | no | exit status 1 | |

## <a id="baseline.judge.candidate"></a> Rule: A judgement is recorded for a candidate

```guarded
When      `judge` takes a [judgement] with the reason <reason> for the [element] <element>
Provided  <element> is one of the current [candidates]
Then      `judge` records for <element> that it stays as it is because of <reason>, with the [fingerprints] of its [parts] as they are now
  And     a [judgement] recorded for <element> before is replaced
Otherwise `judge` writes to "stderr" that <element> is no [candidate], naming it
  And     `judge` exits with `judge.exit-status` 1
  And     `coherence_file` stays as it was
```

| reason | element | provided | outcome | why |
| --- | --- | --- | --- | --- |
| "untouched: it reads no manifest" | `lint.examples.cells`, found through "events" | yes | recorded | |
| "untouched: nothing here builds code" | `lint.run.root`, judged before with another reason | yes | recorded, the former judgement replaced | |
| "untouched" | `lint.run.exit`, changed itself | no | exit status 1 | the element diff lists it |
| "untouched" | `lint.run.nosuch`, an identifier the tree does not hold | no | exit status 1 | |

## <a id="baseline.judge.new-finding"></a> Rule: A new finding gets the next number

```guarded
When      `judge` takes a [finding] without an id, naming the [elements] <elements>
Provided  every [element] of <elements> is in the [spec tree] or in the [baseline]
Then      `judge` records it as open, with the id "f" followed by one more than the highest number among the [findings] `coherence_file` holds, or "f1" if it holds none
  And     the [finding] keeps that id until `coherence_file` is removed
Otherwise `judge` writes to "stderr" that the [finding] names an [element] neither holds, naming it
  And     `judge` exits with `judge.exit-status` 1
  And     `coherence_file` stays as it was
```

| elements | provided | outcome | why |
| --- | --- | --- | --- |
| `baseline.impact.events`, with `f1` and `f2` recorded before | yes | recorded as `f3` | |
| none, with no finding recorded before | yes | recorded as `f1` | a gap may stand beside every element |
| `lint.run.root`, removed by the amendment | yes | recorded | the baseline holds it |
| `lint.run.nosuch` | no | exit status 1 | |

## <a id="baseline.judge.close-finding"></a> Rule: A finding is resolved or dismissed by its number

```guarded
When      `judge` takes a [finding] with the id <id> and the status <status>
Provided  `coherence_file` holds a [finding] with the id <id>
Then      `judge` records that [finding] as <status>, with the reason given for it
  And     the [finding] stays in `coherence_file`, so that the reason can be read at acceptance
Otherwise `judge` writes to "stderr" that no [finding] has the id <id>
  And     `judge` exits with `judge.exit-status` 1
  And     `coherence_file` stays as it was
```

| id | status | provided | outcome | why |
| --- | --- | --- | --- | --- |
| `f2` | "resolved" | yes | `f2` resolved | an edit removed it |
| `f2` | "dismissed", with the reason "decided in system.git" | yes | `f2` dismissed with its reason | |
| `f2`, dismissed before | "resolved" | yes | `f2` resolved | the latest status holds |
| `f9`, with `f1` and `f2` recorded | "resolved" | no | exit status 1 | |

## <a id="baseline.judge.write"></a> Rule: Judge records what the check covered

### Rationale

A run of `judge` that closes no finding is the agent saying that it
checked the amendment as it stands, so every touched element is
recorded as it is now, and as its baseline holds it: a merge that changes the baseline of a touched
element changes what the check was about, though the branch's own text
stays the same. A candidate is recorded only when it is judged, since
only then was it read.

```guarded
When      `judge` has taken every [judgement] and [finding] of the input of a run on the [spec root] <root>
Provided  `coherence_file` can be written into <root>
Then      `judge` writes `coherence_file` into <root>
  And     a [judgement] recorded before for an [element] that is no longer a [candidate] is dropped, and so is a reason that no [judgement] names any more
  And     every [finding] recorded before is kept
  And     `judge` exits with `judge.exit-status` 0
Otherwise `judge` writes to "stderr" why it could not write `coherence_file`
  And     `judge` exits with `judge.exit-status` 2
  And     a former `coherence_file` stays as it was
```

| root | provided | outcome | why |
| --- | --- | --- | --- |
| `repo/proviso/`, without `.coherence.json` | yes | `repo/proviso/.coherence.json` written | |
| `repo/proviso/`, with a judgement recorded for `lint.examples.cells`, edited since | yes | written, that judgement dropped | it is a touched element now |
| `repo/proviso/`, without write permission | no | exit status 2 | |

## <a id="baseline.judge.touched"></a> Rule: Closing a finding records no check

### Rationale

Closing a finding is the engineer's decision, not a check: an edit that
resolves it was read by no agent. A run that closes a finding leaves the
touched elements as the last check recorded them, so that an edit since
then shows as stale, and the next run checks it.

```guarded
When      `judge` records the input <input> in `coherence_file`
Provided  <input> closes no [finding]
Then      every [touched element] is recorded with the [fingerprints] of its [parts] now and in the [baseline]
  And     the check covers every [touched element] as it stands
Otherwise every [touched element] keeps what `coherence_file` recorded for it before, and one recorded for none stays so
  And     an edit of a [touched element] since the last check is left for the next check to read
```

| input | provided | outcome | why |
| --- | --- | --- | --- |
| one judgement and one new finding | yes | every touched element recorded as it is now | |
| no judgements and no findings | yes | every touched element recorded as it is now | the check of an amendment without candidates |
| `f1` resolved, `lint.run.exit` edited since the last run | no | `lint.run.exit` keeps what was recorded before | the edit was read by no agent |
| `f2` dismissed | no | every touched element keeps what was recorded before | |

## <a id="baseline.judge.unreadable"></a> Rule: A coherence file that cannot be read stops the command

### Rationale

`coherence_file` is never edited by hand, so content that `judge` did
not write is damage, such as a merge conflict left in it. Reading past
it could pass an amendment whose check is missing. A symbolic link is
not followed: what it points to may lie outside the branch.

```guarded
When      `judge` or `verify` reads the `coherence_file` <file>
Provided  <file> is a regular file that holds what `judge` writes
Then      the command takes what <file> records
  And     the command goes on
Otherwise the command writes to "stderr" that `coherence_file` cannot be read
  And     the command exits with 2, as `judge.exit-status` and `verify.exit-status` define
  And     nothing goes to "stdout"
```

| file | provided | outcome | why |
| --- | --- | --- | --- |
| written by `judge` | yes | read | |
| `<<<<<<< HEAD` in it | no | exit status 2 | a merge conflict left in the file |
| empty | no | exit status 2 | |
| a folder, beside a tree the manifest records | no | exit status 2 | `verify` reads the file before it asks whether it is left over |
| `<<<<<<< HEAD` in it, with input that is not JSON | no | exit status 2 | the damaged file is reported before the input |
| a symbolic link to a file written by `judge` | no | exit status 2 | a link is not followed |
