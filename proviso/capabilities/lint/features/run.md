# <a id="lint.run"></a> Running lint

- Contact: Bruno Baketarić

## Goal

Running lint checks the right spec tree, reports its findings in a form
people and agents can rely on, and tells by its exit status whether the
tree may be kept as it is.

## Rationale

Agents call lint for every mechanical step and act on what it returns,
so its output must be the same for the same tree, wherever and however
it was run. A random order of findings, such as a map's iteration order,
is exactly the kind of default nobody chose, so the order is specified.

## <a id="lint.run.root"></a> Rule: Lint finds the nearest spec root

### Rationale

Searching upwards, as git does for `.git`, lets lint run from any folder
of the repository, including from inside the tree, without a config
file.

```guarded
When      someone runs `lint` in the directory <directory>
Provided  <directory> or a directory above it holds a [spec root]
Then      lint checks the [spec tree] of the [spec root] nearest to <directory>
  And     lint reports what it finds there
Otherwise lint writes to "stderr" that no [spec root] was found
  And     lint exits with `lint.exit-status` 2
  And     nothing goes to "stdout", not even an empty `report`
```

| directory | provided | outcome | why |
| --- | --- | --- | --- |
| `repo/`, holding `proviso/system.md` | yes | checks `repo/proviso/` | |
| `repo/cmd/proviso/`, below that `repo/` | yes | checks `repo/proviso/` | found upwards |
| `repo/proviso/capabilities/lint/`, inside the tree | yes | checks `repo/proviso/` | found from inside the tree |
| `repo/a/`, where `repo/a/proviso/system.md` and `repo/proviso/system.md` both exist | yes | checks `repo/a/proviso/` | the nearest wins |
| `repo/`, holding `proviso/` without `system.md` | no | exit status 2 | a folder named `proviso/` is not yet a spec root |
| `repo/`, holding `Proviso/System.md` | no | exit status 2 | names are matched exactly, case included, on every platform |
| `repo/`, holding `system.md` but no `proviso/` | no | exit status 2 | the file alone does not mark a spec root |
| `other/`, with no spec root up to the file system's root | no | exit status 2 | |

## <a id="lint.run.unreadable"></a> Rule: Lint checks only a tree it can read completely

### Rationale

A file that cannot be read does not look missing to whatever reads the
tree next: every element in it looks removed. Nothing is derived from a
partial tree, so lint stops rather than reporting on part of it. This is
not an error finding: the notation is not at fault, lint could not run.

Lint reads only regular files and folders. It never follows a symbolic
link, which could loop, read a file twice or reach outside the tree, so
a link counts as unreadable wherever it stands and whatever its name,
and so does a Windows junction or any other entry that is neither a
regular file nor a folder, unless a hidden name has it skipped, which
loses nothing, and the tree stays self-contained. A file the layout
does not list is never read at all: it is reported as unknown, and
leaves no part of the tree unread.

```guarded
Given     lint has found the [spec root] for the run
When      lint reads the files of its [spec tree] that `tree_layout` lists, and finds its symbolic links and other entries that are neither a regular file nor a folder, not skipped as hidden, among them <unreadable>
Provided  <unreadable> is empty
Then      lint checks the whole [spec tree]
  And     lint reports what it finds
Otherwise lint writes every file in <unreadable> to "stderr"
  And     lint exits with `lint.exit-status` 2
  And     nothing goes to "stdout", not even an empty `report`
```

| unreadable | provided | outcome | why |
| --- | --- | --- | --- |
| none | yes | the tree is checked | |
| `glossary.md`, without read permission | no | exit status 2, naming `glossary.md` | |
| `a.md`, not valid UTF-8 | no | exit status 2, naming `a.md` | text that cannot be decoded cannot be read |
| `shared`, a symbolic link to a folder outside the tree | no | exit status 2, naming `shared` | a link is never followed |
| `shared`, a Windows junction to a folder outside the tree | no | exit status 2, naming `shared` | neither a regular file nor a folder |
| `x.md`, a symbolic link to another feature in the tree | no | exit status 2, naming `x.md` | following it would read the feature twice |
| `a.md` and `b.md`, both unreadable | no | exit status 2, naming both | every unreadable file, not only the first |
| `manifest.sum`, without read permission | no | exit status 2, naming `manifest.sum` | the model version is read only from a manifest that can be read |
| none, while `notes.md`, not in the layout, has no read permission | yes | the tree is checked, and `notes.md` is reported as `file-unknown` | a file outside the layout is never read, so it leaves no part of the tree unread |
| none, while `.shared.md` is a symbolic link with a hidden name | yes | the tree is checked, and the link is skipped | a hidden entry never holds spec content |

## <a id="lint.run.model"></a> Rule: Lint checks only a tree of its own model version

### Rationale

The notation changes between versions of the model, and a tree read
with another version's notation is misread without any sign of it. The
manifest records the version a tree was last accepted under, where no
author has to keep it. A tree without a manifest has not been accepted
yet and has nothing to be compatible with, so it is read with the
binary's own version.

