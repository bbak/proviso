# <a id="baseline.manifest"></a> Writing the manifest

- Contact: Bruno Baketarić

## Goal

Writing the manifest records the spec tree as the new baseline, in a
form that merges and reads well in git.

## Rationale

The manifest is written at acceptance, never by hand. Writing it from
the whole tree each time, rather than updating it, means it cannot
drift from the tree it records. One sorted line per part lets two
amendments that touch different elements merge without a conflict. It
holds fingerprints, never text: a copy of the tree would be a second
source of truth, and the former text is in git, at the revision that
wrote the manifest.

## <a id="baseline.manifest.write"></a> Rule: The manifest records every part of every element

```guarded
When      someone runs `manifest`, and it works on the [spec tree] of the [spec root] <root>
Provided  `manifest_file` can be written into <root>
Then      `manifest` writes `manifest_file` into <root>, naming `model_version` on its first line and then one line for every [part] of every [element] of the [spec tree]
  And     a former `manifest_file` is replaced, not updated
  And     `manifest` goes on to remove a `coherence_file` and an `amendment_note` in <root>, since the [spec tree] they served is the [baseline] now
Otherwise `manifest` writes to "stderr" why it could not write `manifest_file`
  And     `manifest` exits with `manifest.exit-status` 2
  And     a former `manifest_file`, a `coherence_file` and an `amendment_note` stay as they were
```

| root | provided | outcome | why |
| --- | --- | --- | --- |
| `repo/proviso/`, without a manifest | yes | `repo/proviso/manifest.sum` written | |
| `repo/proviso/`, with a manifest written before an edit | yes | the manifest replaced | |
| `repo/proviso/`, with a manifest written from the same tree | yes | the same bytes as before | the file depends only on the tree |
| `repo/proviso/`, with `.coherence.json` | yes | the manifest written, `.coherence.json` to be removed | |
| `repo/proviso/`, with `.coherence.json` and `.amendment.md` | yes | the manifest written, both to be removed | what the amendment left out is no concern of the next |
| `repo/proviso/`, without write permission | no | exit status 2 | |

## <a id="baseline.manifest.remove"></a> Rule: What the manifest cannot remove fails it

### Rationale

Once the tree is the baseline, whatever stands under the names of the
coherence file and the amendment note is left over, so a folder goes
with everything in it, and a link goes without what it points to. A
removal can fail only after `manifest_file` is written, so the written
manifest stays, and `verify` then reports what is left over.

```guarded
When      `manifest` has written `manifest_file` into <root> and removes the entries <entries> named as `coherence_file` and `amendment_note` there
Provided  each of <entries> can be removed
Then      each of <entries> is removed, a folder with everything in it and a link without what it points to
  And     `manifest` exits with `manifest.exit-status` 0
Otherwise `manifest` writes to "stderr" which of <entries> it could not remove
  And     each of <entries> that can be removed is removed
  And     `manifest` exits with `manifest.exit-status` 2
  And     `manifest_file` stays as written
```

| root | entries | provided | outcome | why |
| --- | --- | --- | --- | --- |
| `repo/proviso/` | none | yes | exit status 0 | nothing to remove |
| `repo/proviso/` | `.coherence.json` and `.amendment.md`, both files | yes | both removed, exit status 0 | |
| `repo/proviso/` | `.amendment.md`, a folder holding a file | yes | the folder and its file removed, exit status 0 | anything of that name is the note |
| `repo/proviso/` | `.amendment.md`, a link to `capabilities/x.md` | yes | the link removed, `capabilities/x.md` unchanged, exit status 0 | the tree stays as recorded |
| `repo/proviso/` | `.coherence.json`, which cannot be removed | no | exit status 2, `manifest.sum` written | |
| `repo/proviso/` | `.coherence.json`, which cannot be removed, and `.amendment.md`, which can | no | `.amendment.md` removed, exit status 2 | the one that fails holds up no other |
