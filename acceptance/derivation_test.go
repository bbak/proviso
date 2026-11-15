package acceptance

// The rules of baseline.diff and baseline.parts as every command that
// derives the element diff follows them: `impact`, `pack`, `judge` and
// `verify`, beside `diff`, whose own tests are in diff_test.go and
// parts_test.go. Each case is the case of the test named after the rule, run
// with each of these commands (see derivedIn): pack's context_pack holds the
// element diff as element_diff does, so each row's outcome is checked there
// in full; impact, verify and judge show the element diff only in what they
// derive from it.

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

// derivedIn runs `pack --json`, `impact --json`, `verify --json` and `judge`
// in cwd, in that order, where no check is recorded yet, and checks that each
// derived the element diff that lists exactly want, each identifier with its
// change after a colon, as in "cart.add.limit:changed":
//
//   - pack: diff in context_pack lists exactly want;
//   - impact: no element want lists is a candidate, every touched element a
//     reason comes from is one want lists, and nothing is a candidate where
//     want is empty (coherence.impact.listed);
//   - verify: "missing" where want is not empty, and nothing beside
//     "undecided" otherwise (coherence.verify.missing);
//   - judge: a check_result with nothing in it recorded where want is not
//     empty, refused with exit status 1 otherwise (coherence.judge.amendment).
//
// It returns the element diff as pack holds it, so that the assertions of
// the diff tests apply to it.
func derivedIn(t *testing.T, cwd string, want ...string) diffOutcome {
	t.Helper()
	o := packDiffIn(t, cwd)
	var got []string
	for _, e := range o.ed.Elements {
		got = append(got, e.ID+":"+e.Change)
	}
	if strings.Join(sorted(got), " ") != strings.Join(sorted(want), " ") {
		t.Errorf("pack: want the element diff to list %v, got:\n%s", sorted(want), o.dump())
	}
	listed := map[string]bool{}
	for _, w := range want {
		id, _, _ := strings.Cut(w, ":")
		listed[id] = true
	}
	impactDerived(t, cwd, listed)
	verifyDerived(t, cwd, len(listed) > 0)
	judgeDerived(t, cwd, len(listed) > 0)
	return o
}

// derivedAllIn is derivedIn for a tree without manifest.sum: every element of
// tr is listed as added.
func derivedAllIn(t *testing.T, cwd string, tr tree) diffOutcome {
	t.Helper()
	return derivedIn(t, cwd, withChange(treeIDs(tr), "added")...)
}

// packDiffIn runs `pack --json` in cwd, which must succeed, and gives the
// element diff its context_pack holds.
func packDiffIn(t *testing.T, cwd string) diffOutcome {
	t.Helper()
	r := execProviso(t, cwd, "pack", "--json")
	if r.exit != 0 {
		t.Fatalf("pack: want exit status 0, got %d\nstderr: %s", r.exit, r.stderr)
	}
	cp, ok := parseContextPack(t, r.stdout)
	if !ok {
		t.FailNow()
	}
	checkChains(t, cp.Diff)
	return diffOutcome{json: r, ed: elementDiff{Root: cp.Root, Revision: cp.Revision, Elements: cp.Diff}}
}

// impactDerived runs `impact --json` in cwd and checks its candidates against
// the elements listed.
func impactDerived(t *testing.T, cwd string, listed map[string]bool) {
	t.Helper()
	r := execProviso(t, cwd, "impact", "--json")
	if r.exit != 0 {
		t.Errorf("impact: want exit status 0, got %d\nstderr: %s", r.exit, r.stderr)
		return
	}
	is := parseImpactSet(t, r.stdout)
	if len(listed) == 0 && len(is.Candidates) > 0 {
		t.Errorf("impact: the element diff lists nothing, but candidates are found:\n%s", impactOutcome{is: is}.dump())
	}
	for _, c := range is.Candidates {
		if listed[c.ID] {
			t.Errorf("impact: %s is a candidate, but the element diff lists it", c.ID)
		}
		for _, rs := range c.Reasons {
			for _, f := range rs.From {
				if !listed[f] {
					t.Errorf("impact: %s is found from %s, which the element diff does not list", c.ID, f)
				}
			}
		}
	}
}

