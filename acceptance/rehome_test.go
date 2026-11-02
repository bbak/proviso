package acceptance

// Feature upkeep.rehome: proviso/capabilities/upkeep/features/rehome.md, the
// parameter rehome_references, the term identifier home, and the contracts
// system.cli.rehome and system.cli.rehome-plan (proviso/contracts/cli.md).
// What holds for every run (the plan's shape and order, the exit status and
// the invariant upkeep.rehome.writes) is checked by rehomeIn. Every tree is
// accepted in a git repository first, so that it is as its manifest records
// it.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readOnly takes the write permission from the file at the slash-separated
// path p below dir, and gives it back when the test ends, so that the
// temporary folder can be removed.
func readOnly(t *testing.T, dir, p string) {
	t.Helper()
	full := filepath.Join(dir, filepath.FromSlash(p))
	if err := os.Chmod(full, 0o444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(full, 0o644) })
	if f, err := os.OpenFile(full, os.O_WRONLY, 0); err == nil {
		f.Close()
		t.Skipf("%s stays writable without permission (running as root?)", p)
	}
}

func TestUpkeepRehomeLint(t *testing.T) {
	t.Parallel()
	t.Run("row1 repo/, its tree without findings", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		o := r.rehome("--json")
		o.exits(t, 0)
		o.planned(t, mainRehomeTree())
	})
	t.Run("row2 repo/, its tree with an ask only", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, undecidedTree())
		o := r.rehome("--json")
		o.exits(t, 0)
		o.planned(t, undecidedTree())
	})
	t.Run("row3 repo/, its tree with an error, the error on stderr", func(t *testing.T) {
		t.Parallel()
		// notes.md is no file of the layout and no element: lint reports
		// an error, while the tree is still as the manifest records it.
		r, _ := accepted(t, mainRehomeTree())
		r.set(mainRehomeTree().with("proviso/notes.md", "# Notes\n"))
		o := r.rehome("--write", "--json")
		o.exits(t, 2)
		text := execLint(t, r.dir)
		if text.exit != 1 || strings.TrimSpace(text.stdout) == "" {
			t.Fatalf("want lint to report the error, got exit status %d\n%s%s", text.exit, text.stdout, text.stderr)
		}
		norm := func(s string) string { return strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n")) }
		if !strings.Contains(norm(o.stderr), norm(text.stdout)) {
			t.Errorf("upkeep.rehome.lint: want stderr to hold what lint writes as text\n%s\ngot\n%s", text.stdout, o.stderr)
		}
	})
	t.Run("row4 repo/, its tree with an unreadable file", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		r.set(mainRehomeTree().with("proviso/capabilities/cart/features/bad.md", "# <a id=\"cart.bad\"></a> Bad \xff\xfe\n\n- Contact: Dana Reyes\n"))
		r.rehome("--write", "--json").exits(t, 2)
	})
	t.Run("row5 other/, with no spec root up to the file system's root", func(t *testing.T) {
		t.Parallel()
		dir := anywhere(t)
		rehomeIn(t, dir, dir, "", nil, "--write", "--json").exits(t, 2)
	})
}

func TestUpkeepRehomeBaseline(t *testing.T) {
	t.Parallel()
	t.Run("row1 repo/proviso/, its tree as manifest.sum records it", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		o := r.rehome("--json")
		o.exits(t, 0)
		o.renames(t, mainRenames)
	})
	t.Run("row2 repo/proviso/, a paragraph reflowed since manifest.sum was written", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		tr := mainRehomeTree().edit(capPath, "Paying once for many goods is easier.", "Paying once for many\ngoods   is easier.")
		r.set(tr)
		o := r.rehome("--json")
		o.exits(t, 0)
		o.renames(t, mainRenames)
	})
	// The refused rows run with --write, so that a run that went on would
	// have written.
	for _, c := range []struct {
		name string
		tr   func() tree
	}{
		{"row4 repo/proviso/, a rule's Then step edited since manifest.sum was written", func() tree {
			return mainRehomeTree().edit(featurePath, "Then      the cart holds <quantity> more of <product>", "Then      the cart holds <quantity> more pieces of <product>")
		}},
		{"row5 repo/proviso/, a rule added since manifest.sum was written", func() tree {
			return mainRehomeTree().appendTo(featurePath, extraRule)
		}},
		{"row6 repo/proviso/, a rule removed since manifest.sum was written", func() tree {
			empty, _ := elementText(mainRehomeTree()[featurePath], "cart.add.empty")
			return mainRehomeTree().edit(featurePath, empty, "")
		}},
		{"row7 repo/proviso/, holding .coherence.json", func() tree {
			return mainRehomeTree().with(coherencePath, "{}\n")
		}},
		{"row8 repo/proviso/, holding .amendment.md", func() tree {
			return mainRehomeTree().with(notePath, "- refunds\n")
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			r, _ := accepted(t, mainRehomeTree())
			r.set(c.tr())
			r.rehome("--write", "--json").exits(t, 2)
		})
	}
	t.Run("row3 repo/proviso/, without manifest.sum", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		remove(t, r.dir, manifestPath)
		r.rehome("--write", "--json").exits(t, 2)
	})
	t.Run("derived: .amendment.md a folder", func(t *testing.T) {
		t.Parallel()
		// Anything of that name is the note.
		r, _ := accepted(t, mainRehomeTree())
		r.set(mainRehomeTree().with(notePath+"/note.md", "- refunds\n"))
		r.rehome("--write", "--json").exits(t, 2)
	})
}

