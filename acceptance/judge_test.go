package acceptance

// Feature baseline.judge: proviso/capabilities/baseline/features/judge.md,
// and the contracts system.cli.judge, system.cli.check-result and
// system.coherence-file. Every case accepts the impact tree (see
// impact_trees_test.go), amends it and runs `proviso judge`; what judge
// recorded is observed through `proviso verify`, since what coherence_file
// holds is not part of the contract.
//
// The identifiers of the spec's examples stand for elements of the impact
// tree: lint.run.exit, changed, for cart.add.limit, amended by iLimitThen;
// lint.examples.cells, found through "events", for cart.add.full;
// lint.run.root, another candidate, for billing.refund.add; and
// lint.run.nosuch for cart.add.nosuch.

import (
	"os"
	"path/filepath"
	"testing"
)

// The candidates of the impact tree amended by iLimitThen.
var limitCandidates = []string{"billing.refund.add", "cart.add.full", "cart.add.never", "cart.bounded", "system.honest"}

// removeOpen removes the rule cart.pay.open from the impact tree.
func removeOpen() change {
	text, _ := elementText(impactTree()[payPath], "cart.pay.open")
	return change{payPath, text, ""}
}

func TestBaselineJudgeInput(t *testing.T) {
	t.Parallel()
	t.Run("row1 one judgement naming two elements, and one new finding", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		c := judging("untouched: it reads no quantity", "cart.add.full", "billing.refund.add")
		c.Findings = []findingIn{gap("No rule says what a refund of nothing is.", "billing.refund.add")}
		r.judgeOK(c)
		o := r.verify()
		o.only(t, append(unjudgedKeys("cart.add.never", "cart.bounded", "system.honest"), "open f1")...)
	})
	t.Run("row2 no judgements and no findings", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		if res := r.judgeRaw("{}"); res.exit != 0 {
			t.Fatalf("want exit status 0, got %d\nstderr: %s", res.exit, res.stderr)
		}
		r.verify().only(t, unjudgedKeys(limitCandidates...)...)
	})
	t.Run("derived: members given as empty lists", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		if res := r.judgeRaw(`{"judgements": [], "findings": []}`); res.exit != 0 {
			t.Fatalf("want exit status 0, got %d\nstderr: %s", res.exit, res.stderr)
		}
	})
	t.Run("derived: the contract's example", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.judgeOK(checkResult{Findings: []findingIn{gap("one", "cart.pay.open"), gap("two"), gap("three")}})
		// The example of check_result, on the elements of the impact tree.
		res := r.judgeRaw(`{
  "judgements": [
    { "reason": "untouched: nothing here reads the manifest",
      "elements": ["cart.add.full", "billing.refund.add"] }
  ],
  "findings": [
    { "kind": "gap", "elements": ["cart.pay.open"],
      "text": "No rule says what happens when a When has no step text." },
    { "id": "f3", "status": "dismissed", "reason": "decided in system.git" }
  ]
}`)
		if res.exit != 0 {
			t.Fatalf("want exit status 0, got %d\nstderr: %s", res.exit, res.stderr)
		}
		r.verify().only(t, append(unjudgedKeys("cart.add.never", "cart.bounded", "system.honest"), "open f1", "open f2", "open f4")...)
	})
	for _, kind := range []string{"contradiction", "ignored", "gap", "missed"} {
		t.Run("derived: a finding of the kind "+kind, func(t *testing.T) {
			t.Parallel()
			r := amendedRepo(t, iLimitThen)
			r.judgeOK(checkResult{Findings: []findingIn{{Kind: kind, Text: "Found."}}})
			r.verify().has(t, "open f1")
		})
	}

	// Inputs that are no check_result. Each is refused with exit status 1,
	// and coherence_file stays as it was: absent where it was absent (see
	// judgeIn), and the same bytes where judge wrote it before.
	refusedInputs := []struct{ name, input string }{
		{"row3 not JSON", "this is no JSON"},
		{"row4 a finding with both a kind and an id", `{"findings": [{"kind": "gap", "text": "A gap.", "id": "f1", "status": "resolved"}]}`},
		{"row5 one element in two judgements", `{"judgements": [{"reason": "one", "elements": ["cart.add.full"]}, {"reason": "two", "elements": ["cart.add.full", "cart.bounded"]}]}`},
		{"derived: nothing on stdin", ""},
		{"derived: a JSON array", `[]`},
		{"derived: judgements not an array", `{"judgements": {"reason": "one", "elements": ["cart.add.full"]}}`},
		{"derived: a judgement with an empty reason", `{"judgements": [{"reason": "", "elements": ["cart.add.full"]}]}`},
		{"derived: a judgement without a reason", `{"judgements": [{"elements": ["cart.add.full"]}]}`},
		{"derived: a judgement naming no element", `{"judgements": [{"reason": "one", "elements": []}]}`},
		{"derived: a judgement without elements", `{"judgements": [{"reason": "one"}]}`},
		{"derived: a new finding without text", `{"findings": [{"kind": "gap"}]}`},
		{"derived: a new finding with an empty text", `{"findings": [{"kind": "gap", "text": ""}]}`},
		{"derived: a new finding of an unknown kind", `{"findings": [{"kind": "doubt", "text": "A doubt."}]}`},
		{"derived: a new finding with a member the contract does not name", `{"findings": [{"kind": "gap", "text": "A gap.", "severity": "high"}]}`},
		{"derived: a finding with an id and no status", `{"findings": [{"id": "f1"}]}`},
		{"derived: a finding with an unknown status", `{"findings": [{"id": "f1", "status": "closed"}]}`},
		{"derived: a finding dismissed without a reason", `{"findings": [{"id": "f1", "status": "dismissed"}]}`},
		{"derived: a finding dismissed with an empty reason", `{"findings": [{"id": "f1", "status": "dismissed", "reason": ""}]}`},
	}
	for _, c := range refusedInputs {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			r := amendedRepo(t, iLimitThen)
			refused(t, r.judgeRaw(c.input), 1)
			if r.hasCoherence() {
				t.Errorf("want no %s, as nothing was recorded", coherenceName)
			}
		})
		t.Run(c.name+", with a check recorded before", func(t *testing.T) {
			t.Parallel()
			r := amendedRepo(t, iLimitThen)
			r.judgeOK(checkResult{Findings: []findingIn{gap("A gap.")}})
			before := r.coherence()
			refused(t, r.judgeRaw(c.input), 1)
			if r.coherence() != before {
				t.Errorf("want %s as it was", coherenceName)
			}
		})
	}
}

