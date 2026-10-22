# <a id="baseline.diff"></a> The element diff

- Contact: Bruno Baketarić

## Goal

The element diff lists what an amendment did to the spec tree, element
by element, against the baseline: what was added, what changed and in
which part, what was removed, which rules changed through a term, and
which elements may not be built yet.

## Rationale

The element diff is the first thing derived from an amendment, and the
testing and coding agents build from it. It compares fingerprints, so
what changed is decided from the spec tree and the manifest alone. Git
is asked only for former text, which a reader needs to see how a part
changed and which tests a removed element retires. A diff that cannot
show former text is refused rather than shown without it, since a gap
nobody notices is what Proviso exists to prevent.

## <a id="baseline.diff.baseline"></a> Rule: Without a manifest, the baseline is empty

```guarded
When      someone runs `diff`, and it works on the [spec tree] of the [spec root] <root>
Provided  <root> holds `manifest_file`
Then      the [baseline] is what `manifest_file` records
  And     `diff` compares the [spec tree] with it
Otherwise the [baseline] is empty
  And     every [element] of the [spec tree] is added
```

| root | provided | outcome | why |
| --- | --- | --- | --- |
| `repo/proviso/`, with `manifest.sum` | yes | compared with the manifest | |
| `repo/proviso/`, without `manifest.sum` | no | every element added | a tree before its first acceptance |

## <a id="baseline.diff.malformed"></a> Rule: A manifest that cannot be read stops the diff

### Rationale

The manifest is never edited by hand, so a line that does not read as
the contract says is damage, such as a merge conflict left in it. A
guess at what it meant could hide every change it records.

```guarded
When      `diff`, `impact` or `pack` reads the line <line> of `manifest_file`, after its first
Provided  <line> is as `manifest_file` describes
Then      the command takes the [fingerprint] it records
  And     the command goes on with the next line
Otherwise the command writes to "stderr" that `manifest_file` cannot be read, naming <line>
  And     the command exits with 2, as `diff.exit-status`, `impact.exit-status` and `pack.exit-status` define
  And     nothing goes to "stdout"
```

| line | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.run.exit`, tab, `heading`, tab, `h1:` and 64 hexadecimal digits | yes | read | |
| `<<<<<<< HEAD` | no | exit status 2 | a merge conflict left in the file |
| `lint.run.exit`, space, `heading`, space, `h1:` and 64 hexadecimal digits | no | exit status 2 | spaces for tabs |

## <a id="baseline.diff.added"></a> Rule: An element the baseline does not hold is added

```guarded
When      `diff` compares the [element] <element> of the [spec tree] with the [baseline]
Provided  the [baseline] holds no [part] of an [element] with the [identifier] of <element>
Then      <element> is listed as added, in full
  And     everything else stays as it was
Otherwise → baseline.diff.changed
```

| element | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.examples.row-repeated`, new since the manifest was written | yes | added | |
| `lint.run.exit`, unchanged | no | handed over | |
| `lint.run.exit`, moved into another feature's file | no | handed over | the identifier stays |

## <a id="baseline.diff.changed"></a> Rule: An element whose parts differ is changed

```guarded
When      `diff` compares the [element] <element>, whose [identifier] the [baseline] holds, with the [baseline]
Provided  a [part] of <element> is missing from the [baseline], or differs from it in its [fingerprint], or the [baseline] holds a [part] of <element> that the [spec tree] no longer has
Then      <element> is listed as changed, in full, naming each such [part] as added, changed or removed
  And     the [former text] of each changed or removed [part] is listed with it
Otherwise <element> is not listed as changed
  And     everything else stays as it was
```

