# <a id="upkeep.rehome"></a> Rehoming identifiers

- Contact: Bruno Baketarić

## Goal

Rehoming gives every element whose identifier names another place than
its identifier home an identifier that names it, renames every reference to
it that the notation marks, and writes the manifest of the rewritten
tree, so that the tree says the same as before under identifiers that
say where their elements stand. It also lists every other mention of an
identifier it renames, and renames those the engineer picks. It shows
what it would rename before it writes anything.

## Rationale

Renaming an identifier by hand shows in the element diff as one element
removed and one added, and every check of what depends on it starts
over, although no meaning changed. `rehome` renames and writes the
manifest in one step, so the element diff stays empty. For that, it
works only on a tree the manifest records, with no check or amendment
note beside it: a rename never mixes with an amendment's edits.
`rehome` splits the spec tree into element parts and fingerprints them
as `manifest` does.

Only the first segment of an identifier names a place; the rest stays,
so an element moved between features of one capability keeps its
identifier. `rehome` renames all or nothing: a tree half rehomed would
say where some elements stand and not others, so one collision refuses
the whole run, and so does one file it cannot write. Without
`rehome.write` it writes nothing, so that the engineer, and the skill
that runs it, see every rename before the tree changes. Whether a new
identifier was once given to an element since removed is left to the
authors: the manifest forgets removed elements, and the tree's history
is not searched.

An identifier in a code span outside what the notation marks, such as
in an Examples cell, mostly refers to its element, but in the rules
about identifiers it is data, and only a person can tell which. `rehome`
therefore lists every such mention in its plan and renames only those
the engineer picks, in the same step as the rest: a mention left
behind would name an element the tree no longer holds, and no check
would notice. A mention is named by its file, line and identifier, as
findings are, since one line, such as one Examples row, is one case.
A list naming a mention the plan does not hold refuses the run, so
that a stale or mistyped list never writes half of what was meant.

## <a id="upkeep.rehome.lint"></a> Rule: Rehome runs only where lint finds no error

```guarded
When      someone runs `rehome`, with arguments it takes, in the directory <directory>
Provided  `lint`, run in <directory>, would exit with `lint.exit-status` 0
Then      `rehome` works on the [spec tree] that `lint` would check
  And     nothing of what `lint` would write is written
Otherwise `rehome` writes to "stderr" what `lint` would write as text
  And     `rehome` exits with `rehome.exit-status` 2
  And     nothing goes to "stdout", and nothing is written
```

| directory | provided | outcome | why |
| --- | --- | --- | --- |
| `repo/`, its tree without findings | yes | the command runs | |
| `repo/`, its tree with an ask only | yes | the command runs | |
| `repo/`, its tree with an error | no | exit status 2, the error on stderr | |
| `repo/`, its tree with an unreadable file | no | exit status 2 | lint could not run |
| `other/`, with no spec root up to the file system's root | no | exit status 2 | |

## <a id="upkeep.rehome.baseline"></a> Rule: Rehome works only on a tree the manifest records

```guarded
When      `rehome` works on the [spec tree] of the [spec root] <root>
Provided  <root> holds `manifest_file`
  And     every [element part] of the [spec tree] has the [fingerprint] `manifest_file` records for it, and `manifest_file` records no other [element part]
  And     <root> holds neither `coherence_file` nor `amendment_note`
Then      `rehome` plans the renames of the [spec tree]
  And     everything else stays as it was
Otherwise `rehome` writes to "stderr" which of these does not hold
  And     `rehome` exits with `rehome.exit-status` 2
  And     nothing goes to "stdout", and nothing is written
```

| root | provided | outcome | why |
| --- | --- | --- | --- |
| `repo/proviso/`, its tree as `manifest.sum` records it | yes | the renames planned | |
| `repo/proviso/`, a paragraph reflowed since `manifest.sum` was written | yes | the renames planned | a fingerprint ignores whitespace |
| `repo/proviso/`, without `manifest.sum` | no (a manifest) | exit status 2 | a tree before its first acceptance |
| `repo/proviso/`, a rule's `Then` step edited since `manifest.sum` was written | no (the recorded fingerprints) | exit status 2 | an amendment in progress |
| `repo/proviso/`, a rule added since `manifest.sum` was written | no (the recorded fingerprints) | exit status 2 | |
| `repo/proviso/`, a rule removed since `manifest.sum` was written | no (the recorded fingerprints) | exit status 2 | the manifest records element parts the tree no longer has |
| `repo/proviso/`, holding `.coherence.json` | no (neither file) | exit status 2 | a check left over |
| `repo/proviso/`, holding `.amendment.md` | no (neither file) | exit status 2 | an amendment note left over |

