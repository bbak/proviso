package acceptance

// Feature baseline.pack: proviso/capabilities/baseline/features/pack.md, and
// the contracts system.cli.pack, system.cli.pack-text and
// system.cli.context-pack (proviso/contracts/cli.md). Every case
// accepts the pack tree (see pack_trees_test.go), amends it and runs
// `proviso pack`. The Examples rows name elements of Proviso's own tree; each
// case names the element of the pack tree that stands in for it.

import (
	"path/filepath"
	"strings"
	"testing"
)

// Amendments of the pack tree.
var (
	pCapThen     = change{payPath, "Then      the shop takes the order", "Then      the shop accepts the order"}
	pCapProvided = change{payPath, "Provided  <cart> holds at most ´max_quantity´ pieces of each product", "Provided  <cart> holds no more than ´max_quantity´ pieces of each product"}
	pCapRow      = change{payPath, "| the first cart | yes | taken |", "| the first cart | yes | accepted |"}
	pCapOpen     = change{payPath, "Then      the shop takes the order", "Then      ?"}
	pTotalThen   = change{payPath, "Then      the shop charges the [price] of <items>", "Then      the shop charges the whole [price] of <items>"}
	pLinesThen   = change{payPath, "Then      the shop shows ´line.count´ for each line of <order>", "Then      the shop shows ´line.count´ for every line of <order>"}
	pReceiptThen = change{payPath, "Then      <receipt> ends with a [thematic break]", "Then      <receipt> ends with one [thematic break]"}
	pPayGoal     = change{payPath, "Buyers pay for what the cart holds.", "Buyers pay for all the cart holds."}
	pSpokenThen  = change{featurePath, "Then      the cart holds <quantity> more\n", "Then      the cart holds <quantity> more pieces\n"}
	pRefundThen  = change{refundPath, "Then      the shop refunds the [price] of <product>", "Then      the shop refunds the full [price] of <product>"}
	pShelfDef    = change{glossaryPath, "- Definition: Where the shop keeps goods.", "- Definition: Where the shop keeps its goods."}
	pPriceDef    = change{sysGlossaryPath, "- Definition: What a product costs.", "- Definition: What a product costs, taxes included."}
)

// removedFrom is the change that deletes the element id from path in tr.
func removedFrom(tr tree, path, id string) change {
	text, _ := elementText(tr[path], id)
	return change{path, text, ""}
}

// packOf accepts the pack tree, amends it with cs and runs pack. It returns
// the outcome, the amended tree and the accepted commit.
func packOf(t *testing.T, cs ...change) (packOutcome, tree, string) {
	t.Helper()
	return packFrom(t, packTree(), cs...)
}

// packFrom does what packOf does, with the tree from instead of the pack tree.
func packFrom(t *testing.T, from tree, cs ...change) (packOutcome, tree, string) {
	t.Helper()
	r, base := accepted(t, from)
	tr := amended(from.clone(), cs...)
	r.set(tr)
	o := r.pack()
	o.ran(t)
	return o, tr, base
}

// The frames each level brings: the feature and the capability holding an
// element, their invariants, the capability's constraints, and the system's
// invariants and constraints.
var (
	sysFrames     = []string{"system.fast", "system.honest"}
	cartFrames    = append([]string{"cart", "cart.bounded", "cart.light"}, sysFrames...)
	addFrames     = append([]string{"cart.add", "cart.add.never"}, cartFrames...)
	payFrames     = append([]string{"cart.pay"}, cartFrames...)
	refundFrames  = append([]string{"billing", "billing.balanced", "billing.refund"}, sysFrames...)
	billingFrames = append([]string{"billing.balanced"}, sysFrames...)
)

