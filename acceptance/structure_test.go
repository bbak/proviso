package acceptance

// Feature lint.structure: proviso/capabilities/lint/features/structure.md

import "testing"

const invariantMD = `
## <a id="cart.add.bounded"></a> Invariant: The cart never holds too much

### Rationale

Shelves are small.

### Ranges over

Every addition to the cart.

### Property

The cart holds at most ´max_quantity´ pieces of one product.
`

// at the end of add.md, so within the rule cart.add.empty.
func appendFeature(text string) tree { return baseline().appendTo(featurePath, text) }

func TestLintStructureType(t *testing.T) {
	something := appendFeature("\n## <a id=\"cart.add.x\"></a> Something\n\nA paragraph.\n")
	nested := appendFeature("\n### <a id=\"cart.add.y\"></a> Rule: Nested\n\nA paragraph.\n")
	inGlossary := baseline().appendTo(glossaryPath, "\n## <a id=\"cart.rule-like\"></a> Rule: A rule in a glossary\n\nA paragraph.\n")
	inCapability := baseline().appendTo(capPath, "\n## <a id=\"cart.extra\"></a> Extra\n\nA paragraph.\n")
	deeper := baseline().appendTo(glossaryPath, "\n### <a id=\"cart.item.kind\"></a> kind\n\nA paragraph.\n")
	rows(t, "element-unplaced", []row{
		{name: "row1 a rule", tree: baseline(), holds: true},
		{name: "row2 an invariant in a feature", tree: appendFeature(invariantMD), holds: true},
		{name: "row3 invariants.md gives the type", tree: baseline().with("proviso/invariants.md", invariantsMD), holds: true},
		{name: "row4 a member of a contract", tree: baseline().with("proviso/capabilities/cart/contracts/findings.md", findingsContractMD), holds: true},
		{name: "row5 neither Rule: nor Invariant:", tree: something,
			file: featurePath, lines: lines(something.line(featurePath, `id="cart.add.x"`))},
		{name: "row6 a rule at level 3", tree: nested,
			file: featurePath, lines: lines(nested.line(featurePath, `id="cart.add.y"`))},
		// The Examples row says error; tree_layout lists any level 2
		// heading with an anchor in glossary.md as a glossary entry. See
		// the report: the test follows the Examples row.
		{name: "row7 no rules in a glossary (conflicts with tree_layout)", tree: inGlossary,
			file: glossaryPath, lines: lines(inGlossary.line(glossaryPath, `id="cart.rule-like"`))},
		{name: "derived: level 2 in capability.md", tree: inCapability,
			file: capPath, lines: lines(inCapability.line(capPath, `id="cart.extra"`))},
		{name: "derived: level 3 in glossary.md", tree: deeper,
			file: glossaryPath, lines: lines(deeper.line(glossaryPath, `id="cart.item.kind"`))},
	})
	t.Run("row2 the invariant is read as one", func(t *testing.T) {
		check(t, appendFeature(invariantMD)).clean(t)
	})
	t.Run("row3 invariants.md is read", func(t *testing.T) {
		check(t, baseline().with("proviso/invariants.md", invariantsMD)).clean(t)
	})
}

func TestLintStructureSection(t *testing.T) {
	afterRationale := func(heading string) tree {
		return baseline().edit(featurePath, "would empty the shelves.\n", "would empty the shelves.\n\n"+heading+"\n\nSome text.\n")
	}
	scope := afterRationale("## Scope")
	lower := afterRationale("## goal")
	forgotten := afterRationale("## Rule: Something")
	inRule := baseline().edit(featurePath, limitHeading+"\n", limitHeading+"\n\n### Goal\n\nSome text.\n")
	inEntry := baseline().appendTo(glossaryPath, "\n### Notes\n\nSome text.\n")
	rows(t, "heading-unknown", []row{
		{name: "row1 ## Goal in a feature", tree: baseline(), holds: true},
		{name: "row2 ### Ranges over in an invariant", tree: appendFeature(invariantMD), holds: true},
		{name: "row3 only the system has a scope", tree: scope,
			file: featurePath, lines: lines(scope.line(featurePath, "## Scope"))},
		{name: "row4 a rule has only a rationale", tree: inRule,
			file: featurePath, lines: lines(inRule.line(featurePath, "### Goal"))},
		{name: "row5 section names are matched exactly", tree: lower,
			file: featurePath, lines: lines(lower.line(featurePath, "## goal"))},
		{name: "row6 a forgotten anchor", tree: forgotten,
			file: featurePath, lines: lines(forgotten.line(featurePath, "## Rule: Something"))},
		{name: "derived: a glossary entry has no sections", tree: inEntry,
			file: glossaryPath, lines: lines(inEntry.line(glossaryPath, "### Notes"))},
	})
}