func TestUpkeepRehomeHome(t *testing.T) {
	t.Parallel()
	r, _ := accepted(t, mainRehomeTree())
	o := r.rehome("--json")
	o.exits(t, 0)
	o.planned(t, mainRehomeTree())
	// renamed checks that the plan renames from to to; stays that it plans
	// no rename of id.
	renamed := func(t *testing.T, from, to string) {
		t.Helper()
		for _, rn := range o.plan.Renames {
			if rn.From == from {
				if rn.To != to {
					t.Errorf("upkeep.rehome.home: want %s renamed to %s, got %s", from, to, rn.To)
				}
				return
			}
		}
		t.Errorf("upkeep.rehome.home: want %s renamed to %s, got the renames %v", from, to, o.plan.Renames)
	}
	stays := func(t *testing.T, id string) {
		t.Helper()
		for _, rn := range o.plan.Renames {
			if rn.From == id {
				t.Errorf("upkeep.rehome.home: want %s to stay, got it renamed to %s", id, rn.To)
			}
		}
	}
	t.Run("row1 coherence.impact.events, a rule of the capability coherence", func(t *testing.T) {
		t.Parallel()
		stays(t, "cart.add.limit")
		stays(t, "cart.add.show")
	})
	t.Run("row2 baseline.impact.events, a rule of the capability coherence", func(t *testing.T) {
		t.Parallel()
		renamed(t, "billing.x.r", "cart.x.r")
	})
	t.Run("row3 baseline.part, a glossary entry of the system", func(t *testing.T) {
		t.Parallel()
		renamed(t, "cart.item", "system.item")
	})
	t.Run("row4 lint.cli.lint.exit-status, a member in contracts/cli.md of the system", func(t *testing.T) {
		t.Parallel()
		renamed(t, "cart.cli.lint.exit-status", "system.cli.lint.exit-status")
		renamed(t, "cart.cli.lint", "system.cli.lint")
	})
	t.Run("row5 coherence, the capability coherence", func(t *testing.T) {
		t.Parallel()
		stays(t, "cart")
		stays(t, "pay")
	})
	t.Run("row6 system, the system", func(t *testing.T) {
		t.Parallel()
		stays(t, "system")
		stays(t, "system.break")
	})
	t.Run("row7 checks.run, in the folder capabilities/coherence/, whose capability.md is identified as checks", func(t *testing.T) {
		t.Parallel()
		stays(t, "pay.card")
		// Named after the folder, not the capability: renamed.
		renamed(t, "billing.refund", "pay.refund")
	})
	t.Run("derived: a feature moved into another capability's folder", func(t *testing.T) {
		t.Parallel()
		renamed(t, "billing.x", "cart.x")
	})
	t.Run("derived: a reference of the system", func(t *testing.T) {
		t.Parallel()
		renamed(t, "cart.ref.commonmark", "system.ref.commonmark")
	})
}