func union(lists ...[]string) []string {
	seen := map[string]bool{}
	var out []string
	for _, l := range lists {
		for _, id := range l {
			if !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	return out
}

func TestBaselinePackContents(t *testing.T) {
	t.Parallel()
	t.Run("row1 a rule's Then step changed", func(t *testing.T) {
		t.Parallel()
		// lint.run.exit: cart.pay.cap, which no rule hands over to.
		o, tr, base := packOf(t, pCapThen)
		o.exactly(t, "System", "system")
		o.exactly(t, "Amendment", "cart.pay.cap")
		if it := o.item(t, "cart.pay.cap"); it.status != "Changed `cart.pay.cap`: `step Then` changed." {
			t.Errorf("want the line %q, got %q", "Changed `cart.pay.cap`: `step Then` changed.", it.status)
		}
		if o.revision != base {
			t.Errorf("want the line Former text from `%s`. first under Amendment, got the revision %q", base, o.revision)
		}
		system, _ := elementText(tr[systemPath], "system")
		if it := o.item(t, "system"); !sameText(it.text(), system) {
			t.Errorf("want the system as it stands now\n%s\ngot\n%s", system, it.text())
		}
		o.standsAt(t, tr)
	})
	t.Run("row2 a rule removed", func(t *testing.T) {
		t.Parallel()
		// lint.run.root: cart.pay.total.
		o, _, base := packOf(t, removedFrom(packTree(), payPath, "cart.pay.total"))
		o.headingsAre(t, "System", "Amendment")
		o.exactly(t, "Amendment", "cart.pay.total")
		if it := o.item(t, "cart.pay.total"); it.status != "Removed `cart.pay.total`." {
			t.Errorf("want the line %q, got %q", "Removed `cart.pay.total`.", it.status)
		}
		former, _ := elementText(packTree()[payPath], "cart.pay.total")
		if it := o.item(t, "cart.pay.total"); !sameText(unfenced(it.body), former) {
			t.Errorf("want the former text of cart.pay.total in place of its text\n%s\ngot\n%s", former, it.text())
		}
		if deref(o.cp.Revision) != base || o.revision != base {
			t.Errorf("want the revision %s, got %s in context_pack and %q in pack_text", base, deref(o.cp.Revision), o.revision)
		}
	})
	t.Run("row3 none: the tree as the manifest records it", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, packTree())
		o := r.pack()
		o.headingsAre(t, "System")
		o.exactly(t, "System", "system")
	})
	t.Run("derived: an added element", func(t *testing.T) {
		t.Parallel()
		o, _, _ := packOf(t, addTip)
		o.exactly(t, "Amendment", "cart.pay.tip")
		if it := o.item(t, "cart.pay.tip"); it.status != "Added `cart.pay.tip`." {
			t.Errorf("want the line %q, got %q", "Added `cart.pay.tip`.", it.status)
		}
		if deref(o.cp.Revision) != "(none)" || o.revision != "" {
			t.Errorf("want no revision, as no former text was needed, got %s in context_pack and %q in pack_text", deref(o.cp.Revision), o.revision)
		}
	})
	t.Run("derived: elements changed through a dependency, with their chains", func(t *testing.T) {
		t.Parallel()
		o, _, _ := packOf(t, iParamValue)
		o.exactly(t, "Amendment", "cart.max-quantity", "cart.add.limit", "cart.add.spoken", "cart.pay.cap")
		want := map[string]string{
			"cart.max-quantity": "Changed `cart.max-quantity`: `field Value` changed.",
			"cart.add.limit":    "Changed through `cart.add.limit`: `cart.max-quantity`.",
			"cart.add.spoken":   "Changed through `cart.add.spoken`: `cart.max-quantity` → `cart.add.limit`.",
			"cart.pay.cap":      "Changed through `cart.pay.cap`: `cart.max-quantity`.",
		}
		for id, line := range want {
			if it := o.item(t, id); it.status != line {
				t.Errorf("want the line %q, got %q", line, it.status)
			}
		}
	})
	t.Run("derived: every chain that reaches an element", func(t *testing.T) {
		t.Parallel()
		// That the chains stand in diff's order is checked for every run
		// (see checkAmendment).
		o, _, _ := packOf(t, iItemDef, iParamValue)
		st := o.item(t, "cart.add.limit").status
		a, b := "`cart.item`", "`cart.max-quantity`"
		if st != "Changed through `cart.add.limit`: "+a+"; "+b+"." && st != "Changed through `cart.add.limit`: "+b+"; "+a+"." {
			t.Errorf("want cart.add.limit changed through %s and %s, separated by \"; \", got %q", a, b, st)
		}
	})
	t.Run("derived: an element not to be built", func(t *testing.T) {
		t.Parallel()
		o, tr, _ := packOf(t, pCapOpen)
		want := "Not to be built: " + payPath + ":" + itoa(tr.line(payPath, "Then      ?")) + "."
		if it := o.item(t, "cart.pay.cap"); it.notBuilt != want {
			t.Errorf("want the second line %q, got %q", want, it.notBuilt)
		}
	})
	t.Run("derived: without a manifest, every element under Amendment", func(t *testing.T) {
		t.Parallel()
		dir := packTree().write(t)
		o := packIn(t, dir, dir, filepath.Join(dir, "proviso"))
		o.headingsAre(t, "Amendment")
		o.exactly(t, "Amendment", keysOf(packStandsIn)...)
	})
}

