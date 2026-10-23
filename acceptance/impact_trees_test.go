package acceptance

// The impact tree: a spec tree in which the rules and invariants share
// events, levels, terms and references with each other, so that the tests
// of baseline.impact and baseline.diff.dependency can tell each source and
// each dependency apart. Backticks are written as ´ (see md).
//
// What the rules share:
//
//	rule                event head                    terms           references
//	cart.add.limit      buyer adds quantity           item            max_quantity
//	cart.add.full       buyer adds quantity
//	cart.pay.total      buyer pays for items          item, price
//	cart.pay.lines      buyer reviews order                           line.count
//	cart.pay.order      buyer orders line for products                line
//	cart.pay.cap        buyer checks out cart                         max_quantity
//	cart.pay.open       shop opens
//	billing.refund.full buyer returns product         price
//	billing.refund.add  buyer adds quantity
//
// The contract element line represents the term item. The invariants are
// system.honest (system), cart.bounded (capability cart), cart.add.never
// (feature cart.add) and billing.balanced (capability billing); the
// constraint system.fast stands at system level.

import "testing"

const (
	sysConstraintsPath = "proviso/constraints.md"
	sysInvariantsPath  = "proviso/invariants.md"
	sysGlossaryPath    = "proviso/glossary.md"
	apiPath            = "proviso/capabilities/cart/contracts/api.md"
	cartInvariantsPath = "proviso/capabilities/cart/invariants.md"
	payPath            = "proviso/capabilities/cart/features/pay.md"
	billingCapPath     = "proviso/capabilities/billing/capability.md"
	billingInvPath     = "proviso/capabilities/billing/invariants.md"
	refundPath         = "proviso/capabilities/billing/features/refund.md"
)

const impactAddMD = "# <a id=\"cart.add\"></a> Adding to the cart\n" + `
- Contact: Dana Reyes

## Goal

Buyers put goods into the cart.

## Rationale

A cart that takes anything would empty the shelves.

## <a id="cart.add.limit"></a> Rule: The cart takes a limited quantity

### Rationale

A shelf holds ten pieces.

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

## <a id="cart.add.full"></a> Rule: A full cart takes nothing

´´´guarded
When      a buyer adds the quantity <quantity> to the full cart <cart>
Then      <cart> refuses <quantity>
  And     nothing else changes
´´´

| quantity | cart | outcome |
| --- | --- | --- |
| 1 | the first cart | refused |

## <a id="cart.add.never"></a> Invariant: A cart never loses goods

### Rationale

Buyers count on what they put in.

### Ranges over

Every addition to a cart.

### Property

A cart holds every piece a buyer put in until the buyer takes it out.
`

const impactPayMD = "# <a id=\"cart.pay\"></a> Paying for the cart\n" + `
- Contact: Dana Reyes

## Goal

Buyers pay for what the cart holds.

## Rationale

Goods are not given away.

## <a id="cart.pay.total"></a> Rule: The shop charges the total

´´´guarded
When      a buyer pays for the [items] <items>
Then      the shop charges the [price] of <items>
  And     nothing else changes
´´´

| items | outcome |
| --- | --- |
| two apples | charged |

## <a id="cart.pay.lines"></a> Rule: The order shows its lines

´´´guarded
When      a buyer reviews the order <order>
Then      the shop shows ´line.count´ for each line of <order>
  And     nothing else changes
´´´

| order | outcome |
| --- | --- |
| the first order | shown |

## <a id="cart.pay.order"></a> Rule: An order has a line per product

´´´guarded
When      a buyer orders a ´line´ for the products <products>
Then      the order holds one line per product of <products>
  And     nothing else changes
´´´

| products | outcome |
| --- | --- |
| apple and pear | two lines |

## <a id="cart.pay.cap"></a> Rule: The shop takes only bounded orders

´´´guarded
When      a buyer checks out the cart <cart>
Provided  <cart> holds at most ´max_quantity´ pieces of each product
Then      the shop takes the order
  And     nothing else changes
Otherwise the shop refuses the order
  And     nothing else changes
´´´

| cart | provided | outcome |
| --- | --- | --- |
| the first cart | yes | taken |
| the second cart | no | refused |

## <a id="cart.pay.open"></a> Rule: An open shop greets buyers

´´´guarded
When      the shop opens
Then      the shop greets every buyer
  And     nothing else changes
´´´

| outcome |
| --- |
| greeted |
`

