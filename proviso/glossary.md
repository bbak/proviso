# Glossary

## <a id="system.spec-tree"></a> spec tree

- Definition: The files that specify one system: `system.md` in the spec
  root, its siblings, and its capabilities with their features.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: tree

## <a id="system.spec-root"></a> spec root

- Definition: The folder holding a spec tree, named `proviso/` and marked
  by the `system.md` in it.
- Contact: Bruno Baketarić
- Status: defined

## <a id="system.level"></a> level

- Definition: One of system, capability, feature and rule. Each level
  groups the one below it.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: levels

## <a id="system.sibling"></a> sibling

- Definition: A file or folder beside a level's own file that holds
  elements of that level: `glossary.md`, `parameters.md`,
  `references.md`, `questions.md`, `invariants.md`, `constraints.md` or
  `contracts/`.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: siblings

## <a id="system.element"></a> element

- Definition: A part of the spec tree with an identifier: the system, a
  capability, a feature, a rule, an invariant, a constraint, a glossary
  entry, a parameter, a reference or a contract element.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: elements

## <a id="system.element-type"></a> element type

- Definition: Which of system, capability, feature, rule, invariant,
  constraint, glossary entry, parameter, reference or contract element an
  element is.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: element types

## <a id="system.identifier"></a> identifier

- Definition: The permanent name of an element, written as an anchor at
  the start of its heading. No two elements in a spec tree share one.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: identifiers

## <a id="system.section"></a> section

- Definition: A heading without an identifier, written one heading level
  below the element above it, with the text under it up to the next
  heading. It belongs to the nearest element above it.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: sections

## <a id="system.field"></a> field

- Definition: An item of the form `- Key: value` in the list directly
  below an element's heading. Lists further down are prose, even where
  an item looks the same.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: fields

## <a id="system.rule"></a> rule

- Definition: An element at the lowest level, holding one rule block and
  its Examples table.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: rules

## <a id="system.rule-block"></a> rule block

- Definition: A fenced block labelled `guarded`, written in Guarded
  Gherkin.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: rule blocks, guarded block

## <a id="system.guard"></a> guard

- Definition: The `Provided` line of a rule block with its `And` lines:
  the condition that decides whether the `Then` or the `Otherwise`
  applies.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: guards

## <a id="system.hand-over"></a> hand-over

- Definition: An `Otherwise` that names the identifier of another rule
  instead of an outcome, written `Otherwise → <identifier>`.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: hand-overs, hands over

## <a id="system.placeholder"></a> placeholder

- Definition: A value in a rule block written as `<name>`, which the
  Examples table varies.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: placeholders

## <a id="system.examples-table"></a> Examples table

- Definition: The table after a rule block, with one column per
  placeholder, a `provided` column, an `outcome` column, and one row per
  case.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: Examples, Examples row, Examples rows

## <a id="system.term"></a> term

- Definition: A word or phrase defined by a glossary entry, marked as
  `[term]` in a rule block.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: terms, glossary term

## <a id="system.glossary-entry"></a> glossary entry

- Definition: The element that defines a term: its definition, contact,
  status and aliases.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: glossary entries

## <a id="system.contract-element"></a> contract element

- Definition: An element of a contract: an operation, message or entity
  of an interface, or a member of one, referred to in backticks from its
  top-level element, such as `` `customer.id` ``.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: contract elements

## <a id="system.member"></a> member

- Definition: A part of a top-level contract element, such as `id` in
  `customer.id`, written as a heading below it with an anchor of its own.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: members

## <a id="system.reference"></a> reference

- Definition: An element of a `references.md`: an external standard,
  with its version and a link to it, that glossary entries may be
  defined by.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: references

## <a id="system.parameter"></a> parameter

- Definition: An element holding a value that rules refer to in
  backticks, with its unit, whether it is configurable, and why it has
  that value.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: parameters

## <a id="system.undecided"></a> undecided

- Definition: Marked `?` as the first word of a field value, of an
  Examples cell or of a section body, or as a word of its own in a rule
  block. A word is what whitespace separates, so a note after the mark,
  as in `? (ask Treasury)`, leaves the place undecided. Nothing is
  specified there yet.
- Contact: Bruno Baketarić
- Status: defined

## <a id="system.anchor"></a> anchor

- Definition: The HTML element `<a id="…"></a>`: an `a` element with an
  `id` attribute and no content. Inside a code span or a fenced block it
  is text, not an anchor.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: anchors
