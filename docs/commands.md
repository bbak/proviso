# Commands

Every command of `proviso`, with what it is for, its options and its
exit statuses. Engineers run `help`, `init`, `update`, `version`, `lint`,
`diff`, `impact` and `verify`; the skills run all of them, `pack`, `judge` and `manifest`
included.

Every command but `help`, `init`, `update` and `version` runs anywhere inside the
repository and works on the nearest spec root. The commands that compare with the baseline,
`manifest`, `diff`, `impact`, `pack`, `judge` and `verify`, run only on
a tree that `lint` finds no error in, and exit with 2 otherwise.

The exact behaviour of each command is specified in Proviso's own spec
tree: the [command line contract](https://github.com/bbak/proviso/blob/main/proviso/contracts/cli.md),
with every option, exit status and JSON format, and the feature linked
under each command below, with every rule and its examples.

## help

```sh
proviso help
```

Prints one page naming every command and option; `proviso --help` and
`proviso -h` do the same. `proviso` without a command prints the page
to standard error and exits with 2. Every command refuses an argument
it does not take, and `proviso` a command it does not know, with one
line on standard error that points to `proviso help`; it then exits
with 2, having checked and written nothing.

| Exit status | Meaning |
| --- | --- |
| 0 | the page was written |
| 2 | an argument was given, or `proviso` was run without a command or with an unknown one |

Specified in [cli/features/help.md](https://github.com/bbak/proviso/blob/main/proviso/capabilities/cli/features/help.md).

## init

```sh
proviso init
```

Writes the files a repository needs: the start of a spec tree and its
settings, the skills, the agent definitions, and stubs for the
engineers' notes. It never replaces an existing file, and says file by
file what it wrote and what it skipped, so a second run writes only
what is missing. It runs only in the root of a git repository.

| Exit status | Meaning |
| --- | --- |
| 0 | every file is in place |
| 2 | a file could not be written, or the directory is not a repository's root |

Specified in [install/features/init.md](https://github.com/bbak/proviso/blob/main/proviso/capabilities/install/features/init.md).

## update

```sh
proviso update
```

Brings the skills and agent definitions to the version of the binary,
after an upgrade. Each skill and agent carries the version that wrote
it; `update` writes one that is missing, replaces one written by an
older version, by a local build or with an unreadable stamp, and leaves
one of its own version or a newer one alone, warning about a newer one.
It says file by file what it wrote, updated and skipped. It touches
nothing else: the spec tree, its settings and the notes in `.proviso/`
stay as they are.

**A skill or agent you changed by hand is replaced without a question.**
Copy it first if you want to keep your changes; your own notes belong in
`.proviso/`, which `update` never touches.

`update` runs only with a release, only in the root of a git
repository, and only where `init` has installed something.

| Exit status | Meaning |
| --- | --- |
| 0 | every skill and agent is in place |
| 2 | a file could not be written, the binary is no release, the directory is not a repository's root, or nothing is installed |

Specified in [install/features/update.md](https://github.com/bbak/proviso/blob/main/proviso/capabilities/install/features/update.md).

## version

```sh
proviso version
```

Prints two lines: Proviso's own version, which follows semantic
versioning, and the version of the Proviso Model it reads, a whole
number.

```text
proviso 0.1.0
model 0
```

A local build through the Taskfile names the next patch after the
latest release, such as `0.1.1-dev`; a build given no version says
`dev`.

Run in the root of a git repository, it also checks what `init`
installed there, and warns on standard error about every skill or agent
that is missing, carries no version stamp or more than one, or was
installed by another version, naming the model version that one reads
where it can tell; and about a `proviso/manifest.sum` that is missing,
cannot be read, or records another model version or none. Silence means
everything fits. Anywhere else it prints a hint instead. It never
changes a file.

| Exit status | Meaning |
| --- | --- |
| 0 | the version was written, warnings or not |
| 2 | an argument was given |

Specified in [install/features/version.md](https://github.com/bbak/proviso/blob/main/proviso/capabilities/install/features/version.md).

## lint

```sh
proviso lint [--json]
```

Checks the spec tree against the notation: layout, identifiers, the
fields and sections of each element, rule blocks, Examples tables and
glossary entries. It reports errors, and asks about every undecided
place. It never writes the tree.

| Exit status | Meaning |
| --- | --- |
| 0 | no error |
| 1 | at least one error |
| 2 | lint could not run: no spec root was found, a file of the tree cannot be read (a folder where the layout lists a file counts), or the manifest records another model version or none |

Specified in the features of [lint](https://github.com/bbak/proviso/tree/main/proviso/capabilities/lint/features).

## diff

```sh
proviso diff [--json] [--since <revision>]
```

Writes the element diff: what the amendment added, changed, removed, or
changed through a dependency, compared with the baseline, with the
former text of what changed. It lists the elements not to be built yet,
because they hold an undecided place. `--since <revision>` compares with
the tree at that git revision instead.

| Exit status | Meaning |
| --- | --- |
| 0 | the tree is as the manifest records it |
| 1 | the element diff lists at least one element |
| 2 | the command could not run |

Specified in [baseline/features/diff.md](https://github.com/bbak/proviso/blob/main/proviso/capabilities/baseline/features/diff.md).

## impact

```sh
proviso impact [--json]
```

Writes the candidates of the amendment's impact set: the existing rules
and invariants it may affect without having edited them, each with the
reasons it is one.

| Exit status | Meaning |
| --- | --- |
| 0 | the candidates were found, whether there are any or none |
| 2 | the command could not run |

Specified in [baseline/features/impact.md](https://github.com/bbak/proviso/blob/main/proviso/capabilities/baseline/features/impact.md).

## pack

```sh
proviso pack [--json] [--since <revision>]
```

Writes the context pack of the amendment: the system, the element diff,
what the amendment note says the amendment leaves out, and every
element needed to read it. `--since <revision>` writes the pack of what
changed since that revision, for an agent's later round. A note that is
not a readable file in UTF-8 stops it, rather than drop out of the pack.

| Exit status | Meaning |
| --- | --- |
| 0 | the pack was written |
| 2 | the amendment note could not be read, or the command could not run |

Specified in [baseline/features/pack.md](https://github.com/bbak/proviso/blob/main/proviso/capabilities/baseline/features/pack.md).

## judge

```sh
proviso judge < result.json
```

Records a check result, read from standard input, in
`proviso/.coherence.json`. A result is either a check, the checking
agent's judgements and new findings, or the engineer's decisions on
recorded findings, never both. Proviso numbers the findings and adds
the fingerprints of what the check covered. A damaged coherence file is
reported before anything wrong with the input.

| Exit status | Meaning |
| --- | --- |
| 0 | the result was recorded |
| 1 | the result was refused, and nothing was recorded |
| 2 | the coherence file could not be read or written, or the command could not run |

Specified in [baseline/features/judge.md](https://github.com/bbak/proviso/blob/main/proviso/capabilities/baseline/features/judge.md).

## verify

```sh
proviso verify [--json]
```

Tells whether the tree may become the baseline as it stands: either it
is what the manifest records and no coherence file or amendment note is
left over, or the coherence check covers the amendment's final state;
and in either case, no element of the tree is left undecided. It lists
each problem: a check `missing`, a coherence file or note `leftover`, an
element `undecided`, a candidate `unjudged`, an element `stale`, a
finding `open`. A coherence file it cannot read stops it before anything is
reported.

| Exit status | Meaning |
| --- | --- |
| 0 | the tree may become the baseline |
| 1 | at least one problem |
| 2 | the coherence file could not be read, or the command could not run |

Specified in [baseline/features/verify.md](https://github.com/bbak/proviso/blob/main/proviso/capabilities/baseline/features/verify.md).

## manifest

```sh
proviso manifest
```

Records the tree as the baseline in `proviso/manifest.sum`, and removes
the coherence file and the amendment note, whatever stands under their
names. Acceptance runs it; running it by hand records a
tree nobody checked.

| Exit status | Meaning |
| --- | --- |
| 0 | the manifest was written, and the coherence file and the note removed |
| 2 | the manifest could not be written, the coherence file or the note could not be removed, or the command could not run |

If a removal fails, the written manifest stays, and `verify` reports
what is left over.

Specified in [baseline/features/manifest.md](https://github.com/bbak/proviso/blob/main/proviso/capabilities/baseline/features/manifest.md).
