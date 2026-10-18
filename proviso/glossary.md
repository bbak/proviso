# Glossary

## <a id="system.spec-tree"></a> spec tree

- Definition: The files that specify one system: `system.md` in the spec
  root, its siblings, and its capabilities with their features.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: tree

## <a id="system.spec-root"></a> spec root

- Definition: The folder holding a spec tree, named `proviso/` and marked
  by the `system.md` in it. The spec root of a working directory is the
  nearest one above it.
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
  `references.md`, `questions.md`, `invariants/`, `constraints/` or
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

## <a id="system.identifier"></a> identifier

- Definition: The permanent name of an element, written as an anchor at
  the start of its heading. No two elements in a spec tree share one.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: identifiers, id

## <a id="system.section"></a> section

- Definition: A heading without an identifier, with the text under it up
  to the next heading. It belongs to the element above it.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: sections

## <a id="system.field"></a> field

- Definition: A list item of the form `- Key: value` that belongs to the
  element above it.
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

## <a id="system.parameter"></a> parameter

- Definition: An element holding a value that rules refer to in
  backticks, with its unit, whether it is configurable, and why it has
  that value.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: parameters

## <a id="system.undecided"></a> undecided

- Definition: Marked `?` in a whole field value, a whole Examples cell, a
  whole section body, or as a token in a rule block. Nothing is specified
  there yet.
- Contact: Bruno Baketarić
- Status: defined

## <a id="system.finding"></a> finding

- Definition: Something lint reports about the spec tree, with where it
  is: an error, a violation of the notation that the tree must not keep
  and that cannot be overridden; or an ask, which asks an author to look,
  such as at a `?`.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: findings
