# <a id="baseline.run"></a> Running the baseline commands

- Contact: Bruno Baketarić

## Goal

The baseline commands work only on a tree that lint accepts, find it the
way lint does, and change nothing but the manifest.

## Rationale

A misread tree gives wrong fingerprints without any sign of it, so
nothing is derived from a tree with an error. Running lint's checks
first, rather than repeating its rules, means the commands find the spec
root, and refuse an unreadable tree, exactly as lint does. That they
change nothing but the manifest is the capability's invariant. An ask does
not stop them: an undecided place is what the element diff lists as not
to be built.

## <a id="baseline.run.lint"></a> Rule: The baseline commands run only where lint finds no error

```guarded
When      someone runs `manifest`, `diff`, `impact` or `pack` in the directory <directory>
Provided  `lint`, run in <directory>, would exit with `lint.exit-status` 0
Then      the command works on the [spec tree] that `lint` would check
  And     nothing of what `lint` would write is written
Otherwise the command writes to "stderr" what `lint` would write as text
  And     the command exits with 2, as `manifest.exit-status`, `diff.exit-status`, `impact.exit-status` and `pack.exit-status` define
  And     nothing goes to "stdout"
```

| directory | provided | outcome | why |
| --- | --- | --- | --- |
| `repo/`, its tree without findings | yes | the command runs | |
| `repo/`, its tree with an ask only | yes | the command runs | |
| `repo/`, its tree with an error | no | exit status 2, the error on stderr | |
| `repo/`, its tree with an unreadable file | no | exit status 2 | lint could not run |
| `other/`, with no spec root up to the file system's root | no | exit status 2 | |
