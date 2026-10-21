package acceptance

// Test trees of the baseline capability. Each follows the notation in every
// respect, so that lint reports no error about it and the baseline commands
// run (baseline.run.lint). Backticks are written as ´ (see md).

import (
	"strings"
	"testing"
)

const (
	contractPath   = "proviso/capabilities/cart/contracts/cli.md"
	referencesPath = "proviso/references.md"
	manifestPath   = "proviso/manifest.sum"
)

// The plain tree holds no rule, so that the text of every part is fixed by
// part_kinds alone, and the whole manifest can be computed from the
// contract (see plainParts).
const plainGlossaryMD = `# Glossary

## <a id="cart.item"></a> item

- Definition: A product placed
  in the cart.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: items
`

const plainParametersMD = parametersMD + `
## <a id="cart.sizes"></a> sizes

- Configurable: no
- Rationale: carts come in two sizes.

The sizes of a cart.

| size | slots |
| --- | --- |
| small | 5 |
| large | 20 |
`

// featureHead is the feature cart.add without its rules.
var featureHead = featureMD[:strings.Index(featureMD, "## <a id=\"cart.add.limit\"")]

func plainTree() tree {
	return tree{}.with(
		systemPath, systemMD,
		capPath, capabilityMD,
		glossaryPath, plainGlossaryMD,
		paramsPath, plainParametersMD,
		featurePath, featureHead,
		contractPath, cliContractMD,
	)
}

// plainParts is every part of every element of plainTree, with its text as
// part_kinds gives it.
func plainParts() []mpart {
	return []mpart{
		{"system", "heading", "Shop"},
		{"system", "section Goal", "The shop sells goods to buyers."},
		{"system", "section Rationale", "Buyers want goods."},
		{"system", "section Scope", "- Selling goods."},
		{"system", "section Out of scope", "- Shipping goods."},
		{"system", "section Audience", "- Buyers."},
		{"system", "section Design boundary", "- A web browser."},
		{"cart", "heading", "Cart"},
		{"cart", "section Goal", "Buyers collect goods before they pay."},
		{"cart", "section Rationale", "Paying once for many goods is easier."},
		{"cart.item", "heading", "item"},
		{"cart.item", "field Definition", "A product placed in the cart."},
		{"cart.item", "field Contact", "Bruno Baketarić"},
		{"cart.item", "field Status", "defined"},
		{"cart.item", "field Aliases", "items"},
		{"cart.max-quantity", "heading", "max_quantity"},
		{"cart.max-quantity", "field Value", "10"},
		{"cart.max-quantity", "field Unit", "pieces"},
		{"cart.max-quantity", "field Configurable", "no"},
		{"cart.max-quantity", "field Rationale", "a shelf holds ten pieces."},
		{"cart.sizes", "heading", "sizes"},
		{"cart.sizes", "field Configurable", "no"},
		{"cart.sizes", "field Rationale", "carts come in two sizes."},
		{"cart.sizes", "body", "The sizes of a cart.\n\n| size | slots |\n| --- | --- |\n| small | 5 |\n| large | 20 |"},
		{"cart.add", "heading", "Adding to the cart"},
		{"cart.add", "field Contact", "Dana Reyes"},
		{"cart.add", "section Goal", "Buyers put goods into the cart."},
		{"cart.add", "section Rationale", "A cart that takes anything would empty the shelves."},
		{"cart.cli.lint", "heading", "lint"},
		{"cart.cli.lint", "body", "The command."},
		{"cart.cli.lint.exit-status", "heading", "exit-status"},
		{"cart.cli.lint.exit-status", "body", "The exit status."},
	}
}

// The rich tree holds an element of every type that part_kinds tells apart:
// the baseline tree, a parameter with a table, contract elements with a
// member, and x.md with a rule that has a Given and a Rationale, and an
// invariant.
const richXMD = xHead + `## <a id="cart.x.open"></a> Rule: An open shop greets buyers

### Rationale

Closed shops sell nothing.

´´´guarded
Given     the shop is open
When      a buyer enters with the [item] <product>
Then      the shop greets the buyer
  And     nothing else changes
´´´

| product | outcome |
| --- | --- |
| apple | greeted |

## <a id="cart.x.never"></a> Invariant: The cart never overflows

### Rationale

Shelves are small.

### Ranges over

Every addition to the cart.

### Property

The cart holds at most ten pieces of each product.
`

func richTree() tree {
	return baseline().with(
		paramsPath, plainParametersMD,
		contractPath, cliContractMD,
		xPath, richXMD,
	)
}

// The diff tree is what the diff tests accept and then amend: the baseline
// tree, its rule cart.add.limit with a Rationale, x.md with two rules of its
// own, and a term defined by a reference, which the rule cart.x.line marks.
const diffGlossaryMD = glossaryMD + `
## <a id="cart.thematic-break"></a> thematic break

- Defined by: [CommonMark §4.1](../../references.md#system.ref.commonmark)
- Contact: Dana Reyes
- Status: defined
`

const diffXMD = xHead + `## <a id="cart.x.r"></a> Rule: Under test

´´´guarded
` + idBlock + `
´´´

` + idTable + `

## <a id="cart.x.line"></a> Rule: A receipt ends with a break

´´´guarded
When      the shop prints the receipt <receipt>
Then      <receipt> ends with a [thematic break]
  And     nothing else changes
´´´

| receipt | outcome |
| --- | --- |
| the first receipt | a break at the end |
`

const limitRationale = "### Rationale\n\nA shelf holds ten pieces.\n\n"

func diffTree() tree {
	return baseline().
		edit(featurePath, "Rule: The cart takes a limited quantity\n\n", "Rule: The cart takes a limited quantity\n\n"+limitRationale).
		with(
			glossaryPath, diffGlossaryMD,
			referencesPath, systemReferencesMD,
			xPath, diffXMD,
		)
}

// limitBlock is the rule cart.add.limit of diffTree, from its heading to the
// next rule's heading.
func limitBlock() string {
	text, _ := elementText(diffTree()[featurePath], "cart.add.limit")
	return text
}

// TestBaselineFixtures checks the ground the baseline tests stand on: lint
// reports no finding about their trees, or only asks.
func TestBaselineFixtures(t *testing.T) {
	for name, tr := range map[string]tree{"plain": plainTree(), "rich": richTree(), "diff": diffTree()} {
		t.Run(name, func(t *testing.T) {
			check(t, tr).clean(t)
		})
	}
	// The undecided tree holds asks, and no error.
	t.Run("undecided", func(t *testing.T) {
		o := check(t, undecidedTree())
		if o.json.exit != 0 {
			t.Errorf("want asks only, got exit status %d:\n%s%s", o.json.exit, dump(o.report.Findings), o.json.stderr)
		}
	})
}
