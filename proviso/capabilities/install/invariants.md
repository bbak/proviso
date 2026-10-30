# Invariants

## <a id="install.init.keeps"></a> Init changes nothing that existed before it

### Rationale

The files `init` would write are the engineers' own once they exist,
and so is everything else in the repository; a second run, or a run in
a repository that already has a tree, must not cost anyone their
edits.

### Ranges over

Every run of `init`, however it ends.

### Property

Every file that existed below the directory before the run is byte for
byte as it was, and no entry is removed or renamed. The only entries
added are the paths `installed_files` names and the folders above them.

## <a id="install.update.keeps"></a> Update changes only the skills and agent definitions

### Rationale

Everything `init` writes but the skills and agent definitions is the
engineers' own once written, and so is everything else in the
repository; bringing Proviso's own files to a new version must not cost
anyone an edit anywhere else.

### Ranges over

Every run of `update`, however it ends.

### Property

Every file below the directory other than the skills and agent
definitions `installed_files` names is byte for byte as it was, and no
entry is removed or renamed. The only entries added are those paths and
the folders above them.

## <a id="install.version.reads-only"></a> Version changes nothing

### Rationale

`version` reads files the engineers may have changed, and may be run at
any time, by people and by skills; it must never cost anyone an edit.

### Ranges over

Every run of `version`, however it ends.

### Property

Every entry below the directory `version` runs in is byte for byte as it
was, and no entry is added, removed or renamed.