func TestBaselinePackFrames(t *testing.T) {
	t.Parallel()
	t.Run("row1 a rule's Then step changed", func(t *testing.T) {
		t.Parallel()
		// lint.run.exit: cart.pay.cap.
		o, tr, _ := packOf(t, pCapThen)
		o.exactly(t, "Frames", payFrames...)
		o.standsAt(t, tr)
	})
	t.Run("row2 a glossary entry of the capability, its Definition changed", func(t *testing.T) {
		t.Parallel()
		// baseline.former-text: cart.shelf, which no rule marks.
		o, _, _ := packOf(t, pShelfDef)
		o.exactly(t, "Amendment", "cart.shelf")
		o.exactly(t, "Frames", cartFrames...)
	})
	t.Run("row3 a system constraint, its Property changed", func(t *testing.T) {
		t.Parallel()
		// system.git: system.fast.
		o, _, _ := packOf(t, iFastProperty)
		o.exactly(t, "Amendment", "system.fast")
		o.exactly(t, "Frames", "system.honest")
	})
	t.Run("row4 a rule removed", func(t *testing.T) {
		t.Parallel()
		// lint.run.root: cart.pay.total.
		o, _, _ := packOf(t, removedFrom(packTree(), payPath, "cart.pay.total"))
		o.exactly(t, "Frames")
	})
	t.Run("derived: a feature, its Goal changed", func(t *testing.T) {
		t.Parallel()
		o, _, _ := packOf(t, pPayGoal)
		o.exactly(t, "Amendment", "cart.pay")
		o.exactly(t, "Frames", cartFrames...)
	})
	t.Run("derived: a capability, its Goal changed", func(t *testing.T) {
		t.Parallel()
		o, _, _ := packOf(t, iBillingGoal)
		o.exactly(t, "Amendment", "billing")
		o.exactly(t, "Frames", billingFrames...)
	})
	t.Run("derived: an invariant of a feature, its Property changed", func(t *testing.T) {
		t.Parallel()
		o, _, _ := packOf(t, iNeverProp)
		o.exactly(t, "Amendment", "cart.add.never")
		o.exactly(t, "Frames", "cart.add", "cart", "cart.bounded", "cart.light", "system.fast", "system.honest")
	})
	t.Run("derived: elements changed through a dependency bring their frames", func(t *testing.T) {
		t.Parallel()
		// system.price stands at system level; only the rules changed
		// through it bring the frames of cart and billing.
		o, _, _ := packOf(t, pPriceDef)
		o.exactly(t, "Amendment", "system.price", "cart.pay.total", "cart.add.spoken", "billing.refund.full")
		o.exactly(t, "Frames", union(addFrames, payFrames, refundFrames)...)
	})
	t.Run("row5 a rule named by --element", func(t *testing.T) {
		t.Parallel()
		// lint.run.exit: cart.pay.cap, read as an added one.
		r, _ := accepted(t, packTree())
		o := r.packElement("cart.pay.cap")
		o.exactly(t, "Frames", payFrames...)
		o.standsAt(t, packTree())
	})
	t.Run("derived: a glossary entry of the capability named by --element", func(t *testing.T) {
		t.Parallel()
		// As row2: no feature holds it.
		r, _ := accepted(t, packTree())
		r.packElement("cart.shelf").exactly(t, "Frames", cartFrames...)
	})
	t.Run("derived: a system constraint named by --element", func(t *testing.T) {
		t.Parallel()
		// As row3: no feature or capability holds it.
		r, _ := accepted(t, packTree())
		r.packElement("system.fast").exactly(t, "Frames", "system.honest")
	})
}

