package acceptance

// Feature lint.rules: proviso/capabilities/lint/features/rules.md

import (
	"strings"
	"testing"
)

const ruleHeading = `id="cart.x.r"`

// kwTree is the baseline with x.md holding a rule of the given keywords.
func kwTree(keywords ...string) tree {
	return withRule(kwBlock(keywords...))
}

// kwLine is the line of the i-th (0-based) keyword of the block in x.md.
func kwLine(tr tree, i int) int {
	open, _ := tr.fence(xPath, 1)
	return open + 1 + i
}

// blockLines is the block of x.md, from fence to fence.
func blockLines(tr tree) []int {
	open, close := tr.fence(xPath, 1)
	return lines(open, close)
}

func TestLintRulesBlock(t *testing.T) {
	t.Parallel()
	block, table := kwBlock("When", "Then", "And")
	gherkin := withX("## <a id=\"cart.x.r\"></a> Rule: Under test\n\n´´´gherkin\n" + block + "\n´´´\n\n" + table)
	two := withX("## <a id=\"cart.x.r\"></a> Rule: Under test\n\n´´´guarded\n" + block + "\n´´´\n\n´´´guarded\n" + block + "\n´´´\n\n" + table)
	rows(t, "rule-block-count", []row{
		{name: "row1 one block", tree: baseline(), holds: true},
		{name: "row2 only a gherkin block", tree: gherkin, file: xPath, lines: lines(gherkin.line(xPath, ruleHeading))},
		{name: "row3 two guarded blocks", tree: two, file: xPath, lines: lines(two.line(xPath, ruleHeading))},
	})
}

func TestLintRulesKeyword(t *testing.T) {
	t.Parallel()
	const good = `Given     the shop is open
When      a buyer adds <product>
Provided  <product> is in stock
Then      the cart holds <product>
  And     everything else stays as it was
Otherwise → cart.add.empty`
	table := "| product | provided | outcome |\n| --- | --- | --- |\n| apple | yes | added |\n| pear | no | handed over |\n"
	ok := withRule(good, table)
	but := withRule(strings.Replace(good, "  And     everything else stays as it was\n", "  And     everything else stays as it was\nBut       the cart stays as it was\n", 1), table)
	lower := withRule(strings.Replace(good, "Given     the shop", "given     the shop", 1), table)
	continued := withRule(strings.Replace(good, "Then      the cart holds <product>\n", "Then      the cart holds <product>\n          and keeps it\n", 1), table)
	rows(t, "keyword-unknown", []row{
		{name: "row1 Given", tree: ok, holds: true},
		{name: "row2 And may be indented", tree: ok, holds: true},
		{name: "row3 a hand-over", tree: ok, holds: true},
		{name: "derived: a tab before the keyword", tree: withRule(strings.Replace(good, "  And     everything", "\tAnd     everything", 1), table), holds: true},
		{name: "row4 But is not a keyword here", tree: but, file: xPath, lines: lines(but.line(xPath, "But       the cart"))},
		{name: "row5 keywords are matched exactly", tree: lower, file: xPath, lines: lines(lower.line(xPath, "given     the shop"))},
		{name: "row6 a step continued on a new line", tree: continued, file: xPath, lines: lines(continued.line(xPath, "and keeps it"))},
	})
	t.Run("rows1-3 the block is read", func(t *testing.T) {
		t.Parallel()
		check(t, ok).clean(t)
	})
}

func TestLintRulesOrder(t *testing.T) {
	t.Parallel()
	rs := []row{
		{name: "row1 the shortest rule", tree: kwTree("When", "Then", "And"), holds: true},
		{name: "row2 state in two parts, a guard in two parts", tree: kwTree("Given", "And", "When", "Provided", "And", "Then", "And", "Otherwise", "And"), holds: true},
	}
	for _, c := range []struct {
		name     string
		keywords []string
		at       int
	}{
		{"row3 error at Given", []string{"When", "Given", "Then", "And"}, 1},
		{"row4 the guard after the outcome", []string{"Given", "When", "Then", "And", "Provided"}, 4},
		{"row5 error at the first And", []string{"And", "When", "Then", "And"}, 0},
	} {
		tr := kwTree(c.keywords...)
		rs = append(rs, row{name: c.name, tree: tr, file: xPath, lines: lines(kwLine(tr, c.at))})
	}
	rows(t, "keyword-order", rs)
	t.Run("row2 is a well-formed rule", func(t *testing.T) {
		t.Parallel()
		check(t, kwTree("Given", "And", "When", "Provided", "And", "Then", "And", "Otherwise", "And")).clean(t)
	})
}