## <a id="upkeep.rehome.home"></a> Rule: An identifier that names another place is renamed to name its identifier home

```guarded
When      `rehome` plans the renames of the [spec tree], and reads the [identifier] <id> of the [element] <element>
Provided  the first segment of <id>, up to its first ".", is the [identifier home] of <element>
Then      <id> stays as it is
  And     no rename is planned for <element>
Otherwise `rehome` plans to rename <id> to the [identifier home] of <element>, followed by what follows the first segment of <id>
  And     every place `rehome_references` lists that holds <id> is planned to hold the new [identifier]
```

| id | element | provided | outcome | why |
| --- | --- | --- | --- | --- |
| `coherence.impact.events` | a rule of the capability `coherence` | yes | stays | |
| `baseline.impact.events` | a rule of the capability `coherence` | no | renamed to `coherence.impact.events` | only the first segment names a place |
| `baseline.part` | a glossary entry of the system | no | renamed to `system.part` | |
| `lint.cli.lint.exit-status` | a member in `contracts/cli.md` of the system | no | renamed to `system.cli.lint.exit-status` | |
| `coherence` | the capability `coherence` | yes | stays | a capability is its own identifier home |
| `system` | the system | yes | stays | |
| `checks.run`, in the folder `capabilities/coherence/`, whose `capability.md` is identified as `checks` | a feature of that capability | yes | stays | the identifier home is the capability's identifier, not its folder's name |

## <a id="upkeep.rehome.mentions"></a> Rule: The plan lists every mention of an identifier it renames

### Rationale

The event needs no condition: whatever the plan renames, its mentions
are listed, so that the engineer sees them before anything is written.

```guarded
When      `rehome` plans to rename the [identifier] <id>
Then      every [identifier mention] of <id> in the [spec tree] is listed with the plan, once for each line it stands on, with the file and the line, how often it stands on that line, and the [element] whose text holds that line, where one does
  And     no [identifier mention] is renamed unless `rehome.mentions` names it
```

| id | outcome | why |
| --- | --- | --- |
| `baseline.impact.events`, standing as a code span in an Examples cell of `verify.md` | that cell's line listed, once, naming the rule whose table holds it | |
| `baseline.impact.events`, standing twice as a code span in one Examples row | that line listed once, standing there twice | one line is one case |
| `baseline.impact.events`, standing as a code span in the Rationale of a feature | that line listed, naming the feature | prose is listed too |
| `baseline.impact.events`, standing only within the longer code span `` `## <a id="baseline.impact.events"></a> Rule: B` `` in an Examples cell | nothing listed | the code span holds more than the identifier |
| `baseline.impact.events`, standing only in a fenced block of a contract | nothing listed | a fenced block is not read for mentions |
| `baseline.impact.events`, standing only in plain text, without backticks | nothing listed | |
| `lint.run`, not planned for a rename, standing as a code span in an Examples cell | nothing listed | only identifiers the plan renames |
| `baseline.impact.events`, standing as a code span in the text of a question in a `questions.md` | that line listed, naming no element | a `questions.md` holds no element |

## <a id="upkeep.rehome.collision"></a> Rule: A rename that would take an identifier refuses the run

```guarded
When      `rehome` has planned the renames <renames>
Provided  no new [identifier] of <renames> is the [identifier] of an [element] of the [spec tree]
  And     no new [identifier] of <renames> is planned for two [elements]
Then      `rehome` goes on with <renames>
  And     no collision is named
Otherwise `rehome` writes <renames>, naming each collision with the [identifier] it would take and the old [identifiers] planned for it
  And     `rehome` exits with `rehome.exit-status` 1
  And     nothing is written under the [spec root]
```

