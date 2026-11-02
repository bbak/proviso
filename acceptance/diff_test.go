package acceptance

// Feature baseline.diff: proviso/capabilities/baseline/features/diff.md, and
// the contracts system.cli.diff and system.cli.element-diff
// (proviso/contracts/cli.md).

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Edits of diffTree, as old and new text of one file.
type change struct{ path, old, new string }

func (tr tree) apply(cs ...change) tree {
	for _, c := range cs {
		tr = tr.edit(c.path, c.old, c.new)
	}
	return tr
}

var (
	limitThen       = change{featurePath, "Then      the cart holds <quantity> more of <product>", "Then      the cart holds <quantity> more pieces of <product>"}
	limitProvided   = change{featurePath, "Provided  <quantity> is at most ´max_quantity´", "Provided  <quantity> is below ´max_quantity´"}
	limitOutcome    = change{featurePath, "| 1 | apple | yes | added | |", "| 1 | apple | yes | put in the cart | |"}
	limitNewRow     = change{featurePath, "| 1 | apple | yes | added | |\n", "| 1 | apple | yes | added | |\n| 5 | pear | yes | added | |\n"}
	limitNoRatio    = change{featurePath, limitRationale, ""}
	itemDefinition  = change{glossaryPath, "- Definition: A product placed in the cart.", "- Definition: A product a buyer placed in the cart."}
	itemAlias       = change{glossaryPath, "- Aliases: items", "- Aliases: items, cart item"}
	itemContact     = change{glossaryPath, "- Contact: Dana Reyes\n- Status: defined\n- Aliases: items", "- Contact: Bruno Baketarić\n- Status: defined\n- Aliases: items"}
	xThen           = change{xPath, "Then      the shop shows <id>", "Then      the shop lists <id>"}
	breakDefinition = change{glossaryPath, "- Defined by: [CommonMark §4.1](../../references.md#system.ref.commonmark)", "- Definition: A line across the receipt."}
	breakDefinedBy  = change{glossaryPath, "- Defined by: [CommonMark §4.1](", "- Defined by: [CommonMark, section 4.1]("}
	commonmarkNew   = change{referencesPath, "- Version: 0.31.2", "- Version: 0.32"}
)

const extraRule = `
## <a id="cart.add.extra"></a> Rule: A cart takes one product at a time

´´´guarded
When      a buyer adds the products <products>
Then      the cart takes the first of <products>
  And     nothing else changes
´´´

| products | outcome |
| --- | --- |
| apple and pear | apple added |
`

// moved is diffTree with the rule cart.add.limit moved to the end of x.md.
func moved() tree {
	tr := diffTree()
	block := limitBlock()
	return tr.edit(featurePath, block, "").appendTo(xPath, "\n"+strings.TrimRight(block, "\n")+"\n")
}

var diffIDs = []string{"system", "system.ref.commonmark", "cart", "cart.item", "cart.thematic-break", "cart.max-quantity",
	"cart.add", "cart.add.limit", "cart.add.empty", "cart.x", "cart.x.r", "cart.x.line"}

func withChange(ids []string, c string) []string {
	var out []string
	for _, id := range ids {
		out = append(out, id+":"+c)
	}
	return out
}

// checkFull checks that e is listed in full: where it stands now in tr and
// its text from its heading to the next element's heading.
func checkFull(t *testing.T, e diffElement, tr tree, path string) {
	t.Helper()
	text, line := elementText(tr[path], e.ID)
	if e.File == nil || *e.File != path || e.Line == nil || *e.Line != line {
		t.Errorf("want %s at %s:%d, got %v:%v", e.ID, path, line, deref(e.File), derefInt(e.Line))
	}
	if e.Text == nil || !sameText(*e.Text, text) {
		t.Errorf("want the text of %s\n%s\ngot\n%v", e.ID, text, deref(e.Text))
	}
}

func deref(s *string) string {
	if s == nil {
		return "(none)"
	}
	return *s
}

func derefInt(n *int) string {
	if n == nil {
		return "(none)"
	}
	return strconv.Itoa(*n)
}

// checkRevision checks that the former text came from the commit want, or,
// if want is empty, that no former text was needed.
func checkRevision(t *testing.T, o diffOutcome, want string) {
	t.Helper()
	switch {
	case want == "" && o.ed.Revision != nil:
		t.Errorf("want revision null, as no former text was needed, got %q", *o.ed.Revision)
	case want != "" && (o.ed.Revision == nil || *o.ed.Revision != want):
		t.Errorf("want the revision %s, got %s", want, deref(o.ed.Revision))
	}
}

// checkFormer checks the former text of a part against text, up to
// whitespace; for a step, with or without its keyword.
func checkFormer(t *testing.T, p partChange, text string) {
	t.Helper()
	if p.Former == nil {
		t.Errorf("want the former text of %s, got none", p.Part)
		return
	}
	got, want := normWS(*p.Former), normWS(md(text))
	_, rest, _ := strings.Cut(want, " ")
	if got != want && !(strings.HasPrefix(p.Part, "step ") && got == rest) {
		t.Errorf("want the former text of %s to be %q, got %q", p.Part, want, got)
	}
}

func TestBaselineDiffBaseline(t *testing.T) {
	t.Parallel()
	t.Run("row1 with manifest.sum, compared with the manifest", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		o := r.diff()
		o.only(t)
		r.set(diffTree().appendTo(featurePath, extraRule))
		r.diff().only(t, "cart.add.extra:added")
	})
	t.Run("row2 without manifest.sum, every element added", func(t *testing.T) {
		t.Parallel()
		tr := diffTree()
		dir := tr.write(t)
		o := diffIn(t, dir, dir, filepath.Join(dir, "proviso"))
		o.only(t, withChange(diffIDs, "added")...)
		checkRevision(t, o, "")
		paths := map[string]string{"system": systemPath, "cart.add.limit": featurePath, "cart.item": glossaryPath, "system.ref.commonmark": referencesPath, "cart.x.line": xPath}
		for id, path := range paths {
			checkFull(t, o.elem(t, id), tr, path)
		}
	})
}

