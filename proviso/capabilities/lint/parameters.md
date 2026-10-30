# Parameters

## <a id="lint.tree-layout"></a> tree_layout

- Configurable: no
- Rationale: the Proviso Model fixes where each type of element stands,
  so that tooling and readers find it without a config file.

"At any level" means in the spec root or in a capability's folder. The
names are matched exactly, case included, on every platform. A
heading whose text starts with `Rule:` or `Invariant:` gives an element
type only in a feature file; anywhere else it gives none, because a
prefix outside a file that mixes types marks an element in the wrong
place.

| path | heading | element type |
| --- | --- | --- |
| `system.md` | level 1, with an anchor | system |
| `capabilities/<capability>/capability.md` | level 1, with an anchor | capability |
| `capabilities/<capability>/features/<feature>.md` | level 1, with an anchor | feature |
| `capabilities/<capability>/features/<feature>.md` | level 2, with an anchor, text starting `Rule:` | rule |
| `capabilities/<capability>/features/<feature>.md` | level 2, with an anchor, text starting `Invariant:` | invariant |
| `glossary.md`, at any level | level 2, with an anchor | glossary entry |
| `parameters.md`, at any level | level 2, with an anchor | parameter |
| `references.md`, at any level | level 2, with an anchor | reference |
| `invariants.md`, at any level | level 2, with an anchor | invariant |
| `constraints.md`, at any level | level 2, with an anchor | constraint |
| `contracts/<name>.md`, at any level | level 2 or deeper, with an anchor | contract element |
| `questions.md`, at any level | none | none |
| `.markdownlint.jsonc`, in the spec root | none | none |
| `manifest.sum`, in the spec root | none | none |

## <a id="lint.section-names"></a> section_names

- Configurable: no
- Rationale: a fixed set of sections per element type lets lint tell a
  section from a forgotten anchor, and keeps every element of a type
  readable in the same way.

| element type | sections | required |
| --- | --- | --- |
| system | Goal, Rationale, Scope, Out of scope, Audience, Design boundary | all |
| capability | Goal, Rationale | all |
| feature | Goal, Rationale | all |
| rule | Rationale | none |
| invariant | Rationale, Ranges over, Property | all |
| constraint | Rationale, Property | all |
| glossary entry, parameter, reference, contract element | none | none |

## <a id="lint.field-names"></a> field_names

- Configurable: no
- Rationale: a fixed set of fields per element type lets lint tell a
  misspelt key from a field of another type, and keeps out fields that
  do not belong. The tree names a Contact, whom to ask, never an owner,
  a word that means decision authority, accountability or
  responsibility, depending on the reader; a rule names none, since a
  contact per rule goes stale and is rarely kept up. No field lists the
  rules that use a term: the term marks say it, and a list would be a
  second source of truth. A reference has no Rationale: why a standard
  was chosen belongs to the feature or constraint that chose it.

| element type | fields | required |
| --- | --- | --- |
| feature | Contact | Contact |
| glossary entry | Definition, Defined by, Deviations, Contact, Status, Aliases, Question | Contact, Status |
| parameter | Value, Unit, Configurable, Rationale | Configurable, Rationale |
| reference | Version, Link, Contact | Version, Link, Contact |
| contract element | Represents | none |
| system, capability, rule, invariant, constraint | none | none |