// verifyDerived runs `verify --json` in cwd, where no check is recorded, and
// checks that it reports "missing" exactly where the element diff lists an
// element, and nothing else but "undecided", which verify reports across the
// whole tree.
func verifyDerived(t *testing.T, cwd string, amendment bool) {
	t.Helper()
	r := execProviso(t, cwd, "verify", "--json")
	if r.exit == 2 {
		t.Errorf("verify: want it to run, but it exited with 2:\n%s", r.stderr)
		return
	}
	var got []string
	for _, p := range parseVerification(t, r.stdout).Problems {
		if p.Kind != "undecided" {
			got = append(got, p.key())
		}
	}
	want := ""
	if amendment {
		want = "missing"
	}
	if strings.Join(got, "|") != want {
		t.Errorf("verify: want the problems %q beside any undecided, got %q", want, got)
	}
}

// judgeDerived runs `judge` in cwd with a check_result that holds nothing,
// which is recorded only where the element diff lists an element.
func judgeDerived(t *testing.T, cwd string, amendment bool) {
	t.Helper()
	r := execProvisoWith(t, cwd, strings.NewReader("{}"), "judge")
	want := 1
	if amendment {
		want = 0
	}
	if r.exit != want {
		t.Errorf("judge: want exit status %d, as the element diff lists elements: %v, got %d\nstderr: %s", want, amendment, r.exit, r.stderr)
	}
}

// treeIDs lists the identifiers of every element of tr: the anchors of its
// headings.
func treeIDs(tr tree) []string {
	var ids []string
	for _, content := range tr {
		for _, l := range strings.Split(content, "\n") {
			if m := anchorPattern.FindStringSubmatch(l); m != nil {
				ids = append(ids, m[1])
			}
		}
	}
	sort.Strings(ids)
	return ids
}

func TestBaselineDiffBaseline_Commands(t *testing.T) {
	t.Parallel()
	t.Run("row1 with manifest.sum, compared with the manifest", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		tr := diffTree().appendTo(featurePath, extraRule)
		r.set(tr)
		o := derivedIn(t, r.dir, "cart.add.extra:added")
		checkFull(t, o.elem(t, "cart.add.extra"), tr, featurePath)
	})
	t.Run("derived: with manifest.sum, the tree as it records it", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		derivedIn(t, r.dir)
	})
	t.Run("row2 without manifest.sum, every element added", func(t *testing.T) {
		t.Parallel()
		tr := diffTree()
		dir := tr.write(t)
		o := derivedAllIn(t, dir, tr)
		checkRevision(t, o, "")
		paths := map[string]string{"system": systemPath, "cart.add.limit": featurePath, "cart.item": glossaryPath, "system.ref.commonmark": referencesPath, "cart.x.line": xPath}
		for id, path := range paths {
			checkFull(t, o.elem(t, id), tr, path)
		}
	})
}

func TestBaselineDiffAdded_Commands(t *testing.T) {
	t.Parallel()
	t.Run("row1 new since the manifest was written", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		tr := diffTree().appendTo(featurePath, extraRule)
		r.set(tr)
		o := derivedIn(t, r.dir, "cart.add.extra:added")
		checkFull(t, o.elem(t, "cart.add.extra"), tr, featurePath)
		if e := o.elem(t, "cart.add.extra"); len(e.Parts) != 0 {
			t.Errorf("want cart.add.extra listed in full, naming no part, got %v", e.Parts)
		}
		checkRevision(t, o, "")
	})
	t.Run("row2 unchanged, handed over", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		r.set(diffTree().appendTo(featurePath, extraRule))
		derivedIn(t, r.dir, "cart.add.extra:added").lacksElem(t, "cart.add.limit")
	})
	t.Run("row3 moved into another feature's file, handed over", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		r.set(moved())
		derivedIn(t, r.dir)
	})
}