func TestBaselinePackChain(t *testing.T) {
	t.Parallel()
	t.Run("row1 a rule handing over, its Then step changed", func(t *testing.T) {
		t.Parallel()
		// baseline.diff.added: cart.add.spoken, handing over to
		// cart.add.limit; both mark terms.
		o, tr, _ := packOf(t, pSpokenThen)
		o.exactly(t, "Amendment", "cart.add.spoken")
		o.exactly(t, "Context", "cart.add.limit", "cart.item", "cart.max-quantity", "system.price")
		o.standsAt(t, tr)
	})
	t.Run("row2 a rule referring to a member", func(t *testing.T) {
		t.Parallel()
		// findings.anchor-misplaced: line.count, a member of cart.api.line,
		// which represents [item].
		o, _, _ := packOf(t, pLinesThen)
		o.exactly(t, "Amendment", "cart.pay.lines")
		o.exactly(t, "Context", "cart.api.line", "cart.api.line.count", "cart.item")
	})
	t.Run("row3 a rule marking a term defined by a reference", func(t *testing.T) {
		t.Parallel()
		o, _, _ := packOf(t, pReceiptThen)
		o.exactly(t, "Amendment", "cart.pay.receipt")
		o.exactly(t, "Context", "cart.thematic-break", "system.ref.commonmark")
	})
	t.Run("row4 a rule marking a term whose Definition marks another term", func(t *testing.T) {
		t.Parallel()
		// [line ending]: [price], which the Definition of shelf marks;
		// cart.pay.open, which depended on nothing, now marks [shelf].
		shelfPrice := packTree().edit(glossaryPath, "- Definition: Where the shop keeps goods.", "- Definition: Where the shop keeps goods at a [price].")
		o, _, _ := packFrom(t, shelfPrice, change{payPath, "Then      the shop greets every buyer", "Then      the shop greets every buyer at the [shelf]"})
		o.exactly(t, "Amendment", "cart.pay.open")
		o.exactly(t, "Context", "cart.shelf", "system.price")
	})
	t.Run("row5 a rule marking a term other rules mark", func(t *testing.T) {
		t.Parallel()
		// [spec root]: [price], which cart.pay.total and cart.add.spoken
		// mark as well.
		o, _, _ := packOf(t, pRefundThen)
		o.exactly(t, "Amendment", "billing.refund.full")
		o.exactly(t, "Context", "system.price")
	})
	t.Run("row6 a rule removed", func(t *testing.T) {
		t.Parallel()
		// lint.run.root: cart.add.spoken, which marks a term and hands over.
		o, _, _ := packOf(t, removedFrom(packTree(), featurePath, "cart.add.spoken"))
		o.exactly(t, "Amendment", "cart.add.spoken")
		o.exactly(t, "Context")
	})
	t.Run("derived: elements changed through a dependency follow the chain", func(t *testing.T) {
		t.Parallel()
		// cart.max-quantity is under Amendment; cart.add.limit brings
		// cart.item, cart.add.spoken brings system.price.
		o, _, _ := packOf(t, iParamValue)
		o.exactly(t, "Context", "cart.item", "system.price")
	})
	t.Run("derived: a changed member brings its holder and the term it represents", func(t *testing.T) {
		t.Parallel()
		o, _, _ := packOf(t, iCountBody)
		o.exactly(t, "Amendment", "cart.api.line.count", "cart.pay.lines", "cart.pay.order")
		o.exactly(t, "Context", "cart.api.line", "cart.item")
	})
	t.Run("derived: a removed member, while its holder stays, brings nothing", func(t *testing.T) {
		t.Parallel()
		// cart.pay.order depends on the removed cart.api.line.count, which
		// stands under Amendment and is taken no further.
		o, _, _ := packOf(t, removedFrom(packTree(), apiPath, "cart.api.line.count"), linesToLine)
		o.exactly(t, "Amendment", "cart.api.line.count", "cart.pay.lines", "cart.pay.order")
		o.exactly(t, "Context", "cart.api.line", "cart.item")
	})
	t.Run("row7 a rule handing over, named by --element", func(t *testing.T) {
		t.Parallel()
		// baseline.diff.added: cart.add.spoken, as in row1.
		r, _ := accepted(t, packTree())
		o := r.packElement("cart.add.spoken")
		o.exactly(t, "Element", "cart.add.spoken")
		o.exactly(t, "Context", "cart.add.limit", "cart.item", "cart.max-quantity", "system.price")
		o.standsAt(t, packTree())
	})
	t.Run("derived: a rule referring to a member, named by --element", func(t *testing.T) {
		t.Parallel()
		// As row2: the member's holder, and the term the holder represents.
		r, _ := accepted(t, packTree())
		r.packElement("cart.pay.lines").exactly(t, "Context", "cart.api.line", "cart.api.line.count", "cart.item")
	})
}

func TestBaselinePackOnce(t *testing.T) {
	t.Parallel()
	t.Run("row1 the system, its Scope changed", func(t *testing.T) {
		t.Parallel()
		o, _, _ := packOf(t, iSystemScope)
		o.headingsAre(t, "Amendment", "Frames")
		o.exactly(t, "Amendment", "system")
		o.exactly(t, "Frames", sysFrames...)
	})
	t.Run("row2 a feature, its Goal changed, holding a changed rule", func(t *testing.T) {
		t.Parallel()
		o, _, _ := packOf(t, pPayGoal, pCapThen)
		o.exactly(t, "Amendment", "cart.pay", "cart.pay.cap")
		o.exactly(t, "Frames", cartFrames...)
	})
	t.Run("row3 a glossary entry, its Definition changed, marked by a changed rule", func(t *testing.T) {
		t.Parallel()
		o, _, _ := packOf(t, iItemDef, pTotalThen)
		o.exactly(t, "Amendment", "cart.item", "cart.pay.total", "cart.add.limit", "cart.add.spoken", "cart.api.line", "cart.pay.order")
		o.exactly(t, "Context", "cart.api.line.count", "cart.max-quantity", "system.price")
	})
	// That no element appears twice is checked for every run (see packIn).
	t.Run("row4 a feature, unchanged, holding two changed rules", func(t *testing.T) {
		t.Parallel()
		// lint.run: cart.pay, holding cart.pay.cap and cart.pay.total.
		o, _, _ := packOf(t, pCapThen, pTotalThen)
		o.exactly(t, "Amendment", "cart.pay.cap", "cart.pay.total")
		o.exactly(t, "Frames", payFrames...)
	})
	t.Run("row5 a term, unchanged, marked by two changed rules", func(t *testing.T) {
		t.Parallel()
		// [spec root]: [price], marked by cart.pay.total and
		// billing.refund.full.
		o, _, _ := packOf(t, pTotalThen, pRefundThen)
		o.exactly(t, "Amendment", "billing.refund.full", "cart.pay.total")
		o.exactly(t, "Context", "cart.item", "system.price")
	})
	t.Run("row4 the system, named by --element", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, packTree())
		o := r.packElement("system")
		o.headingsAre(t, "Element", "Frames")
		o.exactly(t, "Element", "system")
		o.exactly(t, "System")
		o.exactly(t, "Frames", sysFrames...)
	})
	t.Run("derived: a feature named by --element is not repeated under Frames", func(t *testing.T) {
		t.Parallel()
		// cart.pay is the feature baseline.pack.frames takes for it.
		r, _ := accepted(t, packTree())
		o := r.packElement("cart.pay")
		o.exactly(t, "Element", "cart.pay")
		o.exactly(t, "Frames", cartFrames...)
	})
}

