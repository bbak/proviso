package acceptance

// Feature baseline.verify: proviso/capabilities/baseline/features/verify.md,
// and the contracts system.cli.verify and system.cli.verification. Every case
// accepts the impact tree (see impact_trees_test.go), amends it, records a
// check with `proviso judge` and runs `proviso verify`. The identifiers of
// the spec's examples stand for elements of the impact tree, as in
// judge_test.go; lint.run.root, touched only after the check, stands for
// cart.pay.open.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// Amendments of the impact tree beyond those of impact_test.go.
var (
	iOpenThen = change{payPath, "Then      the shop greets every buyer", "Then      the shop greets every buyer by name"}
	iFullRow  = change{featurePath, "| 1 | the first cart | refused |", "| 1 | the first cart | turned away |"}
	// iLimitAsk leaves the Then step of cart.add.limit undecided.
	iLimitAsk = change{featurePath, "Then      the cart holds <quantity> more of <product>", "Then      ?"}
	// iAddContactAsk leaves the Contact of the feature cart.add undecided.
	iAddContactAsk = change{featurePath, "- Contact: Dana Reyes", "- Contact: ?"}
	// iBasketOpen adds the open glossary entry cart.basket, marked by no
	// rule.
	iBasketOpen = change{glossaryPath, "", "\n## <a id=\"cart.basket\"></a> basket\n\n- Definition: ?\n- Contact: Dana Reyes\n- Status: open\n- Question: Is a basket a cart?\n"}
	// iOpenBasket makes cart.pay.open mark the term basket.
	iOpenBasket = change{payPath, "Then      the shop greets every buyer", "Then      the shop greets every buyer with a [basket]"}
	// iFullAsk leaves the outcome of cart.add.full's Examples row undecided.
	iFullAsk = change{featurePath, "| 1 | the first cart | refused |", "| 1 | the first cart | ? |"}
	// iLimitQuestion asks a question in the prose of cart.add.limit.
	iLimitQuestion = change{featurePath, "A shelf holds ten pieces.", "Does a shelf hold ten pieces? It does."}
)

// covered is the impact tree amended by cs, every candidate judged.
func covered(t *testing.T, cs ...change) *repo {
	t.Helper()
	r := amendedRepo(t, cs...)
	r.judgeAll(nil)
	return r
}

// mergedMainline commits the amendment cs on the branch feature, judged in
// full, accepts mainline on main, and merges main into feature. The
// repository's tree is left as the branch's, since every tree here holds the
// same files.
func mergedMainline(t *testing.T, cs []change, mainline []change) *repo {
	t.Helper()
	r, _ := accepted(t, impactTree())
	r.git("checkout", "-q", "-b", "feature")
	feature := amended(impactTree(), cs...)
	r.set(feature)
	r.judgeAll(nil)
	r.commit("feature, with its check")
	r.git("checkout", "-q", "main")
	r.tr = impactTree()
	r.set(amended(impactTree(), mainline...))
	r.manifest()
	r.commit("accept on the mainline")
	r.git("checkout", "-q", "feature")
	r.tr = feature
	r.git("merge", "-q", "--no-ff", "-m", "merge the mainline", "main")
	if !r.hasCoherence() {
		t.Fatalf("want %s on the branch after the merge", coherenceName)
	}
	return r
}

func TestBaselineVerifyExit(t *testing.T) {
	t.Parallel()
	// The exit status is checked against the problems found for every run
	// (see verifyIn); here each row's outcome is checked as well.
	t.Run("row1 none", func(t *testing.T) {
		t.Parallel()
		o := covered(t, iLimitThen).verify()
		o.only(t)
		o.exits(t, 0)
	})
	t.Run(`row2 one "leftover"`, func(t *testing.T) {
		t.Parallel()
		r := covered(t, iLimitThen)
		r.set(impactTree())
		o := r.verify()
		o.only(t, "leftover")
		o.exits(t, 1)
	})
	t.Run(`row3 one "open"`, func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.judgeAll(nil, gap("A gap."))
		o := r.verify()
		o.only(t, "open f1")
		o.exits(t, 1)
	})
	// baseline.verify Goal: once accepted, the tree passes again.
	t.Run("derived: covered, then accepted", func(t *testing.T) {
		t.Parallel()
		r := covered(t, iLimitThen)
		r.manifest()
		r.commit("accept")
		o := r.verify()
		o.only(t)
		o.exits(t, 0)
	})
}

