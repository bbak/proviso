# <a id="lint.layout"></a> Layout

- Contact: Bruno Baketarić

## Goal

Every file of the spec tree stands where the layout expects it, and the
files the layout requires exist and hold something.

## Rationale

Tooling finds the tree's content by where it stands. A file in an
unexpected place is never read, a missing `capability.md` leaves a
capability without its Goal, and an empty file says there is something
where there is nothing. Each of these is silent unless lint reports it.

## <a id="lint.layout.hidden"></a> Rule: Hidden files are not part of the tree

### Rationale

Operating systems and editors leave files such as `.DS_Store` behind. A
name starting with a dot never holds spec content, because the layout
lists only `.md` files, so skipping such files loses nothing. The one
exception is the tree's own markdownlint settings.

```guarded
When      lint finds the file or folder <entry> in the [spec tree]
Provided  the name of <entry> starts with a dot
  And     <entry> is not `.markdownlint.jsonc` in the [spec root]
Then      lint skips <entry> and everything in it
  And     lint reports nothing about it
Otherwise → lint.layout.file
```

| entry | provided | outcome | why |
| --- | --- | --- | --- |
| `.DS_Store` | yes | skipped | |
| `capabilities/lint/.vscode/` | yes | skipped, with its content | a hidden folder |
| `.markdownlint.jsonc`, in the spec root | no (is the tree's settings) | handed over | the tree's own settings |
| `capabilities/lint/.markdownlint.jsonc` | yes | skipped | only the spec root's settings belong to the tree |
| `glossary.md` | no (no dot) | handed over | |

## <a id="lint.layout.file"></a> Rule: Every file belongs to the layout

```guarded
When      lint finds the file <file> in the [spec tree]
Provided  `tree_layout` lists <file>
Then      lint reads <file>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.file-unknown` at <file>
  And     lint goes on with the rest of the [spec tree]
```

| file | provided | outcome | why |
| --- | --- | --- | --- |
| `capabilities/lint/features/run.md` | yes | read | |
| `.markdownlint.jsonc` | yes | read | |
| `notes.md` | no | error | not part of the layout |
| `capabilities/lint/README.md` | no | error | not part of the layout |
| `capabilities/lint/features/old/run.md` | no | error | features are not nested |
| `invariants/read-only.md` | no | error | invariants are one file per level |

## <a id="lint.layout.capability"></a> Rule: A capability has its own file

```guarded
When      lint finds the capability folder <folder>
Provided  <folder> holds `capability.md`
Then      lint reads the capability from it
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.capability-file-missing` at <folder>
  And     lint goes on with the rest of the [spec tree]
```

| folder | provided | outcome | why |
| --- | --- | --- | --- |
| `capabilities/lint/`, holding `capability.md` | yes | read | |
| `capabilities/lint/`, holding only `glossary.md` | no | error | a capability's Goal is required |

## <a id="lint.layout.title"></a> Rule: A sibling file opens with a title

```guarded
When      lint reads the first heading <heading> of a file of a [sibling]
Provided  <heading> is at level 1
  And     <heading> has no [anchor]
Then      <heading> is the file's title, neither [element] nor [section]
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.title-missing` at <heading>
  And     lint goes on with the rest of the [spec tree]
```

| heading | provided | outcome | why |
| --- | --- | --- | --- |
| `# Glossary` | yes | title | |
| `## Glossary` | no (level 1) | error | |
| `# <a id="lint.glossary"></a> Glossary` | no (no anchor) | error | a title is not an element |
| `## <a id="lint.finding"></a> finding`, as the first heading | no (level 1) | error | the title is missing |

## <a id="lint.layout.empty"></a> Rule: A sibling file is not empty

### Rationale

An optional sibling exists once it has content. An empty file says
there is something where there is nothing.

```guarded
When      lint has read a file <file> of a [sibling]
Provided  <file> holds more than its title
Then      lint reports nothing about <file> being empty
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.sibling-empty` at <file>
  And     lint goes on with the rest of the [spec tree]
```

| file | provided | outcome | why |
| --- | --- | --- | --- |
| `glossary.md` with one entry | yes | nothing reported | |
| `glossary.md` with only `# Glossary` | no | error | |
| `parameters.md` with no content at all | no | error | |