func TestUpkeepRehomeCollision(t *testing.T) {
	t.Parallel()
	t.Run("row1 baseline.impact.events to coherence.impact.events, the tree holding no coherence.impact.events", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		o := r.rehome("--json")
		o.exits(t, 0)
		o.collisions(t, nil)
		o.renames(t, mainRenames)
	})
	t.Run("row2 none", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, baseline())
		o := r.rehome("--json")
		o.exits(t, 0)
		o.renames(t, nil)
		o.collisions(t, nil)
	})
	// The colliding rows run with --write, so that nothing written under
	// the spec root shows (rehomeIn).
	t.Run("row3 baseline.x to coherence.x, the tree holding a rule coherence.x", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, takenRehomeTree())
		o := r.rehome("--write", "--json")
		o.exits(t, 1)
		o.renames(t, []planRename{{"billing.x.r", "cart.x.r"}})
		o.collisions(t, []planCollision{{"cart.x.r", []string{"billing.x.r"}}})
	})
	t.Run("row4 baseline.x to coherence.x, while coherence.x of another capability is renamed to install.x", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, chainRehomeTree())
		o := r.rehome("--write", "--json")
		o.exits(t, 1)
		o.renames(t, []planRename{{"billing.widget", "cart.widget"}, {"cart.widget", "stock.widget"}})
		o.collisions(t, []planCollision{{"cart.widget", []string{"billing.widget"}}})
	})
	t.Run("row5 baseline.x and lint.x, both to coherence.x", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, order2RehomeTree())
		o := r.rehome("--write", "--json")
		o.exits(t, 1)
		o.planned(t, order2RehomeTree())
		for _, c := range o.plan.Collisions {
			if c.Identifier == "cart.gizmo" {
				if strings.Join(c.From, " ") != "billing.gizmo pay.gizmo" {
					t.Errorf("upkeep.rehome.collision: want the collision on cart.gizmo to name billing.gizmo and pay.gizmo, got %v", c.From)
				}
				return
			}
		}
		t.Errorf("upkeep.rehome.collision: want a collision on cart.gizmo, got %v", o.plan.Collisions)
	})
}

func TestUpkeepRehomeWrite(t *testing.T) {
	t.Parallel()
	const rule = "upkeep.rehome.write"
	t.Run("rows 1 to 3", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		o := r.rehome("--write", "--json")
		o.exits(t, 0)
		o.renames(t, mainRenames)
		want := mainRehomed()
		t.Run("row1 baseline.impact.events to coherence.impact.events, its anchor in impact.md and a hand-over to it in verify.md", func(t *testing.T) {
			t.Parallel()
			o.holds(t, rule, xPath, want[xPath])
			o.holds(t, rule, featurePath, want[featurePath])
			o.holds(t, rule, manifestPath, manifestWritten(t, want))
		})
		t.Run("row2 the same, with baseline.impact.events also in an Examples cell of verify.md", func(t *testing.T) {
			t.Parallel()
			got, _ := o.file(featurePath)
			if !strings.Contains(got, "| billing.x.r | yes | shown |") {
				t.Errorf("%s: want the Examples cell to hold billing.x.r as it was, got\n%s", rule, got)
			}
			if !strings.Contains(got, "Otherwise → cart.x.r") {
				t.Errorf("%s: want the hand-over renamed to cart.x.r, got\n%s", rule, got)
			}
		})
		t.Run("row3 baseline.dependencies to coherence.dependencies, tagging a question in questions.md", func(t *testing.T) {
			t.Parallel()
			o.holds(t, rule, cartQuestions, want[cartQuestions])
		})
		t.Run("derived: every file of the spec root as the rewritten tree holds it", func(t *testing.T) {
			t.Parallel()
			for p, content := range want {
				o.holds(t, rule, p, content)
			}
		})
	})
	t.Run("row4 none", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, baseline())
		o := r.rehome("--write", "--json")
		o.exits(t, 0)
		o.renames(t, nil)
		o.unchanged(t, rule, manifestPath)
		checkWrites(t, rule, o.before, o.after, "")
	})
	t.Run("row5 baseline.impact.events to coherence.impact.events, impact.md read-only", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		readOnly(t, r.dir, xPath)
		o := r.rehome("--write", "--json")
		o.exits(t, 2)
		o.names(t, "x.md")
	})
	t.Run("row6 the same, manifest.sum read-only", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		readOnly(t, r.dir, manifestPath)
		o := r.rehome("--write", "--json")
		o.exits(t, 2)
		o.names(t, "manifest.sum")
	})
	t.Run("derived: one rename, the manifest of the rewritten tree", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, oneRehomeTree())
		o := r.rehome("--write", "--json")
		o.exits(t, 0)
		want := oneRehomed()
		o.holds(t, rule, "proviso/glossary.md", want["proviso/glossary.md"])
		o.holds(t, rule, manifestPath, manifestWritten(t, want))
	})
	t.Run("derived: the element diff is empty after the run", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		r.rehome("--write", "--json").exits(t, 0)
		r.diff().only(t)
	})
}