const impactRefundMD = "# <a id=\"billing.refund\"></a> Refunds\n" + `
- Contact: Dana Reyes

## Goal

Buyers get their money back for returned goods.

## Rationale

A shop that keeps the money loses its buyers.

## <a id="billing.refund.full"></a> Rule: A returned product is refunded in full

´´´guarded
When      a buyer returns the product <product>
Then      the shop refunds the [price] of <product>
  And     nothing else changes
´´´

| product | outcome |
| --- | --- |
| apple | refunded |

## <a id="billing.refund.add"></a> Rule: A refund adds up

´´´guarded
When      a buyer adds the quantity <quantity> to the refund <refund>
Then      <refund> grows by <quantity>
  And     nothing else changes
´´´

| quantity | refund | outcome |
| --- | --- | --- |
| 1 | the first refund | grown |
`

func impactTree() tree {
	return tree{}.with(
		systemPath, systemMD,
		sysConstraintsPath, `# Constraints

## <a id="system.fast"></a> Carts answer quickly

### Rationale

Buyers leave slow shops.

### Property

Every cart answers within a second.
`,
		sysInvariantsPath, `# Invariants

## <a id="system.honest"></a> The shop never overcharges

### Rationale

Trust is the shop's capital.

### Ranges over

Every payment.

### Property

No buyer pays more than the listed price.
`,
		sysGlossaryPath, `# Glossary

## <a id="system.price"></a> price

- Definition: What a product costs.
- Contact: Dana Reyes
- Status: defined
`,
		capPath, capabilityMD,
		glossaryPath, glossaryMD,
		paramsPath, parametersMD,
		apiPath, `# API

## <a id="cart.api.line"></a> line

- Represents: [item]

A line of an order.

### <a id="cart.api.line.count"></a> count

How many pieces the line holds.
`,
		cartInvariantsPath, `# Invariants

## <a id="cart.bounded"></a> A cart is bounded

### Rationale

Shelves are small.

### Ranges over

Every addition to a cart.

### Property

No cart holds more than ten pieces of a product.
`,
		featurePath, impactAddMD,
		payPath, impactPayMD,
		billingCapPath, `# <a id="billing"></a> Billing

## Goal

Buyers pay.

## Rationale

The shop needs the money.
`,
		billingInvPath, `# Invariants

## <a id="billing.balanced"></a> The books balance

### Rationale

Auditors check the books.

### Ranges over

Every payment and every refund.

### Property

Payments less refunds equal the money held.
`,
		refundPath, impactRefundMD,
	)
}

// Every rule and invariant of the impact tree, by where it stands.
var (
	addRules     = []string{"cart.add.full", "cart.add.limit", "cart.add.never"}
	payRules     = []string{"cart.pay.cap", "cart.pay.lines", "cart.pay.open", "cart.pay.order", "cart.pay.total"}
	cartAll      = append(append(append([]string(nil), addRules...), payRules...), "cart.bounded")
	billingAll   = []string{"billing.balanced", "billing.refund.add", "billing.refund.full"}
	everyRuleInv = append(append(append([]string(nil), cartAll...), billingAll...), "system.honest")
)

// standsIn gives the file each element of the impact tree stands in.
var standsIn = map[string]string{
	"cart.add.full": featurePath, "cart.add.limit": featurePath, "cart.add.never": featurePath,
	"cart.pay.cap": payPath, "cart.pay.lines": payPath, "cart.pay.open": payPath, "cart.pay.order": payPath, "cart.pay.total": payPath,
	"cart.bounded": cartInvariantsPath, "billing.balanced": billingInvPath, "billing.refund.add": refundPath,
	"billing.refund.full": refundPath, "system.honest": sysInvariantsPath,
}

func TestBaselineImpactFixture(t *testing.T) {
	t.Parallel()
	check(t, impactTree()).clean(t)
}