```guarded
Given     lint has read every file of its [spec tree] that `tree_layout` lists
When      lint starts to check the [spec tree] of the [spec root] <root> it has found for the run
Provided  <root> holds no `manifest_file`, or its first line, without its line ending, names `model_version`
Then      lint checks the [spec tree]
  And     lint checks nothing of `manifest_file` beyond its first line but what it checks of every file it reads
Otherwise → lint.run.model-other
```

| root | provided | outcome | why |
| --- | --- | --- | --- |
| `repo/proviso/`, without `manifest.sum` | yes | the tree is checked | not accepted yet |
| `repo/proviso/`, its `manifest.sum` starting with `model 0` | yes | the tree is checked | |
| `repo/proviso/`, its `manifest.sum` starting with `model 0` and a carriage return | yes | the tree is checked | the line ending is `line-ending`'s to report |
| `repo/proviso/`, its `manifest.sum` starting with `model 1` | no | handed over | a notation this lint does not know |
| `repo/proviso/`, its `manifest.sum` empty | no | handed over | no version is named |
| `repo/proviso/`, its `manifest.sum` starting with `model 00` | no | handed over | `model_version` is written without a leading zero |

## <a id="lint.run.model-other"></a> Rule: Lint names the model version a tree of another version needs

```guarded
When      lint finds that the first line <line> of `manifest_file` does not name `model_version`
Provided  <line>, without its line ending, is "model", a space and a whole number in decimal digits without a leading zero
Then      lint writes to "stderr" that number as the model version the [spec tree] needs, and `model_version` as the one lint reads
  And     lint exits with `lint.exit-status` 2
  And     nothing goes to "stdout", not even an empty `report`
Otherwise lint writes to "stderr" that `manifest_file` names no model version
  And     lint exits with `lint.exit-status` 2
  And     nothing goes to "stdout", not even an empty `report`
```

| line | provided | outcome | why |
| --- | --- | --- | --- |
| `model 1` | yes | the tree needs model 1, lint reads model 0; exit status 2 | |
| `model 12` | yes | the tree needs model 12, lint reads model 0; exit status 2 | |
| empty | no | names no model version; exit status 2 | |
| `model` | no | names no model version; exit status 2 | |
| `model one` | no | names no model version; exit status 2 | |
| `model 00` | no | names no model version; exit status 2 | a leading zero; never "needs 0, reads 0" |
| `<<<<<<< HEAD` | no | names no model version; exit status 2 | a merge conflict left in the file |

## <a id="lint.run.output"></a> Rule: Lint writes text or a report

```guarded
Given     lint has checked a [spec tree] and found the [findings] <findings>
When      lint reports them, run with the options <options>
Provided  <options> include `lint.json`
Then      lint writes the `report` of <findings> to "stdout"
  And     nothing else goes to "stdout"
Otherwise lint writes <findings> to "stdout" as text, one per line
  And     nothing else goes to "stdout"
```

| findings | options | provided | outcome | why |
| --- | --- | --- | --- | --- |
| one error | `--json` | yes | a `report` with one finding | |
| none | `--json` | yes | a `report` with no findings | an empty list, not an empty output |
| one error | none | no | one line of text | |
| none | none | no | nothing written | silence when there is nothing to say |

## <a id="lint.run.order"></a> Rule: Findings are ordered

### Rationale

A fixed order makes two runs on the same tree produce the same output,
so tests can compare it and a person can diff it.

```guarded
When      lint reports the [findings] <findings>
Then      they appear ordered by file path, compared byte by byte, then line, then kind
  And     the order does not depend on the order in which files were read
```

| findings | outcome |
| --- | --- |
| `b.md:3`, `a.md:9`, `a.md:2` | `a.md:2`, `a.md:9`, `b.md:3` |
| `a.md:4 identifier-duplicate`, `a.md:4 anchor-misplaced` | `a.md:4 anchor-misplaced`, then `a.md:4 identifier-duplicate` |
| `a.md:1`, `a-b.md:1` | `a-b.md:1`, then `a.md:1` |

## <a id="lint.run.exit"></a> Rule: The exit status tells whether errors were found

```guarded
When      lint exits, having found the [findings] <findings>
Provided  none of <findings> is an [error]
Then      lint exits with `lint.exit-status` 0
  And     everything else stays as it was
Otherwise lint exits with `lint.exit-status` 1
  And     everything else stays as it was
```

| findings | provided | outcome | why |
| --- | --- | --- | --- |
| none | yes | 0 | |
| one ask | yes | 0 | an ask never fails a run |
| one error | no | 1 | |
| one error and one ask | no | 1 | the ask does not soften the error |

## <a id="lint.run.read-only"></a> Invariant: Lint never writes the tree

### Rationale

Lint is a check. A check that changes what it checks hides what it
found, and would make two runs on the same tree differ.

### Ranges over

Every run of `lint`, whatever it finds and however it ends.

### Property

Every file and folder under the [spec root] is byte for byte and entry
for entry as it was before the run.
