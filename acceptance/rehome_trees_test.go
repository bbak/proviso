package acceptance

// Test trees of the feature upkeep.rehome. Each follows the notation, so that
// lint reports no error about it (upkeep.rehome.lint), and each is accepted,
// so that the tree is as its manifest records it (upkeep.rehome.baseline).
// Backticks are written as ´ (see md).

import (
	"testing"
)

const (
	sysContractPath = "proviso/contracts/cli.md"
	cartQuestions   = "proviso/capabilities/cart/questions.md"
	payCapPath      = "proviso/capabilities/billing/capability.md"
	payCardPath     = "proviso/capabilities/billing/features/card.md"
	payRefundPath   = "proviso/capabilities/billing/features/refund.md"
	stockCapPath    = "proviso/capabilities/stock/capability.md"
	stockGlossPath  = "proviso/capabilities/stock/glossary.md"
)

// The main tree holds an element of every row of upkeep.rehome.home and a
// place of every row of rehome_references:
//
//   - cart.item, a glossary entry of the system, and cart.ref.commonmark, a
//     reference of the system, which the entry system.break links to;
//   - cart.cli.lint and its member cart.cli.lint.exit-status, in the
//     system's contracts/cli.md;
//   - billing.x and its rule billing.x.r, a feature in the folder of the
//     capability cart; the rule cart.add.show hands over to billing.x.r and
//     holds it in an Examples cell, as text, in an inline link and in an
//     autolink; x.md links to it in prose, by an inline link, an autolink
//     and a link reference definition, and names it in a code span, in
//     plain text, in a fenced block (as an anchor, a link and a link
//     reference definition) and in an image; questions.md tags a question
//     with it, and names it in another;
//   - the capability pay in the folder capabilities/billing/, with the
//     features pay.card, at its identifier home, and billing.refund, named
//     after the folder.
const mainSysGlossaryMD = `# Glossary

## <a id="cart.item"></a> item

- Definition: A product placed in the cart.
- Contact: Dana Reyes
- Status: defined
- Aliases: items

## <a id="system.break"></a> thematic break

- Defined by: [CommonMark §4.1](references.md#cart.ref.commonmark)
- Contact: Dana Reyes
- Status: defined
`

const mainReferencesMD = `# References

## <a id="cart.ref.commonmark"></a> CommonMark

- Version: 0.31.2
- Link: https://spec.commonmark.org/0.31.2/
- Contact: Dana Reyes
`

const mainSysContractMD = `# Command line

## <a id="cart.cli.lint"></a> lint

The command.

### <a id="cart.cli.lint.exit-status"></a> exit-status

The exit status.
`

const showRule = `
## <a id="cart.add.show"></a> Rule: A short code shows its product

´´´guarded
When      a buyer enters the code <code>
Provided  <code> is short
Then      the shop shows the product of <code>
  And     nothing else changes
Otherwise → billing.x.r
´´´

| code | provided | outcome | why |
| --- | --- | --- | --- |
| billing.x.r | yes | shown | a code may look like [an identifier](x.md#billing.x.r) |
| a-very-long-code | no | handed over | <https://example.com/x.md#billing.x.r> |
`

const mainXMD = `# <a id="billing.x"></a> Experiments

- Contact: Dana Reyes

## Goal

Try out a rule.

## Rationale

Moved here from billing: see [the rule](x.md#billing.x.r), not
´billing.x.r´ nor ´[the rule](x.md#billing.x.r)´, and billing.x.r in
plain text.

Online at <https://example.com/x.md#billing.x.r>, as the [defined rule][rule],
and drawn in ![a picture](x.md#billing.x.r).

[rule]: x.md#billing.x.r

´´´text
<a id="billing.x.r"></a> [the rule](x.md#billing.x.r)
[rule]: x.md#billing.x.r
´´´

## <a id="billing.x.r"></a> Rule: Under test

´´´guarded
` + idBlock + `
´´´

` + idTable + `
`

