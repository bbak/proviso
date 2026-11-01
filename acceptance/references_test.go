package acceptance

// Feature lint.references: proviso/capabilities/lint/features/references.md

import (
	"fmt"
	"strings"
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
	t.Parallel()
	sys := withSystemGlossary()
	other := withBilling(baseline(), `When      a buyer pays for the [item] <product>
Then      the shop sends a bill for <product>
  And     nothing else changes`, "| product | outcome |\n| --- | --- |\n| apple | billed |\n")
	hyphen := withThenAnd(sys, "the cart stays in the [spec-tree]")
	nbsp := withThenAnd(sys, "the cart stays in the [spec\u00a0tree]")
	underscore := withThenAnd(sys, "the cart stays in the [spec_tree]")
	// api is the tree tr with a contract element that represents mark.
	const apiPath = "proviso/capabilities/cart/contracts/api.md"
	api := func(tr tree, mark string) tree {
		return tr.with(apiPath, "# API\n\n## <a id=\"cart.order\"></a> order\n\n- Represents: "+mark+"\n\nWhat a buyer orders.\n")
	}
	represents := api(baseline(), "[basket]")
	broken := api(sys, "[spec\n  tree]")
	// Rows 5, 10, 12 and 13 need terms of the cart's own, defined as the row
	// says.
	usLaw := withTerms(baseline(), "cart", entryMD("cart.us-law", "U.S. law"))
	ubersicht := withTerms(baseline(), "cart", entryMD("cart.ubersicht", "\u00dcbersicht"))
	strasse := withThenAnd(withTerms(baseline(), "cart", entryMD("cart.strasse", "Stra\u00dfe")), "the cart stays in the [STRASSE]")
	jCaron := withTerms(baseline(), "cart", entryMD("cart.j-caron", "\u01f0"))
	zeroWidth := withThenAnd(sys, "the cart stays in the [spec tree\u200b]")
	noBreakZeroWidth := withThenAnd(sys, "the cart stays in the [spec tree\ufeff]")
	rows(t, "term-unresolved", []row{
		{name: "row1 the system's entry", tree: withThenAnd(sys, "the cart stays in the [spec tree]"), holds: true},
		{name: "row2 by an alias", tree: baseline(), holds: true},
		{name: "row3 names are matched ignoring case", tree: withThenAnd(sys, "the cart stays in the [Spec Tree]"), holds: true},
		{name: "row4 any run of spaces and tabs joins the words of a name", tree: withThenAnd(sys, "the cart stays in the [spec  tree]"), holds: true},
		{name: "row5 every run of spaces and tabs within a name counts as one, next to punctuation too",
			tree: withThenAnd(usLaw, "the cart follows [U.S.  law]"), holds: true},
		{name: "row6 a line break within a field joins words too", tree: broken, holds: true},
		{name: "row7 spaces at either end do not count", tree: withThenAnd(sys, "the cart stays in the [ spec tree ]"), holds: true},
		{name: "row8 any white space at either end does not count", tree: withThenAnd(sys, "the cart stays in the [spec tree\u00a0]"), holds: true},
		{name: "row9 a zero-width space is no white space", tree: zeroWidth, file: featurePath, lines: lines(thenAndLine(zeroWidth))},
		{name: "row10 names are compared in NFC", tree: withThenAnd(ubersicht, "the cart shows the [u\u0308bersicht]"), holds: true},
		{name: "row11 only spaces and tabs join the words of a name", tree: nbsp, file: featurePath, lines: lines(thenAndLine(nbsp))},
		{name: "row12 simple lower case, not full case folding", tree: strasse, file: featurePath, lines: lines(thenAndLine(strasse))},
		{name: "row13 lower case comes before NFC", tree: withThenAnd(jCaron, "the cart shows the [J\u030c]"), holds: true},
		{name: "row14 not visible in another capability", tree: other,
			file: "proviso/capabilities/billing/features/pay.md", lines: lines(other.line("proviso/capabilities/billing/features/pay.md", "[item]"))},
		{name: "row15 neither a name nor an alias", tree: hyphen, file: featurePath, lines: lines(thenAndLine(hyphen))},
		{name: "derived: a tab joins the words of a name", tree: withThenAnd(sys, "the cart stays in the [spec\ttree]"), holds: true},
		{name: "derived: a run of spaces and tabs joins the words of a name", tree: withThenAnd(sys, "the cart stays in the [spec \t \ttree]"), holds: true},
		{name: "derived: tabs at either end do not count", tree: withThenAnd(sys, "the cart stays in the [\tspec tree\t]"), holds: true},
		{name: "derived: other white space at the start does not count", tree: withThenAnd(sys, "the cart stays in the [\u3000\u2003spec tree]"), holds: true},
		{name: "derived: a zero-width no-break space is no white space", tree: noBreakZeroWidth,
			file: featurePath, lines: lines(thenAndLine(noBreakZeroWidth))},
		{name: "derived: a run of spaces in a glossary name counts as one", tree: withThenAnd(
			withTerms(baseline(), "cart", entryMD("cart.shelf-label", "shelf  label")), "the cart shows the [shelf label]"), holds: true},
		{name: "derived: a decomposed glossary name is compared in NFC", tree: withThenAnd(
			withTerms(baseline(), "cart", entryMD("cart.ubersicht", "U\u0308bersicht")), "the cart shows the [\u00fcbersicht]"), holds: true},
		{name: "derived: simple lower case of the capital I with dot above", tree: withThenAnd(
			withTerms(baseline(), "cart", entryMD("cart.istanbul", "istanbul")), "the cart ships to [\u0130stanbul]"), holds: true},
		{name: "derived: no other character joins the words of a name", tree: underscore, file: featurePath, lines: lines(thenAndLine(underscore))},
		{name: "derived: a Represents field that resolves", tree: api(baseline(), "[item]"), holds: true},
		{name: "derived: a Represents field that does not resolve", tree: represents,
			file: apiPath, lines: lines(represents.line(apiPath, "- Represents:"))},
	})
	t.Run("row1 resolves without other findings", func(t *testing.T) {
		t.Parallel()
		check(t, withThenAnd(sys, "the cart stays in the [spec tree]")).clean(t)
	})
	t.Run("row4 resolves without other findings", func(t *testing.T) {
		t.Parallel()
		check(t, withThenAnd(sys, "the cart stays in the [spec  tree]")).clean(t)
	})
	t.Run("row6 resolves without other findings", func(t *testing.T) {
		t.Parallel()
		check(t, broken).clean(t)
	})
}

