# Invariants

## <a id="query.read-only"></a> The query commands never write

### Rationale

A query reads the tree for someone who wants to know what it says; a
query that changed what it reads would change the answer, and would
touch a tree in the middle of an amendment.

### Ranges over

Every run of `show`, whatever it finds and however it ends.

### Property

Every file and folder under the [spec root] is byte for byte and entry
for entry as it was before the run.
