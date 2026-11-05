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

## <a id="baseline.dependencies"></a> dependencies

- Configurable: no
- Rationale: these are the links the notation makes explicit; a link
  only prose makes, such as a definition naming another term without a
  mark, is left to the coherence check. A contract element depends on
  the term it represents, never the reverse: its shape leaves the
  term's meaning as it is, and a rule that relies on the shape refers to
  the element itself. A rule depends on every member of a contract
  element it refers to, a member added later included, and one removed
  by an amendment, since a changed member can affect any such rule:
  coarse on purpose, at the cost of rules listed that only the added or
  removed member touches. A glossary entry depends on the terms its
  definition or deviations mark, and two entries may each mark the
  other, as a glossary's definitions often lean on each other; an entry
  that marks its own term depends on nothing through that mark.

Which element depends on which: when the meaning of the second changes,
so does the meaning of the first.

| element | depends on |
| --- | --- |
| a rule | every glossary entry whose term its rule block marks |
| a rule | every contract element and parameter its rule block refers to, and every member of such a contract element, one removed from it included |
| a rule | the rule its hand-over names |
| a contract element | the glossary entry its field `Represents` marks |
| a glossary entry | the reference its field `Defined by` links to |
| a glossary entry | every other glossary entry whose term its field `Definition` or `Deviations` marks |
