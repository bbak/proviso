package acceptance

// Feature baseline.impact: proviso/capabilities/baseline/features/impact.md,
// and the contracts system.cli.impact and system.cli.impact-set
// (proviso/contracts/cli.md). Every case accepts the impact tree (see
// impact_trees_test.go), amends it and runs `proviso impact`.

import (
	"path/filepath"
	"strings"
	"testing"
)

// Amendments of the impact tree.
var (
	iLimitThen     = change{featurePath, "Then      the cart holds <quantity> more of <product>", "Then      the cart holds <quantity> more pieces of <product>"}
	iLimitProvided = change{featurePath, "Provided  <quantity> is at most ´max_quantity´", "Provided  <quantity> is no more than ´max_quantity´"}
	iLimitUnmark   = change{featurePath, "When      a buyer adds the quantity <quantity> of the [item] <product>", "When      a buyer adds the quantity <quantity> of the product <product>"}
	iLimitRatio    = change{featurePath, "A shelf holds ten pieces.", "A shelf holds no more than ten pieces."}
	iFullThen      = change{featurePath, "Then      <cart> refuses <quantity>", "Then      <cart> turns <quantity> away"}
	iItemDef       = change{glossaryPath, "- Definition: A product placed in the cart.", "- Definition: A product a buyer placed in the cart."}
	iAddContact    = change{featurePath, "- Contact: Dana Reyes", "- Contact: Bruno Baketarić"}
	iAddGoal       = change{featurePath, "Buyers put goods into the cart.", "Buyers put goods into the cart, one by one."}
	iAddRatio      = change{featurePath, "A cart that takes anything would empty the shelves.", "A cart that takes everything would empty the shelves."}
	iCountBody     = change{apiPath, "How many pieces the line holds.", "How many pieces the line holds, at most ten."}
	iNewMember     = change{apiPath, "How many pieces the line holds.\n", "How many pieces the line holds.\n\n### <a id=\"cart.api.line.cost\"></a> cost\n\nWhat the line costs.\n"}
	iParamRatio    = change{paramsPath, "- Rationale: a shelf holds ten pieces.", "- Rationale: a shelf holds no more than ten pieces."}
	iParamValue    = change{paramsPath, "- Value: 10", "- Value: 12"}
	iFastProperty  = change{sysConstraintsPath, "Every cart answers within a second.", "Every cart answers within half a second."}
	iBoundedProp   = change{cartInvariantsPath, "No cart holds more than ten pieces of a product.", "No cart holds more than twelve pieces of a product."}
	iNeverProp     = change{featurePath, "A cart holds every piece a buyer put in until the buyer takes it out.", "A cart holds every piece a buyer put in."}
	iHonestRanges  = change{sysInvariantsPath, "Every payment.", "Every payment and every refund."}
	iSystemScope   = change{systemPath, "- Selling goods.", "- Selling goods.\n- Refunding goods."}
	iCartGoal      = change{capPath, "Buyers collect goods before they pay.", "Buyers collect goods before they pay, in one cart."}
	iBillingGoal   = change{billingCapPath, "Buyers pay.", "Buyers pay, and get refunds."}
)

// addRule appends a rule to the file path.
func addRule(path, id, title, block, table string) change {
	return change{path: path, old: "", new: "\n" + ruleMD(id, title, block, table)}
}

// amended applies changes to tr; a change with no old text appends its new
// text to its file.
func amended(tr tree, cs ...change) tree {
	for _, c := range cs {
		if c.old == "" {
			tr = tr.appendTo(c.path, c.new)
		} else {
			tr = tr.edit(c.path, c.old, c.new)
		}
	}
	return tr
}

// impactOf accepts the impact tree, amends it with cs and runs impact.
func impactOf(t *testing.T, cs ...change) (impactOutcome, tree) {
	t.Helper()
	r, _ := accepted(t, impactTree())
	tr := amended(impactTree(), cs...)
	r.set(tr)
	o := r.impact()
	o.ran(t)
	return o, tr
}

