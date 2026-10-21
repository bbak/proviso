# Manifest

## <a id="system.manifest-file"></a> manifest_file

The file `manifest.sum` in the spec root, written by `manifest` and
never edited by hand. It is text in UTF-8, and every line, the last one
too, ends with a line feed.

- **The first line** is `model`, a space and the `model_version` the
  manifest was written with, such as `model 0`.
- **One line per part** follows, for every part of every element, with
  three fields separated by a tab: the element's identifier, the part as
  `part_kinds` writes it, and the part's fingerprint.
- **A fingerprint** is `h1:` followed by the SHA-256 of the part's text
  in lowercase hexadecimal, after each run of spaces, tabs and line
  feeds in the text has been made one space, and none is left at its
  start or end.
- **A row's key** is the fingerprint, taken the same way, of its
  placeholder values: for each placeholder column, in the byte order of
  the column names, the name, `=` and the cell, joined by tabs. A row
  whose key a row above it in the same table already has gets ` #2`
  appended, the next ` #3`, and so on.
- **The lines of parts are sorted** byte by byte, so the file depends
  only on the tree it records.

```text
model 0
lint.run.exit	heading	h1:5b0d…
lint.run.exit	row h1:9a41…	h1:e5a1…
lint.run.exit	step Otherwise	h1:77d0…
```
