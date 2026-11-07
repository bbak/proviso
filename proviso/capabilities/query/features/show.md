# <a id="query.show"></a> Showing an element

- Contact: Bruno Baketarić

## Goal

`show` writes the text of every element that the identifier or name
it is given stands for, with its type, where it stands and the element
it belongs to.

## Rationale

A reader meets an element in the tree by what the tree writes: an
identifier in a hand-over or an Examples cell, a term in a mark, a
contract element or a parameter in backticks. `show` takes exactly
that, matched as lint matches it, and nothing less: a part of a name
or a word of a heading is a search, which any search tool does over
the tree's files, and a match that depended on what else the tree
holds would give an agent one element today and three tomorrow. A
heading is shown, never matched, since no text of the tree refers to an
element by it.

A name is unique only within a capability and the system level, so the
same name may stand for one element in each of two capabilities, and
the middle of an amendment may leave an identifier twice. `show`
therefore writes a list, usually of one, and `--in` narrows it to what
one capability reads.

`show` runs none of lint's checks, since the tree is read most while
an amendment is being written, when lint may still find errors in it.
It finds and reads the tree as lint does, and refuses it where lint
could not read it, since a tree read in part would make elements look
missing, and one of another model version would be misread without any
sign of it.

The result names the file and line where an element stands, and the
element it belongs to, rather than offering an option to show more: the
element it belongs to is shown by asking for it, and the whole file is
read where the result says it stands.

## <a id="query.show.query"></a> Rule: Show needs a query

```guarded
When      `show` has read every argument it was run with, the arguments <arguments>
Provided  <arguments> hold `show.query`
Then      `show` goes on to find the [spec root]
  And     the rest of the run looks for that `show.query`
Otherwise `show` writes to "stderr" one line saying that `show` needs `show.query`, and that "proviso help" shows the help page
  And     `show` exits with `show.exit-status` 2
  And     nothing goes to "stdout", and `show` checks and writes nothing
```

| arguments | provided | outcome | why |
| --- | --- | --- | --- |
| `spec tree` | yes | looks for "spec tree" | one argument, as the shell passes `"spec tree"` |
| `--json`, `lint.run.exit` | yes | looks for "lint.run.exit" | |
| `lint.run.exit`, `--in`, `lint` | yes | looks for "lint.run.exit" | |
| none | no | refused, exit status 2 | |
| `--json` | no | refused, exit status 2 | |
| `--in`, `lint` | no | refused, exit status 2 | "lint" is the value of `--in` |

## <a id="query.show.tree"></a> Rule: Show reads the tree as lint does, without its checks

```guarded
When      someone runs `show`, with arguments it takes, in the directory <directory>
Provided  `lint`, run in <directory>, would find a [spec root]
  And     would read every file of its [spec tree] that `tree_layout` lists
  And     would find that its `manifest_file`, where it holds one, names `model_version`
Then      `show` works on that [spec tree], whatever `lint` would report about it
  And     nothing of what `lint` would write is written
Otherwise `show` writes to "stderr" what `lint` would write
  And     `show` exits with `show.exit-status` 2
  And     nothing goes to "stdout"
```

| directory | provided | outcome | why |
| --- | --- | --- | --- |
| `repo/`, its tree without findings | yes | `show` runs | |
| `repo/`, its tree with an `identifier-duplicate` error | yes | `show` runs | lint's errors do not stop it |
| `repo/proviso/capabilities/lint/`, inside the tree | yes | works on `repo/proviso/` | found as lint finds it |
| `repo/`, its tree without `manifest.sum` | yes | `show` runs | not accepted yet |
| `other/`, with no spec root up to the file system's root | no (spec root) | exit status 2 | |
| `repo/`, its `glossary.md` without read permission | no (read) | exit status 2, naming `glossary.md` | a tree read in part makes elements look missing |
| `repo/`, its tree holding `x.md`, a symbolic link to another feature | no (read) | exit status 2, naming `x.md` | |
| `repo/`, its `manifest.sum` starting with `model 1` | no (model version) | exit status 2 | another notation is misread |

## <a id="query.show.in"></a> Rule: A capability narrows what show looks among

### Rationale

A capability is named by its folder, so that `--in` works on a
capability in the making, whose `capability.md` is not yet written. A
capability the tree does not hold is refused rather than ignored, since
looking in the whole tree instead would widen what was asked for
without a sign.

```guarded
When      `show`, run with `show.in` naming <capability>, chooses the [elements] it looks among
Provided  the [spec tree] holds a folder named <capability> in the folder "capabilities" of its [spec root]
Then      `show` looks only among the [elements] in a file of that folder and those in a file of the system [level]
  And     no other [element] is shown, whatever it matches
Otherwise `show` writes to "stderr" that the [spec tree] holds no capability <capability>
  And     `show` exits with `show.exit-status` 2
  And     nothing goes to "stdout"
```