func TestBaselinePackOrder(t *testing.T) {
	t.Parallel()
	// That the elements under each heading stand in the order of the tree,
	// with removed elements after them, and the parts of each in diff's
	// order, is checked for every run (see packIn); here each row's outcome
	// is checked as well.
	t.Run("row1 under Frames", func(t *testing.T) {
		t.Parallel()
		// system.git, lint.run, lint: system.fast and system.honest, the
		// feature cart.pay, the capability cart with its constraint
		// cart.light and its invariant cart.bounded. By identifier, cart.bounded
		// would come second.
		o, _, _ := packOf(t, pCapThen)
		o.inOrder(t, "Frames", "cart", "cart.light", "cart.pay", "cart.bounded", "system.fast", "system.honest")
	})
	t.Run("row2 under Amendment, as they stand in one file", func(t *testing.T) {
		t.Parallel()
		// lint.run.exit, lint.run.root: cart.pay.cap and cart.pay.total, of
		// which cart.pay.total stands first in pay.md.
		o, _, _ := packOf(t, pCapThen, pTotalThen)
		o.inOrder(t, "Amendment", "cart.pay.total", "cart.pay.cap")
	})
	t.Run("row3 under Amendment, a removed element follows", func(t *testing.T) {
		t.Parallel()
		// lint.run.root, removed, and lint.run.exit, changed: cart.pay.total,
		// which stood before cart.pay.cap in pay.md.
		o, _, _ := packOf(t, removedFrom(packTree(), payPath, "cart.pay.total"), pCapThen)
		o.inOrder(t, "Amendment", "cart.pay.cap", "cart.pay.total")
	})
	t.Run("row4 under Context, the members as the contract gives them", func(t *testing.T) {
		t.Parallel()
		// system.cli.diff and its members: cart.api.line, with
		// cart.api.line.count before cart.api.line.amount in api.md, and
		// cart.item, which it represents, in glossary.md.
		o, _, _ := packFrom(t, memberTree(), change{payPath, "for each line of <order>", "for every line of <order>"})
		o.inOrder(t, "Amendment", "cart.pay.lines")
		o.inOrder(t, "Context", "cart.api.line", "cart.api.line.count", "cart.api.line.amount", "cart.item")
	})
	t.Run("derived: a removed element follows, though its file and its identifier come first", func(t *testing.T) {
		t.Parallel()
		o, _, _ := packOf(t, removedFrom(packTree(), featurePath, "cart.add.full"), pCapThen)
		o.inOrder(t, "Amendment", "cart.pay.cap", "cart.add.full")
	})
	t.Run("derived: removed elements by identifier, not where they stood", func(t *testing.T) {
		t.Parallel()
		o, _, _ := packOf(t, removedFrom(packTree(), payPath, "cart.pay.total"), removedFrom(packTree(), payPath, "cart.pay.cap"))
		o.inOrder(t, "Amendment", "cart.pay.cap", "cart.pay.total")
	})
	t.Run("derived: elements changed through a dependency stand where the tree holds them", func(t *testing.T) {
		t.Parallel()
		// cart.max-quantity stands in parameters.md, after features/.
		o, _, _ := packOf(t, iParamValue)
		o.inOrder(t, "Amendment", "cart.add.limit", "cart.add.spoken", "cart.pay.cap", "cart.max-quantity")
	})
	t.Run("derived: paths compared byte by byte, not folder by folder", func(t *testing.T) {
		t.Parallel()
		// capabilities/cart-gift/ comes before capabilities/cart/, since -
		// comes before /.
		o, _, _ := packFrom(t, giftTree(), iCartGoal, change{giftPath, "Buyers give carts away.", "Buyers give whole carts away."})
		o.inOrder(t, "Amendment", "cart-gift", "cart")
		o.inOrder(t, "Frames", "cart.light", "cart.bounded", "system.fast", "system.honest")
	})
	t.Run("derived: the chains in diff's order", func(t *testing.T) {
		t.Parallel()
		// cart.pay.receipt marks cart.thematic-break, defined by
		// system.ref.commonmark; both changed.
		o, _, _ := packOf(t, commonmarkNew, breakDefinedBy)
		want := "Changed through `cart.pay.receipt`: `cart.thematic-break`; `system.ref.commonmark` → `cart.thematic-break`."
		if it := o.item(t, "cart.pay.receipt"); it.status != want {
			t.Errorf("want the line %q, got %q", want, it.status)
		}
	})
	t.Run("derived: the lines not decided in diff's order, each once", func(t *testing.T) {
		t.Parallel()
		// cart.pay.cap, undecided in its Provided and its Then step, and in
		// two cells of one Examples row.
		o, tr, _ := packOf(t, pCapOpen,
			change{payPath, "Provided  <cart> holds at most ´max_quantity´ pieces of each product", "Provided  <cart> holds at most ? pieces of each product"},
			change{payPath, "| the second cart | no | refused |", "| the second cart | no | refused |\n| the third cart | ? | ? |"})
		at := func(sub string) string { return payPath + ":" + itoa(tr.line(payPath, sub)) }
		want := "Not to be built: " + at("at most ? pieces") + ", " + at("Then      ?") + ", " + at("| the third cart | ? | ? |") + "."
		if it := o.item(t, "cart.pay.cap"); it.notBuilt != want {
			t.Errorf("want the line %q, got %q", want, it.notBuilt)
		}
	})
	t.Run("derived: the parts in diff's order", func(t *testing.T) {
		t.Parallel()
		o, _, _ := packOf(t, pCapThen, pCapProvided, pCapRow)
		key := "row " + rowKey("cart", "the first cart")
		want := "Changed `cart.pay.cap`: `" + key + "` changed, `step Provided` changed, `step Then` changed."
		if it := o.item(t, "cart.pay.cap"); it.status != want {
			t.Errorf("want the line %q, got %q", want, it.status)
		}
		// Each part's text now, then its former text.
		want = "Now " + key + "|Former " + key + "|Now step Provided|Former step Provided|Now step Then|Former step Then"
		if got := blocksOf(o.item(t, "cart.pay.cap")); got != want {
			t.Errorf("want the lines Now and Former %s, in that order, got %s", want, got)
		}
	})
}