const mainQuestionsMD = `# Questions

- [billing.x.r] Should the shop show pears?
- Who decides on [billing.x.r] later?
`

const payCapMD = `# <a id="pay"></a> Billing

## Goal

Buyers pay.

## Rationale

The shop needs the money.
`

const payCardMD = `# <a id="pay.card"></a> Paying by card

- Contact: Dana Reyes

## Goal

Buyers pay by card.

## Rationale

Most buyers carry a card.
`

const payRefundMD = `# <a id="billing.refund"></a> Refunds

- Contact: Dana Reyes

## Goal

Buyers get their money back.

## Rationale

Goods break.
`

func mainRehomeTree() tree {
	return baseline().without(glossaryPath).
		appendTo(featurePath, showRule).
		with(
			"proviso/glossary.md", mainSysGlossaryMD,
			referencesPath, mainReferencesMD,
			sysContractPath, mainSysContractMD,
			xPath, mainXMD,
			cartQuestions, mainQuestionsMD,
			payCapPath, payCapMD,
			payCardPath, payCardMD,
			payRefundPath, payRefundMD,
		)
}

// mainRenames are the renames of the main tree, ordered by their old
// identifier.
var mainRenames = []planRename{
	{"billing.refund", "pay.refund"},
	{"billing.x", "cart.x"},
	{"billing.x.r", "cart.x.r"},
	{"cart.cli.lint", "system.cli.lint"},
	{"cart.cli.lint.exit-status", "system.cli.lint.exit-status"},
	{"cart.item", "system.item"},
	{"cart.ref.commonmark", "system.ref.commonmark"},
}

// mainRehomed is the main tree as `rehome --write` leaves it: every place
// rehome_references lists renamed, and nothing else.
func mainRehomed() tree {
	return mainRehomeTree().
		edit("proviso/glossary.md", `<a id="cart.item">`, `<a id="system.item">`).
		edit("proviso/glossary.md", "(references.md#cart.ref.commonmark)", "(references.md#system.ref.commonmark)").
		edit(referencesPath, `<a id="cart.ref.commonmark">`, `<a id="system.ref.commonmark">`).
		edit(sysContractPath, `<a id="cart.cli.lint">`, `<a id="system.cli.lint">`).
		edit(sysContractPath, `<a id="cart.cli.lint.exit-status">`, `<a id="system.cli.lint.exit-status">`).
		edit(featurePath, "Otherwise → billing.x.r", "Otherwise → cart.x.r").
		edit(xPath, `# <a id="billing.x"></a>`, `# <a id="cart.x"></a>`).
		edit(xPath, `## <a id="billing.x.r"></a>`, `## <a id="cart.x.r"></a>`).
		edit(xPath, "see [the rule](x.md#billing.x.r), not", "see [the rule](x.md#cart.x.r), not").
		edit(xPath, "Online at <https://example.com/x.md#billing.x.r>", "Online at <https://example.com/x.md#cart.x.r>").
		edit(xPath, "[rule]: x.md#billing.x.r\n\n", "[rule]: x.md#cart.x.r\n\n").
		edit(cartQuestions, "- [billing.x.r] Should", "- [cart.x.r] Should").
		edit(payRefundPath, `<a id="billing.refund">`, `<a id="pay.refund">`)
}

// The one tree is the baseline tree with its glossary entry cart.item moved
// to the system's glossary: one rename, cart.item to system.item.
func oneRehomeTree() tree {
	return baseline().without(glossaryPath).with("proviso/glossary.md", glossaryMD)
}

func oneRehomed() tree {
	return oneRehomeTree().edit("proviso/glossary.md", `<a id="cart.item">`, `<a id="system.item">`)
}

// The taken tree holds the rule cart.x.r, and the rule billing.x.r in the
// feature cart.add, whose rename to cart.x.r collides with it.
func takenRehomeTree() tree {
	return withRule(idBlock, idTable).appendTo(featurePath, "\n"+ruleMD("billing.x.r", "Also under test", idBlock, idTable))
}

