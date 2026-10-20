package acceptance

// Feature lint.references: proviso/capabilities/lint/features/references.md

import (
	"fmt"
	"testing"
)

const limitThenAnd = "  And     nothing else changes\nOtherwise the cart refuses"

// withThenAnd is tr with the first And line of cart.add.limit replaced.
func withThenAnd(tr tree, line string) tree {
	return tr.edit(featurePath, limitThenAnd, "  And     "+line+"\nOtherwise the cart refuses")
}

// thenAndLine is the line of that And line.
func thenAndLine(tr tree) int {
	return tr.line(featurePath, "Otherwise the cart refuses") - 1
}

func withSystemGlossary() tree {
	return baseline().with("proviso/glossary.md", systemGlossaryMD)
}

func TestLintReferencesTerm(t *testing.T) {
	sys := withSystemGlossary()
	other := withBilling(baseline(), `When      a buyer pays for the [item] <product>
Then      the shop sends a bill for <product>
  And     nothing else changes`, "| product | outcome |\n| --- | --- |\n| apple | billed |\n")
	hyphen := withThenAnd(sys, "the cart stays in the [spec-tree]")
	represents := baseline().with("proviso/capabilities/cart/contracts/api.md",
		"# API\n\n## <a id=\"cart.order\"></a> order\n\n- Represents: [basket]\n\nWhat a buyer orders.\n")
	rows(t, "term-unresolved", []row{
		{name: "row1 the system's entry", tree: withThenAnd(sys, "the cart stays in the [spec tree]"), holds: true},
		{name: "row2 by an alias", tree: baseline(), holds: true},
		{name: "row3 names are matched ignoring case", tree: withThenAnd(sys, "the cart stays in the [Spec Tree]"), holds: true},
		{name: "row4 not visible in another capability", tree: other,
			file: "proviso/capabilities/billing/features/pay.md", lines: lines(other.line("proviso/capabilities/billing/features/pay.md", "[item]"))},
		{name: "row5 neither a name nor an alias", tree: hyphen, file: featurePath, lines: lines(thenAndLine(hyphen))},
		{name: "derived: a Represents field that resolves", tree: baseline().with("proviso/capabilities/cart/contracts/api.md",
			"# API\n\n## <a id=\"cart.order\"></a> order\n\n- Represents: [item]\n\nWhat a buyer orders.\n"), holds: true},
		{name: "derived: a Represents field that does not resolve", tree: represents,
			file: "proviso/capabilities/cart/contracts/api.md", lines: lines(represents.line("proviso/capabilities/cart/contracts/api.md", "- Represents:"))},
	})
	t.Run("row1 resolves without other findings", func(t *testing.T) {
		check(t, withThenAnd(sys, "the cart stays in the [spec tree]")).clean(t)
	})
}

func TestLintReferencesMarked(t *testing.T) {
	sys := withSystemGlossary()
	unmarked := func(tr tree, text string) row {
		tr = withThenAnd(tr, text)
		return row{tree: tr, file: featurePath, lines: lines(thenAndLine(tr))}
	}
	r2 := unmarked(sys, "the cart goes on with the rest of the spec tree")
	r2.name = "row2 an unmarked use"
	r3 := unmarked(baseline(), "the cart keeps the item's price")
	r3.name = "row3 a possessive is a use"
	plural := unmarked(baseline(), "the cart keeps the other items")
	plural.name = "derived: a plural, by its alias"
	upper := unmarked(baseline(), "Item counts stay as they were")
	upper.name = "derived: ignoring case"
	rows(t, "term-unmarked", []row{
		{name: "row1 the term is marked", tree: withThenAnd(sys, "the cart goes on with the rest of the [spec tree]"), holds: true},
		r2,
		r3,
		{name: "row4 no term at all", tree: baseline(), holds: true},
		plural,
		upper,
		{name: "derived: only whole words", tree: withThenAnd(baseline(), "the bill stays itemized"), holds: true},
		{name: "derived: a literal is not a use", tree: withThenAnd(baseline(), `the cart shows "item"`), holds: true},
		{name: "derived: the keyword is taken out", tree: baseline().appendTo(glossaryPath,
			"\n## <a id=\"cart.then\"></a> then\n\n- Definition: The time after paying.\n- Contact: Dana Reyes\n- Status: defined\n"), holds: true},
	})
}

