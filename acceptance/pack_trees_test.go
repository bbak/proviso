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
	check(t, packTree()).clean(t)
}

// builderNotesMD is proviso/contracts/builder-notes.md, which holds
// builder_notes, the contract element system.builder-notes: what every
// context pack says first, whatever tree it is assembled from.
const builderNotesMD = `# Builder notes

## <a id="system.builder-notes"></a> builder_notes

What every context pack says first to the implementing agents, before
anything of the spec tree. These are the notes, in full:

- **Goals give the frame.** The Goals of the system, the capability and
  each feature say what it is all for. Where the rules leave a choice
  open, choose in the Goal's sense, and list the choice. Where a rule
  seems to work against its Goal, follow the rule and list the conflict;
  do not resolve it yourself.
- **Scope sets the limit.** Build or test nothing that the system lists
  as Out of scope, and treat what its Design boundary names as given.
- **The pack is the only source.** If it does not say something, it is
  not decided: do not decide it silently, list it.
- A rule is one ´guarded´ block. ´Given´ is state, ´When´ is the event,
  ´Provided´ is the guard. **Both** ´Then´ and ´Otherwise´ are required
  behaviour, and each ´And´ line is part of the outcome, including
  "everything else stays as it was".
- ´[term]´ means exactly its glossary entry: the capability's glossary
  first, then the system's. ´´ ´name´ ´´ is a contract element or
  parameter. ´"…"´ is a literal value. ´<name>´ is a placeholder, varied
  by the rule's Examples table.
- Every Examples row is a case: placeholder columns are inputs, ´outcome´
  is the expectation, ´provided´ says whether the guard holds.
- ´Otherwise → <id>´ hands the decision to the rule with that identifier.
- An invariant holds over every event in its "Ranges over" (tester: test
  it as a property). A constraint is checked by build or review, not at
  run time (coder: it binds how you build). A changed constraint binds
  what was built before as well: review it.
- An element changed through a dependency reads as before but means
  something else: check what was built for it. What was built for a
  removed element is retired.
- **Every ´?´ is a stop, and so is an element listed as not to be built:
  do not build or test it.**
- Rationale explains a choice; it is not behaviour, but do not "fix" or
  contradict what it justifies.
`