// Rules added by the cases.
var (
	addTwice = addRule(featurePath, "cart.add.twice", "A buyer adds twice",
		"When      A buyer adds the quantity, <quantity> twice\nThen      the cart holds twice <quantity>\n  And     nothing else changes",
		"| quantity | outcome |\n| --- | --- |\n| 1 | two held |")
	addAgain = addRule(featurePath, "cart.add.again", "A buyer adds an item again",
		"When      a buyer adds the quantity <quantity> of the [item] <product> again\nThen      the cart holds <quantity> more of <product>\n  And     nothing else changes",
		"| quantity | product | outcome |\n| --- | --- | --- |\n| 1 | apple | added |")
	addTip = addRule(payPath, "cart.pay.tip", "A buyer may tip",
		"When      a buyer tips the shop <amount>\nThen      the shop thanks the buyer for <amount>\n  And     nothing else changes",
		"| amount | outcome |\n| --- | --- |\n| one coin | thanked |")
)

func without(ids []string, drop ...string) []string {
	var out []string
	for _, id := range ids {
		keep := true
		for _, d := range drop {
			keep = keep && id != d
		}
		if keep {
			out = append(out, id)
		}
	}
	return out
}

func TestBaselineImpactListed(t *testing.T) {
	t.Run("row1 unchanged, marking a term a changed rule marks", func(t *testing.T) {
		o, _ := impactOf(t, iLimitThen)
		o.hasReason(t, "cart.pay.total", "terms", "cart.item", "cart.add.limit")
	})
	t.Run("row2 changed itself", func(t *testing.T) {
		// cart.add.full shares the event head of cart.add.limit.
		o, _ := impactOf(t, iLimitThen, iFullThen)
		o.notCand(t, "cart.add.full")
		o.notCand(t, "cart.add.limit")
		o.hasReason(t, "billing.refund.add", "events", "buyer adds quantity", "cart.add.full", "cart.add.limit")
	})
	t.Run("row3 changed by a term", func(t *testing.T) {
		o, _ := impactOf(t, iItemDef)
		o.notCand(t, "cart.add.limit")
		o.notCand(t, "cart.pay.total")
	})
	t.Run("derived: without a manifest, every element is touched and none a candidate", func(t *testing.T) {
		dir := impactTree().write(t)
		o := impactIn(t, dir, dir, filepath.Join(dir, "proviso"))
		o.only(t)
	})
}

func TestBaselineImpactTerms(t *testing.T) {
	t.Run("row1 its Then step changed, marking the term", func(t *testing.T) {
		o, _ := impactOf(t, iLimitThen)
		o.hasReason(t, "cart.pay.total", "terms", "cart.item", "cart.add.limit")
	})
	t.Run("row2 its mark of the term removed", func(t *testing.T) {
		o, _ := impactOf(t, iLimitUnmark)
		o.hasReason(t, "cart.pay.total", "terms", "cart.item", "cart.add.limit")
	})
	t.Run("row3 its Definition changed", func(t *testing.T) {
		// The rules marking item are changed by a term and listed, so none
		// is a candidate through item.
		o, _ := impactOf(t, iItemDef)
		for _, c := range o.is.Candidates {
			for _, r := range c.Reasons {
				if r.Source == "terms" && r.Via == "cart.item" {
					t.Errorf("want no candidate through cart.item, got:\n%s", o.dump())
				}
			}
		}
	})
	t.Run("row4 its Contact changed", func(t *testing.T) {
		o, _ := impactOf(t, iAddContact)
		o.only(t)
	})
	// impact_triggers: every part of a rule changed by a term counts as
	// touched, so its other terms open as well.
	t.Run("derived: a rule changed by a term reaches its other terms", func(t *testing.T) {
		o, _ := impactOf(t, iItemDef)
		o.hasReason(t, "billing.refund.full", "terms", "system.price", "cart.pay.total")
	})
	t.Run("derived: a removed rule reaches the rules marking its terms", func(t *testing.T) {
		total, _ := elementText(impactTree()[payPath], "cart.pay.total")
		o, _ := impactOf(t, change{payPath, total, ""})
		o.hasReason(t, "cart.add.limit", "terms", "cart.item", "cart.pay.total")
		o.hasReason(t, "billing.refund.full", "terms", "system.price", "cart.pay.total")
	})
}