func TestLintReferencesTermAmbiguous(t *testing.T) {
	list := baseline().appendTo(glossaryPath, "\n## <a id=\"cart.item-list\"></a> item list\n\n- Definition: The goods in the cart.\n- Contact: Dana Reyes\n- Status: defined\n- Aliases: items\n")
	redefined := withSystemGlossary().appendTo(glossaryPath, "\n## <a id=\"cart.spec-tree\"></a> spec tree\n\n- Definition: The cart's own files.\n- Contact: Dana Reyes\n- Status: defined\n")
	label := "\n## <a id=\"%s.label\"></a> label\n\n- Definition: A name on a shelf.\n- Contact: Dana Reyes\n- Status: defined\n"
	apart := withBilling(baseline(), findingsBlock, "| findings | outcome |\n| --- | --- |\n| apples | shown |\n").
		appendTo(glossaryPath, fmt.Sprintf(label, "cart")).
		with("proviso/capabilities/billing/glossary.md", "# Glossary\n"+fmt.Sprintf(label, "billing"))

	t.Run("row1 finding", func(t *testing.T) {
		checkGoesOn(t, baseline()).lacks(t, "term-ambiguous", "")
	})
	t.Run("row2 capabilities do not see each other's terms", func(t *testing.T) {
		checkGoesOn(t, apart).lacks(t, "term-ambiguous", "")
	})
	t.Run("row3 the same glossary", func(t *testing.T) {
		o := checkGoesOn(t, list)
		item, other := list.line(glossaryPath, `id="cart.item"`), list.line(glossaryPath, `id="cart.item-list"`)
		relates(t, o.has(t, "term-ambiguous", glossaryPath, item), glossaryPath, other)
		relates(t, o.has(t, "term-ambiguous", glossaryPath, other), glossaryPath, item)
	})
	t.Run("row4 a capability may not redefine a system term", func(t *testing.T) {
		o := checkGoesOn(t, redefined)
		own, sys := redefined.line(glossaryPath, `id="cart.spec-tree"`), redefined.line("proviso/glossary.md", `id="system.spec-tree"`)
		relates(t, o.has(t, "term-ambiguous", glossaryPath, own), "proviso/glossary.md", sys)
		relates(t, o.has(t, "term-ambiguous", "proviso/glossary.md", sys), glossaryPath, own)
	})
}

func TestLintReferencesHandOver(t *testing.T) {
	feature := handOver("Otherwise → cart.add")
	missing := handOver("Otherwise → cart.add.emptyx")
	at := func(tr tree) []int { return lines(tr.line(xPath, "Otherwise →")) }
	rows(t, "hand-over-unresolved", []row{
		{name: "row1 a rule", tree: handOver("Otherwise → cart.add.empty"), holds: true},
		{name: "row2 a feature, not a rule", tree: feature, file: xPath, lines: at(feature)},
		{name: "row3 no such identifier", tree: missing, file: xPath, lines: at(missing)},
	})
}

// chainRule is a guarded rule id whose Otherwise hands over to target.
func chainRule(id, target string) string {
	return ruleMD(id, "Hands over", `When      a buyer adds <product>
Provided  <product> is in stock
Then      the cart holds <product>
  And     nothing else changes
Otherwise → `+target, handOverTable)
}

func TestLintReferencesChain(t *testing.T) {
	self := withX(chainRule("cart.x.a", "cart.x.a"))
	loop := withX(chainRule("cart.x.a", "cart.x.b") + chainRule("cart.x.b", "cart.x.a"))
	ruleA := func(tr tree) []int { return lines(tr.line(xPath, `id="cart.x.a"`)) }
	rows(t, "hand-over-cycle", []row{
		{name: "row1 ends in an outcome", tree: withX(chainRule("cart.x.a", "cart.x.b") + chainRule("cart.x.b", "cart.add.empty")), holds: true},
		{name: "row2 a rule handing over to itself", tree: self, file: xPath, lines: ruleA(self)},
		{name: "row3 a, b, a", tree: loop, file: xPath, lines: ruleA(loop)},
	})
	t.Run("row3 reported once per cycle, with b in related", func(t *testing.T) {
		o := check(t, loop)
		found := o.find("hand-over-cycle", xPath, nil)
		if len(found) != 1 {
			t.Fatalf("want one hand-over-cycle, got:\n%s", dump(o.report.Findings))
		}
		relates(t, found[0], xPath, loop.line(xPath, `id="cart.x.b"`))
	})
}