func TestLintReferencesMarked(t *testing.T) {
	t.Parallel()
	sys := withSystemGlossary()
	unmarked := func(tr tree, text string) row {
		tr = withThenAnd(tr, text)
		return row{tree: tr, file: featurePath, lines: lines(thenAndLine(tr))}
	}
	r2 := unmarked(sys, "the cart goes on with the rest of the spec tree")
	r2.name = "row2 an unmarked use"
	r3 := unmarked(baseline(), "the cart keeps the item's price")
	r3.name = "row3 a possessive is a use"
	// Rows 5, 6, 9 and 10 need terms of two words: the cart's own "shelf
	// label" and "label", and billing's "price label" and "label date".
	shelfLabel := withTerms(baseline(), "cart", entryMD("cart.shelf-label", "shelf label"))
	r6 := unmarked(shelfLabel, "the cart keeps the shelf  label")
	r6.name = "row6 any run of spaces and tabs joins the words of a name"
	r9 := unmarked(withTerms(shelfLabel, "billing", entryMD("billing.label-date", "label date")), "the cart reads the shelf label date")
	r9.name = "row10 the match that starts first wins"
	tab := unmarked(shelfLabel, "the cart keeps the shelf\tlabel")
	tab.name = "derived: a tab joins the words of a name"
	run := unmarked(shelfLabel, "the cart keeps the shelf \t label")
	run.name = "derived: a run of spaces and tabs joins the words of a name"
	// Row 9 as written: two other capabilities, check defining "heading" and
	// bundle "date heading".
	dateHeading := withTerms(withTerms(baseline(), "check", entryMD("check.heading", "heading")),
		"bundle", entryMD("bundle.date-heading", "date heading"))
	priceLabel := withTerms(withTerms(baseline(), "cart", entryMD("cart.label", "label")),
		"billing", entryMD("billing.price-label", "price label"))
	plural := unmarked(baseline(), "the cart keeps the other items")
	plural.name = "derived: a plural, by its alias"
	upper := unmarked(baseline(), "Item counts stay as they were")
	upper.name = "derived: ignoring case"
	// Row 7: the cart defines its name with the composed U+00DC; the text
	// writes its first letter as "u" and the combining diaeresis U+0308.
	r7 := unmarked(withTerms(baseline(), "cart", entryMD("cart.ubersicht", "\u00dcbersicht")), "the cart reads the u\u0308bersicht")
	r7.name = "row7 the combining mark stays in its word, and names are compared in NFC"
	decomposed := unmarked(withTerms(baseline(), "cart", entryMD("cart.ubersicht", "U\u0308bersicht")), "the cart reads the \u00dcbersicht")
	decomposed.name = "derived: a decomposed glossary name is compared in NFC"
	jCaron := unmarked(withTerms(baseline(), "cart", entryMD("cart.j-caron", "\u01f0")), "the cart reads the J\u030c")
	jCaron.name = "derived: lower case comes before NFC"
	rows(t, "term-unmarked", []row{
		{name: "row1 the term is marked", tree: withThenAnd(sys, "the cart goes on with the rest of the [spec tree]"), holds: true},
		r2,
		r3,
		{name: "row4 a hyphen joins item-list into one word", tree: withThenAnd(baseline(), "the cart keeps the item-list"), holds: true},
		{name: "row5 no match spans the placeholder taken out", tree: withThenAnd(shelfLabel, "the cart keeps the shelf <product> label"), holds: true},
		r6,
		r7,
		{name: "row8 no term at all", tree: baseline(), holds: true},
		{name: "row9 the foreign match that starts first wins", tree: withThenAnd(dateHeading, "the cart reads the date heading"), holds: true},
		{name: "row9 over the capability's own term", tree: withThenAnd(priceLabel, "the cart shows the price label"), holds: true},
		r9,
		tab,
		run,
		{name: "derived: a non-breaking space does not join the words of a name", tree: withThenAnd(shelfLabel, "the cart keeps the shelf\u00a0label"), holds: true},
		{name: "derived: a foreign match across a run of spaces wins over the capability's own term", tree: withThenAnd(priceLabel, "the cart shows the price  label"), holds: true},
		plural,
		upper,
		decomposed,
		jCaron,
		{name: "derived: a combining mark does not end a word", tree: withThenAnd(
			withTerms(baseline(), "cart", entryMD("cart.bersicht", "bersicht")), "the cart reads the u\u0308bersicht"), holds: true},
		{name: "derived: simple lower case, not full case folding", tree: withThenAnd(
			withTerms(baseline(), "cart", entryMD("cart.strasse", "Stra\u00dfe")), "the cart reads STRASSE"), holds: true},
		{name: "derived: only whole words", tree: withThenAnd(baseline(), "the bill stays itemized"), holds: true},
		{name: "derived: an underscore is part of a word", tree: withThenAnd(baseline(), "the cart keeps item_count"), holds: true},
		{name: "derived: a literal is not a use", tree: withThenAnd(baseline(), `the cart shows "item"`), holds: true},
		{name: "derived: no match spans a literal taken out", tree: withThenAnd(shelfLabel, `the cart keeps the shelf "x" label`), holds: true},
		{name: "derived: the keyword is taken out", tree: baseline().appendTo(glossaryPath,
			"\n## <a id=\"cart.then\"></a> then\n\n- Definition: The time after paying.\n- Contact: Dana Reyes\n- Status: defined\n"), holds: true},
	})
	t.Run("derived: a word within a match is no use of its own", func(t *testing.T) {
		t.Parallel()
		// "item list" is the match; the "item" in it is not a second one.
		tr := withThenAnd(withTerms(baseline(), "cart", entryMD("cart.item-list", "item list")), "the cart keeps the item list")
		o := checkGoesOn(t, tr)
		if found := o.find("term-unmarked", featurePath, lines(thenAndLine(tr))); len(found) != 1 {
			t.Errorf("want one term-unmarked at %s:%d, got:\n%s", featurePath, thenAndLine(tr), dump(o.report.Findings))
		}
	})
	t.Run("derived: each match is reported", func(t *testing.T) {
		t.Parallel()
		tr := withThenAnd(baseline(), "the item and the items stay")
		o := checkGoesOn(t, tr)
		if found := o.find("term-unmarked", featurePath, lines(thenAndLine(tr))); len(found) != 2 {
			t.Errorf("want two term-unmarked at %s:%d, got:\n%s", featurePath, thenAndLine(tr), dump(o.report.Findings))
		}
	})
}

