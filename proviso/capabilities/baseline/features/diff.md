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
When      someone runs `diff` without `diff.since`, and it works on the [spec tree] of the [spec root] <root>
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

## <a id="baseline.diff.since"></a> Rule: Since a revision, the tree at that revision takes the baseline's place

### Rationale

An amendment is built in rounds, and a resumed agent needs only what
changed since its last round. The tree at the revision it last received
stands in for the baseline: its parts are fingerprinted as the manifest
would record them, so the same comparison finds what changed, and the
former text comes from the same revision. Nothing of the manifest is
read beyond what `lint`, run first, checks. A
tree that lint would reject could be misread, and a misread tree would
show changes nobody made.

```guarded
When      someone runs `diff` with `diff.since`, or `pack` with `pack.since`, with arguments it takes, naming the revision <revision>, and it works on the [spec tree]
Provided  git knows <revision> as a commit
  And     `lint`, run on the [spec tree] at <revision>, would exit with `lint.exit-status` 0, or there is no [spec tree] at <revision>
Then      the command compares the [spec tree] with the [spec tree] at <revision>, every [element part] fingerprinted as `manifest_file` describes, in place of the [baseline]
  And     the text the command lists as former is read from <revision>
  And     nothing of `manifest_file`, in the [spec root] or at <revision>, is read beyond what `lint` checks
Otherwise the command writes to "stderr" why it cannot compare with <revision>
  And     the command exits with 2, as `diff.exit-status` and `pack.exit-status` define
  And     nothing goes to "stdout"
```

| revision | provided | outcome | why |
| --- | --- | --- | --- |
| `a1b2`, the commit an agent received in its last round | yes | compared with the tree at `a1b2`, former text from `a1b2` | |
| `HEAD~1` | yes | compared with the tree one commit back | any name git accepts |
| `e5f6`, from before the tree existed | yes | every element added | an empty tree, as without a manifest |
| `a1b2`, with the `manifest.sum` in the spec root damaged by a merge conflict after its first line | yes | compared with the tree at `a1b2` | the manifest is not read beyond what lint checks |
| `a1b2`, whose `manifest.sum` is damaged by a merge conflict after its first line | yes | compared with the tree at `a1b2` | the same at the revision |
| `nosuch` | no (a commit) | exit status 2 | |
| `c3d4`, whose tree has a lint error | no (no error) | exit status 2 | a misread tree would show changes nobody made |
| `a1b2`, whose `manifest.sum` names another model version | no (no error) | exit status 2 | lint refuses a tree of another model version, at a revision as in the spec root |
| `a1b2`, whose tree holds `x.md` as a symbolic link | no (no error) | exit status 2 | a link is never followed, at a revision as on disk |

## <a id="baseline.diff.malformed"></a> Rule: A manifest that cannot be read stops the diff

### Rationale

The manifest is never edited by hand, so a line that does not read as
the contract says is damage, such as a merge conflict left in it. A
guess at what it meant could hide every change it records.

```guarded
When      `diff`, `impact`, `pack`, `judge`, `verify` or `rehome` reads the line <line> of `manifest_file`, after its first
Provided  <line> is as `manifest_file` describes
Then      the command takes the [fingerprint] it records
  And     the command goes on with the next line
Otherwise the command writes to "stderr" that `manifest_file` cannot be read, naming <line>
  And     the command exits with 2, as `diff.exit-status`, `impact.exit-status`, `pack.exit-status`, `judge.exit-status`, `verify.exit-status` and `rehome.exit-status` define
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
Provided  the [baseline] holds no [element part] of an [element] with the [identifier] of <element>
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
Provided  an [element part] of <element> is missing from the [baseline], or differs from it in its [fingerprint], or the [baseline] holds an [element part] of <element> that the [spec tree] no longer has
Then      <element> is listed as changed, in full, naming each such [element part] as added, changed or removed
  And     the [former text] of each changed or removed [element part] is listed with it
  And     the text of each added or changed [element part] as it stands now is listed with it, so that a part named only by its [fingerprint], such as an [Examples row], can be told from the others
Otherwise <element> is not listed as changed
  And     everything else stays as it was
```