func TestBaselineDiffMalformed(t *testing.T) {
	t.Parallel()
	good := manifestFor(plainParts())
	lines := strings.Split(strings.TrimSuffix(good, "\n"), "\n")
	headingLine := 0
	for i, l := range lines {
		if strings.HasPrefix(l, "cart\theading\t") {
			headingLine = i
		}
	}
	replaced := func(line string) string {
		c := append([]string(nil), lines...)
		c[headingLine] = line
		return strings.Join(c, "\n") + "\n"
	}
	diffWith := func(t *testing.T, manifest string) diffOutcome {
		dir := plainTree().with(manifestPath, manifest).write(t)
		return diffIn(t, dir, dir, filepath.Join(dir, "proviso"))
	}
	// names checks that stderr names the line, by its text or its number.
	names := func(t *testing.T, o diffOutcome, line string, number int) {
		t.Helper()
		for _, r := range o.runs() {
			if !strings.Contains(r.stderr, line) && !strings.Contains(r.stderr, strconv.Itoa(number)) {
				t.Errorf("want stderr to name the line %q (line %d), got: %s", line, number, r.stderr)
			}
		}
	}

	t.Run("row1 an identifier, a part and a fingerprint, separated by tabs", func(t *testing.T) {
		t.Parallel()
		o := diffWith(t, good)
		o.only(t)
	})
	t.Run("row2 <<<<<<< HEAD", func(t *testing.T) {
		t.Parallel()
		c := append(append(append([]string(nil), lines[:2]...), "<<<<<<< HEAD"), lines[2:]...)
		o := diffWith(t, strings.Join(c, "\n")+"\n")
		o.stopped(t)
		names(t, o, "<<<<<<< HEAD", 3)
	})
	t.Run("row3 spaces for tabs", func(t *testing.T) {
		t.Parallel()
		line := "cart heading " + fp("Cart")
		o := diffWith(t, replaced(line))
		o.stopped(t)
		names(t, o, line, headingLine+1)
	})
	// The same rows for impact.
	impactWith := func(t *testing.T, manifest string) impactOutcome {
		dir := plainTree().with(manifestPath, manifest).write(t)
		return impactIn(t, dir, dir, filepath.Join(dir, "proviso"))
	}
	impactNames := func(t *testing.T, o impactOutcome, line string, number int) {
		t.Helper()
		impactStopped(t, o)
		for _, r := range o.runs() {
			if !strings.Contains(r.stderr, line) && !strings.Contains(r.stderr, strconv.Itoa(number)) {
				t.Errorf("want stderr to name the line %q (line %d), got: %s", line, number, r.stderr)
			}
		}
	}
	t.Run("impact/row1 an identifier, a part and a fingerprint, separated by tabs", func(t *testing.T) {
		t.Parallel()
		impactWith(t, good).only(t)
	})
	t.Run("impact/row2 <<<<<<< HEAD", func(t *testing.T) {
		t.Parallel()
		c := append(append(append([]string(nil), lines[:2]...), "<<<<<<< HEAD"), lines[2:]...)
		impactNames(t, impactWith(t, strings.Join(c, "\n")+"\n"), "<<<<<<< HEAD", 3)
	})
	t.Run("impact/row3 spaces for tabs", func(t *testing.T) {
		t.Parallel()
		line := "cart heading " + fp("Cart")
		impactNames(t, impactWith(t, replaced(line)), line, headingLine+1)
	})
	// The same rows for pack.
	packWith := func(t *testing.T, manifest string) packOutcome {
		dir := plainTree().with(manifestPath, manifest).write(t)
		return packIn(t, dir, dir, filepath.Join(dir, "proviso"))
	}
	packNames := func(t *testing.T, o packOutcome, line string, number int) {
		t.Helper()
		o.stopped(t)
		for _, r := range []run{o.json, o.text} {
			if !strings.Contains(r.stderr, line) && !strings.Contains(r.stderr, strconv.Itoa(number)) {
				t.Errorf("want stderr to name the line %q (line %d), got: %s", line, number, r.stderr)
			}
		}
	}
	t.Run("pack/row1 an identifier, a part and a fingerprint, separated by tabs", func(t *testing.T) {
		t.Parallel()
		o := packWith(t, good)
		o.headingsAre(t, "System")
	})
	t.Run("pack/row2 <<<<<<< HEAD", func(t *testing.T) {
		t.Parallel()
		c := append(append(append([]string(nil), lines[:2]...), "<<<<<<< HEAD"), lines[2:]...)
		packNames(t, packWith(t, strings.Join(c, "\n")+"\n"), "<<<<<<< HEAD", 3)
	})
	t.Run("pack/row3 spaces for tabs", func(t *testing.T) {
		t.Parallel()
		line := "cart heading " + fp("Cart")
		packNames(t, packWith(t, replaced(line)), line, headingLine+1)
	})
	// The same rows for judge and verify, on the plain tree with a rule
	// added, so that the element diff lists an element without needing
	// former text, and judge has an amendment to record a check of.
	// judgeIn checks that coherence_file stays as it was unless judge exits
	// with 0, and that nothing goes to stdout; verifyIn that nothing goes to
	// stdout with exit status 2.
	withAdded := func(t *testing.T, manifest string) string {
		return plainTree().appendTo(featurePath, extraRule).with(manifestPath, manifest).write(t)
	}
	namesLine := func(t *testing.T, r run, line string, number int) {
		t.Helper()
		if !strings.Contains(r.stderr, line) && !strings.Contains(r.stderr, strconv.Itoa(number)) {
			t.Errorf("want stderr to name the line %q (line %d), got: %s", line, number, r.stderr)
		}
	}
	judgeWith := func(t *testing.T, manifest string) run {
		dir := withAdded(t, manifest)
		return judgeIn(t, dir, dir, filepath.Join(dir, "proviso"), "{}")
	}
	judgeNames := func(t *testing.T, r run, line string, number int) {
		t.Helper()
		refused(t, r, 2)
		namesLine(t, r, line, number)
	}
	t.Run("judge/row1 an identifier, a part and a fingerprint, separated by tabs", func(t *testing.T) {
		t.Parallel()
		refused(t, judgeWith(t, good), 0)
	})
	t.Run("judge/row2 <<<<<<< HEAD", func(t *testing.T) {
		t.Parallel()
		c := append(append(append([]string(nil), lines[:2]...), "<<<<<<< HEAD"), lines[2:]...)
		judgeNames(t, judgeWith(t, strings.Join(c, "\n")+"\n"), "<<<<<<< HEAD", 3)
	})
	t.Run("judge/row3 spaces for tabs", func(t *testing.T) {
		t.Parallel()
		line := "cart heading " + fp("Cart")
		judgeNames(t, judgeWith(t, replaced(line)), line, headingLine+1)
	})
	verifyWith := func(t *testing.T, manifest string) verifyOutcome {
		dir := withAdded(t, manifest)
		return verifyIn(t, dir, dir, filepath.Join(dir, "proviso"))
	}
	verifyNames := func(t *testing.T, o verifyOutcome, line string, number int) {
		t.Helper()
		o.exits(t, 2)
		for _, r := range o.runs() {
			namesLine(t, r, line, number)
		}
	}
	t.Run("verify/row1 an identifier, a part and a fingerprint, separated by tabs", func(t *testing.T) {
		t.Parallel()
		// The rule added, and no check recorded.
		verifyWith(t, good).only(t, "missing")
	})
	t.Run("verify/row2 <<<<<<< HEAD", func(t *testing.T) {
		t.Parallel()
		c := append(append(append([]string(nil), lines[:2]...), "<<<<<<< HEAD"), lines[2:]...)
		verifyNames(t, verifyWith(t, strings.Join(c, "\n")+"\n"), "<<<<<<< HEAD", 3)
	})
	t.Run("verify/row3 spaces for tabs", func(t *testing.T) {
		t.Parallel()
		line := "cart heading " + fp("Cart")
		verifyNames(t, verifyWith(t, replaced(line)), line, headingLine+1)
	})
	derived := []struct{ name, line string }{
		{"derived: a short fingerprint", "cart\theading\th1:5b0d"},
		{"derived: uppercase hexadecimal", "cart\theading\t" + strings.ToUpper(fp("Cart"))},
		{"derived: a fingerprint without h1:", "cart\theading\t" + strings.TrimPrefix(fp("Cart"), "h1:")},
		{"derived: a part part_kinds does not write", "cart\tchapter Goal\t" + fp("Cart")},
		{"derived: two fields", "cart\t" + fp("Cart")},
		{"derived: four fields", "cart\theading\t" + fp("Cart") + "\textra"},
	}
	for _, c := range derived {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			o := diffWith(t, replaced(c.line))
			o.stopped(t)
		})
	}
}

