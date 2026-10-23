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
	r, base := accepted(t, packTree())
	tr := amended(packTree(), cs...)
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
	t.Run("row1 a rule's Then step changed", func(t *testing.T) {
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
		r, _ := accepted(t, packTree())
		o := r.pack()
		o.headingsAre(t, "System")
		o.exactly(t, "System", "system")
	})
	t.Run("derived: an added element", func(t *testing.T) {
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
		// The order of the chains is checked against element_diff's for
		// every run, as a set (see the report).
		o, _, _ := packOf(t, iItemDef, iParamValue)
		st := o.item(t, "cart.add.limit").status
		a, b := "`cart.item`", "`cart.max-quantity`"
		if st != "Changed through `cart.add.limit`: "+a+"; "+b+"." && st != "Changed through `cart.add.limit`: "+b+"; "+a+"." {
			t.Errorf("want cart.add.limit changed through %s and %s, separated by \"; \", got %q", a, b, st)
		}
	})
	t.Run("derived: an element not to be built", func(t *testing.T) {
		o, tr, _ := packOf(t, pCapOpen)
		want := "Not to be built: " + payPath + ":" + itoa(tr.line(payPath, "Then      ?")) + "."
		if it := o.item(t, "cart.pay.cap"); it.notBuilt != want {
			t.Errorf("want the second line %q, got %q", want, it.notBuilt)
		}
	})
	t.Run("derived: without a manifest, every element under Amendment", func(t *testing.T) {
		dir := packTree().write(t)
		o := packIn(t, dir, dir, filepath.Join(dir, "proviso"))
		o.headingsAre(t, "Amendment")
		o.exactly(t, "Amendment", keysOf(packStandsIn)...)
	})
}

func TestBaselinePackFrames(t *testing.T) {
	t.Run("row1 a rule's Then step changed", func(t *testing.T) {
		// lint.run.exit: cart.pay.cap.
		o, tr, _ := packOf(t, pCapThen)
		o.exactly(t, "Frames", payFrames...)
		o.standsAt(t, tr)
	})
	t.Run("row2 a glossary entry of the capability, its Definition changed", func(t *testing.T) {
		// baseline.part: cart.shelf, which no rule marks.
		o, _, _ := packOf(t, pShelfDef)
		o.exactly(t, "Amendment", "cart.shelf")
		o.exactly(t, "Frames", cartFrames...)
	})
	t.Run("row3 a system constraint, its Property changed", func(t *testing.T) {
		// system.git: system.fast.
		o, _, _ := packOf(t, iFastProperty)
		o.exactly(t, "Amendment", "system.fast")
		o.exactly(t, "Frames", "system.honest")
	})
	t.Run("row4 a rule removed", func(t *testing.T) {
		// lint.run.root: cart.pay.total.
		o, _, _ := packOf(t, removedFrom(packTree(), payPath, "cart.pay.total"))
		o.exactly(t, "Frames")
	})
	t.Run("derived: a feature, its Goal changed", func(t *testing.T) {
		o, _, _ := packOf(t, pPayGoal)
		o.exactly(t, "Amendment", "cart.pay")
		o.exactly(t, "Frames", cartFrames...)
	})
	t.Run("derived: a capability, its Goal changed", func(t *testing.T) {
		o, _, _ := packOf(t, iBillingGoal)
		o.exactly(t, "Amendment", "billing")
		o.exactly(t, "Frames", billingFrames...)
	})
	t.Run("derived: an invariant of a feature, its Property changed", func(t *testing.T) {
		o, _, _ := packOf(t, iNeverProp)
		o.exactly(t, "Amendment", "cart.add.never")
		o.exactly(t, "Frames", "cart.add", "cart", "cart.bounded", "cart.light", "system.fast", "system.honest")
	})
	t.Run("derived: elements changed through a dependency bring their frames", func(t *testing.T) {
		// system.price stands at system level; only the rules changed
		// through it bring the frames of cart and billing.
		o, _, _ := packOf(t, pPriceDef)
		o.exactly(t, "Amendment", "system.price", "cart.pay.total", "cart.add.spoken", "billing.refund.full")
		o.exactly(t, "Frames", union(addFrames, payFrames, refundFrames)...)
	})
}

