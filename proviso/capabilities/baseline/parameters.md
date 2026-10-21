# Parameters

## <a id="baseline.part-kinds"></a> part_kinds

- Configurable: no
- Rationale: the parts are the ones lint already tells apart, so the
  element diff says which part changed in the tree's own terms.

A part's text is what follows its key: a heading's text after its
anchor, a field's value, a section's body, a step's lines with its `And`
lines, a row's line. An element's parts are its own: an element below
it, such as a feature's rule or a contract element's member, has parts
of its own.

| part | key | written as |
| --- | --- | --- |
| the heading | none: one per element | `heading` |
| each field | its key | `field <key>`, such as `field Contact` |
| each section | its name | `section <name>`, such as `section Ranges over` |
| each step of the rule block | its keyword | `step <keyword>`, such as `step Provided` |
| each row of the Examples table | its placeholder values | `row <fingerprint of the values>` |
| everything else the element holds, such as a contract element's prose and blocks or a parameter's table | none: one per element | `body` |
