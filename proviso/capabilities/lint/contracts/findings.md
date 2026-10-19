# Findings

## <a id="lint.findings"></a> findings

- Represents: [finding]

The kinds of finding lint reports. Each kind is a member, and its name is
the value of `kind` in the `report`.

### <a id="lint.findings.anchor-misplaced"></a> anchor-misplaced

An anchor that does not start the text of a heading.

### <a id="lint.findings.identifier-malformed"></a> identifier-malformed

An identifier that is not well-formed.

### <a id="lint.findings.identifier-duplicate"></a> identifier-duplicate

An identifier that another element has too.

### <a id="lint.findings.system-identifier"></a> system-identifier

A system whose identifier is not `system`.

### <a id="lint.findings.file-unknown"></a> file-unknown

A file in the spec tree that the layout does not list.

### <a id="lint.findings.capability-file-missing"></a> capability-file-missing

A capability folder without `capability.md`.

### <a id="lint.findings.title-missing"></a> title-missing

A file of a sibling whose first heading is not a title.

### <a id="lint.findings.element-unplaced"></a> element-unplaced

An element whose heading and file give no element type.

### <a id="lint.findings.heading-unknown"></a> heading-unknown

A heading without an anchor that is no section of its element.

### <a id="lint.findings.section-missing"></a> section-missing

An element without a section its type requires.

### <a id="lint.findings.section-repeated"></a> section-repeated

A section that appears twice in one element.

### <a id="lint.findings.sibling-empty"></a> sibling-empty

A file of a sibling with nothing but its title.

### <a id="lint.findings.rule-block-count"></a> rule-block-count

A rule without exactly one rule block.

### <a id="lint.findings.keyword-unknown"></a> keyword-unknown

A line of a rule block that does not start with a keyword.

### <a id="lint.findings.keyword-order"></a> keyword-order

A keyword out of its order.

### <a id="lint.findings.keyword-count"></a> keyword-count

A keyword that appears too often, or not at all.

### <a id="lint.findings.guard-incomplete"></a> guard-incomplete

A guard without its Otherwise, or an Otherwise without a guard.

### <a id="lint.findings.hand-over-and"></a> hand-over-and

An And line after a hand-over.

### <a id="lint.findings.and-missing"></a> and-missing

An outcome step without an And line.

### <a id="lint.findings.given-refers-to-event"></a> given-refers-to-event

A Given that uses a placeholder of the When.

### <a id="lint.findings.provided-without-event"></a> provided-without-event

A guard that uses no placeholder of the When.

### <a id="lint.findings.placeholder-unbound"></a> placeholder-unbound

A placeholder that neither the When nor a Given introduces.

### <a id="lint.findings.examples-missing"></a> examples-missing

A rule without an Examples table.

### <a id="lint.findings.placeholder-reserved"></a> placeholder-reserved

A placeholder named like a reserved column.

### <a id="lint.findings.examples-column-missing"></a> examples-column-missing

An Examples table without a column it needs.

### <a id="lint.findings.provided-column"></a> provided-column

A `provided` column on a rule without a guard, or none on a rule with one.

### <a id="lint.findings.examples-column-unknown"></a> examples-column-unknown

A column that is neither a placeholder nor reserved.

### <a id="lint.findings.examples-empty"></a> examples-empty

An Examples table without rows.

### <a id="lint.findings.examples-row-incomplete"></a> examples-row-incomplete

A row with a cell missing or empty.

### <a id="lint.findings.provided-value-unknown"></a> provided-value-unknown

A `provided` cell that is not `yes`, `no`, `no (<part>)` or `?`.

### <a id="lint.findings.examples-side-missing"></a> examples-side-missing

A guard never shown holding, or never shown failing.

### <a id="lint.findings.examples-part-missing"></a> examples-part-missing

A part of a guard never shown failing on its own.

### <a id="lint.findings.term-unresolved"></a> term-unresolved

A term mark that names no visible glossary entry.

### <a id="lint.findings.term-unmarked"></a> term-unmarked

A term used in a rule block without its mark.

### <a id="lint.findings.term-ambiguous"></a> term-ambiguous

A term name or alias that names more than one visible entry.

### <a id="lint.findings.hand-over-unresolved"></a> hand-over-unresolved

A hand-over to something other than a rule.

### <a id="lint.findings.hand-over-cycle"></a> hand-over-cycle

A chain of hand-overs that returns to a rule it has passed.

### <a id="lint.findings.reference-unresolved"></a> reference-unresolved

A backtick reference that names no contract element or parameter.

### <a id="lint.findings.reference-ambiguous"></a> reference-ambiguous

A contract or parameter name that names more than one visible element.

### <a id="lint.findings.field-unknown"></a> field-unknown

A field whose key its element type does not have.

### <a id="lint.findings.field-missing"></a> field-missing

An element without a field its type requires, or with that field empty.

### <a id="lint.findings.field-repeated"></a> field-repeated

A field that appears twice in one element.

### <a id="lint.findings.status-unknown"></a> status-unknown

A glossary entry whose status is neither `defined` nor `open`.

### <a id="lint.findings.open-entry-incomplete"></a> open-entry-incomplete

An open glossary entry without its question, or with a definition.

### <a id="lint.findings.defined-entry-open"></a> defined-entry-open

A defined glossary entry with a question left, or an undecided definition.

### <a id="lint.findings.definition-count"></a> definition-count

A glossary entry with both or neither of `Definition` and `Defined by`.

### <a id="lint.findings.defined-by-unresolved"></a> defined-by-unresolved

A `Defined by` field whose link reaches no visible reference.

### <a id="lint.findings.deviations-without-reference"></a> deviations-without-reference

`Deviations` on a glossary entry not defined by a reference.

### <a id="lint.findings.undecided"></a> undecided

A place the tree leaves undecided: a `?` where the structure puts one, or a rule using an open term.
