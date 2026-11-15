package acceptance

// Feature lint.undecided: proviso/capabilities/lint/features/undecided.md

import "testing"

// TestLintUndecidedPlace holds the rows of lint.undecided.place: a field
// value, an Examples cell, a section body and a word of a rule block.
func TestLintUndecidedPlace(t *testing.T) {
	t.Parallel()
	// Fields.
	contactRow := func(name, value string) row {
		tr := baseline().edit(featurePath, "- Contact: Dana Reyes", "- Contact:"+value)
		return row{name: name, tree: tr, file: featurePath, lines: lines(3)}
	}
	open := withEntry("- Definition: ?\n" + contact + "- Status: open\n" + question)
	// Examples cells.
	cell := func(value string) tree {
		return baseline().edit(featurePath, "| 11 | apple | no | refused |", "| 11 | apple | no | "+value+" |")
	}
	atCell := func(tr tree) []int { return lines(tr.line(featurePath, "| 11 | apple |")) }
	cellQ, cellNote := cell("?"), cell("? depends on the limit")
	// Section bodies.
	body := func(text string) tree {
		return baseline().edit(featurePath, "Buyers put goods into the cart.", text)
	}
	// The section: its heading and its body.
	atSection := func(tr tree) []int {
		h := tr.line(featurePath, "## Goal")
		return lines(h, h+2)
	}
	bodyQ, bodyNote := body("?"), body("? to be agreed with the business")
	// Words of a rule block.
	literal := withRule(`When      a buyer enters the code <value>
Provided  <value> is "yes", "no", "no (<part>)" or "?"
Then      the shop accepts <value>
  And     nothing else changes
Otherwise the shop rejects <value>
  And     nothing else changes`, "| value | provided | outcome |\n| --- | --- | --- |\n| yes | yes | accepted |\n| maybe | no | rejected |\n")
	then := withRule("When      a buyer adds <product>\nThen      ?\n  And     nothing else changes", "| product | outcome |\n| --- | --- |\n| apple | added |\n")
	and := withRule("When      a buyer adds <product>\nThen      the cart holds <product>\n  And     ?", "| product | outcome |\n| --- | --- |\n| apple | added |\n")
	comparison := withRule(`When      a buyer adds <count> pieces
Provided  <count> ? the [team limit]
Then      the cart holds <count> pieces
  And     nothing else changes
Otherwise the cart refuses <count> pieces
  And     nothing else changes`, "| count | provided | outcome |\n| --- | --- | --- |\n| 1 | yes | added |\n| 99 | no | refused |\n").
		appendTo(glossaryPath, "\n## <a id=\"cart.team-limit\"></a> team limit\n\n- Definition: The most a team may order.\n- Contact: Dana Reyes\n- Status: defined\n")
	rows(t, "undecided", []row{
		{name: "row1 the field Contact: Dana Reyes", tree: baseline(), holds: true},
		{name: "row3 the field Definition: ?", tree: open, file: glossaryPath, lines: lines(open.line(glossaryPath, "- Definition: ?"))},
		contactRow("row4 the field Contact: ?", " ?"),
		contactRow("row5 a note after the mark leaves the value undecided", " ? (Cash Operations or Treasury)"),
		contactRow("row6 surrounding whitespace does not matter", "   ?   "),
		{name: "row7 the cell error", tree: cell("error"), holds: true},
		{name: "row8 a value shown in a code span", tree: cell("´?´"), holds: true},
		{name: "row9 the cell ?", tree: cellQ, file: featurePath, lines: atCell(cellQ)},
		{name: "row10 the cell with a note after the mark", tree: cellNote, file: featurePath, lines: atCell(cellNote)},
		{name: "row11 the cell ?? is not the mark", tree: cell("??"), holds: true},
		{name: "row12 the body prose", tree: body("Proviso keeps one specification of a system true."), holds: true},
		{name: "row13 a real question in prose", tree: body("Why lowercase? Because links on some hosts differ."), holds: true},
		{name: "row14 the body ?", tree: bodyQ, file: featurePath, lines: atSection(bodyQ)},
		{name: "row15 the body with a note after the mark", tree: bodyNote, file: featurePath, lines: atSection(bodyNote)},
		{name: "row16 the words of Then lint reads <file>", tree: baseline(), holds: true},
		{name: "row18 the ? of Then ?", tree: then, file: xPath, lines: lines(then.line(xPath, "Then      ?"))},
		{name: "row19 the ? of And ?", tree: and, file: xPath, lines: lines(and.line(xPath, "  And     ?"))},
		{name: "row20 the comparison itself is open", tree: comparison, file: xPath, lines: lines(comparison.line(xPath, "<count> ?"))},
		{name: "derived: a word is what whitespace separates", tree: baseline().edit(featurePath, "- Contact: Dana Reyes", "- Contact: ?(Treasury)"), holds: true},
	})
	t.Run("row2 a question mark ending a sentence", func(t *testing.T) {
		t.Parallel()
		o := checkGoesOn(t, open)
		o.lacks(t, "undecided", glossaryPath, open.line(glossaryPath, "- Question:"))
	})
	t.Run("row17 a ? inside a literal is a value", func(t *testing.T) {
		t.Parallel()
		checkGoesOn(t, literal).lacks(t, "undecided", xPath, literal.line(xPath, "Provided"))
	})
	t.Run("derived: a ? in a backtick reference", func(t *testing.T) {
		t.Parallel()
		tr := withRule("When      a buyer adds <product>\nThen      the cart holds <product>, see ´?´\n  And     nothing else changes", "| product | outcome |\n| --- | --- |\n| apple | added |\n")
		checkGoesOn(t, tr).lacks(t, "undecided", xPath, tr.line(xPath, "Then"))
	})
}

func TestLintUndecidedTerm(t *testing.T) {
	t.Parallel()
	mark := func(status, extra string) tree {
		return baseline().
			appendTo(glossaryPath, "\n## <a id=\"cart.daily-limit\"></a> daily limit\n\n"+extra+"- Contact: Dana Reyes\n- Status: "+status+"\n").
			edit(featurePath, limitThenAnd, "  And     the cart keeps to the [daily limit]\nOtherwise the cart refuses")
	}
	open := mark("open", "- Definition: ?\n- Question: What does the daily limit apply to?\n")
	defined := mark("defined", "- Definition: The most a buyer may spend in a day.\n")
	rows(t, "undecided", []row{
		{name: "row1 a defined term", tree: defined, holds: true},
		{name: "row2 the rule waits for the term's meaning", tree: open, file: featurePath, lines: lines(open.line(featurePath, "[daily limit]"))},
	})
	t.Run("row2 naming the entry in related", func(t *testing.T) {
		t.Parallel()
		o := check(t, open)
		f := o.has(t, "undecided", featurePath, open.line(featurePath, "[daily limit]"))
		relates(t, f, glossaryPath, open.line(glossaryPath, `id="cart.daily-limit"`))
	})
	t.Run("row1 nothing reported", func(t *testing.T) {
		t.Parallel()
		check(t, defined).clean(t)
	})
}