func TestBaselineJudgeAmendment(t *testing.T) {
	t.Parallel()
	t.Run("row1 changed", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.judgeOK(judging("untouched", "cart.add.full"))
		r.verify().lacks(t, "unjudged cart.add.full")
	})
	t.Run("row2 none: the tree is as the manifest records it", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, impactTree())
		refused(t, r.judgeRaw("{}"), 1)
		if r.hasCoherence() {
			t.Errorf("want no %s", coherenceName)
		}
	})
	t.Run("derived: none, with a check left over", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.judgeOK(checkResult{})
		r.set(impactTree())
		before := r.coherence()
		refused(t, r.judgeRaw("{}"), 1)
		if r.coherence() != before {
			t.Errorf("want %s as it was", coherenceName)
		}
	})
}

func TestBaselineJudgeCandidate(t *testing.T) {
	t.Parallel()
	t.Run(`row1 "untouched: it reads no manifest", found through events`, func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.judgeOK(judging("untouched: it reads no manifest", "cart.add.full"))
		r.verify().only(t, unjudgedKeys(without(limitCandidates, "cart.add.full")...)...)
	})
	t.Run("row2 judged before with another reason", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.judgeOK(judging("untouched", "billing.refund.add", "cart.add.full"))
		r.judgeOK(judging("untouched: nothing here builds code", "billing.refund.add"))
		// The former judgement of billing.refund.add is replaced; that of
		// cart.add.full, a candidate still, is kept (baseline.judge.write).
		r.verify().only(t, unjudgedKeys("cart.add.never", "cart.bounded", "system.honest")...)
	})
	t.Run("row3 changed itself", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		refused(t, r.judge(judging("untouched", "cart.add.limit")), 1, "cart.add.limit")
		if r.hasCoherence() {
			t.Errorf("want no %s", coherenceName)
		}
	})
	t.Run("row4 an identifier the tree does not hold", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		refused(t, r.judge(judging("untouched", "cart.add.nosuch")), 1, "cart.add.nosuch")
		if r.hasCoherence() {
			t.Errorf("want no %s", coherenceName)
		}
	})
	t.Run("derived: an element of the tree that is no candidate", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		refused(t, r.judge(judging("untouched", "cart.pay.open")), 1, "cart.pay.open")
	})
	t.Run("derived: one refused judgement, and nothing is recorded", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.judgeOK(judging("untouched", "cart.bounded"))
		before := r.coherence()
		c := judging("untouched", "cart.add.full")
		c.Judgements = append(c.Judgements, judgementIn{"untouched", []string{"cart.add.nosuch"}})
		refused(t, r.judge(c), 1, "cart.add.nosuch")
		if r.coherence() != before {
			t.Errorf("want %s as it was", coherenceName)
		}
		r.verify().has(t, "unjudged cart.add.full")
	})
}