// entryMD is a glossary entry of the identifier id that defines name, with
// the aliases given, if any.
func entryMD(id, name string, aliases ...string) string {
	e := "\n## <a id=\"" + id + "\"></a> " + name + "\n\n- Definition: Something of the shop.\n- Contact: Dana Reyes\n- Status: defined\n"
	if len(aliases) > 0 {
		e += "- Aliases: " + strings.Join(aliases, ", ") + "\n"
	}
	return e
}

// withTerms adds entries to the glossary of the capability capID, creating
// the capability, with its capability.md, and the glossary where they do not
// exist yet.
func withTerms(tr tree, capID string, entries ...string) tree {
	dir := "proviso/capabilities/" + capID
	if _, ok := tr[dir+"/capability.md"]; !ok {
		tr = tr.with(dir+"/capability.md", "# <a id=\""+capID+"\"></a> "+strings.ToUpper(capID[:1])+capID[1:]+
			"\n\n## Goal\n\nThe shop works.\n\n## Rationale\n\nBuyers need it.\n")
	}
	if _, ok := tr[glossaryOf(capID)]; !ok {
		tr = tr.with(glossaryOf(capID), "# Glossary\n")
	}
	return tr.appendTo(glossaryOf(capID), strings.Join(entries, ""))
}

// glossaryOf is the path of the glossary of the capability capID.
func glossaryOf(capID string) string { return "proviso/capabilities/" + capID + "/glossary.md" }