func TestBaselineDiffChanged_Commands(t *testing.T) {
	t.Parallel()
	// amendedIn accepts diffTree and applies cs.
	amendedIn := func(t *testing.T, cs ...change) (*repo, tree, string) {
		t.Helper()
		r, base := accepted(t, diffTree())
		tr := diffTree().apply(cs...)
		r.set(tr)
		return r, tr, base
	}
	t.Run("row1 its Then step reworded", func(t *testing.T) {
		t.Parallel()
		r, tr, base := amendedIn(t, limitThen)
		o := derivedIn(t, r.dir, "cart.add.limit:changed")
		e := o.elem(t, "cart.add.limit")
		onlyParts(t, e, "step Then:changed")
		checkNow(t, part(t, e, "step Then"), "Then      the cart holds <quantity> more pieces of <product>\n  And     nothing else changes")
		checkFormer(t, part(t, e, "step Then"), "Then      the cart holds <quantity> more of <product>\n  And     nothing else changes")
		checkFull(t, e, tr, featurePath)
		checkRevision(t, o, base)
	})
	t.Run("row2 its Rationale deleted", func(t *testing.T) {
		t.Parallel()
		r, _, base := amendedIn(t, limitNoRatio)
		o := derivedIn(t, r.dir, "cart.add.limit:changed")
		e := o.elem(t, "cart.add.limit")
		onlyParts(t, e, "section Rationale:removed")
		checkFormer(t, part(t, e, "section Rationale"), "A shelf holds ten pieces.")
		checkRevision(t, o, base)
	})
	t.Run("row3 a new Examples row", func(t *testing.T) {
		t.Parallel()
		r, _, _ := amendedIn(t, limitNewRow)
		e := derivedIn(t, r.dir, "cart.add.limit:changed").elem(t, "cart.add.limit")
		key := "row " + rowKey("quantity", "5", "product", "pear")
		onlyParts(t, e, key+":added")
		checkNow(t, part(t, e, key), "| 5 | pear | yes | added | |")
	})
	t.Run("row4 an Examples row with a new outcome", func(t *testing.T) {
		t.Parallel()
		r, _, _ := amendedIn(t, limitOutcome)
		e := derivedIn(t, r.dir, "cart.add.limit:changed").elem(t, "cart.add.limit")
		key := "row " + rowKey("quantity", "1", "product", "apple")
		onlyParts(t, e, key+":changed")
		checkNow(t, part(t, e, key), "| 1 | apple | yes | put in the cart | |")
		checkFormer(t, part(t, e, key), "| 1 | apple | yes | added | |")
	})
	t.Run("row5 moved into another feature's file", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		r.set(moved())
		derivedIn(t, r.dir)
	})
	t.Run("row6 its text wrapped at other places", func(t *testing.T) {
		t.Parallel()
		r, _, _ := amendedIn(t,
			change{featurePath, "A cart that takes anything would empty the shelves.", "A cart that takes anything\nwould empty   the shelves."},
			change{featurePath, "A shelf holds ten pieces.", "A shelf holds\nten pieces."},
			change{featurePath, "Then      the cart holds <quantity> more of <product>", "Then the cart holds <quantity> more of <product>"},
		)
		derivedIn(t, r.dir)
	})
	t.Run("row7 one of a feature's rules changed", func(t *testing.T) {
		t.Parallel()
		r, _, _ := amendedIn(t, limitThen)
		derivedIn(t, r.dir, "cart.add.limit:changed").lacksElem(t, "cart.add")
	})
}

func TestBaselineDiffRemoved_Commands(t *testing.T) {
	t.Parallel()
	emptyText, _ := elementText(diffTree()[featurePath], "cart.add.empty")
	t.Run("row1 still in the tree, compared", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		r.set(diffTree().edit(featurePath, emptyText, ""))
		derivedIn(t, r.dir, "cart.add.empty:removed").lacksElem(t, "cart.add.limit")
	})
	t.Run("row2 its rule deleted", func(t *testing.T) {
		t.Parallel()
		r, base := accepted(t, diffTree())
		r.set(diffTree().edit(featurePath, emptyText, ""))
		o := derivedIn(t, r.dir, "cart.add.empty:removed")
		e := o.elem(t, "cart.add.empty")
		if e.Former == nil || normWS(*e.Former) != normWS(emptyText) {
			t.Errorf("want the former text of cart.add.empty\n%s\ngot\n%s", emptyText, deref(e.Former))
		}
		checkRevision(t, o, base)
	})
	t.Run("row3 moved to the system glossary", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		item, _ := elementText(diffTree()[glossaryPath], "cart.item")
		r.set(diffTree().edit(glossaryPath, item, "").with("proviso/glossary.md", "# Glossary\n\n"+item))
		derivedIn(t, r.dir)
	})
}