| element | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.run.exit`, its `Then` step reworded | yes | changed, naming `step Then` as changed | |
| `lint.run.root`, its Rationale deleted | yes | changed, naming `section Rationale` as removed | |
| `lint.run.exit`, with a new Examples row | yes | changed, naming that row as added | |
| `lint.run.exit`, an Examples row with a new outcome | yes | changed, naming that row as changed | the placeholder values are the same |
| `lint.run.exit`, moved into another feature's file | no | not listed | |
| `lint.run.exit`, its text wrapped at other places | no | not listed | |
| `lint.run`, one of its rules changed | no | not listed | a rule is not a part of its feature |

## <a id="baseline.diff.removed"></a> Rule: An element the tree no longer has is removed

```guarded
When      `diff` finds the [identifier] <id> in the [baseline]
Provided  the [spec tree] has an [element] with the [identifier] <id>
Then      that [element] is compared with the [baseline]
  And     everything else stays as it was
Otherwise <id> is listed as removed, with the [former text] of its [element]
  And     everything else stays as it was
```

| id | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.run.exit`, still in the tree | yes | compared | |
| `lint.run.exit`, its rule deleted | no | removed, with its former text | its tests can be retired |
| `lint.keyword`, moved to the system glossary | yes | compared | the identifier stays |

## <a id="baseline.diff.dependency"></a> Rule: An element that depends on a changed meaning is changed through it

### Rationale

A rule that marks a term means whatever the term's definition says; one
that refers to a parameter behaves as its value says; a term defined by
a standard means what that version of the standard says. When what an
element depends on changes its meaning, the element's text stays the
same while its meaning does not, and its tests may no longer hold. The
meaning passes on along every dependency, however long the chain, and
stops where nothing depends on a changed meaning: in practice a
reference, a glossary entry, a contract element and a rule at most.
Which elements depend on what is read from the tree as it stands: one
that stopped depending on something was edited, and is listed anyway.

```guarded
When      `diff` takes the [element] <element>, whose meaning changed
Provided  `change_effects` gives "dependents" for a changed [part] of <element>, or <element> depends on an [element] whose meaning changed, whether or not the [element diff] lists <element> as changed
Then      every [element] that depends on <element> by `dependencies`, and is not listed as added, changed or removed, is listed as changed through a dependency, naming the chain of [identifiers] from the first changed [element] to it
  And     every [element] listed so is taken in turn, and every [element] once
Otherwise nothing is listed because of <element>
  And     everything else stays as it was
```

| element | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.step`, its Definition reworded | yes | every rule marking it listed, through `lint.step` | |
| `system.ref.commonmark`, a new `Version` | yes | every term defined by it listed, and every rule marking one of them, through `system.ref.commonmark` and the term | the chain passes on |
| `lint.cli.lint.exit-status`, a new value | yes | every rule referring to `lint.exit-status` listed | |
| `lint.findings.anchor-misplaced`, a new description | yes | every rule referring to `findings.anchor-misplaced` or to `findings` listed | a rule referring to the element holding it depends on it too |
| `lint.run.exit`, its `Then` step changed | yes | every rule handing over to it listed | |
| `lint.run.exit`, its Rationale reworded, marking a term whose Definition changed | yes | every rule handing over to it listed | its meaning changed through the term, though it is listed as changed for its Rationale |
| `lint.step`, a new alias only | no | nothing listed because of it | the meaning is the same |
| `system.ref.commonmark`, a new `Contact` only | no | nothing listed because of it | the standard is the same |

## <a id="baseline.diff.undecided"></a> Rule: An element with an undecided place is not to be built

```guarded
When      `diff` lists the [element] <element> as added, changed or changed through a dependency
Provided  <element> holds no [undecided] place
  And     <element> marks no [term] whose [field] "Status" is "open"
Then      <element> is to be built
  And     everything else stays as it was
Otherwise <element> is listed as not to be built, naming each [undecided] place and each mark of an open [term]
  And     everything else stays as it was
