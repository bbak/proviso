package acceptance

// The capability query: proviso/capabilities/query/, the feature query.show,
// the invariant query.read-only, the term query.element-parent and the
// contracts system.cli.show, show_text and element_view. The Examples rows
// name elements of Proviso's own tree; showTree (show_driver_test.go) holds
// one that stands in for each. Every run goes through showIn, which checks
// query.read-only and the shape of element_view.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFixtureShow: the show tree follows the notation, so that a finding of
// show cannot come from a broken tree.
func TestFixtureShow(t *testing.T) {
	t.Parallel()
	check(t, showTree()).clean(t)
}

// The elements of the show tree as show takes them.
var (
	sLimit    = shown{"cart.add.limit", featurePath, "rule", "cart.add"}
	sSpecTree = shown{"system.spec-tree", showSysGlossaryPath, "glossary entry", "system"}
	sBillFind = shown{"billing.finding", billingGlossaryPath, "glossary entry", "billing"}
	sCartFind = shown{"cart.finding", glossaryPath, "glossary entry", "cart"}
	sSince    = shown{"system.cli.diff.since", showContractPath, "contract element", "system.cli.diff"}
	sDiff     = shown{"system.cli.diff", showContractPath, "contract element", "system"}
	sMaxQty   = shown{"cart.max-quantity", paramsPath, "parameter", "cart"}
	sSystem   = shown{"system", systemPath, "system", ""}
	sCart     = shown{"cart", capPath, "capability", "system"}
	sAdd      = shown{"cart.add", featurePath, "feature", "cart"}
	sItem     = shown{"cart.item", glossaryPath, "glossary entry", "cart"}
)

func TestQueryShowQuery(t *testing.T) {
	t.Parallel()
	tr := showTree()
	t.Run("row1 spec tree", func(t *testing.T) {
		t.Parallel()
		// One argument, as the shell passes "spec tree"; run as the row
		// gives it, without --json.
		o, _ := show(t, tr, "spec tree")
		if o.exit != 0 {
			t.Fatalf("query.show.query: want show to look for \"spec tree\" and find it, got exit status %d\nstderr: %s", o.exit, o.stderr)
		}
		if ids := showTextIDs(o.stdout); strings.Join(ids, " ") != "system.spec-tree" {
			t.Errorf("query.show.query: want system.spec-tree, got %v\n%s", ids, o.stdout)
		}
	})
	t.Run("row2 --json, lint.run.exit", func(t *testing.T) {
		t.Parallel()
		o, dir := show(t, tr, "--json", "cart.add.limit")
		o.elementsAre(t, tr, dir, sLimit)
	})
	t.Run("row3 lint.run.exit, --in, lint", func(t *testing.T) {
		t.Parallel()
		o, _ := show(t, tr, "cart.add.limit", "--in", "cart")
		if o.exit != 0 {
			t.Fatalf("query.show.query: want show to look for cart.add.limit and find it, got exit status %d\nstderr: %s", o.exit, o.stderr)
		}
		if ids := showTextIDs(o.stdout); strings.Join(ids, " ") != "cart.add.limit" {
			t.Errorf("query.show.query: want cart.add.limit, got %v\n%s", ids, o.stdout)
		}
	})
	for _, c := range []struct {
		name string
		args []string
	}{
		{"row4 none", nil},
		{"row5 --json", []string{"--json"}},
		{"row6 --in, lint", []string{"--in", "cart"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			o, _ := show(t, tr, c.args...)
			needsQuery(t, o)
		})
	}
	t.Run("derived: no query is refused before the spec root is looked for", func(t *testing.T) {
		t.Parallel()
		dir := anywhere(t)
		needsQuery(t, showIn(t, dir, dir))
	})
}

// needsQuery checks the refusal of query.show.query: exit status 2, nothing
// on stdout, and one line on stderr saying that "proviso help" shows the
// help page. That nothing was written is checked by showIn.
func needsQuery(t *testing.T, o showOutcome) {
	t.Helper()
	o.refused(t, "query.show.query", "proviso help")
	if lines := stderrLines(o.stderr); len(lines) != 1 {
		t.Errorf("query.show.query: want one line on stderr, got %d:\n%s", len(lines), o.stderr)
	}
}