// blocksOf lists the lines Now and Former of it, in their order, joined by
// "|".
func blocksOf(it packItem) string {
	var got []string
	for _, b := range it.blocks {
		if b.now {
			got = append(got, "Now "+b.part)
		} else {
			got = append(got, "Former "+b.part)
		}
	}
	return strings.Join(got, "|")
}

// blockOf returns the text after the line Now (now true) or Former of the
// part named part of it.
func blockOf(t *testing.T, it packItem, part string, now bool) string {
	t.Helper()
	for _, b := range it.blocks {
		if b.part == part && b.now == now {
			return b.content
		}
	}
	t.Errorf("want a line Now or Former (now: %v) for %s of %s, got %s", now, part, it.id, blocksOf(it))
	return ""
}

// TestSystemCliPackText checks what system.cli.pack-text says of each part
// named under Amendment: the text of an added or changed part as it stands
// now, after a line Now, then the former text of a changed or removed part,
// after a line Former. The layout of both, and that they agree with the
// element diff, is checked for every run (see checkAmendment); here against
// the tree.
func TestSystemCliPackText(t *testing.T) {
	t.Parallel()
	thenNow := "Then      the shop accepts the order\n  And     nothing else changes"
	thenFormer := "Then      the shop takes the order\n  And     nothing else changes"
	t.Run("derived: a changed step, its text now, then its former text", func(t *testing.T) {
		t.Parallel()
		o, _, _ := packOf(t, pCapThen)
		it := o.item(t, "cart.pay.cap")
		if got, want := blocksOf(it), "Now step Then|Former step Then"; got != want {
			t.Errorf("want the lines %s, got %s", want, got)
		}
		checkNow(t, partChange{Part: "step Then", Text: ptr(blockOf(t, it, "step Then", true))}, thenNow)
		checkFormer(t, partChange{Part: "step Then", Former: ptr(blockOf(t, it, "step Then", false))}, thenFormer)
	})
	t.Run("derived: an added Examples row, its text now only", func(t *testing.T) {
		t.Parallel()
		o, _, _ := packOf(t, change{payPath, "| the second cart | no | refused |\n", "| the second cart | no | refused |\n| the third cart | yes | taken |\n"})
		it := o.item(t, "cart.pay.cap")
		key := "row " + rowKey("cart", "the third cart")
		if got, want := blocksOf(it), "Now "+key; got != want {
			t.Errorf("want the lines %s, got %s", want, got)
		}
		checkNow(t, partChange{Part: key, Text: ptr(blockOf(t, it, key, true))}, "| the third cart | yes | taken |")
	})
	t.Run("derived: an Examples row with a new outcome, its text now, then its former text", func(t *testing.T) {
		t.Parallel()
		o, _, _ := packOf(t, pCapRow)
		it := o.item(t, "cart.pay.cap")
		key := "row " + rowKey("cart", "the first cart")
		if got, want := blocksOf(it), "Now "+key+"|Former "+key; got != want {
			t.Errorf("want the lines %s, got %s", want, got)
		}
		checkNow(t, partChange{Part: key, Text: ptr(blockOf(t, it, key, true))}, "| the first cart | yes | accepted |")
		checkFormer(t, partChange{Part: key, Former: ptr(blockOf(t, it, key, false))}, "| the first cart | yes | taken |")
	})
	t.Run("derived: a removed Examples row, its former text only", func(t *testing.T) {
		t.Parallel()
		// The pack tree accepted with a third row, which is then removed.
		third := "| the third cart | yes | taken |\n"
		withThird := amended(packTree(), change{payPath, "| the second cart | no | refused |\n", "| the second cart | no | refused |\n" + third})
		o, _, _ := packFrom(t, withThird, change{payPath, third, ""})
		it := o.item(t, "cart.pay.cap")
		key := "row " + rowKey("cart", "the third cart")
		if got, want := blocksOf(it), "Former "+key; got != want {
			t.Errorf("want the lines %s, got %s", want, got)
		}
		checkFormer(t, partChange{Part: key, Former: ptr(blockOf(t, it, key, false))}, third)
	})
}

