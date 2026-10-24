# Invariants

## <a id="baseline.run.writes"></a> The baseline commands change nothing but the manifest and the coherence file

### Rationale

`diff`, `impact`, `pack` and `verify` read the tree, `manifest` records
what it finds and `judge` what the coherence check found; none may
change what it reads. Git is only read, so a run never moves a branch or
touches the index.

### Ranges over

Every run of `manifest`, `diff`, `impact`, `pack`, `judge` and `verify`,
however it ends.

### Property

Every file and folder under the [spec root] is byte for byte and entry
for entry as it was before the run, except `manifest.sum` and
`.coherence.json` after a run of `manifest` that exits with 0, and
`.coherence.json` after a run of `judge` that exits with 0. The git
repository holding the [spec root] is as it was before the run.
