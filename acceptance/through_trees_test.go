package acceptance

// The through tree: the impact tree, in which cart.add.limit marks two more
// terms beside item, so that the tests of coherence.judge.touched and
// coherence.verify.stale can tell apart what a touched element depends on.
// shelf has a Definition over four lines, so that a branch and the mainline
// can edit it lines apart; label is defined by the system's reference
// CommonMark.
//
// The identifiers of the spec's examples stand for elements of this tree:
// lint.run.exit for cart.add.limit; lint.step for cart.item, or for
// cart.shelf where its Definition is changed by a merge of the mainline, or
// for cart.label where it is defined by system.ref.commonmark; lint.guard for
// cart.shelf; system.ref.commonmark for itself.

import "testing"

const throughGlossaryMD = `
## <a id="cart.shelf"></a> shelf

- Definition: Where the goods of one kind
  stand in the shop,
  each product
  beside the next.
- Contact: Dana Reyes
- Status: defined

## <a id="cart.label"></a> label

- Defined by: [CommonMark §4.1](../../references.md#system.ref.commonmark)
- Deviations: none
- Contact: Dana Reyes
- Status: defined
`

func throughTree() tree {
	return impactTree().
		with(referencesPath, systemReferencesMD).
		appendTo(glossaryPath, throughGlossaryMD).
		edit(featurePath, "  And     the cart stays as it was", "  And     the cart stays as it was, and so does the [shelf]").
		edit(featurePath, "Otherwise the cart refuses <product>", "Otherwise the cart refuses <product> with a [label]")
}

// Amendments of the through tree; iItemDef of impact_test.go applies too.
var (
	// tItemDefAgain rewords the Definition of item once more, after iItemDef.
	tItemDefAgain = change{glossaryPath, "- Definition: A product a buyer placed in the cart.", "- Definition: A product a buyer put in the cart."}
	// tItemContact gives item a new Contact, after iItemDef.
	tItemContact = change{glossaryPath, "- Definition: A product a buyer placed in the cart.\n- Contact: Dana Reyes", "- Definition: A product a buyer placed in the cart.\n- Contact: Bruno Baketarić"}
	// tShelfMiddle rewords a middle line of shelf's Definition.
	tShelfMiddle = change{glossaryPath, "  each product", "  every product"}
	// tShelfFirst and tShelfLast reword the first and the last line of
	// shelf's Definition, two lines apart.
	tShelfFirst = change{glossaryPath, "- Definition: Where the goods of one kind", "- Definition: Where all the goods of one kind"}
	tShelfLast  = change{glossaryPath, "  beside the next.", "  beside the other."}
	// tLabelDev changes the Deviations of label.
	tLabelDev = change{glossaryPath, "- Deviations: none", "- Deviations: only one line"}
	// tRefContact and tRefContactAgain give CommonMark a new Contact, and
	// then another.
	tRefContact      = change{referencesPath, "- Contact: Dana Reyes", "- Contact: Bruno Baketarić"}
	tRefContactAgain = change{referencesPath, "- Contact: Bruno Baketarić", "- Contact: Mia Chen"}
	// tRefVersion and tRefVersionAgain give CommonMark a new Version, and
	// then another.
	tRefVersion      = change{referencesPath, "- Version: 0.31.2", "- Version: 0.32"}
	tRefVersionAgain = change{referencesPath, "- Version: 0.32", "- Version: 0.33"}
)

// removeLimit removes the rule cart.add.limit from the through tree.
func removeLimit() change {
	text, _ := elementText(throughTree()[featurePath], "cart.add.limit")
	return change{featurePath, text, ""}
}

// throughRepo accepts the through tree and amends it with cs.
func throughRepo(t *testing.T, cs ...change) *repo {
	t.Helper()
	r, _ := accepted(t, throughTree())
	r.set(amended(throughTree(), cs...))
	return r
}

// throughCovered is the through tree amended by cs, every candidate judged.
func throughCovered(t *testing.T, cs ...change) *repo {
	t.Helper()
	r := throughRepo(t, cs...)
	r.judgeAll(nil)
	return r
}

func TestFixtureThrough(t *testing.T) {
	t.Parallel()
	check(t, throughTree()).clean(t)
	for _, cs := range [][]change{
		{iItemDef, tItemDefAgain},
		{iItemDef, tItemContact},
		{tShelfFirst, tShelfMiddle, tShelfLast},
		{tLabelDev, tRefContact, tRefContactAgain},
		{tRefVersion, tRefVersionAgain},
		{removeLimit(), iItemDef},
	} {
		check(t, amended(throughTree(), cs...)).clean(t)
	}
}