func TestBaselineDiffAdded(t *testing.T) {
	t.Parallel()
	t.Run("row1 new since the manifest was written", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		tr := diffTree().appendTo(featurePath, extraRule)
		r.set(tr)
		o := r.diff()
		o.only(t, "cart.add.extra:added")
		checkFull(t, o.elem(t, "cart.add.extra"), tr, featurePath)
		checkRevision(t, o, "")
	})
	t.Run("row2 unchanged, handed over", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		r.set(diffTree().appendTo(featurePath, extraRule))
		r.diff().lacksElem(t, "cart.add.limit")
	})
	t.Run("row3 moved into another feature's file, handed over", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		r.set(moved())
		r.diff().only(t)
	})
	t.Run("derived: added needs no former text, and no git", func(t *testing.T) {
		t.Parallel()
		dir := diffTree().write(t)
		manifestOK(t, dir)
		skipIfGitAbove(t, dir)
		diffTree().appendTo(featurePath, extraRule).writeTo(t, dir, false)
		o := diffIn(t, dir, dir, filepath.Join(dir, "proviso"))
		o.only(t, "cart.add.extra:added")
		checkRevision(t, o, "")
	})
}

func TestBaselineDiffChanged(t *testing.T) {
	t.Parallel()
	// amended accepts diffTree, applies cs and diffs.
	amended := func(t *testing.T, cs ...change) (diffOutcome, tree, string) {
		t.Helper()
		r, base := accepted(t, diffTree())
		tr := diffTree().apply(cs...)
		r.set(tr)
		return r.diff(), tr, base
	}

	t.Run("row1 its Then step reworded", func(t *testing.T) {
		t.Parallel()
		o, tr, base := amended(t, limitThen)
		o.only(t, "cart.add.limit:changed")
		e := o.elem(t, "cart.add.limit")
		onlyParts(t, e, "step Then:changed")
		checkFormer(t, part(t, e, "step Then"), "Then      the cart holds <quantity> more of <product>\n  And     nothing else changes")
		checkFull(t, e, tr, featurePath)
		checkRevision(t, o, base)
	})
	t.Run("row2 its Rationale deleted", func(t *testing.T) {
		t.Parallel()
		o, _, base := amended(t, limitNoRatio)
		o.only(t, "cart.add.limit:changed")
		e := o.elem(t, "cart.add.limit")
		onlyParts(t, e, "section Rationale:removed")
		checkFormer(t, part(t, e, "section Rationale"), "A shelf holds ten pieces.")
		checkRevision(t, o, base)
	})
	t.Run("row3 a new Examples row", func(t *testing.T) {
		t.Parallel()
		o, _, _ := amended(t, limitNewRow)
		o.only(t, "cart.add.limit:changed")
		onlyParts(t, o.elem(t, "cart.add.limit"), "row "+rowKey("quantity", "5", "product", "pear")+":added")
	})
	t.Run("row4 an Examples row with a new outcome", func(t *testing.T) {
		t.Parallel()
		o, _, _ := amended(t, limitOutcome)
		o.only(t, "cart.add.limit:changed")
		e := o.elem(t, "cart.add.limit")
		key := "row " + rowKey("quantity", "1", "product", "apple")
		onlyParts(t, e, key+":changed")
		checkFormer(t, part(t, e, key), "| 1 | apple | yes | added | |")
	})
	t.Run("row5 moved into another feature's file", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		r.set(moved())
		r.diff().lacksElem(t, "cart.add.limit")
	})
	t.Run("row6 its text wrapped at other places", func(t *testing.T) {
		t.Parallel()
		o, _, _ := amended(t,
			change{featurePath, "A cart that takes anything would empty the shelves.", "A cart that takes anything\nwould empty   the shelves."},
			change{featurePath, "A shelf holds ten pieces.", "A shelf holds\nten pieces."},
			change{featurePath, "Then      the cart holds <quantity> more of <product>", "Then the cart holds <quantity> more of <product>"},
		)
		o.only(t)
	})
	t.Run("row7 one of a feature's rules changed", func(t *testing.T) {
		t.Parallel()
		o, _, _ := amended(t, limitThen)
		o.lacksElem(t, "cart.add")
	})
	t.Run("derived: several parts, and the heading changed", func(t *testing.T) {
		t.Parallel()
		o, _, _ := amended(t, limitThen, limitProvided, limitOutcome,
			change{xPath, "Rule: Under test", "Rule: Under a new test"},
		)
		o.only(t, "cart.add.limit:changed", "cart.x.r:changed")
		onlyParts(t, o.elem(t, "cart.add.limit"), "step Then:changed", "step Provided:changed", "row "+rowKey("quantity", "1", "product", "apple")+":changed")
		e := o.elem(t, "cart.x.r")
		onlyParts(t, e, "heading:changed")
		checkFormer(t, part(t, e, "heading"), "Rule: Under test")
	})
	t.Run("derived: a field added", func(t *testing.T) {
		t.Parallel()
		o, _, _ := amended(t, change{glossaryPath, "#system.ref.commonmark)\n- Contact: Dana Reyes\n- Status: defined\n", "#system.ref.commonmark)\n- Contact: Dana Reyes\n- Status: defined\n- Aliases: thematic breaks\n"})
		o.only(t, "cart.thematic-break:changed")
		onlyParts(t, o.elem(t, "cart.thematic-break"), "field Aliases:added")
	})
	t.Run("derived: only an added part needs no former text, and no git", func(t *testing.T) {
		t.Parallel()
		dir := diffTree().write(t)
		manifestOK(t, dir)
		skipIfGitAbove(t, dir)
		diffTree().apply(limitNewRow).writeTo(t, dir, false)
		o := diffIn(t, dir, dir, filepath.Join(dir, "proviso"))
		o.only(t, "cart.add.limit:changed")
		checkRevision(t, o, "")
	})
}