func TestLintStructureRequired(t *testing.T) {
	noProperty := appendFeature(invariantMD[:len(invariantMD)-len("### Property\n\nThe cart holds at most ´max_quantity´ pieces of one product.\n")])
	noConstraintProperty := baseline().with("proviso/constraints.md", constraintsMD[:len(constraintsMD)-len("### Property\n\nEvery page works in a browser.\n")])
	rows(t, "section-missing", []row{
		{name: "row1 a feature with Goal and Rationale", tree: baseline(), holds: true},
		{name: "row2 a rule's rationale is optional", tree: baseline(), holds: true},
		{name: "row3 a feature without Rationale", tree: baseline().edit(featurePath, "## Rationale\n\nA cart that takes anything would empty the shelves.\n\n", ""),
			file: featurePath, lines: lines(1)},
		{name: "row4 the system without Design boundary", tree: baseline().edit(systemPath, "\n## Design boundary\n\n- A web browser.\n", ""),
			file: systemPath, lines: lines(1)},
		{name: "derived: an invariant without Property", tree: noProperty,
			file: featurePath, lines: lines(noProperty.line(featurePath, `id="cart.add.bounded"`))},
		{name: "derived: a constraint without Property", tree: noConstraintProperty,
			file: "proviso/constraints.md", lines: lines(noConstraintProperty.line("proviso/constraints.md", `id="system.web"`))},
		{name: "derived: a capability without Goal", tree: baseline().edit(capPath, "## Goal\n\nBuyers collect goods before they pay.\n\n", ""),
			file: capPath, lines: lines(1)},
	})
}

func TestLintStructureOnce(t *testing.T) {
	twice := baseline().edit(featurePath, "would empty the shelves.\n", "would empty the shelves.\n\n## Goal\n\nMore goals.\n")
	rows(t, "section-repeated", []row{
		{name: "row1 Goal, Rationale", tree: baseline(), holds: true},
		{name: "row2 Goal, Rationale, Goal", tree: twice,
			file: featurePath, lines: lines(twice.lineN(featurePath, "## Goal", 2))},
		{name: "row3 the rule's Rationale belongs to the rule", tree: baseline().edit(featurePath, limitHeading+"\n", limitHeading+"\n\n### Rationale\n\nShelves are small.\n"), holds: true},
	})
}

func TestLintStructureFieldKnown(t *testing.T) {
	source := baseline().edit(glossaryPath, "- Status: defined\n", "- Status: defined\n- Source: Maria Keller\n")
	owner := baseline().edit(featurePath, "- Contact: Dana Reyes\n", "- Contact: Dana Reyes\n- Owner: Search team\n")
	lower := baseline().edit(featurePath, "- Contact: Dana Reyes\n", "- Contact: Dana Reyes\n- contact: Dana Reyes\n")
	onCapability := baseline().edit(capPath, "# <a id=\"cart\"></a> Cart\n", "# <a id=\"cart\"></a> Cart\n\n- Contact: Dana Reyes\n")
	rows(t, "field-unknown", []row{
		{name: "row1 Contact on a feature", tree: baseline(), holds: true},
		{name: "row2 Represents on a contract element", tree: baseline().with("proviso/capabilities/cart/contracts/api.md",
			"# API\n\n## <a id=\"cart.order\"></a> order\n\n- Represents: [item]\n\nWhat a buyer orders.\n"), holds: true},
		{name: "row3 Source on a glossary entry", tree: source,
			file: glossaryPath, lines: lines(source.line(glossaryPath, "- Source:"))},
		{name: "row4 Owner on a feature", tree: owner,
			file: featurePath, lines: lines(owner.line(featurePath, "- Owner:"))},
		{name: "row5 keys are matched exactly", tree: lower,
			file: featurePath, lines: lines(lower.line(featurePath, "- contact:"))},
		{name: "derived: a capability has no fields", tree: onCapability,
			file: capPath, lines: lines(onCapability.line(capPath, "- Contact:"))},
		{name: "derived: lists further down are prose", tree: baseline().edit(featurePath, "Buyers put goods into the cart.\n",
			"Buyers put goods into the cart.\n\n- Owner: Search team\n"), holds: true},
	})
}

func TestLintStructureFieldRequired(t *testing.T) {
	noStatus := baseline().appendTo(glossaryPath, "\n## <a id=\"cart.basket\"></a> basket\n\n- Definition: What a buyer carries.\n- Contact: Dana Reyes\n")
	rows(t, "field-missing", []row{
		{name: "row1 Contact", tree: baseline(), holds: true},
		{name: "row2 Contact undecided", tree: baseline().edit(featurePath, "- Contact: Dana Reyes", "- Contact: ?"), holds: true},
		{name: "row3 no Contact", tree: baseline().edit(featurePath, "- Contact: Dana Reyes\n\n", ""),
			file: featurePath, lines: lines(1)},
		{name: "row4 an empty value", tree: baseline().edit(featurePath, "- Contact: Dana Reyes", "- Contact:"),
			file: featurePath, lines: lines(1)},
		{name: "row5 a table-valued parameter needs no Value", tree: baseline().appendTo(paramsPath,
			"\n## <a id=\"cart.sizes\"></a> sizes\n\n- Configurable: no\n- Rationale: the shelves come in these sizes.\n\n| size | pieces |\n| --- | --- |\n| small | 5 |\n"), holds: true},
		{name: "row6 a glossary entry without Status", tree: noStatus,
			file: glossaryPath, lines: lines(noStatus.line(glossaryPath, `id="cart.basket"`))},
	})
}

func TestLintStructureFieldOnce(t *testing.T) {
	twice := baseline().edit(glossaryPath, "- Contact: Dana Reyes\n", "- Contact: Dana Reyes\n- Contact: Lee Park\n")
	rows(t, "field-repeated", []row{
		{name: "row1 Definition, Contact, Status, Aliases", tree: baseline(), holds: true},
		{name: "row2 Definition, Contact, Contact, Status", tree: twice,
			file: glossaryPath, lines: lines(twice.lineN(glossaryPath, "- Contact:", 2))},
	})
}