func TestBaselinePackChain(t *testing.T) {
	t.Run("row1 a rule handing over, its Then step changed", func(t *testing.T) {
		// baseline.diff.added: cart.add.spoken, handing over to
		// cart.add.limit; both mark terms.
		o, tr, _ := packOf(t, pSpokenThen)
		o.exactly(t, "Amendment", "cart.add.spoken")
		o.exactly(t, "Context", "cart.add.limit", "cart.item", "cart.max-quantity", "system.price")
		o.standsAt(t, tr)
	})
	t.Run("row2 a rule referring to a member", func(t *testing.T) {
		// findings.anchor-misplaced: line.count, a member of cart.api.line,
		// which represents [item].
		o, _, _ := packOf(t, pLinesThen)
		o.exactly(t, "Amendment", "cart.pay.lines")
		o.exactly(t, "Context", "cart.api.line", "cart.api.line.count", "cart.item")
	})
	t.Run("row3 a rule marking a term defined by a reference", func(t *testing.T) {
		o, _, _ := packOf(t, pReceiptThen)
		o.exactly(t, "Amendment", "cart.pay.receipt")
		o.exactly(t, "Context", "cart.thematic-break", "system.ref.commonmark")
	})
	t.Run("row4 a rule marking a term other rules mark", func(t *testing.T) {
		// [spec root]: [price], which cart.pay.total and cart.add.spoken
		// mark as well.
		o, _, _ := packOf(t, pRefundThen)
		o.exactly(t, "Amendment", "billing.refund.full")
		o.exactly(t, "Context", "system.price")
	})
	t.Run("row5 a rule removed", func(t *testing.T) {
		// lint.run.root: cart.add.spoken, which marks a term and hands over.
		o, _, _ := packOf(t, removedFrom(packTree(), featurePath, "cart.add.spoken"))
		o.exactly(t, "Amendment", "cart.add.spoken")
		o.exactly(t, "Context")
	})
	t.Run("derived: elements changed through a dependency follow the chain", func(t *testing.T) {
		// cart.max-quantity is under Amendment; cart.add.limit brings
		// cart.item, cart.add.spoken brings system.price.
		o, _, _ := packOf(t, iParamValue)
		o.exactly(t, "Context", "cart.item", "system.price")
	})
	t.Run("derived: a changed member brings its holder and the term it represents", func(t *testing.T) {
		o, _, _ := packOf(t, iCountBody)
		o.exactly(t, "Amendment", "cart.api.line.count", "cart.pay.lines", "cart.pay.order")
		o.exactly(t, "Context", "cart.api.line", "cart.item")
	})
}

func TestBaselinePackOnce(t *testing.T) {
	t.Run("row1 the system, its Scope changed", func(t *testing.T) {
		o, _, _ := packOf(t, iSystemScope)
		o.headingsAre(t, "Amendment", "Frames")
		o.exactly(t, "Amendment", "system")
		o.exactly(t, "Frames", sysFrames...)
	})
	t.Run("row2 a feature, its Goal changed, holding a changed rule", func(t *testing.T) {
		o, _, _ := packOf(t, pPayGoal, pCapThen)
		o.exactly(t, "Amendment", "cart.pay", "cart.pay.cap")
		o.exactly(t, "Frames", cartFrames...)
	})
	t.Run("row3 a glossary entry, its Definition changed, marked by a changed rule", func(t *testing.T) {
		o, _, _ := packOf(t, iItemDef, pTotalThen)
		o.exactly(t, "Amendment", "cart.item", "cart.pay.total", "cart.add.limit", "cart.add.spoken", "cart.api.line", "cart.pay.order")
		o.exactly(t, "Context", "cart.api.line.count", "cart.max-quantity", "system.price")
	})
	// That no element appears twice is checked for every run (see packIn).
	t.Run("row4 a feature, unchanged, holding two changed rules", func(t *testing.T) {
		// lint.run: cart.pay, holding cart.pay.cap and cart.pay.total.
		o, _, _ := packOf(t, pCapThen, pTotalThen)
		o.exactly(t, "Amendment", "cart.pay.cap", "cart.pay.total")
		o.exactly(t, "Frames", payFrames...)
	})
	t.Run("row5 a term, unchanged, marked by two changed rules", func(t *testing.T) {
		// [spec root]: [price], marked by cart.pay.total and
		// billing.refund.full.
		o, _, _ := packOf(t, pTotalThen, pRefundThen)
		o.exactly(t, "Amendment", "billing.refund.full", "cart.pay.total")
		o.exactly(t, "Context", "cart.item", "system.price")
	})
}