func TestBaselineDiffRemoved(t *testing.T) {
	t.Parallel()
	emptyText, _ := elementText(diffTree()[featurePath], "cart.add.empty")

	t.Run("row1 still in the tree, compared", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		r.set(diffTree().edit(featurePath, emptyText, ""))
		o := r.diff()
		o.lacksElem(t, "cart.add.limit")
	})
	t.Run("row2 its rule deleted", func(t *testing.T) {
		t.Parallel()
		r, base := accepted(t, diffTree())
		r.set(diffTree().edit(featurePath, emptyText, ""))
		o := r.diff()
		o.only(t, "cart.add.empty:removed")
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
		r.diff().only(t)
	})
	t.Run("derived: a whole feature file deleted", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		r.set(diffTree().without(xPath))
		r.diff().only(t, "cart.x:removed", "cart.x.r:removed", "cart.x.line:removed")
	})
}

// handTree is diffTree with a rule in x.md that hands over to
// cart.add.limit.
func handTree() tree {
	return diffTree().appendTo(xPath, "\n"+ruleMD("cart.x.hand", "A spoken quantity",
		"When      a buyer says the quantity <quantity> for the product <product>\nProvided  the shop hears <quantity>\nThen      the cart holds <quantity> more of <product>\n  And     nothing else changes\nOtherwise → cart.add.limit",
		"| quantity | product | provided | outcome |\n| --- | --- | --- | --- |\n| 1 | apple | yes | added |\n| 2 | apple | no | handed over |"))
}

// handItemTree is handTree with the rule cart.x.hand marking [item] as
// well, so that it depends on cart.item directly and through cart.add.limit.
func handItemTree() tree {
	return handTree().edit(xPath, "for the product <product>", "for the [item] <product>")
}

// Amendments of the impact tree that remove the member cart.api.line.count,
// while cart.api.line stays, and edit the rule referring to it to refer to
// cart.api.line instead.
var (
	countRemoved = removedFrom(impactTree(), apiPath, "cart.api.line.count")
	linesToLine  = change{payPath, "Then      the shop shows ´line.count´ for each line of <order>", "Then      the shop shows each ´line´ of <order>"}
	amountAdded  = change{apiPath, "How many pieces the line holds.\n", "How many pieces the line holds.\n\n### <a id=\"cart.api.line.amount\"></a> amount\n\nWhat the line costs.\n"}
)

// checkThrough asserts that e is changed through a dependency, along exactly
// the chains given, each written as identifiers joined by " > ", in any
// order.
func checkThrough(t *testing.T, e diffElement, chains ...string) {
	t.Helper()
	if e.Change != "changed-through" {
		t.Errorf("want %s changed-through, got %s", e.ID, e.Change)
	}
	var got []string
	for _, c := range e.Through {
		got = append(got, strings.Join(c, " > "))
	}
	if strings.Join(sorted(got), " | ") != strings.Join(sorted(chains), " | ") {
		t.Errorf("want %s changed through %q, got %q", e.ID, chains, got)
	}
}

