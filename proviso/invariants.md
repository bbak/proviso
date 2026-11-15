# Invariants

## <a id="system.lint-gate"></a> The tree commands run only where lint finds no error

### Rationale

A misread tree gives wrong fingerprints and candidates without any sign
of it, so nothing is derived from a tree with an error, nor rewritten.
Running lint's checks first, rather than repeating its rules, means the
commands find the spec root, and refuse an unreadable tree or one of a
model version they do not read, exactly as lint does, also when they
compare with a revision. An ask does not stop them: an undecided place
is what the element diff lists as not to be built, and an identifier
that names no place is what `rehome` repairs.

### Ranges over

Every run of `manifest`, `diff`, `pack`, `impact`, `judge`, `verify` and
`rehome`, with arguments it takes, in a directory, with or without
`diff.since` or `pack.since`.

### Property

Where `lint`, run in the same directory, would exit with
`lint.exit-status` 0, the run works on the [spec tree] `lint` would
check, and nothing of what `lint` would write is written. Otherwise the
run writes to "stderr" what `lint` would write as text, exits with 2 as
its command's `exit-status` defines, and writes nothing to "stdout".