func TestBaselineVerifyLeftover(t *testing.T) {
	t.Parallel()
	// That "leftover" comes alone, so once, and with exit status 1, is
	// checked for every run (see verifyIn).
	t.Run("row1 changed, with .coherence.json", func(t *testing.T) {
		t.Parallel()
		o := covered(t, iLimitThen).verify()
		o.lacks(t, "leftover")
		o.lacks(t, "missing")
		o.only(t)
	})
	t.Run("row2 changed, with .amendment.md only", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.writeNote("- refunds\n")
		o := r.verify()
		o.lacks(t, "leftover")
		o.only(t, "missing")
	})
	t.Run("row3 none, with .coherence.json", func(t *testing.T) {
		t.Parallel()
		// The check holds an open finding and leaves candidates unjudged;
		// verify compares nothing that it records.
		r := amendedRepo(t, iLimitThen)
		r.judgeOK(checkResult{Findings: []findingIn{gap("A gap.")}})
		r.set(impactTree())
		r.verify().only(t, "leftover")
	})
	t.Run("row4 none, with .amendment.md only", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, impactTree())
		r.writeNote("- refunds\n")
		o := r.verify()
		o.only(t, "leftover")
		o.exits(t, 1)
	})
	t.Run("row5 none, with .coherence.json and .amendment.md", func(t *testing.T) {
		t.Parallel()
		r := covered(t, iLimitThen)
		r.writeNote("- refunds\n")
		r.set(impactTree())
		r.verify().only(t, "leftover")
	})
	t.Run("derived: changed, with .coherence.json and .amendment.md", func(t *testing.T) {
		t.Parallel()
		// The amendment as it stands before acceptance: the note changes
		// nothing of the comparison.
		r := covered(t, iLimitThen)
		r.writeNote("- refunds\n")
		r.verify().only(t)
	})
	t.Run("derived: none, with an empty .amendment.md", func(t *testing.T) {
		t.Parallel()
		// Of the note, verify only asks whether the spec root holds one.
		r, _ := accepted(t, impactTree())
		r.writeNote("")
		r.verify().only(t, "leftover")
	})
	t.Run("derived: none, with .coherence.json, an element left undecided", func(t *testing.T) {
		t.Parallel()
		// verify compares nothing the check records, but still takes every
		// element of the tree (baseline.verify.undecided).
		r, _ := accepted(t, amended(impactTree(), iAddContactAsk))
		r.set(amended(impactTree(), iAddContactAsk, iLimitThen))
		r.judgeOK(checkResult{Findings: []findingIn{gap("A gap.")}})
		r.set(amended(impactTree(), iAddContactAsk))
		o := r.verify()
		o.only(t, "leftover", "undecided cart.add")
		o.exits(t, 1)
	})
	t.Run("derived: a merge that skipped acceptance", func(t *testing.T) {
		t.Parallel()
		// The mainline accepts the amendment's own text; the branch's check
		// is left over.
		r := mergedMainline(t, []change{iLimitThen}, []change{iLimitThen})
		r.verify().only(t, "leftover")
	})
}