// billingWith is tr with billing's rule block reading text in its Then.
func billingWith(tr tree, text string) tree {
	return withBilling(tr, "When      a buyer pays for <product>\nThen      "+text+"\n  And     nothing else changes",
		"| product | outcome |\n| --- | --- |\n| apple | billed |\n")
}

const billingPayPath = "proviso/capabilities/billing/features/pay.md"

func TestLintReferencesForeign(t *testing.T) {
	t.Parallel()
	bill := withTerms(baseline(), "billing", entryMD("billing.bill", "bill", "bills"))
	billNoAlias := withTerms(baseline(), "billing", entryMD("billing.bill", "bill"))

	// asked asserts one term-foreign at line of file, naming each entry,
	// given as capability, identifier pairs, by the line of its heading.
	asked := func(t *testing.T, tr tree, file string, line int, entries ...string) {
		t.Helper()
		o := checkGoesOn(t, tr)
		found := o.find("term-foreign", file, lines(line))
		if len(found) != 1 {
			t.Fatalf("want one term-foreign at %s:%d, got:\n%s", file, line, dump(o.report.Findings))
		}
		for i := 0; i+1 < len(entries); i += 2 {
			relates(t, found[0], glossaryOf(entries[i]), tr.line(glossaryOf(entries[i]), `id="`+entries[i+1]+`"`))
		}
	}
	// atCart asserts one term-foreign at the line text is placed on in the
	// cart, naming the entries.
	atCart := func(t *testing.T, tr tree, text string, entries ...string) {
		t.Helper()
		tr = withThenAnd(tr, text)
		asked(t, tr, featurePath, thenAndLine(tr), entries...)
	}
	notAsked := func(t *testing.T, tr tree) outcome {
		t.Helper()
		o := checkGoesOn(t, tr)
		o.lacks(t, "term-foreign", "")
		return o
	}

	t.Run("row1 no term of another capability", func(t *testing.T) {
		t.Parallel()
		notAsked(t, withThenAnd(bill, "the cart stays as it was"))
	})
	t.Run("row2 by an alias", func(t *testing.T) {
		t.Parallel()
		atCart(t, bill, "the cart keeps the bills", "billing", "billing.bill")
	})
	t.Run("row3 names are matched ignoring case", func(t *testing.T) {
		t.Parallel()
		atCart(t, billNoAlias, "the cart keeps the Bill", "billing", "billing.bill")
	})
	t.Run("row4 not a whole word", func(t *testing.T) {
		t.Parallel()
		notAsked(t, withThenAnd(billNoAlias, "the cart keeps billable goods"))
	})
	t.Run("row5 another capability uses the cart's term", func(t *testing.T) {
		t.Parallel()
		tr := billingWith(baseline(), "the shop bills the item <product>")
		asked(t, tr, billingPayPath, tr.line(billingPayPath, "the shop bills the item"), "cart", "cart.item")
	})
	t.Run("row6 a term of its own as well", func(t *testing.T) {
		t.Parallel()
		tr := withThenAnd(withTerms(billNoAlias, "cart", entryMD("cart.bill", "bill")), "the cart keeps the bill")
		o := notAsked(t, tr)
		// lint.references.marked reports it unmarked instead.
		o.has(t, "term-unmarked", featurePath, thenAndLine(tr))
	})
	t.Run("row7 two other capabilities define it", func(t *testing.T) {
		t.Parallel()
		atCart(t, withTerms(billNoAlias, "stock", entryMD("stock.bill", "bill")), "the cart keeps the bill",
			"billing", "billing.bill", "stock", "stock.bill")
	})
	t.Run("row8 the match that starts first wins over the capability's own term", func(t *testing.T) {
		t.Parallel()
		tr := withTerms(withTerms(baseline(), "cart", entryMD("cart.label", "label")), "billing", entryMD("billing.price-label", "price label"))
		atCart(t, tr, "the cart shows the price label", "billing", "billing.price-label")
	})
	t.Run("row9 the system's term is the longer match", func(t *testing.T) {
		t.Parallel()
		tr := withTerms(withSystemGlossary(), "billing", entryMD("billing.spec", "spec"))
		notAsked(t, withThenAnd(tr, "the cart goes on with the rest of the spec tree"))
	})
	t.Run("row10 the match that starts first leaves no match within it", func(t *testing.T) {
		t.Parallel()
		tr := withTerms(withTerms(baseline(), "cart", entryMD("cart.shelf-label", "shelf label")), "billing", entryMD("billing.label-date", "label date"))
		notAsked(t, withThenAnd(tr, "the cart reads the shelf label date"))
	})

	t.Run("derived: a possessive is a use", func(t *testing.T) {
		t.Parallel()
		atCart(t, billNoAlias, "the cart keeps the bill's total", "billing", "billing.bill")
	})
	t.Run("derived: each match is asked about", func(t *testing.T) {
		t.Parallel()
		tr := withThenAnd(bill, "the cart keeps the bill and the bills")
		o := checkGoesOn(t, tr)
		if found := o.find("term-foreign", featurePath, lines(thenAndLine(tr))); len(found) != 2 {
			t.Errorf("want two term-foreign at %s:%d, got:\n%s", featurePath, thenAndLine(tr), dump(o.report.Findings))
		}
	})
	t.Run("derived: a literal is not asked about", func(t *testing.T) {
		t.Parallel()
		notAsked(t, withThenAnd(billNoAlias, `the cart shows "bill"`))
	})
	t.Run("derived: a term mark is taken out", func(t *testing.T) {
		t.Parallel()
		notAsked(t, withThenAnd(billNoAlias, "the cart keeps the [bill]"))
	})
	t.Run("derived: prose outside a rule block is not asked about", func(t *testing.T) {
		t.Parallel()
		notAsked(t, billNoAlias.edit(featurePath, "Buyers put goods into the cart.\n", "Buyers put goods into the cart and get a bill.\n"))
	})
	t.Run("derived: names are compared in NFC", func(t *testing.T) {
		t.Parallel()
		tr := withTerms(baseline(), "billing", entryMD("billing.ubersicht", "\u00dcbersicht"))
		atCart(t, tr, "the cart keeps the u\u0308bersicht", "billing", "billing.ubersicht")
	})
	t.Run("derived: lower case comes before NFC", func(t *testing.T) {
		t.Parallel()
		tr := withTerms(baseline(), "billing", entryMD("billing.j-caron", "\u01f0"))
		atCart(t, tr, "the cart keeps the J\u030c", "billing", "billing.j-caron")
	})
	t.Run("derived: simple lower case, not full case folding", func(t *testing.T) {
		t.Parallel()
		tr := withTerms(baseline(), "billing", entryMD("billing.strasse", "Stra\u00dfe"))
		notAsked(t, withThenAnd(tr, "the cart keeps the STRASSE"))
	})
	t.Run("derived: an ask leaves the exit status at 0", func(t *testing.T) {
		t.Parallel()
		o := check(t, billingWith(baseline(), "the shop bills the item <product>"))
		o.has(t, "term-foreign", billingPayPath)
		for _, f := range o.report.Findings {
			if f.Kind != "term-foreign" {
				t.Errorf("want only term-foreign, got:\n%s", dump(o.report.Findings))
				break
			}
		}
		if o.json.exit != 0 {
			t.Errorf("exit status %d, want 0 for asks only", o.json.exit)
		}
	})
}