// sameAsLint checks that show, refused with exit status 2, wrote to stderr
// what lint, run in the same directory, writes. The wording of lint's lines
// is for people, so what is compared is what they name: every file of the
// tree lint names, show names too.
func sameAsLint(t *testing.T, o showOutcome, l run) {
	t.Helper()
	o.refused(t, "query.show.tree")
	if l.exit != 2 {
		t.Fatalf("want lint to stop here too, got exit status %d", l.exit)
	}
	for _, line := range stderrLines(l.stderr) {
		if p := strings.TrimSpace(line); strings.HasPrefix(p, "proviso/") && !strings.Contains(o.stderr, p) {
			t.Errorf("query.show.tree: lint names %s, show does not:\n%s", p, o.stderr)
		}
	}
}

func TestQueryShowTree(t *testing.T) {
	t.Parallel()
	tr := showTree()
	t.Run("row1 its tree without findings", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, tr)
		o := showIn(t, r.dir, r.dir, "--json", "cart.add.limit")
		o.elementsAre(t, tr, r.dir, sLimit)
	})
	t.Run("row2 its tree with an identifier-duplicate error", func(t *testing.T) {
		t.Parallel()
		dup := tr.with(xPath, xHead+ruleMD("cart.add.limit", "The cart takes a limited quantity", idBlock, idTable))
		dir := dup.write(t)
		lintIn(t, dir, dir, nil).has(t, "identifier-duplicate", "")
		o := showIn(t, dir, dir, "--json", "system.spec-tree")
		o.elementsAre(t, dup, dir, sSpecTree)
		// Nothing of what lint would write: no finding, on either stream.
		if o.stderr != "" {
			t.Errorf("query.show.tree: nothing of lint's goes to stderr, got %q", o.stderr)
		}
	})
	t.Run("row3 inside the tree", func(t *testing.T) {
		t.Parallel()
		dir := tr.write(t)
		o := showIn(t, dir, filepath.Join(dir, "proviso", "capabilities", "cart"), "--json", "cart.add.limit")
		// The paths are below the directory holding the spec root,
		// wherever show was run.
		o.elementsAre(t, tr, dir, sLimit)
	})
	t.Run("row4 its tree without manifest.sum", func(t *testing.T) {
		t.Parallel()
		o, dir := show(t, tr, "--json", "cart.add.limit")
		o.elementsAre(t, tr, dir, sLimit)
	})
	t.Run("row5 other/, with no spec root up to the file system's root", func(t *testing.T) {
		t.Parallel()
		dir := anywhere(t)
		skipIfSpecRootAbove(t, dir)
		o := showIn(t, dir, dir, "--json", "system")
		sameAsLint(t, o, execLint(t, dir, "--json"))
	})
	t.Run("row6 its glossary.md without read permission", func(t *testing.T) {
		t.Parallel()
		dir := tr.write(t)
		path := filepath.Join(dir, filepath.FromSlash(glossaryPath))
		undo, err := makeUnreadable(path)
		if err != nil {
			t.Skipf("cannot make a file unreadable here: %v", err)
		}
		defer undo()
		o := showIn(t, dir, dir, "--json", "system")
		sameAsLint(t, o, execLint(t, dir, "--json"))
		o.refused(t, "query.show.tree", "glossary.md")
	})
	t.Run("row7 its tree holding x.md, a symbolic link to another feature", func(t *testing.T) {
		t.Parallel()
		dir := tr.write(t)
		link := filepath.Join(dir, filepath.FromSlash(xPath))
		if err := os.Symlink(filepath.Join(dir, filepath.FromSlash(featurePath)), link); err != nil {
			t.Skipf("cannot create a symbolic link here: %v", err)
		}
		o := showIn(t, dir, dir, "--json", "system")
		sameAsLint(t, o, execLint(t, dir, "--json"))
		o.refused(t, "query.show.tree", "x.md")
	})
	t.Run("row8 its manifest.sum starting with model 1", func(t *testing.T) {
		t.Parallel()
		dir := tr.with("proviso/manifest.sum", "model 1\n").write(t)
		o := showIn(t, dir, dir, "--json", "system")
		sameAsLint(t, o, execLint(t, dir, "--json"))
	})
}