func TestBaselinePackOutput(t *testing.T) {
	t.Parallel()
	exit0 := func(t *testing.T, o packOutcome) {
		t.Helper()
		if o.json.exit != 0 || o.text.exit != 0 {
			t.Errorf("want exit status 0, got %d with --json, %d without", o.json.exit, o.text.exit)
		}
	}
	t.Run("row1 one changed rule, --json", func(t *testing.T) {
		t.Parallel()
		o, _, _ := packOf(t, pCapThen)
		exit0(t, o)
		if len(o.cp.Diff) != 1 || o.cp.Diff[0].ID != "cart.pay.cap" {
			t.Errorf("want a context_pack whose diff holds cart.pay.cap, got %s", o.json.stdout)
		}
	})
	t.Run("row2 one changed rule, as text", func(t *testing.T) {
		t.Parallel()
		o, _, _ := packOf(t, pCapThen)
		exit0(t, o)
		o.headingsAre(t, "System", "Amendment", "Frames", "Context")
	})
	t.Run("row3 an empty element diff, without .amendment.md, as text", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, packTree())
		o := r.pack()
		exit0(t, o)
		o.headingsAre(t, "System")
	})
	t.Run("derived: an empty element diff, --json", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, packTree())
		o := r.pack()
		exit0(t, o)
		if len(o.cp.Diff) != 0 {
			t.Errorf("want diff empty, got %s", o.json.stdout)
		}
		var got []string
		for _, e := range o.cp.Elements {
			got = append(got, e.Heading+":"+e.ID)
		}
		if strings.Join(got, " ") != "system:system" {
			t.Errorf("want the system alone among the elements, got %v", got)
		}
	})
	t.Run("row4 lint.run.exit named by --element, --json", func(t *testing.T) {
		t.Parallel()
		// lint.run.exit: cart.pay.cap. packElementIn checks that diff is
		// empty and note null in every pack of a named element.
		r, _ := accepted(t, packTree())
		o := r.packElement("cart.pay.cap")
		exit0(t, o)
		if !o.cpOK || len(o.cp.Diff) != 0 || o.cp.Note != nil {
			t.Errorf("want a context_pack with an empty diff and a null note, got %s", o.json.stdout)
		}
		var got []string
		for _, e := range o.cp.Elements {
			if e.Heading == "element" {
				got = append(got, e.ID)
			}
		}
		if len(got) != 1 || got[0] != "cart.pay.cap" {
			t.Errorf("want cart.pay.cap alone with the heading element, got %v", got)
		}
	})
	t.Run("derived: lint.run.exit named by --element, as text", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, packTree())
		o := r.packElement("cart.pay.cap")
		exit0(t, o)
		o.headingsAre(t, "System", "Element", "Frames", "Context")
	})
}