// recordTwo records the findings f1 and f2 for the impact tree amended by
// iLimitThen, every candidate judged.
func recordTwo(t *testing.T) *repo {
	t.Helper()
	r := amendedRepo(t, iLimitThen)
	r.judgeAll(nil, gap("The first gap."), gap("The second gap.", "cart.pay.open"))
	r.verify().only(t, "open f1", "open f2")
	return r
}

func TestBaselineJudgeCloseFinding(t *testing.T) {
	t.Parallel()
	t.Run(`row1 f2 "resolved"`, func(t *testing.T) {
		t.Parallel()
		r := recordTwo(t)
		r.judgeOK(checkResult{Findings: []findingIn{resolved("f2")}})
		r.verify().only(t, "open f1")
	})
	t.Run(`row2 f2 "dismissed", with the reason "decided in system.git"`, func(t *testing.T) {
		t.Parallel()
		r := recordTwo(t)
		r.judgeOK(checkResult{Findings: []findingIn{dismissed("f2", "decided in system.git")}})
		r.verify().only(t, "open f1")
	})
	t.Run(`row3 f2, dismissed before, "resolved"`, func(t *testing.T) {
		t.Parallel()
		r := recordTwo(t)
		r.judgeOK(checkResult{Findings: []findingIn{dismissed("f2", "decided in system.git")}})
		r.judgeOK(checkResult{Findings: []findingIn{resolved("f2")}})
		r.verify().only(t, "open f1")
	})
	t.Run(`row4 f9, with f1 and f2 recorded, "resolved"`, func(t *testing.T) {
		t.Parallel()
		r := recordTwo(t)
		before := r.coherence()
		refused(t, r.judge(checkResult{Findings: []findingIn{resolved("f9")}}), 1, "f9")
		if r.coherence() != before {
			t.Errorf("want %s as it was", coherenceName)
		}
		r.verify().only(t, "open f1", "open f2")
	})
	t.Run("derived: an id, with no check recorded", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		refused(t, r.judge(checkResult{Findings: []findingIn{resolved("f1")}}), 1, "f1")
		if r.hasCoherence() {
			t.Errorf("want no %s", coherenceName)
		}
	})
	t.Run("derived: the finding stays, and keeps its number", func(t *testing.T) {
		t.Parallel()
		r := recordTwo(t)
		r.judgeOK(checkResult{Findings: []findingIn{resolved("f2")}})
		r.judgeOK(checkResult{Findings: []findingIn{gap("The third gap.")}})
		r.verify().only(t, "open f1", "open f3")
	})
}