func TestUpkeepRehomeOutput(t *testing.T) {
	t.Parallel()
	t.Run("row1 one, none, --json", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, oneRehomeTree())
		o := r.rehome("--json")
		o.exits(t, 0)
		o.renames(t, []planRename{{"cart.item", "system.item"}})
		o.collisions(t, nil)
	})
	t.Run("row2 none, none, --json", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, baseline())
		o := r.rehome("--json")
		o.exits(t, 0)
		// An empty list, not an empty output: parsePlan requires both
		// lists as arrays.
		if len(o.plan.Renames) != 0 || len(o.plan.Collisions) != 0 {
			t.Errorf("upkeep.rehome.output: want no renames and no collisions, got %v", o.plan)
		}
	})
	t.Run("row3 one, none, --write --json", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, oneRehomeTree())
		o := r.rehome("--write", "--json")
		o.exits(t, 0)
		o.renames(t, []planRename{{"cart.item", "system.item"}})
		o.collisions(t, nil)
	})
	t.Run("row4 two, one colliding, one, none", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, chainRehomeTree())
		o := r.rehome()
		o.exits(t, 1)
		if strings.TrimSpace(o.stdout) == "" {
			t.Fatalf("upkeep.rehome.output: want the renames and the collision as text, got nothing")
		}
		if strings.HasPrefix(strings.TrimSpace(o.stdout), "{") {
			t.Errorf("upkeep.rehome.output: want text without --json, got:\n%s", o.stdout)
		}
		for _, id := range []string{"billing.widget", "cart.widget", "stock.widget"} {
			if !strings.Contains(o.stdout, id) {
				t.Errorf("upkeep.rehome.output: want the text to name %s, got:\n%s", id, o.stdout)
			}
		}
	})
	t.Run("row5 none, none, none", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, baseline())
		o := r.rehome()
		o.exits(t, 0)
		if o.stdout != "" {
			t.Errorf("upkeep.rehome.output: nothing to say, but stdout holds %q", o.stdout)
		}
	})
}

func TestUpkeepRehomeOrder(t *testing.T) {
	t.Parallel()
	t.Run("row1 lint.keyword to system.keyword, baseline.part to system.part", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, order1RehomeTree())
		o := r.rehome("--json")
		o.exits(t, 0)
		o.renames(t, []planRename{{"billing.part", "system.part"}, {"stock.keyword", "system.keyword"}})
	})
	t.Run("row2 lint.x and baseline.x, both to coherence.x, and lint.y to coherence.y, taken", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, order2RehomeTree())
		o := r.rehome("--json")
		o.exits(t, 1)
		o.collisions(t, []planCollision{
			{"cart.gizmo", []string{"billing.gizmo", "pay.gizmo"}},
			{"cart.widget", []string{"pay.widget"}},
		})
	})
}

func TestUpkeepRehomeExit(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name string
		tr   func() tree
		want int
	}{
		{"row1 none", mainRehomeTree, 0},
		{"row2 none, and no renames", baseline, 0},
		{"row3 one", takenRehomeTree, 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			r, _ := accepted(t, c.tr())
			r.rehome("--json").exits(t, c.want)
		})
	}
	t.Run("derived: the text form exits as the JSON form", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, takenRehomeTree())
		r.rehome().exits(t, 1)
	})
}

// TestUpkeepRehomeWrites checks the invariant upkeep.rehome.writes on runs
// that end every way; rehomeIn checks it on every run.
func TestUpkeepRehomeWrites(t *testing.T) {
	t.Parallel()
	t.Run("derived: without --write, nothing changes", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		o := r.rehome("--json")
		o.exits(t, 0)
		checkWrites(t, "upkeep.rehome.writes", o.before, o.after, "")
	})
	t.Run("derived: with --write and exit status 0, only the renamed identifiers and manifest.sum", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		o := r.rehome("--write", "--json")
		o.exits(t, 0)
		checkRenamedOnly(t, o.before, o.after, "proviso", mainRenames)
		// The git repository is as it was.
		for p, v := range o.before {
			if strings.HasPrefix(p, ".git/") && o.after[p] != v {
				t.Errorf("upkeep.rehome.writes: %s of the git repository changed", p)
			}
		}
	})
	t.Run("derived: with --write and exit status 1, nothing changes", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, takenRehomeTree())
		o := r.rehome("--write", "--json")
		o.exits(t, 1)
		checkWrites(t, "upkeep.rehome.writes", o.before, o.after, "")
	})
	t.Run("derived: with --write and exit status 2, nothing changes", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		r.set(mainRehomeTree().with(coherencePath, "{}\n"))
		o := r.rehome("--write", "--json")
		o.exits(t, 2)
		checkWrites(t, "upkeep.rehome.writes", o.before, o.after, "")
	})
	t.Run("derived: with --write, in the text form, only the renamed identifiers and manifest.sum", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		o := r.rehome("--write")
		o.exits(t, 0)
		checkRenamedOnly(t, o.before, o.after, "proviso", mainRenames)
	})
}