func TestBaselineDiffDependency(t *testing.T) {
	t.Parallel()
	diffOf := func(t *testing.T, base tree, cs ...change) diffOutcome {
		t.Helper()
		r, _ := accepted(t, base)
		r.set(amended(base, cs...))
		return r.diff()
	}
	t.Run("row1 a glossary entry's Definition reworded", func(t *testing.T) {
		t.Parallel()
		o := diffOf(t, diffTree(), itemDefinition)
		o.only(t, "cart.item:changed", "cart.add.limit:changed-through", "cart.add.empty:changed-through")
		checkThrough(t, o.elem(t, "cart.add.limit"), "cart.item")
		checkThrough(t, o.elem(t, "cart.add.empty"), "cart.item")
	})
	t.Run("row2 a reference with a new Version", func(t *testing.T) {
		t.Parallel()
		o := diffOf(t, diffTree(), change{referencesPath, "- Version: 0.31.2", "- Version: 0.32"})
		o.only(t, "system.ref.commonmark:changed", "cart.thematic-break:changed-through", "cart.x.line:changed-through")
		checkThrough(t, o.elem(t, "cart.thematic-break"), "system.ref.commonmark")
		checkThrough(t, o.elem(t, "cart.x.line"), "system.ref.commonmark > cart.thematic-break")
	})
	t.Run("row3 a member of a contract element with a new value", func(t *testing.T) {
		t.Parallel()
		o := diffOf(t, impactTree(), iCountBody)
		checkThrough(t, o.elem(t, "cart.pay.lines"), "cart.api.line.count")
	})
	t.Run("row4 a member with a new description, reaching the rules referring to the element holding it", func(t *testing.T) {
		t.Parallel()
		o := diffOf(t, impactTree(), iCountBody)
		o.only(t, "cart.api.line.count:changed", "cart.pay.lines:changed-through", "cart.pay.order:changed-through")
		checkThrough(t, o.elem(t, "cart.pay.order"), "cart.api.line.count")
	})
	t.Run("row5 a rule's Then step changed, reaching the rules handing over to it", func(t *testing.T) {
		t.Parallel()
		o := diffOf(t, handTree(), limitThen)
		o.only(t, "cart.add.limit:changed", "cart.x.hand:changed-through")
		checkThrough(t, o.elem(t, "cart.x.hand"), "cart.add.limit")
	})
	t.Run("row6 a rule changed in its Rationale only, marking a term whose Definition changed", func(t *testing.T) {
		t.Parallel()
		o := diffOf(t, handTree(), itemDefinition, limitNoRatio)
		o.only(t, "cart.item:changed", "cart.add.limit:changed", "cart.add.empty:changed-through", "cart.x.hand:changed-through")
		checkThrough(t, o.elem(t, "cart.x.hand"), "cart.item > cart.add.limit")
	})
	t.Run("row7 a reference with a new Version, while the term it defines has its Defined by changed too", func(t *testing.T) {
		t.Parallel()
		// lint.step: cart.thematic-break, marked by cart.x.line.
		o := diffOf(t, diffTree(), commonmarkNew, breakDefinedBy)
		o.only(t, "system.ref.commonmark:changed", "cart.thematic-break:changed", "cart.x.line:changed-through")
		checkThrough(t, o.elem(t, "cart.x.line"), "cart.thematic-break", "system.ref.commonmark > cart.thematic-break")
	})
	t.Run("row8 a member removed, while the element holding it stays", func(t *testing.T) {
		t.Parallel()
		// lint.findings.anchor-misplaced: cart.api.line.count, of
		// cart.api.line; cart.pay.lines, which referred to it, now refers to
		// cart.api.line and is changed itself.
		o := diffOf(t, impactTree(), countRemoved, linesToLine)
		o.only(t, "cart.api.line.count:removed", "cart.pay.lines:changed", "cart.pay.order:changed-through")
		checkThrough(t, o.elem(t, "cart.pay.order"), "cart.api.line.count")
	})
	t.Run("row9 a new alias only", func(t *testing.T) {
		t.Parallel()
		diffOf(t, diffTree(), itemAlias).only(t, "cart.item:changed")
	})
	t.Run("row10 a reference with a new Contact only", func(t *testing.T) {
		t.Parallel()
		diffOf(t, diffTree(), change{referencesPath, "- Contact: Dana Reyes", "- Contact: Bruno Baketarić"}).only(t, "system.ref.commonmark:changed")
	})
	t.Run("derived: a reference with a new Link", func(t *testing.T) {
		t.Parallel()
		o := diffOf(t, diffTree(), change{referencesPath, "- Link: https://spec.commonmark.org/0.31.2/", "- Link: https://spec.commonmark.org/current/"})
		o.only(t, "system.ref.commonmark:changed", "cart.thematic-break:changed-through", "cart.x.line:changed-through")
	})
	t.Run("derived: a new Contact of a glossary entry", func(t *testing.T) {
		t.Parallel()
		diffOf(t, diffTree(), itemContact).only(t, "cart.item:changed")
	})
	t.Run("derived: a rule depending on a changed meaning and changed itself", func(t *testing.T) {
		t.Parallel()
		o := diffOf(t, diffTree(), itemDefinition, limitThen)
		o.only(t, "cart.item:changed", "cart.add.limit:changed", "cart.add.empty:changed-through")
	})
	t.Run("derived: a parameter with a new value", func(t *testing.T) {
		t.Parallel()
		o := diffOf(t, impactTree(), iParamValue)
		o.only(t, "cart.max-quantity:changed", "cart.add.limit:changed-through", "cart.pay.cap:changed-through")
		checkThrough(t, o.elem(t, "cart.pay.cap"), "cart.max-quantity")
	})
	t.Run("derived: a contract element passes on the term it represents", func(t *testing.T) {
		t.Parallel()
		o := diffOf(t, impactTree(), iItemDef)
		o.only(t, "cart.item:changed", "cart.api.line:changed-through", "cart.add.limit:changed-through",
			"cart.pay.total:changed-through", "cart.pay.order:changed-through")
		checkThrough(t, o.elem(t, "cart.api.line"), "cart.item")
		checkThrough(t, o.elem(t, "cart.pay.order"), "cart.item > cart.api.line")
	})
	t.Run("derived: every chain that reaches an element", func(t *testing.T) {
		t.Parallel()
		o := diffOf(t, impactTree(), iItemDef, iParamValue)
		checkThrough(t, o.elem(t, "cart.add.limit"), "cart.item", "cart.max-quantity")
	})
	t.Run("derived: a chain through a hand-over", func(t *testing.T) {
		t.Parallel()
		o := diffOf(t, handTree(), itemDefinition)
		checkThrough(t, o.elem(t, "cart.x.hand"), "cart.item > cart.add.limit")
	})
	t.Run("derived: one rule reached directly and through a hand-over, listed once", func(t *testing.T) {
		t.Parallel()
		o := diffOf(t, handItemTree(), itemDefinition)
		o.only(t, "cart.item:changed", "cart.add.limit:changed-through", "cart.add.empty:changed-through", "cart.x.hand:changed-through")
		checkThrough(t, o.elem(t, "cart.x.hand"), "cart.item", "cart.item > cart.add.limit")
	})
	t.Run("derived: a member added, reaching the rules referring to the element holding it", func(t *testing.T) {
		t.Parallel()
		o := diffOf(t, impactTree(), amountAdded)
		o.only(t, "cart.api.line.amount:added", "cart.pay.order:changed-through")
		checkThrough(t, o.elem(t, "cart.pay.order"), "cart.api.line.amount")
	})
	t.Run("derived: a removed member's change passes on beyond the rules referring to its holder", func(t *testing.T) {
		t.Parallel()
		// cart.pay.ask hands over to cart.pay.order.
		tr := impactTree().appendTo(payPath, "\n"+ruleMD("cart.pay.ask", "A buyer asks for an order",
			"When      a buyer asks for the order of the products <products>\nProvided  the shop holds <products>\nThen      the shop shows the order\n  And     nothing else changes\nOtherwise → cart.pay.order",
			"| products | provided | outcome |\n| --- | --- | --- |\n| apple | yes | shown |\n| pear | no | handed over |"))
		o := diffOf(t, tr, countRemoved, linesToLine)
		o.only(t, "cart.api.line.count:removed", "cart.pay.lines:changed", "cart.pay.order:changed-through", "cart.pay.ask:changed-through")
		checkThrough(t, o.elem(t, "cart.pay.ask"), "cart.api.line.count > cart.pay.order")
	})
}

