# <a id="baseline"></a> Baseline

## Goal

The baseline records the spec tree as it was last accepted, part by part,
and tells what has happened to the tree since: element by element, and in
which part of each, so that a changed meaning shows and a reflowed line
does not, which untouched rules and invariants that may affect,
what the implementing agents and the checking agent need to read it,
and whether the coherence check covers the amendment as it stands.

## Rationale

A line diff cannot tell a moved rule from a new one, or a reflowed
paragraph from a changed guard, and a flag set by hand is forgotten.
Fingerprints of each element's parts, kept in a manifest, let the element
diff find every change mechanically. Everything derived from an amendment
later, impact sets and context packs, starts from this diff.