func TestBaselinePackOrder(t *testing.T) {
	// That the elements under each heading are in byte order, and the parts
	// of each in diff's order, is checked for every run (see packIn); here
	// each row's outcome is checked as well.
	t.Run("row1 under Frames", func(t *testing.T) {
		o, _, _ := packOf(t, pCapThen)
		want := "cart cart.bounded cart.light cart.pay system.fast system.honest"
		if got := strings.Join(o.under("Frames"), " "); got != want {
			t.Errorf("want under Frames %s, got %s", want, got)
		}
	})
	t.Run("row2 under Amendment", func(t *testing.T) {
		o, _, _ := packOf(t, removedFrom(packTree(), payPath, "cart.pay.total"), pCapThen)
		if got := strings.Join(o.under("Amendment"), " "); got != "cart.pay.cap cart.pay.total" {
			t.Errorf("want under Amendment cart.pay.cap, then cart.pay.total, got %s", got)
		}
	})
	t.Run("derived: the parts in diff's order", func(t *testing.T) {
		o, _, _ := packOf(t, pCapThen, pCapProvided, pCapRow)
		key := "row " + rowKey("cart", "the first cart")
		want := "Changed `cart.pay.cap`: `" + key + "` changed, `step Provided` changed, `step Then` changed."
		if it := o.item(t, "cart.pay.cap"); it.status != want {
			t.Errorf("want the line %q, got %q", want, it.status)
		}
		var formers []string
		for _, f := range o.item(t, "cart.pay.cap").formers {
			formers = append(formers, f.part)
		}
		if got := strings.Join(formers, "|"); got != key+"|step Provided|step Then" {
			t.Errorf("want the former text of %s, step Provided and step Then, in that order, got %s", key, got)
		}
	})
}

func TestBaselinePackOutput(t *testing.T) {
	exit0 := func(t *testing.T, o packOutcome) {
		t.Helper()
		if o.json.exit != 0 || o.text.exit != 0 {
			t.Errorf("want exit status 0, got %d with --json, %d without", o.json.exit, o.text.exit)
		}
	}
	t.Run("row1 one changed rule, --json", func(t *testing.T) {
		o, _, _ := packOf(t, pCapThen)
		exit0(t, o)
		if len(o.cp.Diff) != 1 || o.cp.Diff[0].ID != "cart.pay.cap" {
			t.Errorf("want a context_pack whose diff holds cart.pay.cap, got %s", o.json.stdout)
		}
	})
	t.Run("row2 one changed rule, as text", func(t *testing.T) {
		o, _, _ := packOf(t, pCapThen)
		exit0(t, o)
		o.headingsAre(t, "System", "Amendment", "Frames", "Context")
	})
	t.Run("row3 an empty element diff, as text", func(t *testing.T) {
		r, _ := accepted(t, packTree())
		o := r.pack()
		exit0(t, o)
		o.headingsAre(t, "System")
	})
	t.Run("derived: an empty element diff, --json", func(t *testing.T) {
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
}

// TestBaselinePackFormer runs the cases of baseline.diff.former with `pack`.
func TestBaselinePackFormer(t *testing.T) {
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
		dir := packTree().write(t)
		skipIfGitAbove(t, dir)
		manifestOK(t, dir)
		amended(packTree(), pCapThen).writeTo(t, dir, false)
		packIn(t, dir, dir, filepath.Join(dir, "proviso")).stopped(t)
	})
	t.Run("row4 the manifest written but never committed", func(t *testing.T) {
		r := newRepo(t, packTree())
		r.commit("tree")
		r.manifest()
		r.set(amended(packTree(), pCapThen))
		r.pack().stopped(t)
	})
	t.Run("row5 the manifest written again since and not committed", func(t *testing.T) {
		r, _ := accepted(t, packTree())
		r.set(amended(packTree(), pCapThen))
		r.manifest()
		r.set(amended(packTree(), pCapThen, pCapProvided))
		r.pack().stopped(t)
	})
	t.Run("row6 the manifest committed together with a tree it does not record", func(t *testing.T) {
		r := newRepo(t, packTree())
		r.manifest()
		r.set(amended(packTree(), pCapProvided))
		r.commit("a1b2")
		r.set(amended(packTree(), pCapProvided, pCapThen))
		r.pack().stopped(t)
	})
	t.Run("derived: added elements need no former text, and no git", func(t *testing.T) {
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
