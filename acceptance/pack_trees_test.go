package acceptance

// The pack tree: the impact tree (see impact_trees_test.go) with what the
// reading chain of baseline.pack.chain needs besides: a reference and a term
// defined by it, a rule marking that term, a rule handing over to
// cart.add.limit, a term no rule marks, and a constraint of the capability
// cart. Backticks are written as ´ (see md).
//
// What each rule depends on (dependencies):
//
//	rule                 depends on
//	cart.add.limit       cart.item, cart.max-quantity
//	cart.add.full        nothing
//	cart.add.spoken      system.price, cart.add.limit (its hand-over)
//	cart.pay.total       cart.item, system.price
//	cart.pay.lines       cart.api.line.count
//	cart.pay.order       cart.api.line and its member cart.api.line.count
//	cart.pay.cap         cart.max-quantity
//	cart.pay.open        nothing
//	cart.pay.receipt     cart.thematic-break
//	billing.refund.full  system.price
//	billing.refund.add   nothing
//
// cart.api.line represents cart.item; cart.thematic-break is defined by
// system.ref.commonmark; cart.shelf is marked by no rule.

import "testing"

const cartConstraintsPath = "proviso/capabilities/cart/constraints.md"

const packGlossaryMD = diffGlossaryMD + `
## <a id="cart.shelf"></a> shelf

- Definition: Where the shop keeps goods.
- Contact: Dana Reyes
- Status: defined
`

const spokenRule = `## <a id="cart.add.spoken"></a> Rule: A spoken quantity is heard first

´´´guarded
When      a buyer says the quantity <quantity> at the [price] <price>
Provided  the shop hears <quantity>
Then      the cart holds <quantity> more
  And     nothing else changes
Otherwise → cart.add.limit
´´´

| quantity | price | provided | outcome |
| --- | --- | --- | --- |
| 1 | one coin | yes | added |
| 2 | one coin | no | handed over |

`

const receiptRule = `
## <a id="cart.pay.receipt"></a> Rule: A receipt ends with a break

´´´guarded
When      the shop prints the receipt <receipt>
Then      <receipt> ends with a [thematic break]
  And     nothing else changes
´´´

| receipt | outcome |
| --- | --- |
| the first receipt | a break at the end |
`

func packTree() tree {
	return impactTree().
		edit(featurePath, "## <a id=\"cart.add.never\"></a>", spokenRule+"## <a id=\"cart.add.never\"></a>").
		appendTo(payPath, receiptRule).
		with(
			glossaryPath, packGlossaryMD,
			referencesPath, systemReferencesMD,
			cartConstraintsPath, `# Constraints

## <a id="cart.light"></a> Carts stay light

### Rationale

Buyers carry their carts.

### Property

A cart weighs at most five kilograms.
`,
		)
}

// Where each element of the pack tree stands.
var packStandsIn = map[string]string{
	"system": systemPath, "system.fast": sysConstraintsPath, "system.honest": sysInvariantsPath,
	"system.price": sysGlossaryPath, "system.ref.commonmark": referencesPath,
	"cart": capPath, "cart.item": glossaryPath, "cart.thematic-break": glossaryPath, "cart.shelf": glossaryPath,
	"cart.max-quantity": paramsPath, "cart.api.line": apiPath, "cart.api.line.count": apiPath,
	"cart.bounded": cartInvariantsPath, "cart.light": cartConstraintsPath,
	"cart.add": featurePath, "cart.add.limit": featurePath, "cart.add.full": featurePath,
	"cart.add.spoken": featurePath, "cart.add.never": featurePath,
	"cart.pay": payPath, "cart.pay.total": payPath, "cart.pay.lines": payPath, "cart.pay.order": payPath,
	"cart.pay.cap": payPath, "cart.pay.open": payPath, "cart.pay.receipt": payPath,
	"billing": billingCapPath, "billing.balanced": billingInvPath, "billing.refund": refundPath,
	"billing.refund.full": refundPath, "billing.refund.add": refundPath,
}

func TestBaselinePackFixture(t *testing.T) {
	t.Parallel()
	check(t, packTree()).clean(t)
}
