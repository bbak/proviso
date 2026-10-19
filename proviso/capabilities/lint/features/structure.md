# <a id="lint.structure"></a> Structure

- Contact: Bruno Baketarić

## Goal

Every heading in the spec tree is either an element of the type its
place allows, a section that element type has, or a file's title, and
every element has the sections and fields it requires.

## Rationale

Tooling reads each element by its heading and sections. A heading of
the wrong type is read as something it is not, a forgotten anchor turns
an element into an unknown section, and a missing section leaves an
element half-specified. Each of these is silent unless lint reports it.

## <a id="lint.structure.type"></a> Rule: An element's type follows from where it stands

```guarded
When      lint reads the heading <heading> of an [element] in the file <file>
Provided  `tree_layout` gives an [element type] for <heading> in <file>
Then      the element is of that type
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.element-unplaced` at <heading>
  And     lint goes on with the rest of the [spec tree]
```

| heading | file | provided | outcome | why |
| --- | --- | --- | --- | --- |
| `## <a id="lint.run.root"></a> Rule: …` | `capabilities/lint/features/run.md` | yes | rule | |
| `## <a id="lint.run.read-only"></a> Invariant: …` | `capabilities/lint/features/run.md` | yes | invariant | |
| `## <a id="system.read-only"></a> The bundle is never written` | `invariants.md` | yes | invariant | the file gives the type |
| `### <a id="lint.findings.anchor-misplaced"></a> anchor-misplaced` | `capabilities/lint/contracts/findings.md` | yes | contract element | a member |
| `## <a id="lint.run.x"></a> Something` | `capabilities/lint/features/run.md` | no | error | neither `Rule:` nor `Invariant:` |
| `### <a id="lint.run.x"></a> Rule: …` | `capabilities/lint/features/run.md` | no | error | a rule at level 3 |
| `## <a id="lint.x"></a> Rule: …` | `capabilities/lint/glossary.md` | no | error | no rules in a glossary |

## <a id="lint.structure.section"></a> Rule: A heading without an anchor is a known section

```guarded
When      lint reads a heading <heading> without an [anchor], other than a file's title, within an [element] of the [element type] <type>
Provided  `section_names` lists <heading> for <type>
Then      <heading> opens a [section] of that element
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.heading-unknown` at <heading>
  And     lint goes on with the rest of the [spec tree]
```

| heading | type | provided | outcome | why |
| --- | --- | --- | --- | --- |
| `## Goal` | feature | yes | section | |
| `### Ranges over` | invariant | yes | section | |
| `## Scope` | feature | no | error | only the system has a scope |
| `### Goal` | rule | no | error | a rule has only a rationale |
| `## goal` | feature | no | error | section names are matched exactly |
| `## Rule: Something` | feature | no | error | a forgotten anchor |

## <a id="lint.structure.required"></a> Rule: An element has its required sections

```guarded
When      lint has read an [element] <element> of the [element type] <type>, with the sections <sections>
Provided  <sections> include every section `section_names` requires for <type>
Then      lint reports nothing about the sections of <element>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.section-missing` at <element>, naming each missing section
  And     lint goes on with the rest of the [spec tree]
```

| element | type | sections | provided | outcome | why |
| --- | --- | --- | --- | --- | --- |
| `lint.run` | feature | Goal, Rationale | yes | nothing reported | |
| `lint.run.root` | rule | none | yes | nothing reported | a rule's rationale is optional |
| `lint.run` | feature | Goal | no | error, naming Rationale | |
| `system` | system | Goal, Rationale, Scope, Out of scope, Audience | no | error, naming Design boundary | |

## <a id="lint.structure.once"></a> Rule: A section appears once per element

```guarded
When      lint has read an [element] <element> with the sections <sections>
Provided  no section name appears twice in <sections>
Then      lint reports nothing about repeated sections of <element>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.section-repeated` at the second occurrence
  And     lint goes on with the rest of the [spec tree]
```

| element | sections | provided | outcome | why |
| --- | --- | --- | --- | --- |
| `lint.run` | Goal, Rationale | yes | nothing reported | |
| `lint.run` | Goal, Rationale, Goal | no | error at the second Goal | |
| `lint.run`, with a rule that has a Rationale | Goal, Rationale | yes | nothing reported | the rule's Rationale belongs to the rule |

## <a id="lint.structure.contact"></a> Rule: A feature names a contact

```guarded
When      lint has read a feature <feature> with the [fields] <fields>
Provided  <fields> include "Contact" with a value
Then      lint reports nothing about the contact of <feature>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.contact-missing` at <feature>
  And     lint goes on with the rest of the [spec tree]
```

| feature | fields | provided | outcome | why |
| --- | --- | --- | --- | --- |
| `lint.run` | `- Contact: Bruno Baketarić` | yes | nothing reported | |
| `lint.run` | `- Contact: ?` | yes | nothing reported | undecided, reported as an ask elsewhere |
| `lint.run` | none | no | error | |
| `lint.run` | `- Contact:` | no | error | an empty value |