func TestBaselineVerifyMissing(t *testing.T) {
	t.Parallel()
	t.Run("row1 none: the tree is as the manifest records it", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, impactTree())
		o := r.verify()
		o.only(t)
		o.exits(t, 0)
	})
	t.Run("row2 changed", func(t *testing.T) {
		t.Parallel()
		o := amendedRepo(t, iLimitThen).verify()
		o.only(t, "missing")
		o.exits(t, 1)
	})
	t.Run("derived: none, with .amendment.md only", func(t *testing.T) {
		t.Parallel()
		// Nothing is reported as "missing"; the note is left over (see
		// baseline.verify.leftover).
		r, _ := accepted(t, impactTree())
		r.writeNote("- refunds\n")
		o := r.verify()
		o.lacks(t, "missing")
		o.only(t, "leftover")
	})
	t.Run("derived: changed, an element left undecided", func(t *testing.T) {
		t.Parallel()
		// verify reads no judgement, but still takes every element of the
		// tree (baseline.verify.undecided).
		o := amendedRepo(t, iLimitAsk).verify()
		o.only(t, "missing", "undecided cart.add.limit")
		o.exits(t, 1)
	})
}

func TestBaselineVerifyOpen(t *testing.T) {
	t.Parallel()
	// recordThree records f1, f2 and f3 and then resolves f1 and dismisses
	// f2.
	recordThree := func(t *testing.T) verifyOutcome {
		t.Helper()
		r := amendedRepo(t, iLimitThen)
		r.judgeAll(nil, gap("The first gap."), gap("The second gap."), gap("The third gap.", "cart.pay.open"))
		r.judgeOK(checkResult{Findings: []findingIn{resolved("f1"), dismissed("f2", "decided in system.fast")}})
		return r.verify()
	}
	t.Run("row1 f1, resolved", func(t *testing.T) {
		t.Parallel()
		recordThree(t).lacks(t, "open f1")
	})
	t.Run("row2 f2, dismissed with a reason", func(t *testing.T) {
		t.Parallel()
		recordThree(t).lacks(t, "open f2")
	})
	t.Run("row3 f3, as judge recorded it", func(t *testing.T) {
		t.Parallel()
		o := recordThree(t)
		o.only(t, "open f3")
		if p := o.has(t, "open f3"); p.Text != "The third gap." {
			t.Errorf(`want "open" for f3 with its text %q, got %q`, "The third gap.", p.Text)
		}
	})
}

func TestBaselineVerifyOrder(t *testing.T) {
	t.Parallel()
	// That the problems are ordered is checked for every run (see
	// verifyIn); here each row's outcome is checked as well.
	inOrder := func(t *testing.T, o verifyOutcome, want ...string) {
		t.Helper()
		o.ran(t)
		if got := o.keys(); strings.Join(got, "|") != strings.Join(want, "|") {
			t.Errorf("want the problems in the order %q, got %q", want, got)
		}
	}
	t.Run(`row1 "stale" for cart.add.limit, "unjudged" for cart.add.full`, func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.judgeAll([]string{"cart.add.full"})
		r.set(amended(impactTree(), iLimitThen, iLimitProvided))
		inOrder(t, r.verify(), "unjudged cart.add.full", "stale cart.add.limit")
	})
	t.Run(`row2 "open" for f10 and for f2`, func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		var fs []findingIn
		for i := 1; i <= 10; i++ {
			fs = append(fs, gap(fmt.Sprintf("Gap %d.", i)))
		}
		r.judgeAll(nil, fs...)
		var close []findingIn
		for i := 1; i <= 9; i++ {
			if i != 2 {
				close = append(close, resolved(fmt.Sprintf("f%d", i)))
			}
		}
		r.judgeOK(checkResult{Findings: close})
		inOrder(t, r.verify(), "open f2", "open f10")
	})
	t.Run(`row3 "unjudged" for cart.add.full, "undecided" for cart.add.limit`, func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitAsk)
		r.judgeAll([]string{"cart.add.full"})
		inOrder(t, r.verify(), "undecided cart.add.limit", "unjudged cart.add.full")
	})
	t.Run(`row4 "undecided" for cart.add.limit, "missing"`, func(t *testing.T) {
		t.Parallel()
		inOrder(t, amendedRepo(t, iLimitAsk).verify(), "missing", "undecided cart.add.limit")
	})
	t.Run("derived: undecided by identifier, byte by byte", func(t *testing.T) {
		t.Parallel()
		// cart.add < cart.basket < cart.pay.open, byte by byte.
		r, _ := accepted(t, amended(impactTree(), iAddContactAsk, iBasketOpen, iOpenBasket))
		inOrder(t, r.verify(), "undecided cart.add", "undecided cart.basket", "undecided cart.pay.open")
	})
	t.Run("derived: unjudged by identifier, byte by byte", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.judgeOK(checkResult{})
		inOrder(t, r.verify(), unjudgedKeys(sorted(limitCandidates)...)...)
	})
}