// The chain tree holds billing.widget in the cart's glossary and cart.widget
// in the glossary of the capability stock: billing.widget is renamed to
// cart.widget while cart.widget is renamed to stock.widget, which collides.
const stockCapMD = `# <a id="stock"></a> Stock

## Goal

The shop keeps goods.

## Rationale

Buyers want what they see.
`

const chainStockGlossaryMD = `# Glossary

## <a id="cart.widget"></a> gadget

- Definition: A thing the shop keeps.
- Contact: Dana Reyes
- Status: defined
`

func chainRehomeTree() tree {
	return baseline().
		appendTo(glossaryPath, `
## <a id="billing.widget"></a> widget

- Definition: A thing a buyer adds.
- Contact: Dana Reyes
- Status: defined
`).
		with(stockCapPath, stockCapMD, stockGlossPath, chainStockGlossaryMD)
}

// The first order tree holds, in the system's glossary, stock.keyword before
// billing.part, both renamed to system.
func order1RehomeTree() tree {
	return baseline().with("proviso/glossary.md", `# Glossary

## <a id="stock.keyword"></a> keyword

- Definition: A word with a meaning of its own.
- Contact: Dana Reyes
- Status: defined

## <a id="billing.part"></a> part

- Definition: A piece of something.
- Contact: Dana Reyes
- Status: defined
`)
}

// The second order tree holds, in the cart's glossary, pay.widget, then
// pay.gizmo and billing.gizmo, then cart.widget: pay.gizmo and billing.gizmo
// are both renamed to cart.gizmo, and pay.widget to cart.widget, which the
// tree holds.
func order2RehomeTree() tree {
	return baseline().appendTo(glossaryPath, `
## <a id="pay.widget"></a> widget

- Definition: A thing a buyer adds.
- Contact: Dana Reyes
- Status: defined

## <a id="pay.gizmo"></a> gizmo

- Definition: A thing a buyer wants.
- Contact: Dana Reyes
- Status: defined

## <a id="billing.gizmo"></a> gadget

- Definition: A thing a buyer forgets.
- Contact: Dana Reyes
- Status: defined

## <a id="cart.widget"></a> doohickey

- Definition: A thing a buyer finds.
- Contact: Dana Reyes
- Status: defined
`)
}

// TestFixtureRehome checks the ground the rehome tests stand on: lint
// reports nothing about their trees, nor about the trees rehome is expected
// to write, and the oracle of identifier homes plans what each tree's
// comment says.
func TestFixtureRehome(t *testing.T) {
	t.Parallel()
	trees := map[string]tree{
		"main": mainRehomeTree(), "main rehomed": mainRehomed(),
		"one": oneRehomeTree(), "one rehomed": oneRehomed(),
		"taken": takenRehomeTree(), "chain": chainRehomeTree(),
		"order1": order1RehomeTree(), "order2": order2RehomeTree(),
	}
	for name, tr := range trees {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			check(t, tr).clean(t)
		})
	}
	t.Run("oracle", func(t *testing.T) {
		t.Parallel()
		sameRenames(t, homeRenames(mainRehomeTree()), mainRenames)
		sameRenames(t, homeRenames(mainRehomed()), nil)
		sameRenames(t, homeRenames(baseline()), nil)
		sameRenames(t, homeRenames(oneRehomeTree()), []planRename{{"cart.item", "system.item"}})
		sameCollisions(t, collisionsOf(takenRehomeTree()), []planCollision{{"cart.x.r", []string{"billing.x.r"}}})
		sameCollisions(t, collisionsOf(chainRehomeTree()), []planCollision{{"cart.widget", []string{"billing.widget"}}})
		sameCollisions(t, collisionsOf(order2RehomeTree()), []planCollision{
			{"cart.gizmo", []string{"billing.gizmo", "pay.gizmo"}},
			{"cart.widget", []string{"pay.widget"}},
		})
		sameCollisions(t, collisionsOf(mainRehomeTree()), nil)
	})
}