func TestLintReferencesTermAmbiguous(t *testing.T) {
	t.Parallel()
	list := baseline().appendTo(glossaryPath, "\n## <a id=\"cart.item-list\"></a> item list\n\n- Definition: The goods in the cart.\n- Contact: Dana Reyes\n- Status: defined\n- Aliases: items\n")
	redefined := withSystemGlossary().appendTo(glossaryPath, "\n## <a id=\"cart.spec-tree\"></a> spec tree\n\n- Definition: The cart's own files.\n- Contact: Dana Reyes\n- Status: defined\n")
	label := "\n## <a id=\"%s.label\"></a> label\n\n- Definition: A name on a shelf.\n- Contact: Dana Reyes\n- Status: defined\n"
	apart := withBilling(baseline(), findingsBlock, "| findings | outcome |\n| --- | --- |\n| apples | shown |\n").
		appendTo(glossaryPath, fmt.Sprintf(label, "cart")).
		with("proviso/capabilities/billing/glossary.md", "# Glossary\n"+fmt.Sprintf(label, "billing"))

	// spaced is the system's glossary beside an entry of the cart named name.
	spaced := func(name string) tree {
		return withSystemGlossary().appendTo(glossaryPath, "\n## <a id=\"cart.spec-tree\"></a> "+name+"\n\n- Definition: The cart's own files.\n- Contact: Dana Reyes\n- Status: defined\n")
	}
	// levelled asserts term-ambiguous at the cart's entry cart.spec-tree and
	// at the system's, each naming the other.
	levelled := func(t *testing.T, tr tree) {
		t.Helper()
		o := checkGoesOn(t, tr)
		own, sys := tr.line(glossaryPath, `id="cart.spec-tree"`), tr.line("proviso/glossary.md", `id="system.spec-tree"`)
		relates(t, o.has(t, "term-ambiguous", glossaryPath, own), "proviso/glossary.md", sys)
		relates(t, o.has(t, "term-ambiguous", "proviso/glossary.md", sys), glossaryPath, own)
	}

	t.Run("row1 finding", func(t *testing.T) {
		t.Parallel()
		checkGoesOn(t, baseline()).lacks(t, "term-ambiguous", "")
	})
	t.Run("row2 runs of spaces count as one", func(t *testing.T) {
		t.Parallel()
		levelled(t, spaced("Spec  Tree"))
	})
	t.Run("row3 capabilities do not see each other's terms", func(t *testing.T) {
		t.Parallel()
		checkGoesOn(t, apart).lacks(t, "term-ambiguous", "")
	})
	t.Run("row4 the same glossary", func(t *testing.T) {
		t.Parallel()
		o := checkGoesOn(t, list)
		item, other := list.line(glossaryPath, `id="cart.item"`), list.line(glossaryPath, `id="cart.item-list"`)
		relates(t, o.has(t, "term-ambiguous", glossaryPath, item), glossaryPath, other)
		relates(t, o.has(t, "term-ambiguous", glossaryPath, other), glossaryPath, item)
	})
	t.Run("row5 a capability may not redefine a system term", func(t *testing.T) {
		t.Parallel()
		levelled(t, redefined)
	})
	t.Run("derived: a tab between the words of a name", func(t *testing.T) {
		t.Parallel()
		levelled(t, spaced("spec\ttree"))
	})
	t.Run("derived: a non-breaking space does not join the words of a name", func(t *testing.T) {
		t.Parallel()
		checkGoesOn(t, spaced("spec\u00a0tree")).lacks(t, "term-ambiguous", "")
	})
	t.Run("derived: an alias with a run of spaces in the same glossary", func(t *testing.T) {
		t.Parallel()
		tr := withTerms(baseline(), "cart", entryMD("cart.shelf-label", "shelf label"), entryMD("cart.tag", "tag", "Shelf  Label"))
		o := checkGoesOn(t, tr)
		label, tag := tr.line(glossaryPath, `id="cart.shelf-label"`), tr.line(glossaryPath, `id="cart.tag"`)
		relates(t, o.has(t, "term-ambiguous", glossaryPath, label), glossaryPath, tag)
		relates(t, o.has(t, "term-ambiguous", glossaryPath, tag), glossaryPath, label)
	})
	// twins asserts term-ambiguous at the cart's entries a and b, each naming
	// the other.
	twins := func(t *testing.T, tr tree, a, b string) {
		t.Helper()
		o := checkGoesOn(t, tr)
		la, lb := tr.line(glossaryPath, `id="`+a+`"`), tr.line(glossaryPath, `id="`+b+`"`)
		relates(t, o.has(t, "term-ambiguous", glossaryPath, la), glossaryPath, lb)
		relates(t, o.has(t, "term-ambiguous", glossaryPath, lb), glossaryPath, la)
	}
	t.Run("derived: white space at the end of a name does not count", func(t *testing.T) {
		t.Parallel()
		levelled(t, spaced("spec tree\u00a0"))
	})
	t.Run("derived: names are compared in NFC", func(t *testing.T) {
		t.Parallel()
		twins(t, withTerms(baseline(), "cart", entryMD("cart.ubersicht", "\u00dcbersicht"), entryMD("cart.overview", "overview", "U\u0308bersicht")),
			"cart.ubersicht", "cart.overview")
	})
	t.Run("derived: lower case comes before NFC", func(t *testing.T) {
		t.Parallel()
		twins(t, withTerms(baseline(), "cart", entryMD("cart.j-caron", "\u01f0"), entryMD("cart.jay", "jay", "J\u030c")),
			"cart.j-caron", "cart.jay")
	})
	t.Run("derived: simple lower case, not full case folding", func(t *testing.T) {
		t.Parallel()
		tr := withTerms(baseline(), "cart", entryMD("cart.strasse", "Stra\u00dfe"), entryMD("cart.street", "street", "STRASSE"))
		checkGoesOn(t, tr).lacks(t, "term-ambiguous", "")
	})
}