// The undecided tree: diffTree with an open glossary entry, a rule that
// marks it, a rule with an undecided outcome, a rule with an undecided
// Examples cell, a rule with two undecided cells in one Examples row, a rule
// undecided in its When and its Then step, a rule marking the open term in
// its When step and undecided in its Then step, and a rule marking the open
// term twice in one line.
func undecidedTree() tree {
	return diffTree().appendTo(glossaryPath, `
## <a id="cart.basket"></a> basket

- Definition: ?
- Contact: Dana Reyes
- Status: open
- Question: Is a basket a cart?
`).appendTo(xPath, `
## <a id="cart.x.basket"></a> Rule: A basket holds goods

´´´guarded
When      a buyer fills the [basket] <basket>
Then      <basket> holds goods
  And     nothing else changes
´´´

| basket | outcome |
| --- | --- |
| the first basket | full |

## <a id="cart.x.weigh"></a> Rule: The shop weighs goods

´´´guarded
When      a buyer weighs the product <weight>
Then      ?
  And     nothing else changes
´´´

| weight | outcome |
| --- | --- |
| one kilogram | weighed |

## <a id="cart.x.price"></a> Rule: The shop prices goods

´´´guarded
When      a buyer asks for the price of <product>
Then      the shop names the price of <product>
  And     nothing else changes
´´´

| product | outcome |
| --- | --- |
| apple | ? |

## <a id="cart.x.bag"></a> Rule: The shop bags goods

´´´guarded
When      a buyer bags the product <product> in the bag <bag>
Then      <bag> holds <product>
  And     nothing else changes
´´´

| product | bag | outcome |
| --- | --- | --- |
| apple | ? | ? |

## <a id="cart.x.scan"></a> Rule: The shop scans goods

´´´guarded
When      a buyer scans the product <product> with ?
Then      ?
  And     nothing else changes
´´´

| product | outcome |
| --- | --- |
| apple | scanned |

## <a id="cart.x.carry"></a> Rule: A buyer carries a basket

´´´guarded
When      a buyer carries the [basket] <basket>
Then      ?
  And     nothing else changes
´´´

| basket | outcome |
| --- | --- |
| the first basket | carried |

## <a id="cart.x.swap"></a> Rule: A buyer swaps baskets

´´´guarded
When      a buyer swaps the [basket] <basket> for another [basket]
Then      <basket> is given back
  And     nothing else changes
´´´

| basket | outcome |
| --- | --- |
| the first basket | given back |
`)
}

func TestBaselineDiffUndecided(t *testing.T) {
	t.Parallel()
	tr := undecidedTree()
	dir := tr.write(t)
	o := diffIn(t, dir, dir, filepath.Join(dir, "proviso"))
	undecided := func(t *testing.T, id string, want ...place) {
		t.Helper()
		e := o.elem(t, id)
		if len(e.Undecided) != len(want) {
			t.Errorf("want %s to name the undecided places %v, got %v", id, want, e.Undecided)
			return
		}
		for i := range want {
			if e.Undecided[i] != want[i] {
				t.Errorf("want %s to name the undecided places %v, got %v", id, want, e.Undecided)
			}
		}
	}
	t.Run("row1 a rule with every step decided", func(t *testing.T) {
		t.Parallel()
		undecided(t, "cart.add.limit")
	})
	t.Run("row2 a rule with Then ?", func(t *testing.T) {
		t.Parallel()
		undecided(t, "cart.x.weigh", place{xPath, tr.line(xPath, "Then      ?")})
	})
	t.Run("row3 a rule marking an open term", func(t *testing.T) {
		t.Parallel()
		undecided(t, "cart.x.basket", place{xPath, tr.line(xPath, "[basket]")})
	})
	t.Run("row4 a rule with ? in two cells of one Examples row", func(t *testing.T) {
		t.Parallel()
		undecided(t, "cart.x.bag", place{xPath, tr.line(xPath, "| apple | ? | ? |")})
	})
	t.Run("derived: a rule marking an open term twice in one line", func(t *testing.T) {
		t.Parallel()
		undecided(t, "cart.x.swap", place{xPath, tr.line(xPath, "for another [basket]")})
	})
	t.Run("derived: an undecided Examples cell", func(t *testing.T) {
		t.Parallel()
		undecided(t, "cart.x.price", place{xPath, tr.line(xPath, "| apple | ? |")})
	})
	t.Run("derived: an open entry's undecided Definition", func(t *testing.T) {
		t.Parallel()
		undecided(t, "cart.basket", place{glossaryPath, tr.line(glossaryPath, "- Definition: ?")})
	})
	t.Run("derived: a changed element", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		changed := diffTree().edit(featurePath, "Then      the cart holds <quantity> more of <product>", "Then      ?")
		r.set(changed)
		o := r.diff()
		o.only(t, "cart.add.limit:changed")
		e := o.elem(t, "cart.add.limit")
		want := place{featurePath, changed.line(featurePath, "Then      ?")}
		if len(e.Undecided) != 1 || e.Undecided[0] != want {
			t.Errorf("want cart.add.limit to name the undecided place %v, got %v", want, e.Undecided)
		}
	})
	t.Run("derived: a rule changed through an open term", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		open := diffTree().edit(glossaryPath, "- Definition: A product placed in the cart.\n- Contact: Dana Reyes\n- Status: defined\n",
			"- Definition: ?\n- Contact: Dana Reyes\n- Status: open\n- Question: Is a service an item?\n")
		r.set(open)
		o := r.diff()
		e := o.elem(t, "cart.add.empty")
		want := place{featurePath, open.line(featurePath, "[items]")}
		if e.Change != "changed-through" || len(e.Undecided) != 1 || e.Undecided[0] != want {
			t.Errorf("want cart.add.empty changed through a dependency, naming the mark %v, got %s %v", want, e.Change, e.Undecided)
		}
	})
}