func TestLintRulesCount(t *testing.T) {
	t.Parallel()
	bad := func(name string, keywords ...string) row {
		tr := kwTree(keywords...)
		return row{name: name, tree: tr, file: xPath, lines: blockLines(tr)}
	}
	rows(t, "keyword-count", []row{
		{name: "row1 When, Then, And", tree: kwTree("When", "Then", "And"), holds: true},
		{name: "row2 a guarded rule", tree: kwTree("Given", "When", "Provided", "Then", "And", "Otherwise", "And"), holds: true},
		bad("row3 no event", "Given", "Then", "And"),
		bad("row4 two events", "When", "When", "Then", "And"),
		bad("row5 two Provided", "When", "Provided", "Provided", "Then", "And", "Otherwise", "And"),
		bad("row6 two Given", "Given", "Given", "When", "Then", "And"),
		bad("derived: no Then", "When", "Provided", "Otherwise", "And"),
		bad("derived: two Otherwise", "When", "Provided", "Then", "And", "Otherwise", "And", "Otherwise", "And"),
	})
	// lint.rules Rationale: in a block without exactly one When, the checks
	// on Given, guard and placeholders, which take the When as their
	// reference, are skipped.
	for _, c := range []struct{ name, block string }{
		{"rows3-4 no When: no checks on Given, guard or placeholders", `Given     the shop has <product>
Provided  the shop is open
Then      the cart holds <other>
  And     nothing else changes
Otherwise the cart refuses <product>
  And     nothing else changes`},
		{"rows3-4 two Whens: no checks on Given, guard or placeholders", `Given     the shop has <product>
When      a buyer adds <product>
When      a buyer pays for <product>
Provided  the shop is open
Then      the cart holds <other>
  And     nothing else changes
Otherwise the cart refuses <product>
  And     nothing else changes`},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			tr := withRule(c.block, "| product | other | provided | outcome |\n| --- | --- | --- | --- |\n| apple | pear | yes | added |\n| kiwi | pear | no | refused |\n")
			o := checkGoesOn(t, tr)
			o.has(t, "keyword-count", xPath, blockLines(tr)...)
			for _, k := range []string{"given-refers-to-event", "provided-without-event", "placeholder-unbound"} {
				o.lacks(t, k, "")
			}
		})
	}
	// lint.rules Rationale: every other check still applies to such a block.
	t.Run("rows3-4 no When: the other checks still apply", func(t *testing.T) {
		t.Parallel()
		tr := withRule(`Given     the shop has <product>
Then      the cart holds <product>
Provided  <product> is in stock
Otherwise the cart refuses <product>
  And     nothing else changes`, "| product | provided | outcome |\n| --- | --- | --- |\n| apple | yes | added |\n| pear | no | refused |\n")
		o := checkGoesOn(t, tr)
		o.has(t, "keyword-count", xPath, blockLines(tr)...)
		o.has(t, "keyword-order", xPath, tr.line(xPath, "Provided  <product>"))
		o.has(t, "and-missing", xPath, tr.line(xPath, "Then      the cart"))
	})
	t.Run("rows3-4 two Whens: the other checks still apply", func(t *testing.T) {
		t.Parallel()
		tr := withRule(`When      a buyer adds <product>
When      a buyer pays for <product>
Provided  <product> is in stock
Then      the cart holds <product>
  And     nothing else changes
  But     the shelf stays full`, "| product | provided | outcome |\n| --- | --- | --- |\n| apple | yes | added |\n| pear | no | refused |\n")
		o := checkGoesOn(t, tr)
		o.has(t, "keyword-count", xPath, blockLines(tr)...)
		o.has(t, "guard-incomplete", xPath, blockLines(tr)...)
		o.has(t, "keyword-unknown", xPath, tr.line(xPath, "But     the shelf"))
	})
}