func TestQueryShowIn(t *testing.T) {
	t.Parallel()
	tr := showTree()
	t.Run("row1 lint", func(t *testing.T) {
		t.Parallel()
		dir := tr.write(t)
		// The cart's finding, not billing's; and the system's elements.
		showIn(t, dir, dir, "finding", "--in", "cart", "--json").elementsAre(t, tr, dir, sCartFind)
		showIn(t, dir, dir, "spec tree", "--in", "cart", "--json").elementsAre(t, tr, dir, sSpecTree)
	})
	t.Run("row2 new, a folder holding features/a.md but no capability.md", func(t *testing.T) {
		t.Parallel()
		nw := tr.with("proviso/capabilities/new/features/a.md", "# <a id=\"new.a\"></a> A new feature\n"+`
- Contact: Dana Reyes

## Goal

Something new.

## Rationale

It is needed.
`)
		dir := nw.write(t)
		// The parent of a feature of a capability without capability.md
		// is not decided: see the report.
		showIn(t, dir, dir, "new.a", "--in", "new", "--json").elementsAre(t, nw, dir, shown{"new.a", "proviso/capabilities/new/features/a.md", "feature", "*"})
		showIn(t, dir, dir, "spec tree", "--in", "new", "--json").elementsAre(t, nw, dir, sSpecTree)
		showIn(t, dir, dir, "finding", "--in", "new", "--json").none(t)
	})
	for _, c := range []struct {
		name, capability string
		tr               tree
	}{
		{"row3 nosuch", "nosuch", tr},
		{"row4 Lint", "Cart", tr},
		{"row5 .old, a hidden folder holding capability.md", ".old", tr.with("proviso/capabilities/.old/capability.md", `# <a id="old"></a> Old

## Goal

Old things.

## Rationale

Kept for a while.
`)},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			// The query names an element of the system level, so that
			// looking in the whole tree instead would show.
			o, _ := show(t, c.tr, "spec tree", "--in", c.capability, "--json")
			o.refused(t, "query.show.in", c.capability)
		})
	}
}