| renames | provided | outcome | why |
| --- | --- | --- | --- |
| `baseline.impact.events` to `coherence.impact.events`, the tree holding no `coherence.impact.events` | yes | goes on | |
| none | yes | goes on | nothing to rename is no collision |
| `baseline.x` to `coherence.x`, the tree holding a rule `coherence.x` | no (no identifier of an element) | collision on `coherence.x`, exit status 1 | |
| `baseline.x` to `coherence.x`, while `coherence.x` of another capability is renamed to `install.x` | no (no identifier of an element) | collision on `coherence.x`, exit status 1 | an identifier is taken until the run ends, so no two elements swap |
| `baseline.x` and `lint.x`, both to `coherence.x` | no (not planned twice) | collision on `coherence.x`, naming both, exit status 1 | |

## <a id="upkeep.rehome.write"></a> Rule: With write, every planned rename is written, or none

```guarded
When      `rehome`, run with `rehome.write`, goes on with the renames <renames>
Provided  every file of the [spec tree] that holds a place `rehome_references` lists for <renames>, or an [identifier mention] `rehome.mentions` names, and `manifest_file`, can be opened for writing
Then      every place `rehome_references` lists that holds an old [identifier] of <renames> holds the new one, so does every [identifier mention] of an old [identifier] on a line `rehome.mentions` names for it, and nothing else of its file changes
  And     `manifest_file` is what `manifest` would write for the rewritten [spec tree]
Otherwise `rehome` writes to "stderr" each file it cannot open for writing
  And     `rehome` exits with `rehome.exit-status` 2
  And     nothing goes to "stdout", and nothing is written
```

| renames | provided | outcome | why |
| --- | --- | --- | --- |
| `baseline.impact.events` to `coherence.impact.events`, its anchor in `impact.md` and a hand-over to it in `verify.md` | yes | both files and `manifest.sum` rewritten | |
| the same, with `baseline.impact.events` also in an Examples cell of `verify.md` | yes | the anchor and the hand-over renamed, the cell as it was | an Examples cell may hold an identifier as data |
| the same, run with `--mentions` naming that cell's line | yes | the anchor, the hand-over and the cell renamed | the engineer picked the mention |
| the same, the cell's line holding `baseline.impact.events` twice, run with `--mentions` naming that line | yes | the anchor, the hand-over and both code spans renamed | a line is picked whole |
| the same, run with `--mentions` naming no mention | yes | the anchor and the hand-over renamed, the cell as it was | |
| `baseline.dependencies` to `coherence.dependencies`, tagging a question in `questions.md` | yes | the tag renamed | |
| none | yes | nothing rewritten, `manifest.sum` as it was | the manifest of an unchanged tree is the same |
| `baseline.impact.events` to `coherence.impact.events`, `impact.md` read-only | no | exit status 2, naming `impact.md`, nothing written | |
| the same, `manifest.sum` read-only | no | exit status 2, naming `manifest.sum`, nothing written | |

## <a id="upkeep.rehome.picked"></a> Rule: Only mentions the plan lists can be picked

```guarded
When      `rehome`, run with `rehome.write` and `rehome.mentions`, goes on with its renames, and reads the input <input> from "stdin"
Provided  <input> is a `mention_list`
  And     every [identifier mention] <input> names is one its plan lists
Then      `rehome` writes its renames with the [identifier mentions] <input> names
  And     every other [identifier mention] stays as it was
Otherwise `rehome` writes to "stderr" why <input> cannot be taken
  And     `rehome` exits with `rehome.exit-status` 2
  And     nothing goes to "stdout", and nothing is written
```

| input | provided | outcome | why |
| --- | --- | --- | --- |
| a `mention_list` naming one mention the plan lists | yes | that mention renamed with the rest | |
| a `mention_list` with no mentions | yes | no mention renamed | |
| a `mention_list` naming one mention twice | yes | that mention renamed once | the same mention asked for twice |
| `{` | no (a `mention_list`) | exit status 2, nothing written | |
| a `mention_list` naming a line where the identifier stands in plain text only | no (one the plan lists) | exit status 2, nothing written | a stale or mistyped list writes nothing |
| a `mention_list` naming a mention of `lint.run`, which the plan does not rename | no (one the plan lists) | exit status 2, nothing written | |

