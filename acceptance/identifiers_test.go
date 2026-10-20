package acceptance

// Feature lint.identifiers: proviso/capabilities/lint/features/identifiers.md

import "testing"

const limitHeading = `## <a id="cart.add.limit"></a> Rule: The cart takes a limited quantity`

func TestLintIdentifiersPlacement(t *testing.T) {
	before := baseline().edit(featurePath, limitHeading, `<a id="cart.add.limit"></a>`+"\n"+`## Rule: The cart takes a limited quantity`)
	inside := baseline().edit(featurePath, limitHeading, `## Rule: The cart takes a limited quantity <a id="cart.add.limit"></a>`)
	paragraph := baseline().edit(featurePath, "A cart that takes anything would", `A cart that takes anything <a id="cart.add.why"></a> would`)
	rows(t, "anchor-misplaced", []row{
		{name: "row1 the prescribed form", tree: baseline(), holds: true},
		{name: "row2 no space after the anchor", tree: baseline().edit(featurePath, `</a> Rule: The cart takes`, `</a>Rule: The cart takes`), holds: true},
		{name: "row3 on the line before the heading", tree: before,
			file: featurePath, lines: lines(before.line(featurePath, `<a id="cart.add.limit"></a>`))},
		{name: "row4 inside the heading, not at its start", tree: inside,
			file: featurePath, lines: lines(inside.line(featurePath, `<a id="cart.add.limit"></a>`))},
		{name: "row5 within a paragraph", tree: paragraph,
			file: featurePath, lines: lines(paragraph.line(featurePath, `<a id="cart.add.why"></a>`))},
		{name: "derived: in a code span it is text", tree: baseline().edit(featurePath, "A cart that takes anything would",
			"A cart that takes ´<a id=\"cart.add.why\"></a>´ would"), holds: true},
		{name: "derived: in a fenced block it is text", tree: baseline().edit(featurePath, "A cart that takes anything would empty the shelves.\n",
			"A cart that takes anything would empty the shelves.\n\n´´´html\n<a id=\"cart.add.why\"></a>\n´´´\n"), holds: true},
	})
	t.Run("row2 opens the element", func(t *testing.T) {
		check(t, baseline().edit(featurePath, `</a> Rule: The cart takes`, `</a>Rule: The cart takes`)).clean(t)
	})
}

func TestLintIdentifiersForm(t *testing.T) {
	withID := func(id string) tree {
		return baseline().edit(featurePath, `id="cart.add.limit"`, `id="`+id+`"`)
	}
	heading := baseline().line(featurePath, "Rule: The cart takes")
	bad := func(name, id string) row {
		return row{name: name, tree: withID(id), file: featurePath, lines: lines(heading)}
	}
	rows(t, "identifier-malformed", []row{
		{name: "row1 search.attempts.simplified", tree: withID("search.attempts.simplified"), holds: true},
		{name: "row2 system, a single segment", tree: baseline(), holds: true},
		{name: "row3 lint.run.root-2", tree: withID("lint.run.root-2"), holds: true},
		bad("row4 Search.attempts", "Search.attempts"),
		bad("row5 search..attempts", "search..attempts"),
		bad("row6 search.attempts.", "search.attempts."),
		bad("row7 search.2nd", "search.2nd"),
		bad("row8 search.first_attempt", "search.first_attempt"),
		{name: "derived: a capability's identifier", tree: baseline().edit(capPath, `id="cart"`, `id="Cart"`),
			file: capPath, lines: lines(1)},
		{name: "derived: a glossary entry's identifier", tree: baseline().edit(glossaryPath, `id="cart.item"`, `id="cart.Item"`),
			file: glossaryPath, lines: lines(baseline().line(glossaryPath, `id="cart.item"`))},
	})
}

func TestLintIdentifiersUnique(t *testing.T) {
	t.Run("row1 lint.identifiers among lint, lint.run", func(t *testing.T) {
		checkGoesOn(t, baseline()).lacks(t, "identifier-duplicate", "")
	})
	t.Run("row2 reported at each of the two", func(t *testing.T) {
		tr := baseline().edit(featurePath, `id="cart.add.empty"`, `id="cart.add.limit"`)
		first := tr.line(featurePath, "Rule: The cart takes")
		second := tr.line(featurePath, "Rule: A new cart is empty")
		o := checkGoesOn(t, tr)
		relates(t, o.has(t, "identifier-duplicate", featurePath, first), featurePath, second)
		relates(t, o.has(t, "identifier-duplicate", featurePath, second), featurePath, first)
	})
	t.Run("row3 a capability identified as system", func(t *testing.T) {
		o := checkGoesOn(t, baseline().edit(capPath, `id="cart"`, `id="system"`))
		relates(t, o.has(t, "identifier-duplicate", capPath, 1), systemPath, 1)
		relates(t, o.has(t, "identifier-duplicate", systemPath, 1), capPath, 1)
	})
	t.Run("derived: in two files", func(t *testing.T) {
		tr := baseline().edit(glossaryPath, `id="cart.item"`, `id="cart.add"`)
		entry := tr.line(glossaryPath, `id="cart.add"`)
		o := checkGoesOn(t, tr)
		relates(t, o.has(t, "identifier-duplicate", glossaryPath, entry), featurePath, 1)
		relates(t, o.has(t, "identifier-duplicate", featurePath, 1), glossaryPath, entry)
	})
}

func TestLintIdentifiersSystem(t *testing.T) {
	rows(t, "system-identifier", []row{
		{name: "row1 system", tree: baseline(), holds: true},
		{name: "row2 proviso", tree: baseline().edit(systemPath, `id="system"`, `id="proviso"`),
			file: systemPath, lines: lines(1)},
		{name: "row3 system.root", tree: baseline().edit(systemPath, `id="system"`, `id="system.root"`),
			file: systemPath, lines: lines(1)},
	})
}
