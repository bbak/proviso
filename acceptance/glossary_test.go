package acceptance

// Feature lint.glossary: proviso/capabilities/lint/features/glossary-entries.md

import "testing"

const entryID = `id="cart.daily-limit"`

// withEntry is the baseline with a glossary entry daily limit of the cart,
// holding fields.
func withEntry(fields string) tree {
	return baseline().
		with("proviso/references.md", systemReferencesMD).
		appendTo(glossaryPath, "\n## <a id=\"cart.daily-limit\"></a> daily limit\n\n"+fields)
}

// entryRow expects a finding at the heading of daily limit.
func entryRow(name, fields string) row {
	tr := withEntry(fields)
	return row{name: name, tree: tr, file: glossaryPath, lines: lines(tr.line(glossaryPath, entryID))}
}

const (
	definedBy = "- Defined by: [CommonMark §6.3](../../references.md#system.ref.commonmark)\n"
	contact   = "- Contact: Dana Reyes\n"
	question  = "- Question: What does the daily limit apply to?\n"
)

func TestLintGlossaryStatus(t *testing.T) {
	status := func(name, value string, holds bool) row {
		tr := withEntry("- Definition: The amount a buyer may spend per day.\n" + contact + "- Status: " + value + "\n")
		return row{name: name, tree: tr, holds: holds, file: glossaryPath, lines: lines(tr.line(glossaryPath, "- Status: "+value))}
	}
	rows(t, "status-unknown", []row{
		status("row1 defined", "defined", true),
		{name: "row2 open", tree: withEntry("- Definition: ?\n" + contact + "- Status: open\n" + question), holds: true},
		status("row3 ?", "?", false),
		status("row4 draft", "draft", false),
		status("row5 values are matched exactly", "Defined", false),
	})
}

func TestLintGlossaryOpen(t *testing.T) {
	rows(t, "open-entry-incomplete", []row{
		{name: "row1 Definition ? and a Question", tree: withEntry("- Definition: ?\n" + contact + "- Status: open\n" + question), holds: true},
		entryRow("row2 no Question", "- Definition: ?\n"+contact+"- Status: open\n"),
		entryRow("row3 a definition", "- Definition: The amount withdrawn per day.\n"+contact+"- Status: open\n"+question),
		entryRow("row4 defined by a reference", definedBy+contact+"- Status: open\n"+question),
	})
}

func TestLintGlossaryDefined(t *testing.T) {
	rows(t, "defined-entry-open", []row{
		{name: "row1 a Definition", tree: withEntry("- Definition: An amount is composable when it splits.\n" + contact + "- Status: defined\n"), holds: true},
		{name: "row2 defined by a reference", tree: withEntry(definedBy + contact + "- Status: defined\n"), holds: true},
		entryRow("row3 a Question left", "- Definition: An amount is composable when it splits.\n"+contact+"- Status: defined\n"+question),
		entryRow("row4 Definition ?", "- Definition: ?\n"+contact+"- Status: defined\n"),
	})
	t.Run("rows1-2 nothing reported", func(t *testing.T) {
		check(t, withEntry("- Definition: An amount is composable when it splits.\n"+contact+"- Status: defined\n")).clean(t)
		check(t, withEntry(definedBy+contact+"- Status: defined\n")).clean(t)
	})
}

func TestLintGlossaryDefinition(t *testing.T) {
	rows(t, "definition-count", []row{
		{name: "row1 Definition", tree: withEntry("- Definition: An amount is composable when it splits.\n" + contact + "- Status: defined\n"), holds: true},
		{name: "row2 Defined by", tree: withEntry(definedBy + "- Deviations: none\n" + contact + "- Status: defined\n"), holds: true},
		entryRow("row3 both", "- Definition: A target of a link.\n"+definedBy+contact+"- Status: defined\n"),
		entryRow("row4 neither", contact+"- Status: defined\n"),
	})
}

func TestLintGlossaryDefinedBy(t *testing.T) {
	field := func(name, value string, holds bool, extra ...string) row {
		tr := withEntry("- Defined by: " + value + "\n" + contact + "- Status: defined\n").with(extra...)
		return row{name: name, tree: tr, holds: holds, file: glossaryPath, lines: lines(tr.line(glossaryPath, "- Defined by:"))}
	}
	search := []string{
		"proviso/capabilities/search/capability.md", "# <a id=\"search\"></a> Search\n\n## Goal\n\nBuyers find goods.\n\n## Rationale\n\nA shop has many goods.\n",
		"proviso/capabilities/search/references.md", "# References\n\n## <a id=\"search.ref.x\"></a> X\n\n- Version: 1\n- Link: https://example.com/x\n- Contact: Dana Reyes\n",
	}
	own := []string{
		"proviso/capabilities/cart/references.md", "# References\n\n## <a id=\"cart.ref.x\"></a> X\n\n- Version: 1\n- Link: https://example.com/x\n- Contact: Dana Reyes\n",
	}
	rows(t, "defined-by-unresolved", []row{
		field("row1 the system's references, seen from a capability", "[CommonMark §6.3](../../references.md#system.ref.commonmark)", true),
		{name: "row2 from the system's glossary", tree: baseline().with(
			"proviso/references.md", systemReferencesMD,
			"proviso/glossary.md", "# Glossary\n\n## <a id=\"system.link-destination\"></a> link destination\n\n- Defined by: [CommonMark](references.md#system.ref.commonmark)\n"+contact+"- Status: defined\n"), holds: true},
		field("row3 no link to follow", "CommonMark §6.3", false),
		field("row4 a link to a glossary", "[CommonMark](../../glossary.md#system.ref.commonmark)", false),
		field("row5 another capability's references", "[CommonMark](../search/references.md#search.ref.x)", false, search...),
		field("row6 no such identifier", "[CommonMark](../../references.md#system.ref.commonmar)", false),
		field("derived: the capability's own references", "[X](references.md#cart.ref.x)", true, own...),
	})
}

func TestLintGlossaryDeviations(t *testing.T) {
	rows(t, "deviations-without-reference", []row{
		{name: "row1 defined by CommonMark", tree: withEntry(definedBy + "- Deviations: none\n" + contact + "- Status: defined\n"), holds: true},
		entryRow("row2 a Definition only", "- Definition: An amount is composable when it splits.\n- Deviations: none\n"+contact+"- Status: defined\n"),
	})
}