| element | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.run.exit`, its `Then` step reworded | yes | changed, naming `step Then` as changed | |
| `lint.run.root`, its Rationale deleted | yes | changed, naming `section Rationale` as removed | |
| `lint.run.exit`, with a new Examples row | yes | changed, naming that row as added, with its text | |
| `lint.run.exit`, an Examples row with a new outcome | yes | changed, naming that row as changed, with its text now and its former text | the placeholder values are the same |
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
stops where nothing depends on a changed meaning. Every element it
reaches is listed, but each names only the shortest chain from each
changed element that reaches it: one chain shows how a meaning arrives,
and naming every chain grows beyond any reader, since glossary entries
may mark each other's terms, in circles too, and eight entries marking
each other gave over two hundred thousand chains in one diff, and
thousands to a single element. Of
equally short chains the first is named, so that `diff` names the same
chain every time.
Which elements depend on what is read from the tree as it stands: one
that stopped depending on something was edited, and is listed anyway. A
removed member of a contract element is the exception: a rule that
referred to a contract element above it relied on it, and still refers
to that contract element.

```guarded
When      `diff` takes the [element] <element>, whose meaning changed
Provided  `change_effects` gives "dependents" for a changed [element part] of <element>, or <element> depends on an [element] whose meaning changed, whether or not the [element diff] lists <element> as changed
Then      every [element] that depends on <element> by `dependencies`, and is not listed as added, changed or removed, is listed as changed through a dependency, naming, for each [element] listed as added, changed or removed whose changed [element part] `change_effects` gives "dependents" for and from which a chain of [identifiers] reaches it, the shortest such chain, from that [element] to the one it depends on directly, and of equally short ones the first, compared by their first [identifier], then by their second, and so on, each byte by byte
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
| `lint.run.exit`, an Examples row changed | yes | every rule handing over to it listed | a row decides as much as a step |
| `lint.run.exit`, its Rationale reworded, marking a term whose Definition changed | yes | every rule handing over to it listed, through that term, then `lint.run.exit` | its meaning changed through the term, though it is listed as changed for its Rationale |
| `system.ref.commonmark`, a new `Version`, while `lint.step`, which it defines, has its Definition reworded too | yes | every rule marking `lint.step` listed once, through `lint.step` and through `system.ref.commonmark`, `lint.step` | a chain is named from each changed element that reaches it |
| `lint.findings.anchor-misplaced`, removed, while `lint.findings` stays | yes | every rule referring to `findings` listed, through `lint.findings.anchor-misplaced` | a rule relied on every member of `findings` |
| `lint.step`, a new alias only | no | nothing listed because of it | the meaning is the same |
| `system.ref.commonmark`, a new `Contact` only | no | nothing listed because of it | the standard is the same |
| `system.heading`, its Deviations reworded | yes | every rule marking it listed, through `system.heading` | a deviation is part of the term's meaning |
| `system.line-ending`, its Definition reworded, while `system.heading` marks it in its Deviations | yes | `system.heading` listed, through `system.line-ending`, and every rule marking it, through `system.line-ending`, `system.heading` | a deviation relies on the terms it marks |
| `system.line-ending`, its Definition reworded, while `install.version-stamp` marks it in its Definition | yes | `install.version-stamp` listed, through `system.line-ending`, and every rule marking it, through `system.line-ending`, `install.version-stamp` | a definition that marks a term relies on its meaning |
| `lint.step`, its Definition reworded, while `lint.guard` marks it in its Definition and `lint.step` marks `lint.guard` in its own | yes | `lint.guard` listed, through `lint.step`, and every rule marking either, each chain passing no element twice | a circle adds no chain of its own |
| `lint.step`, its Definition reworded, while `lint.guard` marks it, and `lint.guard` and `lint.keyword` mark each other, each in its Definition | yes | `lint.guard` listed, through `lint.step` only, and `lint.keyword`, through `lint.step`, `lint.guard` | a chain never comes back to the element it reaches |
| `system.spec-tree`, its Definition reworded, while `lint.findings` marks `lint.finding` and it in its field `Represents` | yes | `lint.findings` listed, through `system.spec-tree`, and every rule referring to `findings`, through `system.spec-tree`, `lint.findings` | a contract element depends on every term it represents |
| `lint.step`, its Definition reworded, while `lint.guard` marks it in its Definition and a rule marks both | yes | the rule listed, through `lint.step` only | only the shortest chain from each changed element is named |
| `system.line-ending`, its Definition reworded, while `system.heading` marks it in its Deviations, `install.version-stamp` marks it in its Definition, and a rule marks both of them but not it | yes | the rule listed, through `system.line-ending`, `install.version-stamp` only | of equally short chains the first is named |
| `lint.step`, its Definition reworded, while it marks itself in its Definition | yes | every rule marking it listed, through `lint.step` | an entry marking its own term depends on nothing through that mark |
| `system.change-effects`, its Rationale reworded only | no | nothing listed because of it | a reworded rationale cannot alter behaviour |

## <a id="baseline.diff.undecided"></a> Rule: An element with an undecided place is not to be built

### Rationale

Only the element's own places and the marks in its own rule block
count, not an open term reached through a definition that marks it:
such a term is undecided itself, `verify` reports it across the whole
tree, and no amendment is accepted while it stays open, so nothing
built around it is accepted unnoticed.