func TestBaselineDiffDependency_Commands(t *testing.T) {
	t.Parallel()
	// depOf accepts base, amends it with cs, and checks that each command
	// lists want.
	depOf := func(t *testing.T, base tree, cs []change, want ...string) diffOutcome {
		t.Helper()
		r, _ := accepted(t, base)
		r.set(amended(base, cs...))
		return derivedIn(t, r.dir, want...)
	}
	t.Run("row1 a glossary entry's Definition reworded", func(t *testing.T) {
		t.Parallel()
		o := depOf(t, diffTree(), []change{itemDefinition}, "cart.item:changed", "cart.add.limit:changed-through", "cart.add.empty:changed-through")
		checkThrough(t, o.elem(t, "cart.add.limit"), "cart.item")
		checkThrough(t, o.elem(t, "cart.add.empty"), "cart.item")
	})
	t.Run("row2 a reference with a new Version", func(t *testing.T) {
		t.Parallel()
		o := depOf(t, diffTree(), []change{commonmarkNew}, "system.ref.commonmark:changed", "cart.thematic-break:changed-through", "cart.x.line:changed-through")
		checkThrough(t, o.elem(t, "cart.thematic-break"), "system.ref.commonmark")
		checkThrough(t, o.elem(t, "cart.x.line"), "system.ref.commonmark > cart.thematic-break")
	})
	t.Run("row3 a member of a contract element with a new value", func(t *testing.T) {
		t.Parallel()
		o := depOf(t, impactTree(), []change{iCountBody}, "cart.api.line.count:changed", "cart.pay.lines:changed-through", "cart.pay.order:changed-through")
		checkThrough(t, o.elem(t, "cart.pay.lines"), "cart.api.line.count")
	})
	t.Run("row4 a member with a new description, reaching the rules referring to the element holding it", func(t *testing.T) {
		t.Parallel()
		o := depOf(t, impactTree(), []change{iCountBody}, "cart.api.line.count:changed", "cart.pay.lines:changed-through", "cart.pay.order:changed-through")
		checkThrough(t, o.elem(t, "cart.pay.order"), "cart.api.line.count")
	})
	t.Run("row5 a rule's Then step changed, reaching the rules handing over to it", func(t *testing.T) {
		t.Parallel()
		o := depOf(t, handTree(), []change{limitThen}, "cart.add.limit:changed", "cart.x.hand:changed-through")
		checkThrough(t, o.elem(t, "cart.x.hand"), "cart.add.limit")
	})
	t.Run("row6 a rule's Examples row changed, reaching the rules handing over to it", func(t *testing.T) {
		t.Parallel()
		o := depOf(t, handTree(), []change{limitOutcome}, "cart.add.limit:changed", "cart.x.hand:changed-through")
		checkThrough(t, o.elem(t, "cart.x.hand"), "cart.add.limit")
	})
	t.Run("row7 a rule changed in its Rationale only, marking a term whose Definition changed", func(t *testing.T) {
		t.Parallel()
		o := depOf(t, handTree(), []change{itemDefinition, limitNoRatio},
			"cart.item:changed", "cart.add.limit:changed", "cart.add.empty:changed-through", "cart.x.hand:changed-through")
		checkThrough(t, o.elem(t, "cart.x.hand"), "cart.item > cart.add.limit")
	})
	t.Run("row8 a reference with a new Version, while the term it defines has its Defined by changed too", func(t *testing.T) {
		t.Parallel()
		o := depOf(t, diffTree(), []change{commonmarkNew, breakDefinedBy},
			"system.ref.commonmark:changed", "cart.thematic-break:changed", "cart.x.line:changed-through")
		checkThrough(t, o.elem(t, "cart.x.line"), "cart.thematic-break", "system.ref.commonmark > cart.thematic-break")
	})
	t.Run("row9 a member removed, while the element holding it stays", func(t *testing.T) {
		t.Parallel()
		o := depOf(t, impactTree(), []change{countRemoved, linesToLine},
			"cart.api.line.count:removed", "cart.pay.lines:changed", "cart.pay.order:changed-through")
		checkThrough(t, o.elem(t, "cart.pay.order"), "cart.api.line.count")
	})
	t.Run("row10 a new alias only", func(t *testing.T) {
		t.Parallel()
		depOf(t, diffTree(), []change{itemAlias}, "cart.item:changed")
	})
	t.Run("row11 a reference with a new Contact only", func(t *testing.T) {
		t.Parallel()
		depOf(t, diffTree(), []change{{referencesPath, "- Contact: Dana Reyes", "- Contact: Bruno Baketarić"}}, "system.ref.commonmark:changed")
	})
	t.Run("row12 a glossary entry's Deviations reworded", func(t *testing.T) {
		t.Parallel()
		o := depOf(t, deviationsTree(), []change{breakDeviations}, "cart.thematic-break:changed", "cart.x.line:changed-through")
		checkThrough(t, o.elem(t, "cart.x.line"), "cart.thematic-break")
	})
	t.Run("row13 a Definition reworded, while another entry marks it in its Deviations", func(t *testing.T) {
		t.Parallel()
		o := depOf(t, lineEndingDeviationsTree(), []change{lineEndingDefinition},
			"system.line-ending:changed", "cart.thematic-break:changed-through", "cart.x.line:changed-through")
		checkThrough(t, o.elem(t, "cart.thematic-break"), "system.line-ending")
		checkThrough(t, o.elem(t, "cart.x.line"), "system.line-ending > cart.thematic-break")
	})
	t.Run("row14 a Definition reworded, while another entry marks it in its Definition", func(t *testing.T) {
		t.Parallel()
		o := depOf(t, lineEndingTree(), []change{lineEndingDefinition},
			"system.line-ending:changed", "cart.receipt-line:changed-through", "cart.x.print:changed-through")
		checkThrough(t, o.elem(t, "cart.receipt-line"), "system.line-ending")
		checkThrough(t, o.elem(t, "cart.x.print"), "system.line-ending > cart.receipt-line")
	})
	t.Run("row15 a Definition reworded, while another entry holds it escaped in its Definition", func(t *testing.T) {
		t.Parallel()
		o := depOf(t, escapedEndingTree(), []change{lineEndingDefinition}, "system.line-ending:changed", "cart.x.end:changed-through")
		checkThrough(t, o.elem(t, "cart.x.end"), "system.line-ending")
	})
	t.Run("row16 a Definition reworded, while two entries mark each other", func(t *testing.T) {
		t.Parallel()
		o := depOf(t, circleTree(), []change{itemFromShelf}, "cart.item:changed", "cart.shelf:changed-through",
			"cart.add.limit:changed-through", "cart.add.empty:changed-through", "cart.x.stock:changed-through")
		checkThrough(t, o.elem(t, "cart.shelf"), "cart.item")
		checkThrough(t, o.elem(t, "cart.add.limit"), "cart.item")
		checkThrough(t, o.elem(t, "cart.add.empty"), "cart.item")
		checkThrough(t, o.elem(t, "cart.x.stock"), "cart.item > cart.shelf")
	})
	t.Run("row17 a Definition reworded, while the entry marking it and a third mark each other", func(t *testing.T) {
		t.Parallel()
		o := depOf(t, aisleTree(), []change{itemDefinition}, "cart.item:changed", "cart.shelf:changed-through",
			"cart.aisle:changed-through", "cart.add.limit:changed-through", "cart.add.empty:changed-through")
		checkThrough(t, o.elem(t, "cart.shelf"), "cart.item")
		checkThrough(t, o.elem(t, "cart.aisle"), "cart.item > cart.shelf")
	})
	t.Run("row18 a Definition reworded, while a contract element marks another term and it in its Represents", func(t *testing.T) {
		t.Parallel()
		tr := impactTree().edit(apiPath, "- Represents: [item]\n", "- Represents: [item] and [price]\n")
		o := depOf(t, tr, []change{priceDefinition}, "system.price:changed", "cart.api.line:changed-through",
			"cart.pay.order:changed-through", "cart.pay.total:changed-through", "billing.refund.full:changed-through")
		checkThrough(t, o.elem(t, "cart.api.line"), "system.price")
		checkThrough(t, o.elem(t, "cart.pay.order"), "system.price > cart.api.line")
	})
	t.Run("row19 a new line in a contract element's body, while another refers to it in its field Builds on", func(t *testing.T) {
		t.Parallel()
		o := depOf(t, slipTree(), []change{lineBody}, "cart.api.line:changed", "cart.pay.order:changed-through",
			"cart.api.slip:changed-through", "cart.pay.slip:changed-through")
		checkThrough(t, o.elem(t, "cart.pay.order"), "cart.api.line")
		checkThrough(t, o.elem(t, "cart.api.slip"), "cart.api.line")
		checkThrough(t, o.elem(t, "cart.pay.slip"), "cart.api.line > cart.api.slip")
	})
	t.Run("row20 a new line in a contract element's body, while another names it only in its prose", func(t *testing.T) {
		t.Parallel()
		o := depOf(t, proseSlipTree(), []change{lineBody}, "cart.api.line:changed", "cart.pay.order:changed-through")
		checkThrough(t, o.elem(t, "cart.pay.order"), "cart.api.line")
	})
	t.Run("row21 a Definition reworded, while another entry marks it and a rule marks both", func(t *testing.T) {
		t.Parallel()
		o := depOf(t, shelfItemTree(), []change{itemDefinition}, "cart.item:changed", "cart.shelf:changed-through",
			"cart.add.limit:changed-through", "cart.add.empty:changed-through", "cart.x.stock:changed-through")
		checkThrough(t, o.elem(t, "cart.shelf"), "cart.item")
		checkThrough(t, o.elem(t, "cart.x.stock"), "cart.item")
	})
	t.Run("row22 a Definition reworded, while two entries mark it and a rule marks both of them but not it", func(t *testing.T) {
		t.Parallel()
		o := depOf(t, bothEndingsTree(), []change{lineEndingDefinition}, "system.line-ending:changed", "cart.thematic-break:changed-through",
			"cart.receipt-line:changed-through", "cart.x.line:changed-through", "cart.x.print:changed-through")
		checkThrough(t, o.elem(t, "cart.x.print"), "system.line-ending > cart.receipt-line")
	})
	t.Run("row23 a Definition reworded, while it marks its own term", func(t *testing.T) {
		t.Parallel()
		o := depOf(t, selfTree(), []change{itemSelfDefinition}, "cart.item:changed", "cart.add.limit:changed-through", "cart.add.empty:changed-through")
		checkThrough(t, o.elem(t, "cart.add.limit"), "cart.item")
		checkThrough(t, o.elem(t, "cart.add.empty"), "cart.item")
	})
	t.Run("row24 a parameter's Rationale reworded only", func(t *testing.T) {
		t.Parallel()
		depOf(t, diffTree(), []change{{paramsPath, "- Rationale: a shelf holds ten pieces.", "- Rationale: a shelf holds no more than ten pieces."}}, "cart.max-quantity:changed")
	})
}

