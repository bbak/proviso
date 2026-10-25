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
		// verify looks no further.
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
	// checked for every run (see verifyIn).
	oneUnjudged := func(t *testing.T) verifyOutcome {
		t.Helper()
		// iNeverProp makes cart.add.full and cart.add.limit the candidates.
		r := amendedRepo(t, iNeverProp)
		r.judgeOK(judging("untouched", "cart.add.full"))
		return r.verify()
	}
	t.Run(`row1 one "unjudged", --json`, func(t *testing.T) {
		t.Parallel()
		o := oneUnjudged(t)
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
	t.Run(`row3 one "unjudged", as text`, func(t *testing.T) {
		t.Parallel()
		o := oneUnjudged(t)
		if strings.TrimSpace(o.text.stdout) == "" {
			t.Errorf("want the problem as text, got nothing")
		}
	})
	t.Run("row4 none, as text", func(t *testing.T) {
		t.Parallel()
		o := covered(t, iLimitThen).verify()
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
