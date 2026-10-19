# Parameters

## <a id="lint.tree-layout"></a> tree_layout

- Configurable: no
- Rationale: the Proviso Model fixes where each type of element stands,
  so that tooling and readers find it without a config file.

"At any level" means in the spec root or in a capability's folder.

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