func TestLintRulesGuard(t *testing.T) {
	t.Parallel()
	bad := func(name string, keywords ...string) row {
		tr := kwTree(keywords...)
		return row{name: name, tree: tr, file: xPath, lines: blockLines(tr)}
	}
	rows(t, "guard-incomplete", []row{
		{name: "row1 Provided and Otherwise", tree: kwTree("When", "Provided", "Then", "And", "Otherwise", "And"), holds: true},
		{name: "row2 a rule that always applies", tree: kwTree("When", "Then", "And"), holds: true},
		bad("row3 a guard without Otherwise", "When", "Provided", "Then", "And"),
		bad("row4 an Otherwise of nothing", "When", "Then", "And", "Otherwise", "And"),
	})
}

const handOverTable = "| product | provided | outcome |\n| --- | --- | --- |\n| apple | yes | added |\n| pear | no | handed over |\n"

// handOver is a guarded rule whose Otherwise is tail.
func handOver(tail string) tree {
	return withRule(`When      a buyer adds <product>
Provided  <product> is in stock
Then      the cart holds <product>
  And     nothing else changes
`+tail, handOverTable)
}

func TestLintRulesHandOver(t *testing.T) {
	t.Parallel()
	bare := handOver("Otherwise → cart.add.empty")
	one := handOver("Otherwise → cart.add.empty\n  And     everything else stays as it was")
	two := handOver("Otherwise → cart.add.empty\n  And     nothing else changes\n  And     the cart stays as it was")
	stated := handOver("Otherwise the cart refuses <product>\n  And     the cart stays as it was")
	otherwise := func(tr tree) int { return tr.line(xPath, "Otherwise") }
	rows(t, "hand-over-and", []row{
		{name: "row1 a hand-over without And lines", tree: bare, holds: true},
		{name: "row2 an And line after a hand-over", tree: one, file: xPath, lines: lines(otherwise(one) + 1)},
		{name: "row3 not a hand-over", tree: stated, holds: true},
		{name: "derived: the first of two And lines", tree: two, file: xPath, lines: lines(otherwise(two) + 1)},
		{name: "derived: the second of two And lines", tree: two, file: xPath, lines: lines(otherwise(two) + 2)},
	})
	t.Run("row1 nothing reported", func(t *testing.T) {
		t.Parallel()
		check(t, bare).clean(t)
	})
	t.Run("row3 handed over to lint.rules.and", func(t *testing.T) {
		t.Parallel()
		check(t, stated).clean(t)
	})
}

func TestLintRulesAnd(t *testing.T) {
	t.Parallel()
	noThenAnd := withRule("When      a buyer opens the new cart <cart>\nThen      <cart> is empty", "| cart | outcome |\n| --- | --- |\n| the first cart | empty |\n")
	noOtherwiseAnd := handOver("Otherwise the cart refuses <product>")
	everything := withRule("When      a buyer opens the new cart <cart>\nThen      <cart> is empty\n  And     everything else stays as it was", "| cart | outcome |\n| --- | --- |\n| the first cart | empty |\n")
	rows(t, "and-missing", []row{
		{name: "row1 Then with And", tree: baseline(), holds: true},
		{name: "row2 everything else is a statement too", tree: everything, holds: true},
		{name: "row3 And ? is undecided, not missing", tree: handOver("Otherwise the cart refuses <product>\n  And     ?"), holds: true},
		{name: "row4 Then without And", tree: noThenAnd, file: xPath, lines: lines(noThenAnd.line(xPath, "Then      <cart> is empty"))},
		{name: "row5 Otherwise without And", tree: noOtherwiseAnd, file: xPath, lines: lines(noOtherwiseAnd.line(xPath, "Otherwise the cart refuses"))},
	})
}

const givenTable = "| others | id | outcome |\n| --- | --- | --- |\n| apple, pear | apple | shown |\n"