func TestBaselineDiffUndecided_Commands(t *testing.T) {
	t.Parallel()
	// Each tree is written once, without manifest.sum, so that every element
	// is listed as added, and the commands run on it once, before the
	// subtests read what pack listed; judge writes into the folder, which is
	// why no subtest runs a command of its own on it.
	tr := undecidedTree()
	o := derivedAllIn(t, tr.write(t), tr)
	undecided := func(t *testing.T, o diffOutcome, id string, want ...place) {
		t.Helper()
		if got := o.elem(t, id).Undecided; fmt.Sprint(got) != fmt.Sprint(want) {
			t.Errorf("want %s to name the undecided places %v, got %v", id, want, got)
		}
	}
	rt := undecidedTree().with("proviso/capabilities/cart/contracts/order.md",
		"# Order\n\n## <a id=\"cart.order\"></a> order\n\n- Represents: [basket]\n\nWhat a buyer orders.\n")
	ro := derivedAllIn(t, rt.write(t), rt)
	gt := undecidedTree().appendTo(glossaryPath, trolleyEntry)
	gto := derivedAllIn(t, gt.write(t), gt)
	lt := undecidedTree().appendTo(xPath, literalRule).appendTo(glossaryPath, wagonEntry)
	lo := derivedAllIn(t, lt.write(t), lt)

	t.Run("row1 a rule with every step decided", func(t *testing.T) {
		t.Parallel()
		undecided(t, o, "cart.add.limit")
	})
	t.Run("row2 a rule with Then ?", func(t *testing.T) {
		t.Parallel()
		undecided(t, o, "cart.x.weigh", place{xPath, tr.line(xPath, "Then      ?")})
	})
	t.Run("row3 a rule marking an open term", func(t *testing.T) {
		t.Parallel()
		undecided(t, o, "cart.x.basket", place{xPath, tr.line(xPath, "[basket]")})
	})
	t.Run("row4 a contract element whose Represents field marks an open term", func(t *testing.T) {
		t.Parallel()
		undecided(t, ro, "cart.order")
	})
	t.Run("row5 a glossary entry whose Definition field marks an open term", func(t *testing.T) {
		t.Parallel()
		undecided(t, gto, "cart.trolley")
	})
	t.Run("row6 a rule with ? in two cells of one Examples row", func(t *testing.T) {
		t.Parallel()
		undecided(t, o, "cart.x.bag", place{xPath, tr.line(xPath, "| apple | ? | ? |")})
	})
	t.Run("row7 a rule with a ? inside a literal", func(t *testing.T) {
		t.Parallel()
		undecided(t, lo, "cart.x.code")
	})
	t.Run("row8 a glossary entry with Contact: ?, whose Definition marks an open term", func(t *testing.T) {
		t.Parallel()
		undecided(t, lo, "cart.wagon", place{glossaryPath, lt.line(glossaryPath, "- Contact: ?")})
	})
	// The When step names elements listed as changed and as changed through
	// a dependency as well.
	t.Run("derived: a changed element", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		changed := diffTree().edit(featurePath, "Then      the cart holds <quantity> more of <product>", "Then      ?")
		r.set(changed)
		o := derivedIn(t, r.dir, "cart.add.limit:changed")
		undecided(t, o, "cart.add.limit", place{featurePath, changed.line(featurePath, "Then      ?")})
	})
	t.Run("derived: a rule changed through an open term", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		open := diffTree().edit(glossaryPath, "- Definition: A product placed in the cart.\n- Contact: Dana Reyes\n- Status: defined\n",
			"- Definition: ?\n- Contact: Dana Reyes\n- Status: open\n- Question: Is a service an item?\n")
		r.set(open)
		o := derivedIn(t, r.dir, "cart.item:changed", "cart.add.limit:changed-through", "cart.add.empty:changed-through")
		e := o.elem(t, "cart.add.empty")
		if e.Change != "changed-through" {
			t.Errorf("want cart.add.empty changed through a dependency, got %s", e.Change)
		}
		undecided(t, o, "cart.add.empty", place{featurePath, open.line(featurePath, "[items]")})
	})
}

