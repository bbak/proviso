package acceptance

// The baseline tree: a small spec tree that follows the notation in every
// respect, so that lint reports nothing about it. Each test changes one
// thing about it. Backticks are written as ´ (see md).

import (
	"strings"
	"testing"
)

const (
	systemPath   = "proviso/system.md"
	capPath      = "proviso/capabilities/cart/capability.md"
	glossaryPath = "proviso/capabilities/cart/glossary.md"
	paramsPath   = "proviso/capabilities/cart/parameters.md"
	featurePath  = "proviso/capabilities/cart/features/add.md"
	xPath        = "proviso/capabilities/cart/features/x.md"
)

const systemMD = `# <a id="system"></a> Shop

## Goal

The shop sells goods to buyers.

## Rationale

Buyers want goods.

## Scope

- Selling goods.

## Out of scope

- Shipping goods.

## Audience

- Buyers.

## Design boundary

- A web browser.
`

const capabilityMD = `# <a id="cart"></a> Cart

## Goal

Buyers collect goods before they pay.

## Rationale

Paying once for many goods is easier.
`

const glossaryMD = `# Glossary

## <a id="cart.item"></a> item

- Definition: A product placed in the cart.
- Contact: Dana Reyes
- Status: defined
- Aliases: items
`

const parametersMD = `# Parameters

## <a id="cart.max-quantity"></a> max_quantity

- Value: 10
- Unit: pieces
- Configurable: no
- Rationale: a shelf holds ten pieces.
`

const featureMD = "# <a id=\"cart.add\"></a> Adding to the cart\n" + `
- Contact: Dana Reyes

## Goal

Buyers put goods into the cart.

## Rationale

A cart that takes anything would empty the shelves.

## <a id="cart.add.limit"></a> Rule: The cart takes a limited quantity

´´´guarded
When      a buyer adds the quantity <quantity> of the [item] <product>
Provided  <quantity> is at most ´max_quantity´
Then      the cart holds <quantity> more of <product>
  And     nothing else changes
Otherwise the cart refuses <product>
  And     the cart stays as it was
´´´

| quantity | product | provided | outcome | why |
| --- | --- | --- | --- | --- |
| 1 | apple | yes | added | |
| 11 | apple | no | refused | more than a shelf holds |

## <a id="cart.add.empty"></a> Rule: A new cart is empty

´´´guarded
When      a buyer opens the new cart <cart>
Then      <cart> holds no [items]
  And     nothing else changes
´´´

| cart | outcome |
| --- | --- |
| the first cart | empty |
`

func baseline() tree {
	return tree{}.with(
		systemPath, systemMD,
		capPath, capabilityMD,
		glossaryPath, glossaryMD,
		paramsPath, parametersMD,
		featurePath, featureMD,
	)
}

// The head of x.md, a second feature of the cart that holds the rules a
// test is about.
const xHead = `# <a id="cart.x"></a> Experiments

- Contact: Dana Reyes

## Goal

Try out a rule.

## Rationale

To see what lint says about it.

`

// ruleMD writes a rule with one guarded block and the table after it.
func ruleMD(id, title, block, table string) string {
	return "## <a id=\"" + id + "\"></a> Rule: " + title + "\n\n´´´guarded\n" +
		strings.TrimRight(block, "\n") + "\n´´´\n\n" + strings.TrimRight(table, "\n") + "\n\n"
}

// withRule is the baseline with x.md holding one rule, cart.x.r.
func withRule(block, table string) tree {
	return baseline().with(xPath, xHead+ruleMD("cart.x.r", "Under test", block, table))
}

// withX is the baseline with x.md holding body after its head.
func withX(body string) tree {
	return baseline().with(xPath, xHead+body)
}

// A guarded block on the placeholder <id>, and its table.
const idBlock = `When      a buyer looks up the product <id>
Provided  <id> is listed
Then      the shop shows <id>
  And     nothing else changes
Otherwise the shop shows nothing
  And     nothing else changes`

const idTable = `| id | provided | outcome | why |
| --- | --- | --- | --- |
| apple | yes | shown | |
| pear | no | nothing shown | |`

// kwBlock writes a rule block with one line per keyword, in the given order,
// and a table that suits it.
func kwBlock(keywords ...string) (string, string) {
	var b []string
	step := ""
	parts := 0
	guarded := false
	for i, k := range keywords {
		switch k {
		case "Given":
			b = append(b, "Given     the shop is open")
		case "When":
			b = append(b, "When      a buyer adds <product>")
		case "Provided":
			guarded = true
			parts++
			b = append(b, "Provided  <product> is in stock")
		case "Then":
			b = append(b, "Then      the cart holds <product>")
		case "Otherwise":
			b = append(b, "Otherwise the cart refuses <product>")
		case "And":
			switch {
			case i == 0:
				b = append(b, "  And     the shop is open")
			case step == "Given":
				b = append(b, "  And     the shop is lit")
			case step == "Provided":
				parts++
				b = append(b, "  And     <product> is cheap")
			default:
				b = append(b, "  And     nothing else changes")
			}
		}
		if k != "And" {
			step = k
		}
	}
	var table string
	if guarded {
		table = "| product | provided | outcome |\n| --- | --- | --- |\n| apple | yes | added |\n"
		// A product of its own per failing part, so that no row repeats
		// another (lint.examples.row-repeated).
		failing := []string{"pear", "kiwi", "plum", "fig"}
		for p := 1; p <= parts; p++ {
			table += "| " + failing[p-1] + " | no (part " + string(rune('0'+p)) + ") | refused |\n"
		}
	} else {
		table = "| product | outcome |\n| --- | --- |\n| apple | added |\n"
	}
	return strings.Join(b, "\n"), table
}

// The system's glossary, defining "spec tree".
const systemGlossaryMD = `# Glossary

## <a id="system.spec-tree"></a> spec tree

- Definition: The files that specify the shop.
- Contact: Dana Reyes
- Status: defined
- Aliases: tree
`

// The cart's contracts: findings with a member, and lint with a member.
const findingsContractMD = `# Findings

## <a id="cart.findings"></a> findings

The kinds of finding.

### <a id="cart.findings.anchor-misplaced"></a> anchor-misplaced

An anchor in the wrong place.
`

const cliContractMD = `# Command line

## <a id="cart.cli.lint"></a> lint

The command.

### <a id="cart.cli.lint.exit-status"></a> exit-status

The exit status.
`

// The system's references.
const systemReferencesMD = `# References

## <a id="system.ref.commonmark"></a> CommonMark

- Version: 0.31.2
- Link: https://spec.commonmark.org/0.31.2/
- Contact: Dana Reyes
`

// billing is a second capability, with one feature holding block and table.
func withBilling(tr tree, block, table string) tree {
	return tr.with(
		"proviso/capabilities/billing/capability.md", `# <a id="billing"></a> Billing

## Goal

Buyers pay.

## Rationale

The shop needs the money.
`,
		"proviso/capabilities/billing/features/pay.md", `# <a id="billing.pay"></a> Paying

- Contact: Dana Reyes

## Goal

Buyers pay for the cart.

## Rationale

Goods are not given away.

`+ruleMD("billing.pay.r", "A buyer pays", block, table))
}

// TestFixtureLint is the ground every other test stands on: the
// baseline tree follows the notation, so lint reports nothing and exits
// with 0.
func TestFixtureLint(t *testing.T) {
	t.Parallel()
	check(t, baseline()).clean(t)
}