func TestBaselineImpactReferences(t *testing.T) {
	t.Run("row1 a member of a contract element with a new value", func(t *testing.T) {
		o, _ := impactOf(t, iCountBody)
		o.hasReason(t, "cart.pay.lines", "references", "line.count", "cart.api.line.count")
	})
	t.Run("row2 a rule's step changed, referring to a parameter", func(t *testing.T) {
		o, _ := impactOf(t, iLimitProvided)
		o.hasReason(t, "cart.pay.cap", "references", "max_quantity", "cart.add.limit")
	})
	t.Run("row3 a new member of a contract element that represents a term", func(t *testing.T) {
		o, _ := impactOf(t, iNewMember)
		o.hasReason(t, "cart.add.limit", "references", "", "cart.api.line.cost")
		o.hasReason(t, "cart.pay.total", "references", "", "cart.api.line.cost")
	})
	t.Run("row4 a parameter's Rationale reworded", func(t *testing.T) {
		o, _ := impactOf(t, iParamRatio)
		o.only(t)
	})
	t.Run("derived: a parameter's Value changed", func(t *testing.T) {
		o, _ := impactOf(t, iParamValue)
		o.only(t, "cart.add.limit", "cart.pay.cap")
		o.hasReason(t, "cart.add.limit", "references", "max_quantity", "cart.max-quantity")
		o.hasReason(t, "cart.pay.cap", "references", "max_quantity", "cart.max-quantity")
	})
	// The link between a contract element and the term it represents is
	// followed both ways.
	t.Run("derived: a changed definition reaches the rules referring to its contract element", func(t *testing.T) {
		o, _ := impactOf(t, iItemDef)
		o.hasReason(t, "cart.pay.order", "references", "line", "cart.item")
	})
}

func TestBaselineImpactEvents(t *testing.T) {
	t.Run("row1 an added rule with an event head others have", func(t *testing.T) {
		o, _ := impactOf(t, addTwice)
		o.only(t, "billing.refund.add", "cart.add.full", "cart.add.limit", "cart.add.never", "cart.bounded", "system.honest")
		for _, id := range []string{"billing.refund.add", "cart.add.full", "cart.add.limit"} {
			o.hasReason(t, id, "events", "buyer adds quantity", "cart.add.twice")
		}
	})
	t.Run("row2 its Rationale reworded", func(t *testing.T) {
		o, _ := impactOf(t, iLimitRatio)
		o.only(t)
	})
}

func TestBaselineImpactEventHead(t *testing.T) {
	t.Run("rows1-2 the When up to its first placeholder", func(t *testing.T) {
		// cart.add.full and cart.add.limit qualify the event after the
		// placeholder each in its own words.
		o, _ := impactOf(t, addRule(payPath, "cart.pay.more", "A buyer adds more",
			"When      a buyer adds the quantity <quantity> once more\nThen      the cart holds <quantity> more\n  And     nothing else changes",
			"| quantity | outcome |\n| --- | --- |\n| 1 | held |"))
		o.hasReason(t, "cart.add.full", "events", "buyer adds quantity", "cart.pay.more")
		o.hasReason(t, "cart.add.limit", "events", "buyer adds quantity", "cart.pay.more")
	})
	t.Run("row3 case, articles, punctuation, term marks and backticks ignored", func(t *testing.T) {
		// cart.pay.order: "a buyer orders a `line` for the products <products>".
		o, _ := impactOf(t, addRule(payPath, "cart.pay.again", "A buyer orders again",
			"When      The Buyer orders the line, for the products: <products>\nThen      the order holds <products>\n  And     nothing else changes",
			"| products | outcome |\n| --- | --- |\n| apple | held |"))
		o.hasReason(t, "cart.pay.order", "events", "buyer orders line for products", "cart.pay.again")
	})
	t.Run("derived: a term mark ignored, its words kept", func(t *testing.T) {
		// cart.pay.total: "a buyer pays for the [items] <items>".
		o, _ := impactOf(t, addRule(payPath, "cart.pay.again", "A buyer pays again",
			"When      the buyer pays for [items] <items> again\nThen      the shop charges <items> again\n  And     nothing else changes",
			"| items | outcome |\n| --- | --- |\n| apple | charged |"))
		o.hasReason(t, "cart.pay.total", "events", "buyer pays for items", "cart.pay.again")
	})
	t.Run("row4 a When without a placeholder", func(t *testing.T) {
		// cart.pay.open: "the shop opens".
		o, _ := impactOf(t, addRule(payPath, "cart.pay.dawn", "The shop opens at dawn",
			"When      The shop opens.\nThen      the shop lights its lamps\n  And     nothing else changes",
			"| outcome |\n| --- |\n| lit |"))
		o.hasReason(t, "cart.pay.open", "events", "shop opens", "cart.pay.dawn")
	})
	t.Run("derived: other words are another event", func(t *testing.T) {
		o, _ := impactOf(t, addTwice)
		for _, id := range []string{"cart.pay.lines", "cart.pay.cap", "cart.pay.open", "cart.pay.total", "billing.refund.full"} {
			o.notCand(t, id)
		}
	})
}