func TestBaselineJudgeNewFinding(t *testing.T) {
	t.Parallel()
	t.Run("row1 an element, with f1 and f2 recorded before", func(t *testing.T) {
		t.Parallel()
		r := recordTwo(t)
		r.judgeOK(checkResult{Findings: []findingIn{gap("A third gap.", "cart.pay.cap")}})
		o := r.verify()
		o.only(t, "open f1", "open f2", "open f3")
		if p := o.has(t, "open f3"); p.Text != "A third gap." {
			t.Errorf("want f3 with its text %q, got %q", "A third gap.", p.Text)
		}
	})
	t.Run("row2 none, with no finding recorded before", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.judgeAll(nil, gap("A gap beside every element."))
		o := r.verify()
		o.only(t, "open f1")
		if p := o.has(t, "open f1"); p.Text != "A gap beside every element." {
			t.Errorf("want f1 with its text, got %q", p.Text)
		}
	})
	t.Run("row3 an element removed by the amendment", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, removeOpen())
		r.judgeOK(checkResult{Findings: []findingIn{gap("Nothing greets the buyers now.", "cart.pay.open")}})
		r.verify().has(t, "open f1")
	})
	t.Run("row4 an element neither holds", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		refused(t, r.judge(checkResult{Findings: []findingIn{gap("A gap.", "cart.add.nosuch")}}), 1, "cart.add.nosuch")
		if r.hasCoherence() {
			t.Errorf("want no %s", coherenceName)
		}
	})
	t.Run("derived: several new findings, numbered as they come", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.judgeAll(nil, gap("First."), gap("Second."), gap("Third."))
		o := r.verify()
		o.only(t, "open f1", "open f2", "open f3")
		for id, text := range map[string]string{"f1": "First.", "f2": "Second.", "f3": "Third."} {
			if p := o.has(t, "open "+id); p.Text != text {
				t.Errorf("want %s with the text %q, got %q", id, text, p.Text)
			}
		}
	})
	t.Run("derived: the numbers start again once coherence_file is removed", func(t *testing.T) {
		t.Parallel()
		r := recordTwo(t)
		r.judgeOK(checkResult{Findings: []findingIn{resolved("f1"), resolved("f2")}})
		r.manifest()
		r.commit("accept")
		r.set(amended(r.tr, iFullThen))
		r.judgeAll(nil, gap("A new gap."))
		r.verify().only(t, "open f1")
	})
}

func TestBaselineJudgeUnreadable(t *testing.T) {
	t.Parallel()
	damaged := map[string]func(written string) string{
		"row2 <<<<<<< HEAD in it": func(w string) string {
			return "<<<<<<< HEAD\n" + w + "=======\n" + w + ">>>>>>> main\n"
		},
		"row3 empty":              func(string) string { return "" },
		"derived: not UTF-8":      func(string) string { return "\xff\xfe{}" },
		"derived: cut off midway": func(w string) string { return w[:len(w)/2] },
	}
	for _, cmd := range []string{"judge", "verify"} {
		t.Run(cmd, func(t *testing.T) {
			t.Parallel()
			t.Run("row1 written by judge", func(t *testing.T) {
				t.Parallel()
				r := amendedRepo(t, iLimitThen)
				r.judgeOK(checkResult{Findings: []findingIn{gap("A gap.")}})
				if cmd == "judge" {
					r.judgeOK(judging("untouched", "cart.add.full"))
					return
				}
				r.verify().has(t, "open f1")
			})
			for name, damage := range damaged {
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					r := amendedRepo(t, iLimitThen)
					r.judgeAll(nil)
					p := filepath.Join(r.root(), coherenceName)
					if err := os.WriteFile(p, []byte(damage(r.coherence())), 0o644); err != nil {
						t.Fatal(err)
					}
					// That nothing goes to stdout, and that the file stays
					// as it was, is checked for every run (see judgeIn and
					// verifyIn).
					if cmd == "judge" {
						refused(t, r.judge(checkResult{}), 2)
						return
					}
					r.verify().exits(t, 2)
				})
			}
		})
	}
}