func TestBaselinePartsSplit_Commands(t *testing.T) {
	t.Parallel()
	// Every case writes the rich tree with a manifest that records only the
	// heading of the element, with its fingerprint: every other part of it is
	// then listed as added, which needs no former text, and every other
	// element as added in full. The parts of the element are its heading and
	// the parts listed.
	tr := richTree().with(sysParamsPath, sysParamsMD)
	limitRows := []string{"row " + rowKey("quantity", "1", "product", "apple"), "row " + rowKey("quantity", "11", "product", "apple")}
	cases := []struct {
		name, id, heading string
		want              []string
	}{
		{"row1 a rule without a Rationale", "cart.add.limit", "Rule: The cart takes a limited quantity",
			append([]string{"step When", "step Provided", "step Then", "step Otherwise"}, limitRows...)},
		{"row2 a feature with rules", "cart.add", "Adding to the cart",
			[]string{"field Contact", "section Goal", "section Rationale"}},
		{"row3 a glossary entry", "cart.item", "item",
			[]string{"field Definition", "field Contact", "field Status", "field Aliases"}},
		{"row4 a parameter with a table", "system.shelf-layout", "shelf_layout",
			[]string{"field Configurable", "field Rationale", "body"}},
		{"row5 a contract element with members", "cart.cli.lint", "lint",
			[]string{"body"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			dir := tr.with(manifestPath, manifestFor([]mpart{{c.id, "heading", c.heading}})).write(t)
			want := withChange(without(treeIDs(tr), c.id), "added")
			o := derivedIn(t, dir, append(want, c.id+":changed")...)
			onlyParts(t, o.elem(t, c.id), withChange(c.want, "added")...)
		})
	}
}