func TestBaselineImpactRanges(t *testing.T) {
	t.Run("row1 a rule's Then step changed", func(t *testing.T) {
		o, _ := impactOf(t, iLimitThen)
		o.hasReason(t, "cart.add.never", "ranges", "feature", "cart.add.limit")
		o.hasReason(t, "cart.bounded", "ranges", "capability", "cart.add.limit")
		o.hasReason(t, "system.honest", "ranges", "system", "cart.add.limit")
		o.notCand(t, "billing.balanced")
	})
	t.Run("row2 an added rule", func(t *testing.T) {
		o, _ := impactOf(t, addTip)
		o.only(t, "cart.bounded", "system.honest")
		o.hasReason(t, "cart.bounded", "ranges", "capability", "cart.pay.tip")
		o.hasReason(t, "system.honest", "ranges", "system", "cart.pay.tip")
	})
	t.Run("row3 a rule's Rationale reworded", func(t *testing.T) {
		o, _ := impactOf(t, iLimitRatio)
		o.only(t)
	})
	// Every source of a changed step at once, and nothing else.
	t.Run("derived: every source of a changed step", func(t *testing.T) {
		o, tr := impactOf(t, iLimitThen)
		o.only(t, "billing.refund.add", "cart.add.full", "cart.add.never", "cart.bounded", "cart.pay.cap", "cart.pay.total", "system.honest")
		o.hasReason(t, "cart.pay.cap", "references", "max_quantity", "cart.add.limit")
		o.hasReason(t, "cart.add.full", "events", "buyer adds quantity", "cart.add.limit")
		o.hasReason(t, "billing.refund.add", "events", "buyer adds quantity", "cart.add.limit")
		o.at(t, tr)
		if len(o.is.Scopes) != 0 {
			t.Errorf("want no scopes, as nothing opened scope, got %v", o.is.Scopes)
		}
	})
}