func TestBaselineDiffFormer(t *testing.T) {
	t.Parallel()
	formerThen := "Then      the cart holds <quantity> more of <product>\n  And     nothing else changes"

	t.Run("row1 the manifest committed, the tree edited since", func(t *testing.T) {
		t.Parallel()
		r, base := accepted(t, diffTree())
		r.set(diffTree().apply(limitThen))
		o := r.diff()
		checkRevision(t, o, base)
		checkFormer(t, part(t, o.elem(t, "cart.add.limit"), "step Then"), formerThen)
	})
	t.Run("derived: the edit committed since", func(t *testing.T) {
		t.Parallel()
		r, base := accepted(t, diffTree())
		r.set(diffTree().apply(limitThen))
		r.commit("amend")
		o := r.diff()
		checkRevision(t, o, base)
		checkFormer(t, part(t, o.elem(t, "cart.add.limit"), "step Then"), formerThen)
	})
	t.Run("row2 a merge that took the manifest from the mainline", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		r.git("checkout", "-q", "-b", "feature")
		feature := diffTree().apply(xThen)
		r.set(feature)
		r.commit("feature")
		r.git("checkout", "-q", "main")
		r.tr = diffTree()
		mainline := diffTree().apply(limitThen)
		r.set(mainline)
		r.manifest()
		e5f6 := r.commit("accept on the mainline")
		r.git("checkout", "-q", "feature")
		r.tr = feature
		r.git("merge", "-q", "--no-ff", "-m", "merge the mainline", "main")
		merged := diffTree().apply(xThen, limitThen)
		r.tr = merged
		r.set(merged.edit(featurePath, "more pieces of", "at most ten pieces of"))
		o := r.diff()
		o.only(t, "cart.add.limit:changed", "cart.x.r:changed")
		checkRevision(t, o, e5f6)
		checkFormer(t, part(t, o.elem(t, "cart.add.limit"), "step Then"),
			"Then      the cart holds <quantity> more pieces of <product>\n  And     nothing else changes")
	})
	t.Run("row3 the spec root in no git repository", func(t *testing.T) {
		t.Parallel()
		dir := diffTree().write(t)
		skipIfGitAbove(t, dir)
		manifestOK(t, dir)
		diffTree().apply(limitThen).writeTo(t, dir, false)
		diffIn(t, dir, dir, filepath.Join(dir, "proviso")).stopped(t)
	})
	t.Run("row4 the manifest written but never committed", func(t *testing.T) {
		t.Parallel()
		r := newRepo(t, diffTree())
		r.commit("tree")
		r.manifest()
		r.set(diffTree().apply(limitThen))
		r.diff().stopped(t)
	})
	t.Run("row5 the manifest written again since and not committed", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		r.set(diffTree().apply(limitThen))
		r.manifest()
		r.set(diffTree().apply(limitThen, limitProvided))
		r.diff().stopped(t)
	})
	t.Run("row6 the manifest committed together with a tree that differs from it in a part the element diff lists as changed", func(t *testing.T) {
		t.Parallel()
		r := newRepo(t, diffTree())
		r.manifest()
		r.set(diffTree().apply(limitProvided))
		r.commit("a1b2")
		r.set(diffTree().apply(limitProvided, limitThen))
		r.diff().stopped(t)
	})
	t.Run("row7 the manifest committed together with a tree that differs from it only in parts the element diff does not list as changed or removed", func(t *testing.T) {
		t.Parallel()
		// The tree in a1b2 has the Then step of cart.x.r changed, which the
		// tree now has as the manifest records it again.
		r := newRepo(t, diffTree())
		r.manifest()
		r.set(diffTree().apply(xThen))
		a1b2 := r.commit("a1b2")
		r.set(diffTree().apply(limitThen))
		o := r.diff()
		o.only(t, "cart.add.limit:changed")
		checkRevision(t, o, a1b2)
		checkFormer(t, part(t, o.elem(t, "cart.add.limit"), "step Then"), formerThen)
	})
	t.Run("derived: the tree in a1b2 differs in another part of an element listed as changed", func(t *testing.T) {
		t.Parallel()
		// Only the parts listed as changed or removed are checked, not
		// every part of the elements listed.
		r := newRepo(t, diffTree())
		r.manifest()
		r.set(diffTree().apply(limitProvided))
		a1b2 := r.commit("a1b2")
		r.set(diffTree().apply(limitThen))
		o := r.diff()
		o.only(t, "cart.add.limit:changed")
		onlyParts(t, o.elem(t, "cart.add.limit"), "step Then:changed")
		checkRevision(t, o, a1b2)
		checkFormer(t, part(t, o.elem(t, "cart.add.limit"), "step Then"), formerThen)
	})
	t.Run("derived: the tree in a1b2 holds an element listed as added", func(t *testing.T) {
		t.Parallel()
		r := newRepo(t, diffTree())
		r.manifest()
		withExtra := diffTree().appendTo(featurePath, extraRule)
		r.set(withExtra)
		a1b2 := r.commit("a1b2")
		r.set(withExtra.apply(limitThen))
		o := r.diff()
		o.only(t, "cart.add.extra:added", "cart.add.limit:changed")
		checkRevision(t, o, a1b2)
		checkFormer(t, part(t, o.elem(t, "cart.add.limit"), "step Then"), formerThen)
	})
	t.Run("derived: the tree in a1b2 differs in a part the element diff lists as removed", func(t *testing.T) {
		t.Parallel()
		r := newRepo(t, diffTree())
		r.manifest()
		r.set(diffTree().apply(change{featurePath, "A shelf holds ten pieces.", "A shelf holds ten pieces at most."}))
		r.commit("a1b2")
		r.set(diffTree().apply(limitNoRatio))
		r.diff().stopped(t)
	})
	t.Run("derived: an added part needs no former text, and no git", func(t *testing.T) {
		t.Parallel()
		dir := diffTree().write(t)
		skipIfGitAbove(t, dir)
		manifestOK(t, dir)
		diffTree().apply(limitNewRow).writeTo(t, dir, false)
		o := diffIn(t, dir, dir, filepath.Join(dir, "proviso"))
		o.only(t, "cart.add.limit:changed")
		onlyParts(t, o.elem(t, "cart.add.limit"), "row "+rowKey("quantity", "5", "product", "pear")+":added")
		checkRevision(t, o, "")
	})
	t.Run("derived: the spec root in a folder of the repository", func(t *testing.T) {
		t.Parallel()
		// The spec root is found upwards from where a command runs
		// (lint.run.root), so both run inside svc/.
		r := newRepo(t, diffTree().prefixed("svc"))
		svc := filepath.Join(r.dir, "svc")
		if m := manifestIn(t, r.dir, svc, filepath.Join(svc, "proviso"), nil); m.exit != 0 {
			t.Fatalf("want manifest to exit with 0, got %d\nstderr: %s", m.exit, m.stderr)
		}
		base := r.commit("accept")
		r.set(diffTree().apply(limitThen).prefixed("svc"))
		o := diffIn(t, r.dir, svc, filepath.Join(svc, "proviso"))
		o.only(t, "cart.add.limit:changed")
		e := o.elem(t, "cart.add.limit")
		if e.File == nil || *e.File != featurePath {
			t.Errorf("want the path %s, relative to the folder holding the spec root, got %s", featurePath, deref(e.File))
		}
		checkFormer(t, part(t, e, "step Then"), formerThen)
		checkRevision(t, o, base)
	})
}