```guarded
When      `diff` lists the [element] <element> as added, changed or changed through a dependency
Provided  <element> holds no [undecided] place
  And     <element> marks no [term] whose [field] "Status" is "open" in its [rule block]
Then      <element> is to be built
  And     everything else stays as it was
Otherwise <element> is listed as not to be built, naming each line that holds an [undecided] place or a mark of an open [term], once
  And     everything else stays as it was
```

| element | provided | outcome | why |
| --- | --- | --- | --- |
| a rule with every step decided | yes | to be built | |
| a rule with `Then      ?` | no (an undecided place) | not to be built, naming the `Then` step | |
| a rule marking a term whose Status is `open` | no (an open term) | not to be built, naming the mark | its meaning is not decided |
| a contract element whose "Represents" field marks a term whose Status is `open` | yes | to be built | only a rule block's marks count |
| a glossary entry whose "Definition" field marks a term whose Status is `open` | yes | to be built | only a rule block's marks count |
| a rule with `?` in two cells of one Examples row | no (an undecided place) | not to be built, naming that row's line once | a place is named by its line |

## <a id="baseline.diff.former"></a> Rule: Former text comes from the revision that wrote the manifest

### Rationale

The manifest is written at acceptance and committed with the tree it
records, so the revision that last changed it holds the baseline's text.
A merge that brought in a newer manifest changed nothing of its own:
the revision that wrote that manifest is the one that counts. If the
manifest in the tree differs from the one in that revision, the former
text found there would not belong to the baseline. Only the former text
the element diff needs is checked against the manifest: a difference
anywhere else changes nothing a command derives from it.

```guarded
When      `diff`, `impact`, `pack`, `judge` or `verify`, run without `diff.since` or `pack.since`, derives the [element diff] and needs the [former text] of a removed [element] or of a changed or removed [element part], in the git history <history>
Provided  <history> names the newest revision that changed `manifest_file`, not counting a merge that took it unchanged from one of its parents
  And     `manifest_file` in that revision is the same as in the [spec root]
  And     the [spec tree] in that revision has the [fingerprints] `manifest_file` records for each removed [element] and each changed or removed [element part]
Then      the command reads the [former text] from the [spec tree] in that revision
  And     `diff`, `impact` and `pack` name that revision in what they write
Otherwise the command writes to "stderr" why it could not read the [former text]
  And     the command exits with 2, as `diff.exit-status`, `impact.exit-status`, `pack.exit-status`, `judge.exit-status` and `verify.exit-status` define
  And     nothing goes to "stdout", not even an empty `element_diff`, `impact_set`, `context_pack` or `verification`
```

| history | provided | outcome | why |
| --- | --- | --- | --- |
| the manifest committed in `a1b2`, the tree edited since | yes | former text from `a1b2` | |
| a merge `c3d4` that took the manifest from the mainline's `e5f6` | yes | former text from `e5f6` | the merge only took it over |
| none: the spec root is in no git repository | no (a revision) | exit status 2 | |
| the manifest written but never committed | no (a revision) | exit status 2 | |
| the manifest committed in `a1b2`, written again since and not committed | no (the same manifest) | exit status 2 | the text in `a1b2` is not the baseline's |
| the manifest committed in `a1b2` together with a tree that differs from it in a part the element diff lists as changed | no (the recorded fingerprints) | exit status 2 | the former text in `a1b2` is not the baseline's either |
| the manifest committed in `a1b2` together with a tree that differs from it only in parts the element diff does not list as changed or removed | yes | former text from `a1b2` | the text read is the baseline's |

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
  And     the [element parts] named for each appear in the order of their lines in `manifest_file`
  And     the chains of [identifiers] named for each appear ordered by their first [identifier], compared byte by byte
  And     the [undecided] places and the marks of open [terms] named for each appear in the order of their lines
```

| elements | outcome |
| --- | --- |
| `lint.run.root`, `lint.run.exit`, `baseline.parts.split` | `baseline.parts.split`, `lint.run.exit`, `lint.run.root` |
| `lint.run`, `lint.run.exit` | `lint.run`, then `lint.run.exit` |
| a rule changed through the chains `system.ref.commonmark`, `lint.step` and `lint.step` | `lint.step` first, then `system.ref.commonmark`, `lint.step` |
| a rule with `?` in its `Then` step and in its `When` step | the `When` step first, then the `Then` step |
| a rule marking an open term in its `When` step, with `?` in its `Then` step | the mark first, then the `Then` step |

## <a id="baseline.diff.exit"></a> Rule: The exit status tells whether the tree is as the manifest records it

### Rationale

The mainline's tree must be what its manifest records, so that no
amendment reaches it without acceptance. This exit status tells it in
one command; `verify`, which CI and the local checks run, builds on the
same comparison.

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
