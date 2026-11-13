package acceptance

// Feature lint.identifiers: proviso/capabilities/lint/features/identifiers.md

import (
	"fmt"
	"testing"
)

const limitHeading = `## <a id="cart.add.limit"></a> Rule: The cart takes a limited quantity`

func TestLintIdentifiersPlacement(t *testing.T) {
	t.Parallel()
	before := baseline().edit(featurePath, limitHeading, `<a id="cart.add.limit"></a>`+"\n"+`## Rule: The cart takes a limited quantity`)
	inside := baseline().edit(featurePath, limitHeading, `## Rule: The cart takes a limited quantity <a id="cart.add.limit"></a>`)
	paragraph := baseline().edit(featurePath, "A cart that takes anything would", `A cart that takes anything <a id="cart.add.why"></a> would`)
	upperID := baseline().edit(featurePath, `<a id="cart.add.limit"></a>`, `<a ID="cart.add.limit"></a>`)
	rows(t, "anchor-misplaced", []row{
		{name: "row1 the prescribed form", tree: baseline(), holds: true},
		{name: "row2 no space after the anchor", tree: baseline().edit(featurePath, `</a> Rule: The cart takes`, `</a>Rule: The cart takes`), holds: true},
		{name: "row3 HTML attribute names ignore case", tree: upperID, holds: true},
		{name: "row4 on the line before the heading", tree: before,
			file: featurePath, lines: lines(before.line(featurePath, `<a id="cart.add.limit"></a>`))},
		{name: "row5 inside the heading, not at its start", tree: inside,
			file: featurePath, lines: lines(inside.line(featurePath, `<a id="cart.add.limit"></a>`))},
		{name: "row6 within a paragraph", tree: paragraph,
			file: featurePath, lines: lines(paragraph.line(featurePath, `<a id="cart.add.why"></a>`))},
		{name: "derived: in a code span it is text", tree: baseline().edit(featurePath, "A cart that takes anything would",
			"A cart that takes ´<a id=\"cart.add.why\"></a>´ would"), holds: true},
		{name: "derived: in a fenced block it is text", tree: baseline().edit(featurePath, "A cart that takes anything would empty the shelves.\n",
			"A cart that takes anything would empty the shelves.\n\n´´´html\n<a id=\"cart.add.why\"></a>\n´´´\n"), holds: true},
	})
	t.Run("row2 opens the element", func(t *testing.T) {
		t.Parallel()
		check(t, baseline().edit(featurePath, `</a> Rule: The cart takes`, `</a>Rule: The cart takes`)).clean(t)
	})
	t.Run("row3 opens the element", func(t *testing.T) {
		t.Parallel()
		check(t, upperID).clean(t)
	})
}

func TestLintIdentifiersForm(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	t.Run("row1 lint.identifiers among lint, lint.run", func(t *testing.T) {
		t.Parallel()
		checkGoesOn(t, baseline()).lacks(t, "identifier-duplicate", "")
	})
	t.Run("row2 reported at each of the two", func(t *testing.T) {
		t.Parallel()
		tr := baseline().edit(featurePath, `id="cart.add.empty"`, `id="cart.add.limit"`)
		first := tr.line(featurePath, "Rule: The cart takes")
		second := tr.line(featurePath, "Rule: A new cart is empty")
		o := checkGoesOn(t, tr)
		relates(t, o.has(t, "identifier-duplicate", featurePath, first), featurePath, second)
		relates(t, o.has(t, "identifier-duplicate", featurePath, second), featurePath, first)
	})
	t.Run("row3 a capability identified as system", func(t *testing.T) {
		t.Parallel()
		o := checkGoesOn(t, baseline().edit(capPath, `id="cart"`, `id="system"`))
		relates(t, o.has(t, "identifier-duplicate", capPath, 1), systemPath, 1)
		relates(t, o.has(t, "identifier-duplicate", systemPath, 1), capPath, 1)
	})
	t.Run("derived: in two files", func(t *testing.T) {
		t.Parallel()
		tr := baseline().edit(glossaryPath, `id="cart.item"`, `id="cart.add"`)
		entry := tr.line(glossaryPath, `id="cart.add"`)
		o := checkGoesOn(t, tr)
		relates(t, o.has(t, "identifier-duplicate", glossaryPath, entry), featurePath, 1)
		relates(t, o.has(t, "identifier-duplicate", featurePath, 1), glossaryPath, entry)
	})
}

func TestLintIdentifiersSystem(t *testing.T) {
	t.Parallel()
	rows(t, "system-identifier", []row{
		{name: "row1 system", tree: baseline(), holds: true},
		{name: "row2 proviso", tree: baseline().edit(systemPath, `id="system"`, `id="proviso"`),
			file: systemPath, lines: lines(1)},
		{name: "row3 system.root", tree: baseline().edit(systemPath, `id="system"`, `id="system.root"`),
			file: systemPath, lines: lines(1)},
	})
}

func TestLintIdentifiersCapability(t *testing.T) {
	t.Parallel()
	rows(t, "capability-identifier", []row{
		// The baseline's features and rules have several segments, and
		// are no capabilities: nothing is reported anywhere.
		{name: "row1 coherence", tree: baseline(), holds: true},
		{name: "row2 cash.withdrawal", tree: baseline().edit(capPath, `id="cart"`, `id="cart.shop"`),
			file: capPath, lines: lines(1)},
		// identifier-duplicate reports it (TestLintIdentifiersUnique).
		{name: "row3 system", tree: baseline().edit(capPath, `id="cart"`, `id="system"`), holds: true},
		{name: "derived: three segments, in a second capability", tree: withBilling(baseline(), idBlock, idTable).
			edit("proviso/capabilities/billing/capability.md", `id="billing"`, `id="billing.pay.now"`),
			file: "proviso/capabilities/billing/capability.md", lines: lines(1)},
	})
}

