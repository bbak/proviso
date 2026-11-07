# Parameters

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