// TestUpkeepRehomeReferences checks the parameter rehome_references: each
// place it lists is renamed, and an identifier anywhere else stays.
func TestUpkeepRehomeReferences(t *testing.T) {
	t.Parallel()
	r, _ := accepted(t, mainRehomeTree())
	o := r.rehome("--write", "--json")
	o.exits(t, 0)
	get := func(p string) string { s, _ := o.file(p); return s }
	holds := func(t *testing.T, p, sub string) {
		t.Helper()
		if !strings.Contains(get(p), md(sub)) {
			t.Errorf("upkeep.rehome-references: want %s to hold %q, got\n%s", p, md(sub), get(p))
		}
	}
	lacks := func(t *testing.T, p, sub string) {
		t.Helper()
		if strings.Contains(get(p), md(sub)) {
			t.Errorf("upkeep.rehome-references: want %s not to hold %q, got\n%s", p, md(sub), get(p))
		}
	}
	t.Run("row1 an anchor", func(t *testing.T) {
		t.Parallel()
		holds(t, xPath, `## <a id="cart.x.r"></a> Rule: Under test`)
		holds(t, payRefundPath, `# <a id="pay.refund"></a> Refunds`)
		lacks(t, payRefundPath, "billing.refund")
	})
	t.Run("row2 a hand-over", func(t *testing.T) {
		t.Parallel()
		holds(t, featurePath, "Otherwise → cart.x.r\n")
		lacks(t, featurePath, "Otherwise → billing.x.r")
	})
	t.Run("row3 a link, outside code spans and fenced blocks", func(t *testing.T) {
		t.Parallel()
		holds(t, "proviso/glossary.md", "(references.md#system.ref.commonmark)")
		holds(t, xPath, "see [the rule](x.md#cart.x.r), not")
	})
	t.Run("row4 a question in a questions.md", func(t *testing.T) {
		t.Parallel()
		holds(t, cartQuestions, "- [cart.x.r] Should the shop show pears?")
	})
	t.Run("derived: an identifier in plain text stays", func(t *testing.T) {
		t.Parallel()
		holds(t, xPath, "and billing.x.r in\nplain text.")
		holds(t, cartQuestions, "- Who decides on [billing.x.r] later?")
	})
	t.Run("derived: an identifier in an Examples cell stays", func(t *testing.T) {
		t.Parallel()
		holds(t, featurePath, "| billing.x.r | yes | shown |")
	})
	t.Run("derived: an identifier in a code span stays", func(t *testing.T) {
		t.Parallel()
		holds(t, xPath, "´billing.x.r´ nor ´[the rule](x.md#billing.x.r)´")
	})
	t.Run("derived: an identifier in a fenced block stays", func(t *testing.T) {
		t.Parallel()
		holds(t, xPath, "´´´text\n<a id=\"billing.x.r\"></a> [the rule](x.md#billing.x.r)\n´´´")
	})
}

// TestUpkeepIdentifierHome checks the term identifier home through the
// plans of trees whose elements stand at every kind of place: the oracle
// identifierHome gives each element's home, and every element whose first
// segment names another place is planned to be renamed.
func TestUpkeepIdentifierHome(t *testing.T) {
	t.Parallel()
	for name, tr := range map[string]func() tree{
		"main": mainRehomeTree, "one": oneRehomeTree, "order1": order1RehomeTree, "chain": chainRehomeTree,
	} {
		t.Run("derived: "+name, func(t *testing.T) {
			t.Parallel()
			r, _ := accepted(t, tr())
			o := r.rehome("--json")
			if o.exit == 2 {
				t.Fatalf("want rehome to plan, got exit status 2\n%s", o.stderr)
			}
			o.planned(t, tr())
		})
	}
}

// TestSystemCliRehome checks the contract system.cli.rehome: it runs
// anywhere inside the system's repository.
func TestSystemCliRehome(t *testing.T) {
	t.Parallel()
	t.Run("derived: run in a folder of the spec tree, it plans for the spec root", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		cwd := filepath.Join(r.dir, "proviso", "capabilities", "cart")
		o := rehomeIn(t, r.dir, cwd, r.root(), nil, "--json")
		o.exits(t, 0)
		o.renames(t, mainRenames)
	})
	t.Run("derived: run in a folder beside the spec root, it plans for the spec root", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree().with("src/main.go", "package main\n"))
		o := rehomeIn(t, r.dir, filepath.Join(r.dir, "src"), r.root(), nil, "--json")
		o.exits(t, 0)
		o.renames(t, mainRenames)
	})
}