func TestBaselineVerifyOutput(t *testing.T) {
	t.Parallel()
	// That the text form is written for problems, and nothing for none, is
	// checked for every run as text (see runVerify).
	oneUnjudged := func(t *testing.T, text bool) verifyOutcome {
		t.Helper()
		// iNeverProp makes cart.add.full and cart.add.limit the candidates.
		r := amendedRepo(t, iNeverProp)
		r.judgeOK(judging("untouched", "cart.add.full"))
		if text {
			return r.verifyText()
		}
		return r.verify()
	}
	t.Run(`row1 one "unjudged", --json`, func(t *testing.T) {
		t.Parallel()
		o := oneUnjudged(t, false)
		o.only(t, "unjudged cart.add.limit")
	})
	t.Run("row2 none, --json", func(t *testing.T) {
		t.Parallel()
		o := covered(t, iLimitThen).verify()
		o.ran(t)
		var raw map[string]json.RawMessage
		if err := json.Unmarshal([]byte(o.json.stdout), &raw); err != nil {
			t.Fatalf("want a verification, got %q", o.json.stdout)
		}
		if p := bytes.TrimSpace(raw["problems"]); string(p) != "[]" {
			t.Errorf("want an empty list of problems, got %s", p)
		}
	})
	t.Run(`derived: one "undecided", --json`, func(t *testing.T) {
		t.Parallel()
		// system.cli.verification now lists the kind undecided, with the
		// element it is about.
		r, _ := accepted(t, amended(impactTree(), iAddContactAsk))
		o := r.verify()
		o.only(t, "undecided cart.add")
	})
	t.Run(`row3 one "unjudged", as text`, func(t *testing.T) {
		t.Parallel()
		o := oneUnjudged(t, true)
		if strings.TrimSpace(o.text.stdout) == "" {
			t.Errorf("want the problem as text, got nothing")
		}
	})
	t.Run("row4 none, as text", func(t *testing.T) {
		t.Parallel()
		o := covered(t, iLimitThen).verifyText()
		if o.text.stdout != "" {
			t.Errorf("want nothing written, got %q", o.text.stdout)
		}
	})
}

func TestBaselineVerifyStale(t *testing.T) {
	t.Parallel()
	t.Run("row1 touched, unchanged since the check", func(t *testing.T) {
		t.Parallel()
		covered(t, iLimitThen).verify().lacks(t, "stale cart.add.limit")
	})
	t.Run("row2 touched, its Then step edited after the check", func(t *testing.T) {
		t.Parallel()
		r := covered(t, iLimitProvided)
		r.set(amended(impactTree(), iLimitProvided, iLimitThen))
		r.verify().only(t, "stale cart.add.limit")
	})
	t.Run("row3 touched, its baseline changed by a merge of the mainline", func(t *testing.T) {
		t.Parallel()
		// The branch changes the Then step and the Rationale; the mainline
		// accepts the Rationale alone, so the branch's own text stays the
		// same. The two edits stand lines apart, so that git merges them
		// without a conflict.
		r := mergedMainline(t, []change{iLimitThen, iLimitRatio}, []change{iLimitRatio})
		r.verify().only(t, "stale cart.add.limit")
	})
	t.Run("row4 touched only after the check", func(t *testing.T) {
		t.Parallel()
		r := covered(t, iLimitThen)
		r.set(amended(impactTree(), iLimitThen, iOpenThen))
		r.verify().only(t, "stale cart.pay.open")
	})
	t.Run("row5 judged, its Examples row edited by a merge of the mainline", func(t *testing.T) {
		t.Parallel()
		r := mergedMainline(t, []change{iLimitThen}, []change{iFullRow})
		r.verify().only(t, "stale cart.add.full")
	})
}