func TestBaselinePartsFingerprint_Commands(t *testing.T) {
	t.Parallel()
	goal := "Buyers put goods into the cart."
	t.Run("row1 a section's body, and the same body wrapped", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, plainTree())
		r.set(plainTree().edit(featurePath, goal, "Buyers put goods\ninto the cart."))
		derivedIn(t, r.dir)
	})
	t.Run("row2 the padding after a keyword", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, baseline())
		r.set(baseline().edit(featurePath, "Then      the cart holds", "Then the cart holds"))
		derivedIn(t, r.dir)
	})
	t.Run("row3 *a* and _a_", func(t *testing.T) {
		t.Parallel()
		base := plainTree().edit(featurePath, goal, "Buyers put *some* goods into the cart.")
		r, _ := accepted(t, base)
		r.set(plainTree().edit(featurePath, goal, "Buyers put _some_ goods into the cart."))
		onlyParts(t, derivedIn(t, r.dir, "cart.add:changed").elem(t, "cart.add"), "section Goal:changed")
	})
	t.Run("row4 case counts", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, plainTree())
		r.set(plainTree().edit(glossaryPath, "- Contact: Bruno", "- Contact: bruno"))
		onlyParts(t, derivedIn(t, r.dir, "cart.item:changed").elem(t, "cart.item"), "field Contact:changed")
	})
}

