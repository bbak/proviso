# Parameters

## <a id="baseline.part-kinds"></a> part_kinds

- Configurable: no
- Rationale: the parts are the ones lint already tells apart, so the
  element diff says which part changed in the tree's own terms.

A part's text is what follows its key: a heading's text after its
anchor, a field's value, a section's body, a step's lines with its `And`
lines, a row's line. The header and separator rows of an Examples table
belong to no part; `body` is a part only where it holds text. An
element's parts are its own: an element below it, such as a feature's
rule or a contract element's member, has parts of its own.

| part | key | written as |
| --- | --- | --- |
| the heading | none: one per element | `heading` |
| each field | its key | `field <key>`, such as `field Contact` |
| each section | its name | `section <name>`, such as `section Ranges over` |
| each step of the rule block | its keyword | `step <keyword>`, such as `step Provided` |
| each row of the Examples table | its placeholder values | `row <fingerprint of the values>` |
| everything else the element holds, such as a contract element's prose and blocks or a parameter's table | none: one per element | `body` |

## <a id="baseline.impact-triggers"></a> impact_triggers

- Configurable: no
- Rationale: only a change that can alter behaviour, or the frame in
  which agents fill the gaps between rules, may affect other elements; a
  reworded rationale or a new contact cannot.

Which sources a touched element's parts open. A part of an added or
removed element counts as touched, and so does every part of a rule
changed by a term, whose meaning changed while its text did not.

| touched part | sources |
| --- | --- |
| a step or Examples row of a rule | terms, references, events, ranges |
| the field `Definition` or `Defined by` of a glossary entry | terms |
| any part of a contract element, and any part of a parameter but its field `Rationale` | references |
| the section `Property` or `Ranges over` of an invariant or a constraint | scope |
| any section of the system | scope |
| the section `Goal` of a capability or a feature | scope |
| any other part, such as a field `Contact` or a section `Rationale` | none |
