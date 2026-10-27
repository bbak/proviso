# <a id="lint.layout"></a> Layout

- Contact: Bruno Baketarić

## Goal

Every file of the spec tree stands where the layout expects it, under a
name and with line endings that read the same on every platform, and
the files the layout requires exist and hold something.

## Rationale

Tooling finds the tree's content by where it stands. A file in an
unexpected place is never read, a missing `capability.md` leaves a
capability without its Goal, and an empty file says there is something
where there is nothing. Each of these is silent unless lint reports it.
The same holds for a name or a line ending that one platform reads
differently from another.

## <a id="lint.layout.hidden"></a> Rule: Hidden files are not part of the tree

### Rationale

Operating systems and editors leave files such as `.DS_Store` behind. A
name starting with a dot never holds spec content, because the layout
lists only `.md` files, so skipping such files loses nothing. The one
exception is the tree's own markdownlint settings.

Hidden means a name starting with a dot, on every platform. A file
system's own mark, such as the hidden attribute of Windows, does not
travel with the tree through version control: a file hidden on one
machine would be found on the next, and lint would pass on one and fail
on the other.

```guarded
When      lint finds the file or folder <entry> in the [spec tree]
Provided  the name of <entry> starts with a dot
  And     <entry> is not ".markdownlint.jsonc" in the [spec root]
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
| `notes.md`, marked hidden by Windows | no (no dot) | handed over | the mark does not travel with the tree |

## <a id="lint.layout.name"></a> Rule: Names read the same on every platform

### Rationale

Windows and macOS do not tell `Glossary.md` from `glossary.md` by
default, Linux does. A name with an uppercase letter would be one file
on one machine and two on the next, so names are lowercase, as the names
the layout gives already are. A letter counts as lowercase when it has
no other lowercase form, so the letters of a script without case count
too. Windows refuses characters such as `:` or `?` and a name ending in
a dot, and a space would have to be wrapped in every Markdown link, so
a name holds lowercase letters with their marks, digits, hyphens and
dots, and nothing else. Digits, like letters, may come from any script.

A letter such as `ü` can be stored as one character or as `u` followed
by a mark, and file systems differ in which form they keep. Lint reads
every name in the composed form, Unicode's NFC, so the same tree gives
the same paths on every platform. Two names in one folder that are the
same in that form would be one file on macOS and two on Linux, as with
case. They share one path in what lint reports, so lint reports them
once: a second finding at the same place would only look like a
duplicate.

```guarded
When      lint finds the file or folder <entry> in the [spec tree], not skipped as hidden
Provided  the name of <entry> holds only lowercase letters, marks, digits, hyphens and dots
  And     the name of <entry> does not end with a dot
  And     no other entry in the folder of <entry> has the same name in NFC
Then      lint reports nothing about the name of <entry>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.name-malformed` at <entry>, once for all entries whose names are the same in NFC
  And     lint goes on with the rest of the [spec tree]
```

| entry | provided | outcome | why |
| --- | --- | --- | --- |
| `capabilities/lint/features/run.md` | yes | nothing reported | |
| `capabilities/lint/features/übersicht.md` | yes | nothing reported | a letter outside ASCII |
| `capabilities/lint/features/übersicht.md`, stored as `u` followed by a mark | yes | nothing reported, and its path is reported in the composed form | read as NFC |
| `capabilities/lint/features/検索.md` | yes | nothing reported | a script without case |
| `.markdownlint.jsonc`, in the spec root | yes | nothing reported | dots are allowed |
| `capabilities/lint/features/Run.md` | no (only lowercase letters, marks, digits, hyphens and dots) | error | |
| `capabilities/Lint`, a folder | no (only lowercase letters, marks, digits, hyphens and dots) | error | a folder, too |
| `capabilities/lint/features/Übersicht.md` | no (only lowercase letters, marks, digits, hyphens and dots) | error | an uppercase letter outside ASCII |
| `capabilities/lint/features/my run.md` | no (only lowercase letters, marks, digits, hyphens and dots) | error | a space |
| `capabilities/lint/features/run_old.md` | no (only lowercase letters, marks, digits, hyphens and dots) | error | an underscore |
| `capabilities/lint/features/run.md.` | no (does not end with a dot) | error | Windows drops a trailing dot |
| `capabilities/lint/features/übersicht.md`, once composed and once as `u` followed by a mark | no (no other entry has the same name in NFC) | error, once, at the name both share | one file on macOS, two on Linux |
| `capabilities/lint/features/run٢.md` | yes | nothing reported | a digit from another script |

## <a id="lint.layout.line-ending"></a> Rule: Lines end with a line feed only

### Rationale

Windows' line ending, a carriage return before the line feed, changes
every line of a file without changing what it says. The tree uses the
line feed alone, which every current editor keeps.

```guarded
When      lint reads the file <file> of the [spec tree]
Provided  <file> holds no carriage return
Then      lint reports nothing about the [line endings] of <file>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.line-ending` at the first line of <file> that holds a carriage return
  And     lint goes on with the rest of the [spec tree]
```

| file | provided | outcome | why |
| --- | --- | --- | --- |
| `glossary.md`, every line ending in a line feed | yes | nothing reported | |
| `glossary.md`, lines 3 and 7 ending in a carriage return and a line feed | no | error at line 3 | once per file, at the first |
| `.markdownlint.jsonc`, ending its lines in a carriage return and a line feed | no | error at line 1 | every file lint reads |
| `manifest.sum`, ending its lines in a carriage return and a line feed | no | error at line 1 | as checked out with `core.autocrlf` |

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
Provided  <folder> holds "capability.md"
Then      lint reads the capability from it
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.capability-file-missing` at <folder>
  And     lint goes on with the rest of the [spec tree]
```

| folder | provided | outcome | why |
| --- | --- | --- | --- |
| `capabilities/lint/`, holding `capability.md` | yes | read | |
| `capabilities/lint/`, holding only `glossary.md` | no | error | a capability's Goal is required |
| `capabilities/lint/`, holding `Capability.md` | no | error | names are matched exactly, case included |

## <a id="lint.layout.title"></a> Rule: A sibling file opens with a title

```guarded
When      lint reads the first heading <heading> of a file of a [sibling]
Provided  the [heading level] of <heading> is 1
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