func withContracts() tree {
	return baseline().with(
		"proviso/capabilities/cart/contracts/findings.md", findingsContractMD,
		"proviso/capabilities/cart/contracts/cli.md", cliContractMD,
	)
}

func TestLintReferencesReference(t *testing.T) {
	ref := func(r string) tree { return withThenAnd(withContracts(), "the shop logs ´"+r+"´") }
	bad := func(name, r string) row {
		tr := ref(r)
		return row{name: name, tree: tr, file: featurePath, lines: lines(thenAndLine(tr))}
	}
	rows(t, "reference-unresolved", []row{
		{name: "row1 findings.anchor-misplaced", tree: ref("findings.anchor-misplaced"), holds: true},
		{name: "row2 lint.exit-status", tree: ref("lint.exit-status"), holds: true},
		{name: "row3 a parameter", tree: ref("max_quantity"), holds: true},
		bad("row4 the term's name, not the contract's", "finding.anchor-misplaced"),
		bad("row5 no such member", "findings.anchor-missing"),
		bad("row6 not qualified from its top-level element", "exit-status"),
		bad("row7 an identifier, not a name", "cart.cli.lint"),
		{name: "derived: a system-level parameter", tree: withThenAnd(baseline().with("proviso/parameters.md",
			"# Parameters\n\n## <a id=\"system.currency\"></a> currency\n\n- Value: EUR\n- Configurable: no\n- Rationale: the shop sells in one country.\n"),
			"the shop logs ´currency´"), holds: true},
	})
	t.Run("rows1-3 resolve without other findings", func(t *testing.T) {
		for _, r := range []string{"findings.anchor-misplaced", "lint.exit-status", "max_quantity"} {
			check(t, ref(r)).clean(t)
		}
	})
}

func TestLintReferencesReferenceAmbiguous(t *testing.T) {
	report := "# Report\n\n## <a id=\"%s\"></a> report\n\nWhat the shop writes.\n"
	contract := "proviso/capabilities/cart/contracts/report.md"
	sameLevel := withContracts().
		with(contract, fmt.Sprintf(report, "cart.report")).
		appendTo(paramsPath, "\n## <a id=\"cart.report-size\"></a> report\n\n- Value: 1\n- Configurable: no\n- Rationale: one report a day.\n")
	levels := withContracts().
		with(contract, fmt.Sprintf(report, "cart.report")).
		with("proviso/contracts/report.md", fmt.Sprintf(report, "system.report"))

	t.Run("row1 findings", func(t *testing.T) {
		checkGoesOn(t, withContracts()).lacks(t, "reference-ambiguous", "")
	})
	t.Run("row2 contracts and parameters share one namespace", func(t *testing.T) {
		o := checkGoesOn(t, sameLevel)
		c, p := sameLevel.line(contract, `id="cart.report"`), sameLevel.line(paramsPath, `id="cart.report-size"`)
		relates(t, o.has(t, "reference-ambiguous", contract, c), paramsPath, p)
		relates(t, o.has(t, "reference-ambiguous", paramsPath, p), contract, c)
	})
	t.Run("row3 no shadowing", func(t *testing.T) {
		o := checkGoesOn(t, levels)
		c, s := levels.line(contract, `id="cart.report"`), levels.line("proviso/contracts/report.md", `id="system.report"`)
		relates(t, o.has(t, "reference-ambiguous", contract, c), "proviso/contracts/report.md", s)
		relates(t, o.has(t, "reference-ambiguous", "proviso/contracts/report.md", s), contract, c)
	})
}
