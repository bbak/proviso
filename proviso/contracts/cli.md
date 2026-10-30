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
repository; it takes no argument. It writes `manifest.sum` into the spec root, removes the
coherence file and the amendment note there, and writes nothing to
standard output.

### <a id="system.cli.manifest.exit-status"></a> exit-status

The process exit status.

```yaml
type: integer
enum: [0, 2]
```

- `0`: the manifest was written, and the coherence file and the
  amendment note were removed.
- `2`: the manifest could not be written, the coherence file or the
  amendment note could not be removed, or the command could not run.

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
implementing agents and the checking agent. Unlike the text of the other commands, it is part
of the contract.

- **Five headings** of level 1 divide it, in this order: `# System`,
  `# Amendment`, `# Not in this amendment`, `# Frames` and `# Context`.
  A heading under which nothing appears is left out.
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
- **Under Not in this amendment,** the text of `amendment_note` stands
  as the file holds it, with a blank line before and after.
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
required: [root, revision, diff, note, elements]
properties:
  root:
    type: string
  revision:
    type: [string, "null"]
  diff:
    type: array
  note:
    type: [string, "null"]
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
- `note` is the text `pack_text` holds under Not in this amendment, or
  null where it has no such heading.
- `elements` holds every other element of the pack, under the heading
  `heading` names, in the order of `pack_text`.
- `file`, `line` and `text` mean what they mean in `element_diff`.

## <a id="system.cli.judge"></a> judge

The command `proviso judge`, run anywhere inside the system's
repository; it takes no argument. It reads a `check_result` from standard input, records it
in `coherence_file` and writes nothing to standard output.

### <a id="system.cli.judge.exit-status"></a> exit-status

The process exit status.

```yaml
type: integer
enum: [0, 1, 2]
```

- `0`: the check result was recorded.
- `1`: the check result was refused, and nothing was recorded.
- `2`: `coherence_file` could not be read or written, or the command
  could not run.

## <a id="system.cli.check-result"></a> check_result

What `judge` reads, in UTF-8: either a run of the coherence check, its
judgements and new findings, or the engineer's decisions on findings
recorded before, by their `id`; never both.

```yaml
type: object
properties:
  judgements:
    type: array
    items:
      type: object
      required: [reason, elements]
      properties:
        reason:
          type: string
          minLength: 1
        elements:
          type: array
          minItems: 1
          items:
            type: string
  findings:
    type: array
    items:
      oneOf:
        - type: object
          required: [kind, text]
          additionalProperties: false
          properties:
            kind:
              enum: [contradiction, ignored, gap, missed]
            elements:
              type: array
              items:
                type: string
            text:
              type: string
              minLength: 1
        - type: object
          required: [id, status]
          additionalProperties: false
          properties:
            id:
              type: string
            status:
              enum: [resolved, dismissed]
            reason:
              type: string
              minLength: 1
```

- A judgement says that every element it names stays as it is, for its
  `reason`. One reason may cover many elements; an element appears in
  at most one judgement.
- A finding without an `id` is new. `kind` says what was found: a
  `contradiction`, a term, invariant or constraint `ignored`, a `gap`,
  or a candidate the impact sources `missed`. `elements` names the
  elements it concerns, if any.
- A finding with an `id` changes the status of one recorded before:
  `resolved` when an edit removed it, `dismissed` when the engineer
  dismissed it, which needs a `reason`.
- A member left out is an empty list.

A run of the coherence check:

```json
{
  "judgements": [
    { "reason": "untouched: nothing here reads the manifest",
      "elements": ["lint.examples.cells", "lint.run.root"] }
  ],
  "findings": [
    { "kind": "gap", "elements": ["baseline.impact.events"],
      "text": "No rule says what happens when a When has no step text." }
  ]
}
```

A decision of the engineer:

```json
{
  "findings": [
    { "id": "f3", "status": "dismissed", "reason": "decided in system.git" }
  ]
}
```

## <a id="system.cli.verify"></a> verify

The command `proviso verify`, run anywhere inside the system's
repository. It writes the problems it finds as text by default, or a
`verification` with `--json`. CI and the local checks run it on every
branch.

### <a id="system.cli.verify.json"></a> json

The option `--json`: write the `verification` to standard output
instead of text. The text form is for people and not part of the
contract.

