# Commands

Every command of `proviso`, with what it is for, its options and its
exit statuses. Engineers run `init`, `lint`, `diff`, `impact` and
`verify`; the skills run all of them, `pack`, `judge` and `manifest`
included.

Every command but `init` runs anywhere inside the repository and works
on the nearest spec root. The commands that compare with the baseline,
`manifest`, `diff`, `impact`, `pack`, `judge` and `verify`, run only on
a tree that `lint` finds no error in, and exit with 2 otherwise.

The exact behaviour of each command is specified in Proviso's own spec
tree: the [command line contract](https://github.com/bbak/proviso/blob/main/proviso/contracts/cli.md),
with every option, exit status and JSON format, and the feature linked
under each command below, with every rule and its examples.

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
| 2 | lint could not run, such as for a file it cannot read |

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
and every element needed to read it. `--since <revision>` writes the
pack of what changed since that revision, for an agent's later round.

| Exit status | Meaning |
| --- | --- |
| 0 | the pack was written |
| 2 | the command could not run |

Specified in [baseline/features/pack.md](https://github.com/bbak/proviso/blob/main/proviso/capabilities/baseline/features/pack.md).

## judge

```sh
proviso judge < result.json
```

Records a check result, read from standard input, in
`proviso/.coherence.json`. A result is either a check, the checking
agent's judgements and new findings, or the engineer's decisions on
recorded findings, never both. Proviso numbers the findings and adds
the fingerprints of what the check covered.

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
is what the manifest records and no coherence file is left over, or the
coherence check covers the amendment's final state. It lists each
problem: a check `missing` or `leftover`, a candidate `unjudged`, an
element `stale`, a finding `open`.

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
the coherence file. Acceptance runs it; running it by hand records a
tree nobody checked.

| Exit status | Meaning |
| --- | --- |
| 0 | the manifest was written |
| 2 | the manifest could not be written, or the command could not run |

Specified in [baseline/features/manifest.md](https://github.com/bbak/proviso/blob/main/proviso/capabilities/baseline/features/manifest.md).
