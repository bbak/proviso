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

### <a id="system.cli.diff.since"></a> since

The option `--since <revision>`: compare the spec tree with the spec tree
at that git revision instead of with the baseline, and read former text
from there. `<revision>` is any name git accepts for a commit, such as
`a1b2c3d` or `HEAD~1`.

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
          enum: [added, changed, removed, changed-through]
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
        through:
          type: array
          items:
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
- `through` lists, for an element changed through a dependency, every
  chain of identifiers that reaches it, each from the first changed
  element to the one the element depends on directly, as
  `dependencies` gives them.
- `undecided` lists every undecided place in the element and every mark
  of an open term; an element with any is not to be built.

## <a id="system.cli.impact"></a> impact

The command `proviso impact`, run anywhere inside the system's
repository. It writes the impact candidates of the amendment as text by
default, or an `impact_set` with `--json`.

### <a id="system.cli.impact.json"></a> json

The option `--json`: write the `impact_set` to standard output instead
of text. The text form is for people and not part of the contract.

### <a id="system.cli.impact.exit-status"></a> exit-status

The process exit status.

```yaml
type: integer
enum: [0, 2]
```

- `0`: the candidates were found, whether there are any or none.
- `2`: the command could not run.

## <a id="system.cli.impact-set"></a> impact_set

What `impact --json` writes: the spec root that was compared, the
revision the former text came from, every candidate with every reason
it is one, and the candidates found through scope, grouped by what framed
them. Paths, `root` and `revision` follow the same conventions as in
`element_diff`.

```yaml
type: object
required: [root, revision, candidates, scopes]
properties:
  root:
    type: string
  revision:
    type: [string, "null"]
  candidates:
    type: array
    items:
      type: object
      required: [id, file, line, reasons]
      properties:
        id:
          type: string
        file:
          type: string
        line:
          type: integer
          minimum: 1
        reasons:
          type: array
          items:
            type: object
            required: [source, via, from]
            properties:
              source:
                enum: [events, ranges, scope]
              via:
                type: string
              from:
                type: array
                items:
                  type: string
  scopes:
    type: array
    items:
      type: object
      required: [id, candidates]
      properties:
        id:
          type: string
        candidates:
          type: array
          items:
            type: string
```

- `file` and `line` say where the candidate stands, at the line of its
  heading.
- `via` names what the candidate shares with the touched elements: the
  event head for `events`, `feature`, `capability` or `system` for
  `ranges`, and the identifier of the framing element for `scope`.
  Reasons with the same source and `via` are one reason, whose `from`
  lists every touched element behind it.
- `from` lists the identifiers of the touched elements the reason comes
  from, in byte order.
- `scopes` lists, for each touched element that opened `scope`, the
  identifiers of the candidates found through it, in byte order, so that
  they can be judged in bulk.

## <a id="system.cli.pack"></a> pack

The command `proviso pack`, run anywhere inside the system's repository.
It writes the context pack of the amendment as `pack_text` by default,
or a `context_pack` with `--json`.

### <a id="system.cli.pack.json"></a> json

The option `--json`: write the `context_pack` to standard output instead
of `pack_text`.

### <a id="system.cli.pack.since"></a> since

The option `--since <revision>`: assemble the pack of what changed since
that git revision, as `diff.since` compares, instead of since the
baseline. A resumed implementing agent receives it for a later round.

### <a id="system.cli.pack.exit-status"></a> exit-status

The process exit status.

```yaml
type: integer
enum: [0, 2]
```

- `0`: the pack was written, whether the element diff lists any element
  or none.
- `2`: the command could not run.

## <a id="system.cli.pack-text"></a> pack_text

What `pack` writes by default: Markdown in UTF-8, read by the
implementing agents. Unlike the text of the other commands, it is part
of the contract.

- **Four headings** of level 1 divide it, in this order: `# System`,
  `# Amendment`, `# Frames` and `# Context`. A heading under which no
  element appears is left out.
- **An element** appears as its own text as it stands now, from its
  heading to the next element's heading, as `text` in `element_diff`
  holds it, with a blank line before and after.
- **The revision** the former text came from is named in the first
  line under Amendment, ``Former text from `<revision>`.``, with its
  full name. A pack that needed no former text has no such line.
- **Under Amendment,** a line saying what happened to it comes before
  each element:
  - ``Added `<id>`.``
  - ``Changed `<id>`:``, then each part named for it, as `part_kinds`
    writes it, in backticks, followed by `added`, `changed` or
    `removed`, separated by a comma and a space, and a full stop.
  - ``Removed `<id>`.``; its former text, in a fenced block like all
    former text, stands in place of its text.
  - ``Changed through `<id>`:``, then each chain that reaches it, as
    identifiers in backticks joined by ` → `, separated by `; `, and a
    full stop.
  - An element not to be built gets a second line: `Not to be built:`,
    then each undecided place as its path and line joined by `:`,
    separated by a comma and a space, and a full stop.
- **The former text** of each changed or removed part follows its
  element, each after a line ``Former `<part>`:``.
- **Former text** stands in a fenced block labelled `text`, its fence
  longer than any run of backticks in it, so that nothing in it reads
  as Markdown.
- Paths follow the same conventions as in `report`.

~~~~text
# System

# <a id="system"></a> Proviso
…

# Amendment

Former text from `1edea75c…`.

Changed `lint.run.exit`: `step Then` changed.

## <a id="lint.run.exit"></a> Rule: The exit status tells whether errors were found
…

Former `step Then`:

```text
Then      lint exits with `lint.exit-status` 0
```

# Frames
…
~~~~

## <a id="system.cli.context-pack"></a> context_pack

What `pack --json` writes: the spec root the pack was assembled from,
the revision the former text came from, the element diff, and every
other element of the pack. Paths, `root` and `revision` follow the same
conventions as in `element_diff`.

```yaml
type: object
required: [root, revision, diff, elements]
properties:
  root:
    type: string
  revision:
    type: [string, "null"]
  diff:
    type: array
  elements:
    type: array
    items:
      type: object
      required: [id, heading, file, line, text]
      properties:
        id:
          type: string
        heading:
          enum: [system, frames, context]
        file:
          type: string
        line:
          type: integer
          minimum: 1
        text:
          type: string
```

- `diff` holds the elements under Amendment, each exactly as an item of
  `elements` in `element_diff`.
- `elements` holds every other element of the pack, under the heading
  `heading` names, in the order of `pack_text`.
- `file`, `line` and `text` mean what they mean in `element_diff`.

## <a id="system.cli.init"></a> init

The command `proviso init`, run in the root of a git repository. It
writes the installed files that are missing, and names each installed
file on standard output: one line each, in the byte order of their
paths, `written` or `skipped`, a space and the path, as
`installed_files` gives it.

```text
written .claude/agents/proviso-coder.md
skipped proviso/system.md
```

### <a id="system.cli.init.exit-status"></a> exit-status

The process exit status.

```yaml
type: integer
enum: [0, 2]
```

- `0`: every installed file is in place, written or skipped.
- `2`: the command could not run, or an installed file could not be
  written.