func TestBaselineDiffOutput(t *testing.T) {
	t.Parallel()
	// That the text form is written when elements are listed, and nothing
	// when none are, is checked for every diff run as text (see runDiff).
	t.Run("row1 one changed, --json", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		r.set(diffTree().apply(limitThen))
		o := r.diff()
		if len(o.ed.Elements) != 1 {
			t.Errorf("want an element_diff with one element, got:\n%s", o.dump())
		}
	})
	t.Run("row2 none, --json", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		o := r.diff()
		if strings.TrimSpace(o.json.stdout) == "" || len(o.ed.Elements) != 0 {
			t.Errorf("want an element_diff with no elements, got %q", o.json.stdout)
		}
	})
	t.Run("row3 one changed, as text", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		r.set(diffTree().apply(limitThen))
		o := r.diffText()
		if strings.TrimSpace(o.text.stdout) == "" {
			t.Errorf("want the element as text, got nothing")
		}
	})
	t.Run("row4 none, as text", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		if o := r.diffText(); o.text.stdout != "" {
			t.Errorf("want nothing written, got %q", o.text.stdout)
		}
	})
}

func TestBaselineDiffOrder(t *testing.T) {
	t.Parallel()
	ids := func(o diffOutcome) string {
		var s []string
		for _, e := range o.ed.Elements {
			s = append(s, e.ID)
		}
		return strings.Join(s, ", ")
	}
	t.Run("row1 by identifier, byte by byte", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		r.set(diffTree().apply(xThen, limitThen, itemContact,
			change{systemPath, "The shop sells goods to buyers.", "The shop sells goods to its buyers."}))
		o := r.diff()
		if got, want := ids(o), "cart.add.limit, cart.item, cart.x.r, system"; got != want {
			t.Errorf("want %s, got %s", want, got)
		}
	})
	t.Run("row2 a feature before its rule", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		r.set(diffTree().apply(limitThen, change{featurePath, "Buyers put goods into the cart.", "Buyers put goods in the cart."}))
		o := r.diff()
		if got, want := ids(o), "cart.add, cart.add.limit"; got != want {
			t.Errorf("want %s, got %s", want, got)
		}
	})
	t.Run("row3 a rule changed through the chains system.ref.commonmark, lint.step and lint.step", func(t *testing.T) {
		t.Parallel()
		// lint.step: cart.thematic-break, marked by cart.x.line.
		r, _ := accepted(t, diffTree())
		r.set(diffTree().apply(commonmarkNew, breakDefinedBy))
		chainsAre(t, r.diff().elem(t, "cart.x.line"), "cart.thematic-break", "system.ref.commonmark > cart.thematic-break")
	})
	t.Run("row4 a rule changed through the chains lint.step, lint.keyword and lint.step", func(t *testing.T) {
		t.Parallel()
		// lint.step: cart.item; lint.keyword: cart.add.limit, which
		// cart.x.hand hands over to.
		r, _ := accepted(t, handItemTree())
		r.set(handItemTree().apply(itemDefinition))
		chainsAre(t, r.diff().elem(t, "cart.x.hand"), "cart.item", "cart.item > cart.add.limit")
	})
	// The undecided tree, every element added.
	utr := undecidedTree()
	udir := utr.write(t)
	uo := diffIn(t, udir, udir, filepath.Join(udir, "proviso"))
	t.Run("row5 a rule with ? in its Then step and in its When step", func(t *testing.T) {
		t.Parallel()
		linesAre(t, uo.elem(t, "cart.x.scan"),
			place{xPath, utr.line(xPath, "with ?")}, place{xPath, utr.lineN(xPath, "Then      ?", 2)})
	})
	t.Run("row6 a rule marking an open term in its When step, with ? in its Then step", func(t *testing.T) {
		t.Parallel()
		linesAre(t, uo.elem(t, "cart.x.carry"),
			place{xPath, utr.line(xPath, "carries the [basket]")}, place{xPath, utr.lineN(xPath, "Then      ?", 3)})
	})
	t.Run("derived: the parts in the order of their lines", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		r.set(diffTree().apply(limitThen, limitProvided, limitOutcome, limitNoRatio))
		e := r.diff().elem(t, "cart.add.limit")
		var got []string
		for _, p := range e.Parts {
			got = append(got, p.Part)
		}
		want := []string{"row " + rowKey("quantity", "1", "product", "apple"), "section Rationale", "step Provided", "step Then"}
		if strings.Join(got, "|") != strings.Join(want, "|") {
			t.Errorf("want the parts %q, got %q", want, got)
		}
	})
}

// chainsAre asserts that e names exactly the chains given, in that order,
// each written as identifiers joined by " > ".
func chainsAre(t *testing.T, e diffElement, chains ...string) {
	t.Helper()
	var got []string
	for _, c := range e.Through {
		got = append(got, strings.Join(c, " > "))
	}
	if strings.Join(got, " | ") != strings.Join(chains, " | ") {
		t.Errorf("want %s changed through %q, in that order, got %q", e.ID, chains, got)
	}
}

// linesAre asserts that e names exactly the lines not decided given, in that
// order.
func linesAre(t *testing.T, e diffElement, want ...place) {
	t.Helper()
	if fmt.Sprint(e.Undecided) != fmt.Sprint(want) {
		t.Errorf("want %s to name the lines %v, in that order, got %v", e.ID, want, e.Undecided)
	}
}

func TestBaselineDiffExit(t *testing.T) {
	t.Parallel()
	// The exit status is checked against the elements listed for every
	// diff (see diffIn); here each row's outcome is checked as well.
	exit := func(t *testing.T, o diffOutcome, want int) {
		t.Helper()
		if o.json.exit != want {
			t.Errorf("want exit status %d, got %d", want, o.json.exit)
		}
	}
	t.Run("row1 none", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		exit(t, r.diff(), 0)
	})
	t.Run("row2 one removed", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		text, _ := elementText(diffTree()[featurePath], "cart.add.empty")
		r.set(diffTree().edit(featurePath, text, ""))
		exit(t, r.diff(), 1)
	})
	t.Run("row3 a glossary entry with a new Definition, and one rule changed through it", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		// thematic break, marked by cart.x.line alone, gets a Definition
		// in place of its reference.
		r.set(diffTree().apply(breakDefinition))
		o := r.diff()
		o.only(t, "cart.thematic-break:changed", "cart.x.line:changed-through")
		exit(t, o, 1)
	})
}
