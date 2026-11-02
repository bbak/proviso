# Invariants

## <a id="coherence.run.writes"></a> The coherence commands change nothing but the coherence file

### Rationale

`impact` and `verify` read the tree, and `judge` records what the
coherence check found; none may change what it reads. Git is only read,
so a run never moves a branch or touches the index.

### Ranges over

Every run of `impact`, `judge` and `verify`, however it ends.

### Property

Every file and folder under the [spec root] is byte for byte and entry
for entry as it was before the run, except `.coherence.json` after a run
of `judge` that exits with 0. The git repository holding the [spec root]
is as it was before the run.
