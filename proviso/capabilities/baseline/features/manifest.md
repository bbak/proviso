# <a id="baseline.manifest"></a> Writing the manifest

- Contact: Bruno Baketarić

## Goal

Writing the manifest records the spec tree as the new baseline, in a
form that merges and reads well in git.

## Rationale

The manifest is written at acceptance, never by hand. Writing it from
the whole tree each time, rather than updating it, means it cannot
drift from the tree it records. One sorted line per part lets two
amendments that touch different elements merge without a conflict.

## <a id="baseline.manifest.write"></a> Rule: The manifest records every part of every element

```guarded
When      someone runs `manifest`, and it works on the [spec tree] of the [spec root] <root>
Provided  `manifest_file` can be written into <root>
Then      `manifest` writes `manifest_file` into <root>, naming `model_version` on its first line and then one line for every [part] of every [element] of the [spec tree]
  And     a former `manifest_file` is replaced, not updated
  And     a `coherence_file` and an `amendment_note` in <root> are removed, since the [spec tree] they served is the [baseline] now
  And     `manifest` exits with `manifest.exit-status` 0
Otherwise `manifest` writes to "stderr" why it could not write `manifest_file`
  And     `manifest` exits with `manifest.exit-status` 2
  And     a former `manifest_file`, a `coherence_file` and an `amendment_note` stay as they were
```

| root | provided | outcome | why |
| --- | --- | --- | --- |
| `repo/proviso/`, without a manifest | yes | `repo/proviso/manifest.sum` written | |
| `repo/proviso/`, with a manifest written before an edit | yes | the manifest replaced | |
| `repo/proviso/`, with a manifest written from the same tree | yes | the same bytes as before | the file depends only on the tree |
| `repo/proviso/`, with `.coherence.json` | yes | the manifest written, `.coherence.json` removed | |
| `repo/proviso/`, with `.coherence.json` and `.amendment.md` | yes | the manifest written, both removed | what the amendment left out is no concern of the next |
| `repo/proviso/`, without write permission | no | exit status 2 | |