| capability | provided | outcome | why |
| --- | --- | --- | --- |
| `lint` | yes | looks among lint's elements and the system's | |
| `new`, a folder holding `features/a.md` but no `capability.md` | yes | looks among new's elements and the system's | a capability in the making |
| `nosuch` | no | exit status 2 | never the whole tree instead |
| `Lint` | no | exit status 2 | names are matched exactly, case included |
| `.old`, a hidden folder holding `capability.md` | no | exit status 2 | a hidden folder is not part of the tree |

## <a id="query.show.match"></a> Rule: Show takes every element the query names

```guarded
When      `show` looks for `show.query` <query> among the [elements] <elements>, every [element] of the [spec tree] unless `show.in` chose fewer
Provided  at least one of <elements> has <query> as its [identifier], character by character, or as its name: as a [glossary entry], its name or an alias, compared as `name_comparison` says; as a [contract element] or a [parameter], its name, character by character, as a [rule] refers to it in backticks: a [member] after the name of the [contract element] holding it and a "."
Then      `show` takes every such [element] once, with its [element type], or none where `tree_layout` gives none, the text of its [heading] after its [anchor], the file and line where its [heading] stands, and its [element parent], in the order the [spec tree] holds them: by the path of their file below the [spec root], compared byte by byte, then by where they stand in that file
  And     `show` goes on to write them
Otherwise `show` writes to "stdout" an `element_view` without [elements], where it was run with `show.json`
  And     `show` writes to "stderr" one line saying that <query> names no [element], within the capability `show.in` names where it names one, and nothing to "stdout", where it was run without `show.json`
  And     `show` exits with `show.exit-status` 1
```

| query | elements | provided | outcome | why |
| --- | --- | --- | --- | --- |
| `lint.run.exit` | the tree | yes | `lint.run.exit` | |
| `spec tree` | the tree | yes | `system.spec-tree` | a term by its name |
| `tree` | the tree | yes | `system.spec-tree` | a term by its alias |
| `Spec  Tree`, with two spaces | the tree | yes | `system.spec-tree` | compared as `name_comparison` says |
| `finding` | the tree | yes | `coherence.finding`, then `lint.finding` | one name in two capabilities, in the tree's order |
| `finding` | lint's and the system's | yes | `lint.finding` | |
| `diff.since` | the tree | yes | `system.cli.diff.since` | a member as a rule refers to it |
| `tree_layout` | the tree | yes | `system.tree-layout`, a parameter of the system | |
| `lint.run.twice`, the identifier of two rules in the middle of an amendment | the tree | yes | both | |
| `lint.run.odd`, a level-2 heading with an anchor in a feature file, its text starting neither `Rule:` nor `Invariant:` | the tree | yes | `lint.run.odd`, with no element type | |
| `system` | the tree | yes | `system`, with no element parent | |
| `since` | the tree | no | exit status 1 | a member is qualified |
| `Lint.Run.Exit` | the tree | no | exit status 1 | an identifier is matched character by character |
| `Tree_Layout` | the tree | no | exit status 1 | so is a parameter's name |
| `spec` | the tree | no | one line on stderr, nothing on stdout, exit status 1 | no part of a name |
| `Rule: Lint finds the nearest spec root` | the tree | no | exit status 1 | a heading is shown, never matched |
| `baseline.diff.since` | lint's and the system's | no | exit status 1 | `--in` narrows an identifier too |
| `nosuch`, with `--json` | the tree | no | an `element_view` without elements, nothing on stderr, exit status 1 | an empty list, not an empty output |

## <a id="query.show.output"></a> Rule: Show writes text or JSON

```guarded
Given     `show` has taken the [elements] <elements>
When      `show` writes them, run with the options <options>
Provided  <options> include `show.json`
Then      `show` writes the `element_view` of <elements> to "stdout"
  And     `show` exits with `show.exit-status` 0
Otherwise `show` writes <elements> to "stdout" as `show_text`
  And     `show` exits with `show.exit-status` 0
```

| elements | options | provided | outcome | why |
| --- | --- | --- | --- | --- |
| `lint.run.exit` | `--json` | yes | an `element_view` with one element, exit status 0 | |
| `system` | `--json` | yes | an `element_view` whose element has a null `parent`, exit status 0 | |
| `lint.run.exit` | none | no | `show_text` with one element, exit status 0 | |
| `coherence.finding`, `lint.finding` | none | no | `show_text` with two elements, exit status 0 | |