func TestBaselineJudgeWrite(t *testing.T) {
	t.Parallel()
	t.Run("row1 without .coherence.json", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.judgeOK(checkResult{})
		if !r.hasCoherence() {
			t.Errorf("want %s written", coherencePath)
		}
	})
	t.Run("derived: run inside the tree, written into the spec root", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		cwd := filepath.Join(r.root(), "capabilities", "cart")
		if res := judgeIn(t, r.dir, cwd, r.root(), "{}"); res.exit != 0 {
			t.Fatalf("want exit status 0, got %d\nstderr: %s", res.exit, res.stderr)
		}
	})
	t.Run("row2 with a judgement recorded for an element edited since", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.judgeOK(judging("untouched", "cart.add.full"))
		// cart.add.full is a touched element now, and no candidate.
		r.set(amended(impactTree(), iLimitThen, iFullThen))
		r.judgeOK(checkResult{})
		// Undone, it is a candidate again: its judgement was dropped.
		r.set(amended(impactTree(), iLimitThen))
		r.verify().has(t, "unjudged cart.add.full")
	})
	t.Run("derived: a judgement of a candidate still is kept", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.judgeOK(judging("untouched", "cart.add.full"))
		r.judgeOK(checkResult{})
		r.verify().lacks(t, "unjudged cart.add.full")
	})
	t.Run("derived: every finding recorded before is kept", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.judgeAll(nil, gap("A gap."))
		r.judgeOK(checkResult{})
		r.verify().only(t, "open f1")
	})
	t.Run("row3 without write permission", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.judgeOK(checkResult{Findings: []findingIn{gap("A gap.")}})
		former := r.coherence()
		undo, err := makeCoherenceUnwritable(r.root())
		if err != nil {
			t.Skipf("cannot take the write permission here: %v", err)
		}
		// That the former file stays as it was is checked for every run
		// that does not exit with 0 (see judgeIn).
		res := r.judge(judging("untouched", "cart.add.full"))
		undo()
		refused(t, res, 2)
		if r.coherence() != former {
			t.Errorf("want the former %s as it was", coherenceName)
		}
	})
}

// TestSystemCoherenceFile checks what system.coherence-file says beyond the
// rules: lint skips the file, and the manifest does not record it.
func TestSystemCoherenceFile(t *testing.T) {
	t.Parallel()
	t.Run("lint skips it", func(t *testing.T) {
		t.Parallel()
		check(t, plainTree().with(coherencePath, "{}\n")).clean(t)
	})
	t.Run("the manifest does not record it", func(t *testing.T) {
		t.Parallel()
		if got, want := manifestOf(t, plainTree().with(coherencePath, "{}\n")), manifestFor(plainParts()); got != want {
			t.Errorf("want the manifest\n%s\ngot\n%s", want, got)
		}
	})
	t.Run("judge writes it as JSON in UTF-8", func(t *testing.T) {
		t.Parallel()
		// Checked for every run of judge that exits with 0 (see judgeIn);
		// here with text beyond ASCII.
		r := amendedRepo(t, iLimitThen)
		r.judgeOK(checkResult{
			Judgements: []judgementIn{{"unberührt: Bruno Baketarić", []string{"cart.add.full"}}},
			Findings:   []findingIn{gap("Lücke — „quoted“")},
		})
		if p := r.verify().has(t, "open f1"); p.Text != "Lücke — „quoted“" {
			t.Errorf("want the finding's text as given, got %q", p.Text)
		}
	})
}