```

| element | provided | outcome | why |
| --- | --- | --- | --- |
| a rule with every step decided | yes | to be built | |
| a rule with `Then      ?` | no (an undecided place) | not to be built, naming the `Then` step | |
| a rule marking a term whose Status is `open` | no (an open term) | not to be built, naming the mark | its meaning is not decided |

## <a id="baseline.diff.former"></a> Rule: Former text comes from the revision that wrote the manifest

### Rationale

The manifest is written at acceptance and committed with the tree it
records, so the revision that last changed it holds the baseline's text.
A merge that brought in a newer manifest changed nothing of its own:
the revision that wrote that manifest is the one that counts. If the
manifest in the tree differs from the one in that revision, the former
text found there would not belong to the baseline.

```guarded
When      `diff`, `impact` or `pack` needs the [former text] of a removed [element] or of a [part], in the git history <history>
Provided  <history> names the newest revision that changed `manifest_file`, not counting a merge that took it unchanged from one of its parents
  And     `manifest_file` in that revision is the same as in the [spec root]
  And     the [spec tree] in that revision has the [fingerprints] `manifest_file` records
Then      the command reads the [former text] from the [spec tree] in that revision
  And     the command names that revision in what it writes
Otherwise the command writes to "stderr" why it could not read the [former text]
  And     the command exits with 2, as `diff.exit-status`, `impact.exit-status` and `pack.exit-status` define
  And     nothing goes to "stdout", not even an empty `element_diff`, `impact_set` or `context_pack`
```

| history | provided | outcome | why |
| --- | --- | --- | --- |
| the manifest committed in `a1b2`, the tree edited since | yes | former text from `a1b2` | |
| a merge `c3d4` that took the manifest from the mainline's `e5f6` | yes | former text from `e5f6` | the merge only took it over |
| none: the spec root is in no git repository | no (a revision) | exit status 2 | |
| the manifest written but never committed | no (a revision) | exit status 2 | |
| the manifest committed in `a1b2`, written again since and not committed | no (the same manifest) | exit status 2 | the text in `a1b2` is not the baseline's |
| the manifest committed in `a1b2` together with a tree it does not record | no (the recorded fingerprints) | exit status 2 | the former text in `a1b2` is not the baseline's either |

## <a id="baseline.diff.output"></a> Rule: The element diff is written as text or as JSON

```guarded
Given     `diff` has listed the [elements] <elements>
When      `diff` writes them, run with the options <options>
Provided  <options> include `diff.json`
Then      `diff` writes the `element_diff` of <elements> to "stdout"
  And     nothing else goes to "stdout"
Otherwise `diff` writes <elements> to "stdout" as text
  And     nothing else goes to "stdout"
```

| elements | options | provided | outcome | why |
| --- | --- | --- | --- | --- |
| one changed | `--json` | yes | an `element_diff` with one element | |
| none | `--json` | yes | an `element_diff` with no elements | an empty list, not an empty output |
| one changed | none | no | that element as text | |
| none | none | no | nothing written | silence when there is nothing to say |

## <a id="baseline.diff.order"></a> Rule: The element diff is ordered

```guarded
When      `diff` writes the listed [elements] <elements>
Then      they appear ordered by [identifier], compared byte by byte
  And     the [parts] named for each appear in the order of their lines in `manifest_file`
```

| elements | outcome |
| --- | --- |
| `lint.run.root`, `lint.run.exit`, `baseline.parts.split` | `baseline.parts.split`, `lint.run.exit`, `lint.run.root` |
| `lint.run`, `lint.run.exit` | `lint.run`, then `lint.run.exit` |

## <a id="baseline.diff.exit"></a> Rule: The exit status tells whether the tree is as the manifest records it

### Rationale

The mainline's tree must be what its manifest records, so that no
amendment reaches it without acceptance. With this exit status, `diff`
is that check, in CI and in the local checks, without a command of its
own.

```guarded
When      `diff` exits, having listed the [elements] <elements>
Provided  <elements> is empty
Then      `diff` exits with `diff.exit-status` 0
  And     everything else stays as it was
Otherwise `diff` exits with `diff.exit-status` 1
  And     everything else stays as it was
```

| elements | provided | outcome | why |
| --- | --- | --- | --- |
| none | yes | 0 | |
| one removed | no | 1 | |
| a glossary entry with a new Definition, and one rule changed through it | no | 1 | |