func TestBaselinePartsRowKey_Commands(t *testing.T) {
	t.Parallel()
	head := "| id | provided | outcome | why |\n| --- | --- | --- | --- |\n"
	tail := "| b | no | nothing shown | |\n"
	rowA := "| a | yes | error | |"
	key := "row " + rowKey("id", "a")
	amendedIn := func(t *testing.T, from, to string) diffOutcome {
		t.Helper()
		r, _ := accepted(t, withRule(idBlock, from))
		r.set(withRule(idBlock, to))
		return derivedIn(t, r.dir, "cart.x.r:changed")
	}
	t.Run("row1 keyed by its placeholder values", func(t *testing.T) {
		t.Parallel()
		// The table holds the holding side without row a as well.
		c := "| c | yes | shown | |\n"
		e := amendedIn(t, head+c+tail, head+c+rowA+"\n"+tail).elem(t, "cart.x.r")
		onlyParts(t, e, key+":added")
		checkNow(t, part(t, e, key), rowA)
	})
	t.Run("row2 a new outcome keeps the key", func(t *testing.T) {
		t.Parallel()
		changed := "| a | yes | nothing reported | |"
		e := amendedIn(t, head+rowA+"\n"+tail, head+changed+"\n"+tail).elem(t, "cart.x.r")
		onlyParts(t, e, key+":changed")
		checkNow(t, part(t, e, key), changed)
		checkFormer(t, part(t, e, key), rowA)
	})
	t.Run("row3 the order of the columns does not change the key", func(t *testing.T) {
		t.Parallel()
		swapped := "| outcome | id | provided | why |\n| --- | --- | --- | --- |\n| error | a | yes | |\n| nothing shown | b | no | |\n"
		e := amendedIn(t, head+rowA+"\n"+tail, swapped).elem(t, "cart.x.r")
		onlyParts(t, e, key+":changed", "row "+rowKey("id", "b")+":changed")
	})
	t.Run("row4 a row below another with the same values", func(t *testing.T) {
		t.Parallel()
		first, second := "| ? | yes | shown | |", "| ? | no | nothing shown | |"
		e := amendedIn(t, head+first+"\n"+tail, head+first+"\n"+second+"\n"+tail).elem(t, "cart.x.r")
		q := "row " + rowKey("id", "?")
		onlyParts(t, e, q+" #2:added")
		checkNow(t, part(t, e, q+" #2"), second)
	})
}