// TestLintIdentifiersHome checks lint.identifiers.home: an identifier whose
// first segment names no place, neither system nor the first segment of a
// capability's identifier, is asked about, but where another rule reports
// the identifier. Every row compares the findings identifier-unhomed with
// the oracle unhomed as well.
func TestLintIdentifiersHome(t *testing.T) {
	t.Parallel()
	const sysGlossaryPath = "proviso/glossary.md"
	payPath := "proviso/capabilities/billing/features/pay.md"
	billing := withBilling(baseline(), idBlock, idTable)
	for _, c := range []struct {
		name  string
		tree  tree
		file  string
		id    string // the identifier the row is about, at the start of its heading
		holds bool
	}{
		// lint.run.root: cart.add.limit, a rule of the capability cart.
		{"row1 lint.run.root, a rule of the capability lint", baseline(), featurePath, "cart.add.limit", true},
		{"row2 system.read-only, an invariant of the system", baseline().with("proviso/invariants.md", invariantsMD),
			"proviso/invariants.md", "system.read-only", true},
		// coherence.impact.events: billing.moved, a rule of the cart, while
		// a capability is identified as billing.
		{"row3 coherence.impact.events, a rule of the capability baseline, while a capability is identified as coherence",
			billing.with(xPath, xHead+ruleMD("billing.moved", "Moved here", idBlock, idTable)), xPath, "billing.moved", true},
		{"row4 system", baseline(), systemPath, "system", true},
		{"row5 proviso, the system's identifier", baseline().edit(systemPath, `id="system"`, `id="proviso"`), systemPath, "proviso", true},
		{"row6 Lint.run, a rule of the capability lint", baseline().edit(featurePath, `id="cart.add.limit"`, `id="Cart.add.limit"`),
			featurePath, "Cart.add.limit", true},
		{"row7 lint, the capability so identified", baseline(), capPath, "cart", true},
		// cash.withdrawal.pay: cart.add.limit, while the capability is
		// identified as cart.shop.
		{"row8 cash.withdrawal.pay, while a capability is identified as cash.withdrawal",
			baseline().edit(capPath, `id="cart"`, `id="cart.shop"`), featurePath, "cart.add.limit", true},
		{"row9 lint.run.exit, in the folder capabilities/lint/, which holds no capability.md", baseline().without(capPath),
			featurePath, "cart.add.limit", true},
		{"row10 coherence.x, in a capability identified as Coherence", baseline().edit(capPath, `id="cart"`, `id="Cart"`),
			featurePath, "cart.add.limit", true},
		// okf-get: a contract of the cart.
		{"row11 okf-get, a contract element of the capability access", baseline().with(placeAPIPath, placeAPIMD),
			placeAPIPath, "okf-get", false},
		{"row12 severity.error, a member of a contract of the system", baseline().with(placeFindingsPath, placeFindingsMD),
			placeFindingsPath, "severity.error", false},
		// checks.run: billing.pay, the feature in the folder
		// capabilities/billing/, whose capability.md is identified as pay.
		{"row13 checks.run, in the folder capabilities/checks/, whose capability.md is identified as coherence, no capability identified as checks",
			billing.edit("proviso/capabilities/billing/capability.md", `id="billing"`, `id="pay"`), payPath, "billing.pay", false},
		{"row14 systems.x, no capability identified as systems", baseline().with(sysGlossaryPath, `# Glossary

## <a id="systems.x"></a> shelf

- Definition: Where goods stand.
- Contact: Dana Reyes
- Status: defined
`), sysGlossaryPath, "systems.x", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			o := checkGoesOn(t, c.tree)
			line := c.tree.line(c.file, `id="`+c.id+`"`)
			if c.holds {
				o.lacks(t, "identifier-unhomed", c.file, line)
			} else {
				o.has(t, "identifier-unhomed", c.file, line)
			}
			// Exactly the identifiers the oracle gives, one finding each.
			want := map[string]int{}
			for _, u := range unhomed(c.tree) {
				want[fmt.Sprintf("%s:%d", u[1], c.tree.line(u[1], `id="`+u[0]+`"`))]++
			}
			got := map[string]int{}
			for _, f := range o.find("identifier-unhomed", "", nil) {
				got[fmt.Sprintf("%s:%d", f.File, f.Line)]++
			}
			if fmt.Sprint(got) != fmt.Sprint(want) {
				t.Errorf("lint.identifiers.home: want identifier-unhomed at %v, got %v", want, got)
			}
		})
	}
	t.Run("derived: an ask, lint exits with 0", func(t *testing.T) {
		t.Parallel()
		tr := baseline().with(placeAPIPath, placeAPIMD)
		o := check(t, tr)
		o.has(t, "identifier-unhomed", placeAPIPath, tr.line(placeAPIPath, `id="okf-get"`))
		if o.json.exit != 0 {
			t.Errorf("lint.identifiers.home: an ask, but lint exits with %d", o.json.exit)
		}
	})
}
