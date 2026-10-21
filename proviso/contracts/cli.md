# Command line

## <a id="lint.cli.lint"></a> lint

The command `proviso lint`, run anywhere inside the system's repository.
It writes findings as text by default, or a `report` with `--json`.

### <a id="lint.cli.lint.json"></a> json

The option `--json`: write the `report` to standard output instead of
text. The text form is for people and not part of the contract.

### <a id="lint.cli.lint.exit-status"></a> exit-status

The process exit status.

```yaml
type: integer
enum: [0, 1, 2]
```

- `0`: lint ran and reported no error.
- `1`: lint ran and reported at least one error.
- `2`: lint could not run.

## <a id="lint.cli.report"></a> report

What `lint --json` writes: the spec root that was linted, and every
finding. The paths of findings are relative to the directory holding
the spec root, so they read the same from wherever lint was run. Every
path uses `/` as its separator, and every name in it is in Unicode's
composed form, NFC, on every platform, so that the order of findings is
the same everywhere.

```yaml
type: object
required: [root, findings]
properties:
  root:
    type: string
  findings:
    type: array
    items:
      type: object
      required: [kind, severity, file, line, message]
      properties:
        kind:
          type: string
        severity:
          enum: [error, ask]
        file:
          type: string
        line:
          type: integer
          minimum: 1
        message:
          type: string
        related:
          type: array
          items:
            type: object
            required: [file, line]
            properties:
              file:
                type: string
              line:
                type: integer
                minimum: 1
```

- `root` is the absolute path of the spec root that was linted, with
  `/` as well and, on Windows, its drive first, as an uppercase letter,
  as in `C:/repo/proviso`, and every name in NFC, so it shows which tree
  was checked when several are nested.
- `kind` names one of the members of `findings`.
- `message` is for people and not part of the contract: tests check
  `kind`, never the wording.
- `line` is the line the finding stands on, counting lines by line
  feeds only. A finding about a whole file or folder has line 1, and a
  folder's path has no trailing `/`.
- `related` lists the other places the finding concerns. Wherever a rule
  says a finding names where else something is used, or names another
  element, those places are listed here, an element at the line of its
  heading.

## <a id="system.cli.manifest"></a> manifest

The command `proviso manifest`, run anywhere inside the system's
repository. It writes `manifest.sum` into the spec root and nothing to
standard output.

### <a id="system.cli.manifest.exit-status"></a> exit-status

The process exit status.

```yaml
type: integer
enum: [0, 2]
```

- `0`: the manifest was written.
- `2`: the manifest could not be written, or the command could not run.

## <a id="system.cli.diff"></a> diff

The command `proviso diff`, run anywhere inside the system's repository.
It writes the element diff as text by default, or an `element_diff` with
`--json`.

### <a id="system.cli.diff.json"></a> json

The option `--json`: write the `element_diff` to standard output instead
of text. The text form is for people and not part of the contract.

### <a id="system.cli.diff.exit-status"></a> exit-status

The process exit status.

```yaml
type: integer
enum: [0, 1, 2]
```

- `0`: the spec tree is as the manifest records it.
- `1`: the element diff lists at least one element.
- `2`: the command could not run.

## <a id="system.cli.element-diff"></a> element_diff

What `diff --json` writes: the spec root that was compared, the revision
the former text came from, and every element listed. Paths, `root` and
the order of elements follow the same conventions as in `report`.

```yaml
type: object
required: [root, revision, elements]
properties:
  root:
    type: string
  revision:
    type: [string, "null"]
  elements:
    type: array
    items:
      type: object
      required: [id, change, undecided]
      properties:
        id:
          type: string
        change:
          enum: [added, changed, removed, changed-by-term]
        file:
          type: string
        line:
          type: integer
          minimum: 1
        text:
          type: string
        former:
          type: string
        parts:
          type: array
          items:
            type: object
            required: [part, change]
            properties:
              part:
                type: string
              change:
                enum: [added, changed, removed]
              former:
                type: string
        terms:
          type: array
          items:
            type: string
        undecided:
          type: array
          items:
            type: object
            required: [file, line]
            properties:
              file:
                type: string
              line:
                type: integer
                minimum: 1
```

- `revision` is the full name of the git revision the former text came
  from, or null when no former text was needed.
- `file` and `line` say where the element stands now, at the line of its
  heading; a removed element has neither.
- `text` is the element's own text as it stands now, from its heading to
  the next element's heading; a removed element has none. `former` is
  its former text, for a removed element only.
- `parts` lists, for a changed element, each part that was added,
  changed or removed, written as `part_kinds` writes it, with the former
  text of a changed or removed part.
- `terms` lists the identifiers of the glossary entries through which a
  rule is changed by a term.
- `undecided` lists every undecided place in the element and every mark
  of an open term; an element with any is not to be built.