func TestQueryShowMatch(t *testing.T) {
	t.Parallel()
	tr := showTree()
	found := []struct {
		name string
		tr   tree
		args []string
		want []shown
	}{
		{"row1 lint.run.exit, the tree", tr, []string{"cart.add.limit"}, []shown{sLimit}},
		{"row2 spec tree, the tree", tr, []string{"spec tree"}, []shown{sSpecTree}},
		{"row3 tree, the tree", tr, []string{"tree"}, []shown{sSpecTree}},
		{"row4 Spec  Tree, with two spaces, the tree", tr, []string{"Spec  Tree"}, []shown{sSpecTree}},
		// coherence.finding: billing.finding; lint.finding: cart.finding.
		{"row5 finding, the tree", tr, []string{"finding"}, []shown{sBillFind, sCartFind}},
		{"row6 finding, lint's and the system's", tr, []string{"finding", "--in", "cart"}, []shown{sCartFind}},
		{"row7 diff.since, the tree", tr, []string{"diff.since"}, []shown{sSince}},
		// tree_layout: max_quantity.
		{"row8 tree_layout, the tree", tr, []string{"max_quantity"}, []shown{sMaxQty}},
		// lint.run.twice: cart.add.limit, again in x.md.
		{"row9 lint.run.twice, the identifier of two rules", tr.with(xPath, xHead+ruleMD("cart.add.limit", "The cart takes a limited quantity", idBlock, idTable)),
			[]string{"cart.add.limit"}, []shown{sLimit, {"cart.add.limit", xPath, "rule", "cart.x"}}},
		// lint.run.odd: cart.x.odd.
		{"row10 lint.run.odd, a level-2 heading with an anchor in a feature file, neither Rule: nor Invariant:", tr.with(xPath, xHead+"## <a id=\"cart.x.odd\"></a> Odd notes\n\nSome notes.\n"),
			[]string{"cart.x.odd"}, []shown{{"cart.x.odd", xPath, "", "cart.x"}}},
		{"row11 system, the tree", tr, []string{"system"}, []shown{sSystem}},
		{"derived: a contract element by its name", tr, []string{"diff"}, []shown{sDiff}},
		{"derived: a member by its identifier", tr, []string{"system.cli.diff.since"}, []shown{sSince}},
		{"derived: a glossary entry by its identifier", tr, []string{"cart.item"}, []shown{sItem}},
		{"derived: a glossary entry by its plural alias", tr, []string{"items"}, []shown{sItem}},
		{"derived: an element matched by its name and an alias is taken once", tr.edit(glossaryPath, "- Aliases: items", "- Aliases: items, Item"),
			[]string{"item"}, []shown{sItem}},
		{"derived: a system-level entry after those of the capabilities, by path", tr.appendTo(showSysGlossaryPath, `
## <a id="system.finding"></a> finding

- Definition: Something the shop reports.
- Contact: Dana Reyes
- Status: defined
`), []string{"finding"}, []shown{sBillFind, sCartFind, {"system.finding", showSysGlossaryPath, "glossary entry", "system"}}},
		{"derived: two of one file in the order they stand", tr.edit(glossaryPath, "## <a id=\"cart.item\"></a> item", "## <a id=\"cart.finding-2\"></a> finding\n\n- Definition: Another.\n- Contact: Dana Reyes\n- Status: defined\n\n## <a id=\"cart.item\"></a> item"),
			[]string{"finding"}, []shown{sBillFind, {"cart.finding-2", glossaryPath, "glossary entry", "cart"}, sCartFind}},
	}
	for _, c := range found {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			o, dir := show(t, c.tr, append(c.args, "--json")...)
			if o.exit != 0 {
				t.Fatalf("query.show.match: want exit status 0, got %d\nstderr: %s", o.exit, o.stderr)
			}
			o.elementsAre(t, c.tr, dir, c.want...)
		})
	}

	// Compared as name_comparison says, step by step.
	terms := tr.appendTo(glossaryPath, "\n"+strings.Join([]string{
		termMD("cart.shelf-label", "shelf label"),
		termMD("cart.strasse", "Stra\u00dfe"),
		termMD("cart.kelvin", "kelvin"),
		termMD("cart.overview", "\u00dcbersicht"),
	}, "\n"))
	for _, c := range []struct {
		name, query, id string // id "" for none
	}{
		{"derived: name_comparison step1, white space at either end", "\u00a0shelf label ", "cart.shelf-label"},
		{"derived: name_comparison step2, a tab within", "shelf\tlabel", "cart.shelf-label"},
		{"derived: name_comparison step2, a non-breaking space joins no words", "shelf\u00a0label", ""},
		{"derived: name_comparison step3, the Kelvin sign", "\u212aelvin", "cart.kelvin"},
		{"derived: name_comparison step3, no full case folding of sharp s", "STRASSE", ""},
		{"derived: name_comparison step4, decomposed", "u\u0308bersicht", "cart.overview"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			o, dir := show(t, terms, c.query, "--json")
			if c.id == "" {
				o.none(t)
				return
			}
			o.elementsAre(t, terms, dir, shown{c.id, glossaryPath, "glossary entry", "cart"})
		})
	}

	for _, c := range []struct {
		name string
		args []string
	}{
		{"row12 since, the tree", []string{"since"}},
		{"row13 Lint.Run.Exit, the tree", []string{"Cart.Add.Limit"}},
		{"row14 Tree_Layout, the tree", []string{"Max_Quantity"}},
		{"row16 Rule: Lint finds the nearest spec root, the tree", []string{"Rule: The cart takes a limited quantity"}},
		// baseline.diff.since: billing.finding, an identifier outside cart.
		{"row17 baseline.diff.since, lint's and the system's", []string{"billing.finding", "--in", "cart"}},
		{"row18 nosuch, with --json, the tree", []string{"nosuch"}},
		{"derived: a capability's heading is never matched", []string{"Cart"}},
		{"derived: a part of an identifier", []string{"cart.a"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			o, _ := show(t, tr, append(c.args, "--json")...)
			o.none(t)
			if o.stderr != "" {
				t.Errorf("query.show.match: with --json, nothing goes to stderr, got %q", o.stderr)
			}
		})
	}
	t.Run("row15 spec, the tree", func(t *testing.T) {
		t.Parallel()
		o, _ := show(t, tr, "spec")
		if o.exit != 1 {
			t.Fatalf("query.show.match: want exit status 1, got %d\nstdout: %s\nstderr: %s", o.exit, o.stdout, o.stderr)
		}
		if o.stdout != "" {
			t.Errorf("query.show.match: without --json, nothing goes to stdout, got %q", o.stdout)
		}
		lines := stderrLines(o.stderr)
		if len(lines) != 1 || !strings.Contains(lines[0], "spec") {
			t.Errorf("query.show.match: want one line on stderr naming spec, got:\n%s", o.stderr)
		}
	})
	t.Run("derived: without --json, the line names the capability --in names", func(t *testing.T) {
		t.Parallel()
		o, _ := show(t, tr, "billing.finding", "--in", "cart")
		if o.exit != 1 {
			t.Fatalf("query.show.match: want exit status 1, got %d\nstderr: %s", o.exit, o.stderr)
		}
		if o.stdout != "" {
			t.Errorf("query.show.match: without --json, nothing goes to stdout, got %q", o.stdout)
		}
		lines := stderrLines(o.stderr)
		if len(lines) != 1 || !strings.Contains(lines[0], "billing.finding") || !strings.Contains(lines[0], "cart") {
			t.Errorf("query.show.match: want one line on stderr naming billing.finding and cart, got:\n%s", o.stderr)
		}
	})
}

