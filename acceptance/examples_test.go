package acceptance

// Feature lint.examples: proviso/capabilities/lint/features/examples.md

import (
	"strings"
	"testing"
)

// tableLines is the table after the block of x.md, from header to last row.
func tableLines(tr tree) []int {
	_, close := tr.fence(xPath, 1)
	first, last := tr.table(xPath, close)
	return lines(first, last)
}

// headerLine is the header row of the table after the block of x.md.
func headerLine(tr tree) int {
	return tableLines(tr)[0]
}

// A rule without a guard, on the placeholder <findings>.
const findingsBlock = `When      a buyer lists the goods <findings>
Then      the shop shows <findings>
  And     nothing else changes`

func TestLintExamplesPresent(t *testing.T) {
	noTable := withX(ruleMD("cart.x.r", "Under test", findingsBlock, ""))
	before := withX("## <a id=\"cart.x.r\"></a> Rule: Under test\n\n| findings | outcome |\n| --- | --- |\n| apples | shown |\n\n´´´guarded\n" + findingsBlock + "\n´´´\n")
	rows(t, "examples-missing", []row{
		{name: "row1 a table after the block", tree: baseline(), holds: true},
		{name: "row2 a block and no table", tree: noTable, file: xPath, lines: lines(noTable.line(xPath, ruleHeading))},
		{name: "row3 a table only before its block", tree: before, file: xPath, lines: lines(before.line(xPath, ruleHeading))},
	})
}

func TestLintExamplesReserved(t *testing.T) {
	reserved := func(name string) tree {
		return withRule(`When      a buyer adds <product> on <`+name+`>
Provided  <product> is in stock
Then      the cart holds <product>
  And     nothing else changes
Otherwise the cart refuses <product>
  And     nothing else changes`, "| product | "+name+" | provided | outcome |\n| --- | --- | --- | --- |\n| apple | x | yes | added |\n| pear | x | no | refused |\n")
	}
	at := func(tr tree) []int { return lines(tr.line(xPath, "When      a buyer adds")) }
	provided, outcome, why := reserved("provided"), reserved("outcome"), reserved("why")
	rows(t, "placeholder-reserved", []row{
		{name: "row1 <guard>", tree: reserved("guard"), holds: true},
		{name: "row2 <provided>", tree: provided, file: xPath, lines: at(provided)},
		{name: "row3 <outcome>", tree: outcome, file: xPath, lines: at(outcome)},
		{name: "derived: <why>", tree: why, file: xPath, lines: at(why)},
	})
}

func TestLintExamplesColumns(t *testing.T) {
	othersBlock := `Given     the shop lists the products <others>
When      a buyer looks up the product <id>
Provided  <id> is listed
Then      the shop shows <id>
  And     nothing else changes
Otherwise the shop shows nothing
  And     nothing else changes`
	noID := withRule(idBlock, "| provided | outcome | why |\n| --- | --- | --- |\n| yes | shown | |\n| no | nothing shown | |\n")
	asFound := withRule(findingsBlock, "| findings, as found | outcome |\n| --- | --- |\n| apples | shown |\n")
	noOutcome := withRule(idBlock, "| id | provided | why |\n| --- | --- | --- |\n| apple | yes | |\n| pear | no | |\n")
	rows(t, "examples-column-missing", []row{
		{name: "row1 id, provided, outcome, why", tree: withRule(idBlock, idTable), holds: true},
		{name: "row2 a Given's placeholder too", tree: withRule(othersBlock,
			"| others | id | provided | outcome | why |\n| --- | --- | --- | --- | --- |\n| apple, pear | apple | yes | shown | |\n| apple | kiwi | no | nothing shown | |\n"), holds: true},
		{name: "row3 no column for <id>", tree: noID, file: xPath, lines: tableLines(noID)},
		{name: "row4 a column is named exactly as its placeholder", tree: asFound, file: xPath, lines: tableLines(asFound)},
		{name: "row5 no outcome", tree: noOutcome, file: xPath, lines: tableLines(noOutcome)},
	})
	t.Run("row1 is a well-formed rule", func(t *testing.T) {
		check(t, withRule(idBlock, idTable)).clean(t)
	})
}

func TestLintExamplesProvidedColumn(t *testing.T) {
	missing := withRule(idBlock, "| id | outcome |\n| --- | --- |\n| apple | shown |\n| pear | nothing shown |\n")
	extra := withRule(findingsBlock, "| findings | provided | outcome |\n| --- | --- | --- |\n| apples | yes | shown |\n")
	rows(t, "provided-column", []row{
		{name: "row1 a guarded rule with provided", tree: withRule(idBlock, idTable), holds: true},
		{name: "row2 a rule that always applies", tree: withRule(findingsBlock, "| findings | outcome |\n| --- | --- |\n| apples | shown |\n"), holds: true},
		{name: "row3 the guard's truth value is missing", tree: missing, file: xPath, lines: tableLines(missing)},
		{name: "row4 there is no guard to be true or false", tree: extra, file: xPath, lines: tableLines(extra)},
	})
}