func TestBaselineImpactScope(t *testing.T) {
	cases := []struct {
		name  string
		edit  change
		frame string
		want  []string
	}{
		{"row1 a system constraint's Property changed", iFastProperty, "system.fast", everyRuleInv},
		{"row2 a capability's invariant's Property changed", iBoundedProp, "cart.bounded", without(cartAll, "cart.bounded")},
		{"row3 a feature's Goal changed", iAddGoal, "cart.add", addRules},
		{"row4 the system's Scope changed", iSystemScope, "system", everyRuleInv},
		{"derived: a feature's invariant's Property changed", iNeverProp, "cart.add.never", without(addRules, "cart.add.never")},
		{"derived: a system invariant's Ranges over changed", iHonestRanges, "system.honest", without(everyRuleInv, "system.honest")},
		{"derived: a capability's Goal changed", iCartGoal, "cart", cartAll},
		{"derived: another capability's Goal changed", iBillingGoal, "billing", billingAll},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			o, tr := impactOf(t, c.edit)
			o.only(t, c.want...)
			for _, id := range c.want {
				o.hasReason(t, id, "scope", c.frame, c.frame)
			}
			o.at(t, tr)
			found := false
			for _, g := range o.is.Scopes {
				if g.ID == c.frame {
					found = true
					if strings.Join(g.Candidates, " ") != strings.Join(sorted(c.want), " ") {
						t.Errorf("want the scope %s to list %v, got %v", c.frame, sorted(c.want), g.Candidates)
					}
				}
			}
			if !found {
				t.Errorf("want the scope %s listed, got %v", c.frame, o.is.Scopes)
			}
		})
	}
	t.Run("row5 a feature's Rationale reworded", func(t *testing.T) {
		o, _ := impactOf(t, iAddRatio)
		o.only(t)
		if len(o.is.Scopes) != 0 {
			t.Errorf("want no scopes, got %v", o.is.Scopes)
		}
	})
}

func sorted(s []string) []string {
	c := append([]string(nil), s...)
	for i := 1; i < len(c); i++ {
		for j := i; j > 0 && c[j-1] > c[j]; j-- {
			c[j-1], c[j] = c[j], c[j-1]
		}
	}
	return c
}

func TestBaselineImpactOrder(t *testing.T) {
	// That candidates and reasons are ordered is checked for every run
	// (see parseImpactSet); here each row's outcome is checked as well.
	t.Run("row1 by identifier, byte by byte", func(t *testing.T) {
		o, _ := impactOf(t, iFastProperty)
		var got []string
		for _, c := range o.is.Candidates {
			got = append(got, c.ID)
		}
		if strings.Join(got, " ") != strings.Join(sorted(everyRuleInv), " ") {
			t.Errorf("want %v, got %v", sorted(everyRuleInv), got)
		}
	})
	t.Run("row2 found through events and through terms", func(t *testing.T) {
		o, _ := impactOf(t, addAgain)
		var sources []string
		for _, r := range o.cand(t, "cart.add.limit").Reasons {
			sources = append(sources, r.Source)
		}
		if strings.Join(sources, " ") != "terms events" {
			t.Errorf("want the terms reason first, then the events reason, got %v", sources)
		}
	})
}

func TestBaselineImpactOutput(t *testing.T) {
	// That the text form is written for candidates, and nothing for none,
	// is checked for every run (see impactIn).
	exit0 := func(t *testing.T, o impactOutcome) {
		t.Helper()
		if o.json.exit != 0 || o.text.exit != 0 {
			t.Errorf("want exit status 0, got %d with --json, %d without", o.json.exit, o.text.exit)
		}
	}
	t.Run("row1 two, --json", func(t *testing.T) {
		o, _ := impactOf(t, iNeverProp)
		o.only(t, "cart.add.full", "cart.add.limit")
		exit0(t, o)
	})
	t.Run("row2 none, --json", func(t *testing.T) {
		r, _ := accepted(t, impactTree())
		o := r.impact()
		o.only(t)
		if len(o.is.Scopes) != 0 {
			t.Errorf("want no scopes, got %v", o.is.Scopes)
		}
		exit0(t, o)
	})
	t.Run("row3 two, as text", func(t *testing.T) {
		o, _ := impactOf(t, iNeverProp)
		if strings.TrimSpace(o.text.stdout) == "" {
			t.Errorf("want the two candidates as text, got nothing")
		}
		exit0(t, o)
	})
	t.Run("row4 none, as text", func(t *testing.T) {
		r, _ := accepted(t, impactTree())
		o := r.impact()
		if o.text.stdout != "" {
			t.Errorf("want nothing written, got %q", o.text.stdout)
		}
		exit0(t, o)
	})
}