// termMD writes a glossary entry of the cart.
func termMD(id, name string) string {
	return "## <a id=\"" + id + "\"></a> " + name + "\n\n- Definition: A word.\n- Contact: Dana Reyes\n- Status: defined\n"
}

func TestQueryShowOutput(t *testing.T) {
	t.Parallel()
	tr := showTree()
	t.Run("row1 lint.run.exit, --json", func(t *testing.T) {
		t.Parallel()
		o, dir := show(t, tr, "cart.add.limit", "--json")
		if o.exit != 0 {
			t.Fatalf("query.show.output: want exit status 0, got %d\nstderr: %s", o.exit, o.stderr)
		}
		o.elementsAre(t, tr, dir, sLimit)
	})
	t.Run("row2 system, --json", func(t *testing.T) {
		t.Parallel()
		o, dir := show(t, tr, "system", "--json")
		if o.exit != 0 {
			t.Fatalf("query.show.output: want exit status 0, got %d\nstderr: %s", o.exit, o.stderr)
		}
		o.elementsAre(t, tr, dir, sSystem)
		if e := o.view.Elements[0]; e.Parent != nil {
			t.Errorf("query.show.output: want a null parent, got %q", *e.Parent)
		}
	})
	for _, c := range []struct {
		name, query string
		want        []shown
	}{
		{"row3 lint.run.exit, none", "cart.add.limit", []shown{sLimit}},
		// coherence.finding, lint.finding: billing.finding, cart.finding.
		{"row4 coherence.finding, lint.finding, none", "finding", []shown{sBillFind, sCartFind}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			dir := tr.write(t)
			j := showIn(t, dir, dir, c.query, "--json")
			j.elementsAre(t, tr, dir, c.want...)
			o := showIn(t, dir, dir, c.query)
			if o.exit != 0 {
				t.Fatalf("query.show.output: want exit status 0, got %d\nstderr: %s", o.exit, o.stderr)
			}
			isShowText(t, o.stdout, *j.view)
		})
	}
}

// TestQueryElementParent checks the definition of element parent on an
// element of each kind the definition names.
func TestQueryElementParent(t *testing.T) {
	t.Parallel()
	tr := showTree()
	for _, c := range []struct {
		name string
		el   shown
	}{
		{"a rule belongs to its feature", sLimit},
		{"a feature belongs to its capability", sAdd},
		{"a capability belongs to the system", sCart},
		{"a member belongs to its contract element", sSince},
		{"a top-level contract element of the system level belongs to the system", sDiff},
		{"a glossary entry of a capability belongs to it", sItem},
		{"a parameter of a capability belongs to it", sMaxQty},
		{"a glossary entry of the system level belongs to the system", sSpecTree},
		{"the system belongs to none", sSystem},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			o, dir := show(t, tr, c.el.id, "--json")
			o.elementsAre(t, tr, dir, c.el)
		})
	}
}

// TestQueryReadOnly checks query.read-only over runs of show that end in
// every way: showIn compares every spec root below the test's directory
// before and after the run.
func TestQueryReadOnly(t *testing.T) {
	t.Parallel()
	tr := showTree()
	for _, c := range []struct {
		name string
		args []string
	}{
		{"found, as element_view", []string{"cart.add.limit", "--json"}},
		{"found, as show_text", []string{"finding"}},
		{"none found", []string{"nosuch"}},
		{"no query", nil},
		{"a capability the tree does not hold", []string{"system", "--in", "nosuch"}},
		{"an argument show does not take", []string{"system", "--write"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			show(t, tr, c.args...)
		})
	}
	t.Run("in an accepted repository in the middle of an amendment", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, tr)
		r.set(tr.edit(glossaryPath, "A product placed in the cart.", "A product put into the cart."))
		showIn(t, r.dir, r.dir, "cart.item", "--json")
	})
	t.Run("a tree lint cannot read", func(t *testing.T) {
		t.Parallel()
		dir := tr.with("proviso/manifest.sum", "model 1\n").write(t)
		showIn(t, dir, dir, "system", "--json")
	})
}

