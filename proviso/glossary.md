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

## <a id="system.request"></a> request

- Definition: What the business wants changed in a system, written as a
  document that can be checked. It is the input to an amendment and never
  becomes part of the spec tree.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: requests

## <a id="system.amendment"></a> amendment

- Definition: An edit of the spec tree on one branch, usually bringing in
  one request, that ends when it is accepted. Edits made by hand, by a
  decision found while building or by a merge belong to it as well.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: amendments

## <a id="system.baseline"></a> baseline

- Definition: The spec tree as it was last accepted, against which an
  amendment is compared.
- Contact: Bruno Baketarić
- Status: defined

## <a id="system.manifest"></a> manifest

- Definition: The file that records the baseline as a fingerprint of
  every element.
- Contact: Bruno Baketarić
- Status: defined

## <a id="system.element-diff"></a> element diff

- Definition: What an amendment did to the spec tree, element by element:
  the elements added, those changed with the parts that changed, those
  removed, and those changed by a term whose definition changed. It is
  derived mechanically from the baseline and the tree, so rewording
  that leaves every element's normalised text the same shows nothing.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: element diffs

## <a id="system.impact-set"></a> impact set

- Definition: The existing elements an amendment may affect, each with its
  judgement. Its candidates are found mechanically; judging them takes
  an agent and the engineer.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: impact sets

## <a id="system.judgement"></a> judgement

- Definition: The decision on one candidate of an impact set: changed by
  the amendment, or untouched, with the reason why.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: judgements

## <a id="system.coherence-check"></a> coherence check

- Definition: The check by an agent that an amendment keeps the spec tree
  sane in meaning: every candidate of its impact set has a judgement, and
  no contradiction, gap or ignored term, invariant or constraint is left.
  It reads the amendment against its capability and the system level,
  never the whole spec tree, and runs again after every later edit.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: coherence checks

## <a id="system.context-pack"></a> context pack

- Definition: What an implementing agent receives for an amendment: the
  element diff and everything needed to read it, such as the glossary
  entries, invariants, constraints and parameters it uses and
  `system.md`, and nothing else of the spec tree.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: context packs

## <a id="system.implementing-agent"></a> implementing agent

- Definition: The testing agent or the coding agent: the agents that build
  from a context pack, kept apart from each other.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: implementing agents
