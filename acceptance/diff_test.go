package acceptance

// Feature baseline.diff: proviso/capabilities/baseline/features/diff.md, and
// the contracts system.cli.diff and system.cli.element-diff
// (proviso/contracts/cli.md).

import (
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
		for _, r := range []run{o.json, o.text} {
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
		for _, r := range []run{o.json, o.text} {
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
	t.Run("row7 a new alias only", func(t *testing.T) {
		t.Parallel()
		diffOf(t, diffTree(), itemAlias).only(t, "cart.item:changed")
	})
	t.Run("row8 a reference with a new Contact only", func(t *testing.T) {
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
}

// The undecided tree: diffTree with an open glossary entry, a rule that
// marks it, a rule with an undecided outcome and a rule with an undecided
// Examples cell.
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
	t.Run("row6 the manifest committed together with a tree it does not record", func(t *testing.T) {
		t.Parallel()
		r := newRepo(t, diffTree())
		r.manifest()
		r.set(diffTree().apply(limitProvided))
		r.commit("a1b2")
		r.set(diffTree().apply(limitProvided, limitThen))
		r.diff().stopped(t)
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
	// when none are, is checked for every diff (see diffIn).
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
		o := r.diff()
		if strings.TrimSpace(o.text.stdout) == "" {
			t.Errorf("want the element as text, got nothing")
		}
	})
	t.Run("row4 none, as text", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		if o := r.diff(); o.text.stdout != "" {
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

func TestBaselineDiffExit(t *testing.T) {
	t.Parallel()
	// The exit status is checked against the elements listed for every
	// diff (see diffIn); here each row's outcome is checked as well.
	exit := func(t *testing.T, o diffOutcome, want int) {
		t.Helper()
		if o.json.exit != want || o.text.exit != want {
			t.Errorf("want exit status %d, got %d with --json, %d without", want, o.json.exit, o.text.exit)
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