### <a id="system.cli.verify.exit-status"></a> exit-status

The process exit status.

```yaml
type: integer
enum: [0, 1, 2]
```

- `0`: the spec tree may become the baseline as it stands.
- `1`: `verify` found at least one problem.
- `2`: `coherence_file` could not be read, or the command could not
  run.

## <a id="system.cli.verification"></a> verification

What `verify --json` writes: the spec root that was verified, and every
problem found. `root` follows the same conventions as in `report`.

```yaml
type: object
required: [root, problems]
properties:
  root:
    type: string
  problems:
    type: array
    items:
      type: object
      required: [kind, message]
      properties:
        kind:
          enum: [missing, leftover, undecided, unjudged, stale, open]
        element:
          type: string
        finding:
          type: string
        text:
          type: string
        message:
          type: string
```

- `kind` says what is wrong: the element diff lists elements but no
  coherence check is recorded (`missing`); a check or an amendment note
  is left beside a tree the manifest records (`leftover`); an element
  holds an undecided place or marks an open term (`undecided`); a
  candidate has no judgement (`unjudged`); an element changed since the
  check covered it (`stale`); a finding is neither resolved nor
  dismissed (`open`).
- `element` is the identifier of the element an `undecided`,
  `unjudged` or `stale` problem is about; `finding` and `text` are the
  `id` and the text of an `open` finding.
- `message` is for people and not part of the contract.

## <a id="system.cli.init"></a> init

The command `proviso init`, run in the root of a git repository; it
takes no argument. It writes the installed files that are missing, and names each installed
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

## <a id="install.cli.update"></a> update

The command `proviso update`, run in the root of a git repository; it
takes no argument. It brings the skills and agent definitions that
`installed_files` names to the binary's version, and names each on
standard output: one line each, in the byte order of their paths,
`written`, `updated` or `skipped`, a space and the path, as
`installed_files` gives it.

```text
updated .claude/agents/proviso-coder.md
skipped .claude/skills/proviso-accept/SKILL.md
```

### <a id="install.cli.update.exit-status"></a> exit-status

The process exit status.

```yaml
type: integer
enum: [0, 2]
```

- `0`: every skill and agent definition is in place, written, updated
  or skipped.
- `2`: the binary is no release, the directory is not the root of a
  git repository, no skill or agent definition is installed, an
  argument was given, or a file could not be written.

## <a id="system.cli.version"></a> version

The command `proviso version`, run anywhere; it takes no argument. On
standard output it writes two lines, the Proviso version and the model
version:

```text
proviso 0.1.0
model 0
```

With an argument, it refuses it as every command refuses an argument it
does not take, and checks nothing. Without one, it writes to
standard error, outside the root of a git repository, one line of hint.
In the root it writes, first, either one line saying that
no skill or agent is installed, or one line for each skill or agent
that `installed_files` names and that is missing, holds more than one
stamp, names no Proviso version or was installed by another version, in
the byte order of their paths; then at most one line about the spec tree's manifest. In the
root, nothing on standard error means everything fits. The wording of
these lines is for people and not part of the contract.

### <a id="system.cli.version.exit-status"></a> exit-status

The process exit status.

```yaml
type: integer
enum: [0, 2]
```

- `0`: the version was written, whatever went to standard error.
- `2`: an argument was given.

## <a id="cli.cli.help"></a> help

The command `proviso help`, also `proviso --help` and `proviso -h`, run
anywhere; it takes no argument. It writes `page` to standard output.

### <a id="cli.cli.help.page"></a> page

The help page: what Proviso is, in one line; every command with its
options and one sentence on what it does, first those engineers run,
`init`, `update`, `version`, `lint`, `diff`, `impact`, `verify` and
`help`, then those the skills run, `pack`, `judge` and `manifest`; the
exit statuses the commands share;
and where the documentation is. It names every command and every
option. Its wording is for people and not part of the contract.

### <a id="cli.cli.help.exit-status"></a> exit-status

The process exit status of `help`, and of `proviso` run without a
command or with one it does not know.

```yaml
type: integer
enum: [0, 2]
```

- `0`: the page was written to standard output.
- `2`: an argument was given to `help`, or `proviso` was run without a
  command or with an unknown one.