// TestSystemCliShow checks the contract system.cli.show: run anywhere inside
// the system's repository.
func TestSystemCliShow(t *testing.T) {
	t.Parallel()
	tr := showTree().with("cmd/shop/main.go", "package main\n")
	dir := tr.write(t)
	showIn(t, dir, filepath.Join(dir, "cmd", "shop"), "cart.add.limit", "--json").elementsAre(t, tr, dir, sLimit)
}

func TestSystemCliShowExitStatus(t *testing.T) {
	t.Parallel()
	tr := showTree()
	for _, c := range []struct {
		name string
		args []string
		want int
	}{
		{"0: at least one element was written", []string{"system", "--json"}, 0},
		{"1: the query names no element", []string{"nosuch", "--json"}, 1},
		{"2: no query", []string{"--json"}, 2},
		{"2: --in names no capability of the spec tree", []string{"system", "--in", "nosuch", "--json"}, 2},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if o, _ := show(t, tr, c.args...); o.exit != c.want {
				t.Errorf("system.cli.show.exit-status: want %d, got %d\nstderr: %s", c.want, o.exit, o.stderr)
			}
		})
	}
	t.Run("2: the command could not run", func(t *testing.T) {
		t.Parallel()
		dir := anywhere(t)
		skipIfSpecRootAbove(t, dir)
		if o := showIn(t, dir, dir, "system", "--json"); o.exit != 2 {
			t.Errorf("system.cli.show.exit-status: want 2, got %d", o.exit)
		}
	})
}

// TestSystemCliShowText checks show_text against its example, written out
// in full, independently of element_view.
func TestSystemCliShowText(t *testing.T) {
	t.Parallel()
	tr := showTree()
	o, _ := show(t, tr, "cart.finding")
	if o.exit != 0 {
		t.Fatalf("want exit status 0, got %d\nstderr: %s", o.exit, o.stderr)
	}
	text, line := elementText(tr[glossaryPath], "cart.finding")
	exp := "---\n" +
		"id: \"cart.finding\"\n" +
		"type: \"glossary entry\"\n" +
		"heading: \"finding\"\n" +
		"file: \"proviso/capabilities/cart/glossary.md\"\n" +
		"line: " + itoa(line) + "\n" +
		"parent: \"cart\"\n" +
		"---\n\n" +
		strings.TrimRight(text, "\n")
	if got := strings.TrimRight(strings.ReplaceAll(o.stdout, "\r\n", "\n"), "\n"); got != exp {
		t.Errorf("system.cli.show-text: want\n%s\ngot\n%s", exp, got)
	}
	t.Run("derived: null where there is none", func(t *testing.T) {
		t.Parallel()
		o, _ := show(t, tr, "system")
		if o.exit != 0 {
			t.Fatalf("want exit status 0, got %d\nstderr: %s", o.exit, o.stderr)
		}
		lines := strings.Split(strings.ReplaceAll(o.stdout, "\r\n", "\n"), "\n")
		if len(lines) < 8 || lines[6] != "parent: null" {
			t.Errorf("system.cli.show-text: want the line \"parent: null\" last in the frontmatter, got:\n%s", o.stdout)
		}
	})
}

// TestSystemCliElementView checks element_view: the spec root that was read,
// and the paths, as in report.
func TestSystemCliElementView(t *testing.T) {
	t.Parallel()
	tr := showTree()
	t.Run("derived: elements empty where the query names none", func(t *testing.T) {
		t.Parallel()
		o, dir := show(t, tr, "nosuch", "--json")
		o.none(t)
		checkRoot(t, outcome{report: report{Root: o.view.Root}}, filepath.Join(dir, "proviso"))
	})
	t.Run("derived: a heading with characters JSON escapes", func(t *testing.T) {
		t.Parallel()
		q := tr.appendTo(glossaryPath, "\n"+termMD("cart.quoted", `the "best" price`))
		o, dir := show(t, q, "cart.quoted", "--json")
		o.elementsAre(t, q, dir, shown{"cart.quoted", glossaryPath, "glossary entry", "cart"})
	})
}