## <a id="upkeep.rehome.output"></a> Rule: The plan is written as text or as JSON

```guarded
Given     `rehome` has planned the renames <renames>, with their [identifier mentions], and found the collisions <collisions>
When      `rehome` writes them, run with the options <options>
Provided  <options> include `rehome.json`
Then      `rehome` writes the `rehome_plan` of <renames>, their [identifier mentions] and <collisions> to "stdout"
  And     nothing else goes to "stdout"
Otherwise `rehome` writes <renames>, their [identifier mentions] and <collisions> to "stdout" as text
  And     nothing else goes to "stdout"
```

| renames | collisions | options | provided | outcome | why |
| --- | --- | --- | --- | --- | --- |
| one | none | `--json` | yes | a `rehome_plan` with one rename | |
| one, its identifier in two Examples cells | none | `--json` | yes | a `rehome_plan` with one rename and two mentions | |
| none | none | `--json` | yes | a `rehome_plan` with no renames | an empty list, not an empty output |
| one | none | `--write --json` | yes | a `rehome_plan` with one rename | the renames written; the mentions listed as planned, picked or not |
| one, its identifier in two Examples cells | none | `--write --mentions --json`, naming one of them | yes | a `rehome_plan` with one rename and both mentions | the output shows the plan; the list passed in says what was picked |
| two, one colliding | one | none | no | both renames and the collision as text | |
| none | none | none | no | nothing written | silence when there is nothing to say |

## <a id="upkeep.rehome.order"></a> Rule: The renames and collisions are ordered

```guarded
When      `rehome` writes the renames <renames> and the collisions <collisions>
Then      <renames> appear ordered by their old [identifier], compared byte by byte
  And     <collisions> appear ordered by the [identifier] each would take, compared so too, each naming its old [identifiers] in that order
  And     the [identifier mentions] of <renames> appear ordered by their file's path, then their line, then their [identifier], paths and [identifiers] compared byte by byte
```

| renames | collisions | outcome |
| --- | --- | --- |
| `lint.keyword` to `system.keyword`, `baseline.part` to `system.part` | none | `baseline.part` first, then `lint.keyword` |
| `lint.x` and `baseline.x`, both to `coherence.x`, and `lint.y` to `coherence.y`, taken | on `coherence.x` and on `coherence.y` | `coherence.x`, naming `baseline.x` then `lint.x`, then `coherence.y` |
| `lint.keyword` to `system.keyword`, mentioned in `capabilities/lint/features/rules.md` lines 9 and 3 and in `capabilities/baseline/features/diff.md` line 40 | none | `baseline/features/diff.md` 40, then `lint/features/rules.md` 3, then 9 |

## <a id="upkeep.rehome.exit"></a> Rule: The exit status tells whether the plan holds

### Rationale

A run that stops earlier, because lint finds an error, the tree is not
as the manifest records it, or a file cannot be opened for writing,
writes no plan and exits with 2 by its own rule.

```guarded
When      `rehome` exits, having written the plan of its renames, with the collisions <collisions>
Provided  <collisions> is empty
Then      `rehome` exits with `rehome.exit-status` 0
  And     everything else stays as it was
Otherwise `rehome` exits with `rehome.exit-status` 1
  And     everything else stays as it was
```

| collisions | provided | outcome | why |
| --- | --- | --- | --- |
| none | yes | 0 | |
| none, and no renames | yes | 0 | a tree whose identifiers all name their identifier home |
| one | no | 1 | |

## <a id="upkeep.rehome.writes"></a> Invariant: Rehome changes nothing but the renamed identifiers and the manifest

### Rationale

`rehome` renames; any other change would mix a meaning into a step that
must change none. Git is only read, so a run never moves a branch or
touches the index.

### Ranges over

Every run of `rehome`, however it ends.

### Property

Every file and folder under the [spec root] is byte for byte and entry
for entry as it was before the run, except, after a run with
`rehome.write` that exits with 0, `manifest.sum`, the identifiers
renamed in the places `rehome_references` lists, and the
[identifier mentions] `rehome.mentions` names. The git repository
holding the [spec root] is as it was before the run.
