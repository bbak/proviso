package acceptance

// Feature lint.undecided: proviso/capabilities/lint/features/undecided.md

import "testing"

func TestLintUndecidedField(t *testing.T) {
	contactRow := func(name, value string) row {
		tr := baseline().edit(featurePath, "- Contact: Dana Reyes", "- Contact:"+value)
		return row{name: name, tree: tr, file: featurePath, lines: lines(3)}
	}
	open := withEntry("- Definition: ?\n" + contact + "- Status: open\n" + question)
	rows(t, "undecided", []row{
		{name: "row1 Contact: Dana Reyes", tree: baseline(), holds: true},
		{name: "row3 an open entry's definition", tree: open, file: glossaryPath, lines: lines(open.line(glossaryPath, "- Definition: ?"))},
		contactRow("row4 Contact: ?", " ?"),
		contactRow("row5 a note after the mark", " ? (Cash Operations or Treasury)"),
		contactRow("row6 surrounding whitespace does not matter", "   ?   "),
	})
	t.Run("row2 a question mark ending a sentence", func(t *testing.T) {
		o := checkGoesOn(t, open)
		o.lacks(t, "undecided", glossaryPath, open.line(glossaryPath, "- Question:"))
	})
}

func TestLintUndecidedCell(t *testing.T) {
	cell := func(value string) tree {
		return baseline().edit(featurePath, "| 11 | apple | no | refused |", "| 11 | apple | no | "+value+" |")
	}
	at := func(tr tree) []int { return lines(tr.line(featurePath, "| 11 | apple |")) }
	q, note := cell("?"), cell("? depends on the limit")
	rows(t, "undecided", []row{
		{name: "row1 error", tree: cell("error"), holds: true},
		{name: "row2 a value shown in a code span", tree: cell("´?´"), holds: true},
		{name: "row3 an outcome nobody has decided", tree: q, file: featurePath, lines: at(q)},
		{name: "row4 a note after the mark", tree: note, file: featurePath, lines: at(note)},
		{name: "row5 ?? is not the mark", tree: cell("??"), holds: true},
	})
}

func TestLintUndecidedSection(t *testing.T) {
	body := func(text string) tree {
		return baseline().edit(featurePath, "Buyers put goods into the cart.", text)
	}
	// The section: its heading and its body.
	at := func(tr tree) []int {
		h := tr.line(featurePath, "## Goal")
		return lines(h, h+2)
	}
	q, note := body("?"), body("? to be agreed with the business")
	rows(t, "undecided", []row{
		{name: "row1 prose", tree: body("Proviso keeps one specification of a system true."), holds: true},
		{name: "row2 a real question in prose", tree: body("Why lowercase? Because links on some hosts differ."), holds: true},
		{name: "row3 a skeleton's section", tree: q, file: featurePath, lines: at(q)},
		{name: "row4 a note after the mark", tree: note, file: featurePath, lines: at(note)},
	})
}

func TestLintUndecidedToken(t *testing.T) {
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
		{name: "row1 Then lint reads <file>", tree: baseline(), holds: true},
		{name: "row3 an outcome nobody has decided", tree: then, file: xPath, lines: lines(then.line(xPath, "Then      ?"))},
		{name: "row4 what else happens is undecided", tree: and, file: xPath, lines: lines(and.line(xPath, "  And     ?"))},
		{name: "row5 the comparison itself is open", tree: comparison, file: xPath, lines: lines(comparison.line(xPath, "<count> ?"))},
	})
	t.Run("row2 a ? inside a literal is a value", func(t *testing.T) {
		checkGoesOn(t, literal).lacks(t, "undecided", xPath, literal.line(xPath, "Provided"))
	})
	t.Run("derived: a ? in a backtick reference", func(t *testing.T) {
		tr := withRule("When      a buyer adds <product>\nThen      the cart holds <product>, see ´?´\n  And     nothing else changes", "| product | outcome |\n| --- | --- |\n| apple | added |\n")
		checkGoesOn(t, tr).lacks(t, "undecided", xPath, tr.line(xPath, "Then"))
	})
}

func TestLintUndecidedTerm(t *testing.T) {
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
	t.Run("row1 nothing reported", func(t *testing.T) {
		check(t, defined).clean(t)
	})
}
