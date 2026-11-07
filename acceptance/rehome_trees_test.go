package acceptance

// Test trees of the feature upkeep.rehome. Each follows the notation, so that
// lint reports no error about it (upkeep.rehome.lint), and each is accepted,
// so that the tree is as its manifest records it (upkeep.rehome.baseline).
// Backticks are written as ´ (see md).

import (
	"strings"
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

// The mentions tree holds the rule billing.x.r in the feature cart.x, renamed
// to cart.x.r, with an identifier mention of it on every kind of line
// upkeep.rehome.mentions names, and the same identifier where it is none:
//
//   - the rule cart.add.show hands over to billing.x.r and holds it as a
//     code span in an Examples cell (A), twice in one row (B), only within a
//     longer code span (C) and in plain text; and ´lint.run´, which is not
//     renamed, as a code span in a cell (D);
//   - x.md names it as a code span in the Rationale of cart.x (F), in plain
//     text on the next line, and as a code span in an Examples cell of the
//     rule billing.x.r itself (H);
//   - the system's contracts/cli.md names it only in a fenced block;
//   - questions.md names it as a code span in the text of a question (G).
const mentionShowRule = `
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
| ´billing.x.r´ | yes | shown | a code may look like an identifier |
| ´billing.x.r´ again | yes | shown | as ´billing.x.r´ twice |
| ´## <a id="billing.x.r"></a> Rule: B´ | no | handed over | a heading as data |
| ´lint.run´ | yes | shown | an identifier not renamed |
| billing.x.r in plain text | yes | shown | |
`

// The lines of the mentions tree its comment names.
const (
	mentionLineA     = "| ´billing.x.r´ | yes | shown | a code may look like an identifier |"
	mentionLineB     = "| ´billing.x.r´ again | yes | shown | as ´billing.x.r´ twice |"
	mentionLineC     = "| ´## <a id=\"billing.x.r\"></a> Rule: B´ | no | handed over | a heading as data |"
	mentionLineD     = "| ´lint.run´ | yes | shown | an identifier not renamed |"
	mentionLinePlain = "| billing.x.r in plain text | yes | shown | |"
	mentionLineF     = "Moved here from billing: see ´billing.x.r´."
	mentionLineXText = "Also billing.x.r in plain text."
	mentionLineG     = "- [cart.add.limit] Should ´billing.x.r´ show pears?"
	mentionLineH     = "| pear | no | nothing shown | ´billing.x.r´ names this rule |"
)

const mentionXMD = `# <a id="cart.x"></a> Experiments

- Contact: Dana Reyes

## Goal

Try out a rule.

## Rationale

Moved here from billing: see ´billing.x.r´.
Also billing.x.r in plain text.

## <a id="billing.x.r"></a> Rule: Under test

´´´guarded
` + idBlock + `
´´´

| id | provided | outcome | why |
| --- | --- | --- | --- |
| apple | yes | shown | |
| pear | no | nothing shown | ´billing.x.r´ names this rule |
`

const mentionContractMD = `# Command line

## <a id="system.cli.lint"></a> lint

The command.

´´´text
´billing.x.r´
´´´
`

const mentionQuestionsMD = `# Questions

- [cart.add.limit] Should ´billing.x.r´ show pears?
`

func mentionsTree() tree {
	return baseline().
		appendTo(featurePath, mentionShowRule).
		with(
			xPath, mentionXMD,
			sysContractPath, mentionContractMD,
			cartQuestions, mentionQuestionsMD,
		)
}

// mentionsRehomed is the mentions tree as `rehome --write` leaves it with
// the mentions picked, each given as the line that holds it: the anchor and
// the hand-over renamed, and every code span of billing.x.r on those lines.
func mentionsRehomed(picked ...string) tree {
	tr := mentionsTree().
		edit(xPath, `<a id="billing.x.r">`, `<a id="cart.x.r">`).
		edit(featurePath, "Otherwise → billing.x.r", "Otherwise → cart.x.r")
	for _, l := range picked {
		for _, p := range []string{featurePath, xPath, cartQuestions} {
			if strings.Contains(tr[p], md(l)) {
				tr = tr.edit(p, l, strings.ReplaceAll(l, "´billing.x.r´", "´cart.x.r´"))
			}
		}
	}
	return tr
}

// mentionsOfTree are the mentions of the mentions tree, as
// upkeep.rehome.mentions lists them.
func mentionsOfTree() []planMention {
	tr := mentionsTree()
	return []planMention{
		mention("billing.x.r", featurePath, tr.line(featurePath, mentionLineA), 1, "cart.add.show"),
		mention("billing.x.r", featurePath, tr.line(featurePath, mentionLineB), 2, "cart.add.show"),
		mention("billing.x.r", xPath, tr.line(xPath, mentionLineF), 1, "cart.x"),
		mention("billing.x.r", xPath, tr.line(xPath, mentionLineH), 1, "billing.x.r"),
		mention("billing.x.r", cartQuestions, tr.line(cartQuestions, mentionLineG), 1, ""),
	}
}

// The two-cells tree is the one tree with its rename's old identifier,
// cart.item, as a code span in the why cells of both Examples rows of
// cart.add.limit.
const (
	twoCellsLine1 = "| 1 | apple | yes | added | ´cart.item´ |"
	twoCellsLine2 = "| 11 | apple | no | refused | more than one ´cart.item´ holds |"
)

func twoCellsTree() tree {
	return oneRehomeTree().
		edit(featurePath, "| 1 | apple | yes | added | |", twoCellsLine1).
		edit(featurePath, "| 11 | apple | no | refused | more than a shelf holds |", twoCellsLine2)
}

// twoCellsRehomed is the two-cells tree as `rehome --write` leaves it with
// the first cell's mention picked.
func twoCellsRehomed() tree {
	return twoCellsTree().
		edit("proviso/glossary.md", `<a id="cart.item">`, `<a id="system.item">`).
		edit(featurePath, twoCellsLine1, "| 1 | apple | yes | added | ´system.item´ |")
}

func twoCellsMentions() []planMention {
	tr := twoCellsTree()
	return []planMention{
		mention("cart.item", featurePath, tr.line(featurePath, twoCellsLine1), 1, "cart.add.limit"),
		mention("cart.item", featurePath, tr.line(featurePath, twoCellsLine2), 1, "cart.add.limit"),
	}
}

// The third order tree is the first with the capability stock: stock.keyword
// stands as a code span in cart's capability.md on line 9, and in the stock
// feature shelf.md on lines 7 and 12, line 12 holding billing.part after
// it; so the path decides before the line, a line of two digits comes after
// one of one, and the identifier decides last.
const shelfPath = "proviso/capabilities/stock/features/shelf.md"

const shelfMD = `# <a id="stock.shelf"></a> Shelves

- Contact: Dana Reyes

## Goal

Shelves hold every ´stock.keyword´ the shop keeps.

## Rationale

Goods need a place.
A ´stock.keyword´ or a ´billing.part´ marks it.
`

func order3RehomeTree() tree {
	return order1RehomeTree().
		edit(capPath, "Paying once for many goods is easier.", "Paying once for many goods, each a ´stock.keyword´, is easier.").
		with(stockCapPath, stockCapMD, shelfPath, shelfMD)
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
		"mentions": mentionsTree(), "mentions rehomed": mentionsRehomed(),
		"two cells": twoCellsTree(), "two cells rehomed": twoCellsRehomed(),
		"order3": order3RehomeTree(), "mentions all picked": mentionsRehomed(mentionLineA, mentionLineB, mentionLineF, mentionLineG, mentionLineH),
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
		sameRenames(t, homeRenames(mentionsTree()), []planRename{{"billing.x.r", "cart.x.r"}})
		sameMentions(t, mentionsOf(mentionsTree()), mentionsOfTree())
		sameMentions(t, mentionsOf(twoCellsTree()), twoCellsMentions())
		sameRenames(t, homeRenames(order3RehomeTree()), []planRename{{"billing.part", "system.part"}, {"stock.keyword", "system.keyword"}})
		tr := order3RehomeTree()
		if tr.line(capPath, "´stock.keyword´") != 9 || tr.line(shelfPath, "´stock.keyword´ the shop") != 7 || tr.line(shelfPath, "´stock.keyword´ or") != 12 {
			t.Errorf("the third order tree's mentions stand on other lines than its comment says")
		}
	})
}