// TestBaselineDiffFormer_Pack runs the cases of baseline.diff.former with `pack`.
func TestBaselineDiffFormer_Pack(t *testing.T) {
	t.Parallel()
	formerThen := "Then      the shop takes the order\n  And     nothing else changes"
	formerOf := func(t *testing.T, o packOutcome) string {
		t.Helper()
		for _, f := range o.item(t, "cart.pay.cap").formers {
			if f.part == "step Then" {
				return f.content
			}
		}
		t.Fatalf("want the former text of step Then of cart.pay.cap, got none")
		return ""
	}
	checkThen := func(t *testing.T, o packOutcome, want string) {
		t.Helper()
		checkFormer(t, partChange{Part: "step Then", Former: ptr(formerOf(t, o))}, want)
	}
	t.Run("row1 the manifest committed, the tree edited since", func(t *testing.T) {
		t.Parallel()
		r, base := accepted(t, packTree())
		r.set(amended(packTree(), pCapThen))
		o := r.pack()
		o.ran(t)
		if deref(o.cp.Revision) != base || o.revision != base {
			t.Errorf("want the revision %s, got %s in context_pack and %q in pack_text", base, deref(o.cp.Revision), o.revision)
		}
		checkThen(t, o, formerThen)
	})
	t.Run("row2 a merge that took the manifest from the mainline", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, packTree())
		r.git("checkout", "-q", "-b", "feature")
		feature := amended(packTree(), iCountBody)
		r.set(feature)
		r.commit("feature")
		r.git("checkout", "-q", "main")
		r.tr = packTree()
		r.set(amended(packTree(), pCapThen))
		r.manifest()
		e5f6 := r.commit("accept on the mainline")
		r.git("checkout", "-q", "feature")
		r.tr = feature
		r.git("merge", "-q", "--no-ff", "-m", "merge the mainline", "main")
		r.tr = amended(packTree(), iCountBody, pCapThen)
		r.set(amended(packTree(), iCountBody, pCapThen,
			change{payPath, "Then      the shop accepts the order", "Then      the shop gladly accepts the order"}))
		o := r.pack()
		o.ran(t)
		if deref(o.cp.Revision) != e5f6 || o.revision != e5f6 {
			t.Errorf("want the revision %s, got %s in context_pack and %q in pack_text", e5f6, deref(o.cp.Revision), o.revision)
		}
		checkThen(t, o, "Then      the shop accepts the order\n  And     nothing else changes")
	})
	t.Run("row3 the spec root in no git repository", func(t *testing.T) {
		t.Parallel()
		dir := packTree().write(t)
		skipIfGitAbove(t, dir)
		manifestOK(t, dir)
		amended(packTree(), pCapThen).writeTo(t, dir, false)
		packIn(t, dir, dir, filepath.Join(dir, "proviso")).stopped(t)
	})
	t.Run("row4 the manifest written but never committed", func(t *testing.T) {
		t.Parallel()
		r := newRepo(t, packTree())
		r.commit("tree")
		r.manifest()
		r.set(amended(packTree(), pCapThen))
		r.pack().stopped(t)
	})
	t.Run("row5 the manifest written again since and not committed", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, packTree())
		r.set(amended(packTree(), pCapThen))
		r.manifest()
		r.set(amended(packTree(), pCapThen, pCapProvided))
		r.pack().stopped(t)
	})
	t.Run("row6 the manifest committed together with a tree that differs from it in a part the element diff lists as changed", func(t *testing.T) {
		t.Parallel()
		r := newRepo(t, packTree())
		r.manifest()
		r.set(amended(packTree(), pCapProvided))
		r.commit("a1b2")
		r.set(amended(packTree(), pCapProvided, pCapThen))
		r.pack().stopped(t)
	})
	t.Run("row7 the manifest committed together with a tree that differs from it only in parts the element diff does not list as changed or removed", func(t *testing.T) {
		t.Parallel()
		r := newRepo(t, packTree())
		r.manifest()
		r.set(amended(packTree(), iCountBody))
		a1b2 := r.commit("a1b2")
		r.set(amended(packTree(), pCapThen))
		o := r.pack()
		o.ran(t)
		if deref(o.cp.Revision) != a1b2 || o.revision != a1b2 {
			t.Errorf("want the revision %s, got %s in context_pack and %q in pack_text", a1b2, deref(o.cp.Revision), o.revision)
		}
		checkThen(t, o, formerThen)
	})
	t.Run("derived: added elements need no former text, and no git", func(t *testing.T) {
		t.Parallel()
		dir := packTree().write(t)
		manifestOK(t, dir)
		skipIfGitAbove(t, dir)
		amended(packTree(), addTip).writeTo(t, dir, false)
		o := packIn(t, dir, dir, filepath.Join(dir, "proviso"))
		o.exactly(t, "Amendment", "cart.pay.tip")
		if o.cp.Revision != nil {
			t.Errorf("want revision null, as no former text was needed, got %q", *o.cp.Revision)
		}
	})
}

func ptr(s string) *string { return &s }

func itoa(n int) string { return derefInt(&n) }

func keysOf(m map[string]string) []string {
	var k []string
	for s := range m {
		k = append(k, s)
	}
	return k
}