func TestLintReferencesHandOver(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	self := withX(chainRule("cart.x.a", "cart.x.a"))
	loop := withX(chainRule("cart.x.a", "cart.x.b") + chainRule("cart.x.b", "cart.x.a"))
	ruleA := func(tr tree) []int { return lines(tr.line(xPath, `id="cart.x.a"`)) }
	rows(t, "hand-over-cycle", []row{
		{name: "row1 ends in an outcome", tree: withX(chainRule("cart.x.a", "cart.x.b") + chainRule("cart.x.b", "cart.add.empty")), holds: true},
		{name: "row2 a rule handing over to itself", tree: self, file: xPath, lines: ruleA(self)},
		{name: "row3 a, b, a", tree: loop, file: xPath, lines: ruleA(loop)},
	})
	t.Run("row3 reported once per cycle, with b in related", func(t *testing.T) {
		t.Parallel()
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
	t.Parallel()
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
		t.Parallel()
		for _, r := range []string{"findings.anchor-misplaced", "lint.exit-status", "max_quantity"} {
			check(t, ref(r)).clean(t)
		}
	})
}

func TestLintReferencesReferenceAmbiguous(t *testing.T) {
	t.Parallel()
	report := "# Report\n\n## <a id=\"%s\"></a> report\n\nWhat the shop writes.\n"
	contract := "proviso/capabilities/cart/contracts/report.md"
	sameLevel := withContracts().
		with(contract, fmt.Sprintf(report, "cart.report")).
		appendTo(paramsPath, "\n## <a id=\"cart.report-size\"></a> report\n\n- Value: 1\n- Configurable: no\n- Rationale: one report a day.\n")
	levels := withContracts().
		with(contract, fmt.Sprintf(report, "cart.report")).
		with("proviso/contracts/report.md", fmt.Sprintf(report, "system.report"))

	t.Run("row1 findings", func(t *testing.T) {
		t.Parallel()
		checkGoesOn(t, withContracts()).lacks(t, "reference-ambiguous", "")
	})
	t.Run("row2 contracts and parameters share one namespace", func(t *testing.T) {
		t.Parallel()
		o := checkGoesOn(t, sameLevel)
		c, p := sameLevel.line(contract, `id="cart.report"`), sameLevel.line(paramsPath, `id="cart.report-size"`)
		relates(t, o.has(t, "reference-ambiguous", contract, c), paramsPath, p)
		relates(t, o.has(t, "reference-ambiguous", paramsPath, p), contract, c)
	})
	t.Run("row3 no shadowing", func(t *testing.T) {
		t.Parallel()
		o := checkGoesOn(t, levels)
		c, s := levels.line(contract, `id="cart.report"`), levels.line("proviso/contracts/report.md", `id="system.report"`)
		relates(t, o.has(t, "reference-ambiguous", contract, c), "proviso/contracts/report.md", s)
		relates(t, o.has(t, "reference-ambiguous", "proviso/contracts/report.md", s), contract, c)
	})
}