func TestBaselineVerifyUnjudged(t *testing.T) {
	t.Parallel()
	t.Run("row1 judged", func(t *testing.T) {
		t.Parallel()
		covered(t, iLimitThen).verify().lacks(t, "unjudged cart.add.full")
	})
	t.Run("row2 found through an event added after the check", func(t *testing.T) {
		t.Parallel()
		// cart.add.full is no candidate of the Goal of billing; the edit
		// of cart.add.limit after the check makes it one.
		r := covered(t, iBillingGoal)
		r.set(amended(impactTree(), iBillingGoal, iLimitThen))
		o := r.verify()
		o.has(t, "unjudged cart.add.full")
		o.lacks(t, "unjudged billing.refund.add")
	})
}

func TestBaselineVerifyUndecided(t *testing.T) {
	t.Parallel()
	// That "undecided" comes with exit status 1, and that the run changes
	// nothing, is checked for every run (see verifyIn).
	t.Run("row1 a rule with every step decided", func(t *testing.T) {
		t.Parallel()
		o := covered(t, iLimitThen).verify()
		o.lacks(t, "undecided cart.add.limit")
		o.only(t)
	})
	t.Run("row2 a rule with Then ?", func(t *testing.T) {
		t.Parallel()
		o := covered(t, iLimitAsk).verify()
		o.only(t, "undecided cart.add.limit")
		o.exits(t, 1)
	})
	t.Run("row3 a rule marking a term whose Status is open", func(t *testing.T) {
		t.Parallel()
		// The open entry is undecided itself (row5).
		r, _ := accepted(t, amended(impactTree(), iBasketOpen, iOpenBasket))
		r.verify().only(t, "undecided cart.pay.open", "undecided cart.basket")
	})
	t.Run("row4 a feature with Contact: ?, unchanged since the manifest", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, amended(impactTree(), iAddContactAsk))
		o := r.verify()
		o.only(t, "undecided cart.add")
		o.exits(t, 1)
	})
	t.Run("row5 an open glossary entry, marked by no rule", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, amended(impactTree(), iBasketOpen))
		r.verify().only(t, "undecided cart.basket")
	})
	t.Run("derived: an undecided place outside what the amendment touched", func(t *testing.T) {
		t.Parallel()
		// The whole tree is verified: cart.add is no candidate of the
		// amendment, and the check covers it.
		r, _ := accepted(t, amended(impactTree(), iAddContactAsk))
		r.set(amended(impactTree(), iAddContactAsk, iBillingGoal))
		r.judgeAll(nil)
		r.verify().only(t, "undecided cart.add")
	})
	t.Run("derived: an undecided Examples cell", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, amended(impactTree(), iFullAsk))
		r.verify().only(t, "undecided cart.add.full")
	})
	t.Run("derived: a question in prose is not undecided", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, amended(impactTree(), iLimitQuestion))
		r.verify().only(t)
	})
	t.Run("derived: decided again before acceptance", func(t *testing.T) {
		t.Parallel()
		// The amendment decides what the accepted tree left open.
		r, _ := accepted(t, amended(impactTree(), iAddContactAsk))
		r.set(impactTree())
		r.judgeAll(nil)
		r.verify().only(t)
	})
}
