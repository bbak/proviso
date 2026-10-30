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

## <a id="baseline.change-effects"></a> change_effects

- Configurable: no
- Rationale: only a change that can alter behaviour, or the frame in
  which agents fill the gaps between rules, may affect other elements; a
  reworded rationale or a new contact cannot.

What a changed part does beyond its own element. "Dependents" passes the
changed meaning on to every element that depends on it, which the
element diff lists as changed through a dependency; the other entries
are sources of impact candidates. A part of an added or removed element
counts as changed, and so does every part of an element changed through
a dependency, whose meaning changed while its text did not.

| changed part | dependents | candidate sources |
| --- | --- | --- |
| a step or Examples row of a rule | yes | events, ranges |
| the field `Definition` or `Defined by` of a glossary entry | yes | none |
| the field `Version` or `Link` of a reference | yes | none |
| any part of a contract element, and any part of a parameter but its field `Rationale` | yes | none |
| the section `Property` or `Ranges over` of an invariant or a constraint | no | scope |
| any section of the system | no | scope |
| the section `Goal` of a capability or a feature | no | scope |
| any other part, such as a field `Contact` or a section `Rationale` | no | none |

## <a id="baseline.dependencies"></a> dependencies

- Configurable: no
- Rationale: these are the links the notation makes explicit; a link
  only prose makes, such as a definition naming another term without a
  mark, is left to the coherence check. A contract element depends on
  the term it represents, never the reverse: its shape leaves the
  term's meaning as it is, and a rule that relies on the shape refers to
  the element itself. A rule depends on every member of a contract
  element it refers to, a member added later included, since a changed
  member can affect any such rule: coarse on purpose, at the cost of
  rules listed that only the added member touches.

Which element depends on which: when the meaning of the second changes,
so does the meaning of the first.

| element | depends on |
| --- | --- |
| a rule | every glossary entry whose term its rule block marks |
| a rule | every contract element and parameter its rule block refers to, and every member of such a contract element |
| a rule | the rule its hand-over names |
| a contract element | the glossary entry its field `Represents` marks |
| a glossary entry | the reference its field `Defined by` links to |