func TestLintExamplesUnknownColumn(t *testing.T) {
	extra := func(column string) tree {
		return withRule(idBlock, "| id | provided | outcome | "+column+" |\n| --- | --- | --- | --- |\n| apple | yes | shown | x |\n| pear | no | nothing shown | x |\n")
	}
	expected, notes := extra("expected"), extra("notes")
	rows(t, "examples-column-unknown", []row{
		{name: "row1 id", tree: withRule(idBlock, idTable), holds: true},
		{name: "row2 why", tree: withRule(idBlock, idTable), holds: true},
		{name: "row3 an outcome under another name", tree: expected, file: xPath, lines: lines(headerLine(expected))},
		{name: "row4 notes", tree: notes, file: xPath, lines: lines(headerLine(notes))},
	})
}

func TestLintExamplesRows(t *testing.T) {
	empty := withRule(findingsBlock, "| findings | outcome |\n| --- | --- |\n")
	rows(t, "examples-empty", []row{
		{name: "row1 one row", tree: withRule(findingsBlock, "| findings | outcome |\n| --- | --- |\n| apples | shown |\n"), holds: true},
		{name: "row2 a header alone", tree: empty, file: xPath, lines: tableLines(empty)},
	})
}

// withRow is the rule on <id> with one row under test, besides a failing row.
func withRow(cells string) tree {
	return withRule(idBlock, "| id | provided | outcome | why |\n| --- | --- | --- | --- |\n"+cells+"\n| pear | no | nothing shown | |\n")
}

func TestLintExamplesCells(t *testing.T) {
	short := withRow("| system | yes | nothing reported |")
	empty := withRow("| system | yes | | |")
	at := func(tr tree) []int { return lines(tr.line(xPath, "| system |")) }
	rows(t, "examples-row-incomplete", []row{
		{name: "row1 why may stay empty", tree: withRow("| system | yes | nothing reported | |"), holds: true},
		{name: "row2 an undecided outcome", tree: withRow("| system | yes | ? | |"), holds: true},
		{name: "row3 a cell short", tree: short, file: xPath, lines: at(short)},
		{name: "row4 an outcome left empty", tree: empty, file: xPath, lines: at(empty)},
	})
}

func TestLintExamplesProvidedValue(t *testing.T) {
	value := func(v string) tree { return withRow("| system | " + v + " | shown | |") }
	at := func(tr tree) []int { return lines(tr.line(xPath, "| system |")) }
	isTrue, upper := value("true"), value("Yes")
	rows(t, "provided-value-unknown", []row{
		{name: "row1 yes", tree: value("yes"), holds: true},
		{name: "row2 no (level 1)", tree: value("no (level 1)"), holds: true},
		{name: "row3 ? is undecided", tree: value("?"), holds: true},
		{name: "row4 true", tree: isTrue, file: xPath, lines: at(isTrue)},
		{name: "row5 values are matched exactly", tree: upper, file: xPath, lines: at(upper)},
	})
}

// withValues is the rule on <id> with the given provided column.
func withValues(values ...string) tree {
	var b strings.Builder
	b.WriteString("| id | provided | outcome | why |\n| --- | --- | --- | --- |\n")
	for i, v := range values {
		b.WriteString("| product" + string(rune('a'+i)) + " | " + v + " | shown | |\n")
	}
	return withRule(idBlock, b.String())
}

func TestLintExamplesSides(t *testing.T) {
	bad := func(name string, values ...string) row {
		tr := withValues(values...)
		return row{name: name, tree: tr, file: xPath, lines: tableLines(tr)}
	}
	rows(t, "examples-side-missing", []row{
		{name: "row1 yes, no", tree: withValues("yes", "no"), holds: true},
		{name: "row2 yes, yes, no (level 1)", tree: withValues("yes", "yes", "no (level 1)"), holds: true},
		bad("row3 never failing", "yes", "yes"),
		bad("row4 never holding", "no", "no"),
		bad("row5 an undecided row shows neither side", "yes", "?"),
	})
}

func TestLintExamplesParts(t *testing.T) {
	twoParts := func(values ...string) tree {
		var b strings.Builder
		b.WriteString("| id | provided | outcome | why |\n| --- | --- | --- | --- |\n")
		for i, v := range values {
			b.WriteString("| product" + string(rune('a'+i)) + " | " + v + " | shown | |\n")
		}
		return withRule(`When      a buyer looks up the product <id>
Provided  <id> is listed
  And     <id> is in stock
Then      the shop shows <id>
  And     nothing else changes
Otherwise the shop shows nothing
  And     nothing else changes`, b.String())
	}
	missing := twoParts("yes", "no (level 1)")
	rows(t, "examples-part-missing", []row{
		{name: "row1 one part, one no", tree: withValues("yes", "no"), holds: true},
		{name: "row2 two parts, two no", tree: twoParts("yes", "no (level 1)", "no (no anchor)"), holds: true},
		{name: "row3 the second part never shown failing", tree: missing, file: xPath, lines: tableLines(missing)},
	})
	t.Run("row2 is a well-formed rule", func(t *testing.T) {
		check(t, twoParts("yes", "no (level 1)", "no (no anchor)")).clean(t)
	})
}