// TestBaselineImpactFormer runs the cases of baseline.diff.former that need
// former text with `impact`.
func TestBaselineImpactFormer(t *testing.T) {
	t.Run("row1 the manifest committed, the tree edited since", func(t *testing.T) {
		r, base := accepted(t, impactTree())
		r.set(amended(impactTree(), iLimitThen))
		o := r.impact()
		o.ran(t)
		if o.is.Revision == nil || *o.is.Revision != base {
			t.Errorf("want the revision %s, got %s", base, deref(o.is.Revision))
		}
	})
	t.Run("row2 a merge that took the manifest from the mainline", func(t *testing.T) {
		r, _ := accepted(t, impactTree())
		r.git("checkout", "-q", "-b", "feature")
		feature := amended(impactTree(), iCountBody)
		r.set(feature)
		r.commit("feature")
		r.git("checkout", "-q", "main")
		r.tr = impactTree()
		r.set(amended(impactTree(), iLimitThen))
		r.manifest()
		e5f6 := r.commit("accept on the mainline")
		r.git("checkout", "-q", "feature")
		r.tr = feature
		r.git("merge", "-q", "--no-ff", "-m", "merge the mainline", "main")
		r.tr = amended(impactTree(), iCountBody, iLimitThen)
		r.set(amended(impactTree(), iCountBody, iLimitThen, iLimitProvided))
		o := r.impact()
		o.ran(t)
		if o.is.Revision == nil || *o.is.Revision != e5f6 {
			t.Errorf("want the revision %s, got %s", e5f6, deref(o.is.Revision))
		}
	})
	t.Run("row3 the spec root in no git repository", func(t *testing.T) {
		dir := impactTree().write(t)
		skipIfGitAbove(t, dir)
		manifestOK(t, dir)
		amended(impactTree(), iLimitThen).writeTo(t, dir, false)
		impactStopped(t, impactIn(t, dir, dir, filepath.Join(dir, "proviso")))
	})
	t.Run("row4 the manifest written but never committed", func(t *testing.T) {
		r := newRepo(t, impactTree())
		r.commit("tree")
		r.manifest()
		r.set(amended(impactTree(), iLimitThen))
		impactStopped(t, r.impact())
	})
	t.Run("row5 the manifest written again since and not committed", func(t *testing.T) {
		r, _ := accepted(t, impactTree())
		r.set(amended(impactTree(), iLimitThen))
		r.manifest()
		r.set(amended(impactTree(), iLimitThen, iLimitProvided))
		impactStopped(t, r.impact())
	})
	t.Run("row6 the manifest committed together with a tree it does not record", func(t *testing.T) {
		r := newRepo(t, impactTree())
		r.manifest()
		r.set(amended(impactTree(), iLimitProvided))
		r.commit("a1b2")
		r.set(amended(impactTree(), iLimitProvided, iLimitThen))
		impactStopped(t, r.impact())
	})
	t.Run("derived: added rules need no former text, and no git", func(t *testing.T) {
		dir := impactTree().write(t)
		manifestOK(t, dir)
		skipIfGitAbove(t, dir)
		amended(impactTree(), addTip).writeTo(t, dir, false)
		o := impactIn(t, dir, dir, filepath.Join(dir, "proviso"))
		o.only(t, "cart.bounded", "system.honest")
		if o.is.Revision != nil {
			t.Errorf("want revision null, as no former text was needed, got %q", *o.is.Revision)
		}
	})
	// impact derives the element diff as diff does (baseline.impact).
	t.Run("derived: a malformed manifest stops impact as it stops diff", func(t *testing.T) {
		good := manifestFor(plainParts())
		dir := plainTree().with(manifestPath, good+"<<<<<<< HEAD\n").write(t)
		impactStopped(t, impactIn(t, dir, dir, filepath.Join(dir, "proviso")))
	})
}

// impactStopped asserts exit status 2 for both runs; that nothing went to
// stdout is checked for every run with exit status 2.
func impactStopped(t *testing.T, o impactOutcome) {
	t.Helper()
	for _, r := range []run{o.json, o.text} {
		if r.exit != 2 {
			t.Errorf("want exit status 2, got %d\nstdout: %s\nstderr: %s", r.exit, r.stdout, r.stderr)
		}
	}
}