func TestLintRulesGiven(t *testing.T) {
	t.Parallel()
	refers := withRule(`Given     the shop has <id>
When      a buyer looks up the product <id>
Then      the shop shows <id>
  And     nothing else changes`, "| id | outcome |\n| --- | --- |\n| apple | shown |\n")
	rows(t, "given-refers-to-event", []row{
		{name: "row1 a Given of its own", tree: withRule(`Given     the shop lists the products <others>
When      a buyer looks up the product <id>
Then      the shop shows <id>
  And     nothing else changes`, givenTable), holds: true},
		{name: "row2 a Given without placeholders", tree: withRule(`Given     the shop is open
When      a buyer opens the new cart <cart>
Then      <cart> is empty
  And     nothing else changes`, "| cart | outcome |\n| --- | --- |\n| the first cart | empty |\n"), holds: true},
		{name: "row3 a condition on the event", tree: refers, file: xPath, lines: lines(refers.line(xPath, "Given     the shop has <id>"))},
	})
}

func TestLintRulesProvided(t *testing.T) {
	t.Parallel()
	without := withRule(`When      a buyer adds the product <file>
Provided  the shop is open
Then      the cart holds <file>
  And     nothing else changes
Otherwise the cart refuses <file>
  And     nothing else changes`, "| file | provided | outcome |\n| --- | --- | --- |\n| apple | yes | added |\n| pear | no | refused |\n")
	rows(t, "provided-without-event", []row{
		{name: "row1 Provided <count> is 1", tree: withRule(`When      a buyer adds <count> of the product <rule>
Provided  <count> is 1
Then      the cart holds <rule>
  And     nothing else changes
Otherwise the cart refuses <rule>
  And     nothing else changes`, "| rule | count | provided | outcome |\n| --- | --- | --- | --- |\n| apple | 1 | yes | added |\n| apple | 2 | no | refused |\n"), holds: true},
		{name: "row2 a guard in two parts", tree: withRule(`When      a buyer adds the product <entry>
Provided  the name of <entry> is known
  And     <entry> is not sold out
Then      the cart holds <entry>
  And     nothing else changes
Otherwise the cart refuses <entry>
  And     nothing else changes`, "| entry | provided | outcome |\n| --- | --- | --- |\n| apple | yes | added |\n| kiwi | no (known) | refused |\n| pear | no (sold out) | refused |\n"), holds: true},
		{name: "row3 perhaps state for a Given", tree: without, file: xPath, lines: lines(without.line(xPath, "Provided  the shop is open"))},
	})
	t.Run("row3 an ask does not fail the run", func(t *testing.T) {
		t.Parallel()
		o := check(t, without)
		if o.json.exit != 0 {
			t.Errorf("want exit status 0, got %d:\n%s", o.json.exit, dump(o.report.Findings))
		}
	})
}

func TestLintRulesBound(t *testing.T) {
	t.Parallel()
	unbound := withRule(`When      a buyer reads the heading <heading>
Then      the shop shows <file>
  And     nothing else changes`, "| heading | file | outcome |\n| --- | --- | --- |\n| Fruit | apples.md | shown |\n")
	inGuard := withRule(`When      a buyer adds <product>
Provided  <amount> is at most ´max_quantity´
Then      the cart holds <product>
  And     nothing else changes
Otherwise the cart refuses <product>
  And     nothing else changes`, "| product | amount | provided | outcome |\n| --- | --- | --- | --- |\n| apple | 1 | yes | added |\n| apple | 11 | no | refused |\n")
	rows(t, "placeholder-unbound", []row{
		{name: "row1 introduced by the When", tree: withRule(idBlock, idTable), holds: true},
		{name: "row2 introduced by a Given", tree: withRule(`Given     the shop lists the products <others>
When      a buyer looks up the product <id>
Then      the shop shows <id>
  And     nothing else changes`, givenTable), holds: true},
		{name: "row3 used in an outcome, introduced nowhere", tree: unbound, file: xPath, lines: lines(unbound.line(xPath, "<file>"))},
		{name: "derived: used in the guard, introduced nowhere", tree: inGuard, file: xPath, lines: lines(inGuard.line(xPath, "<amount>"))},
	})
}
